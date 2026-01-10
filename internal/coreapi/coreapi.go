package coreapi

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	fbauth "firebase.google.com/go/v4/auth"
	"github.com/go-chi/chi/v5"
	"github.com/linkasu/linka.type-backend/internal/auth"
	"github.com/linkasu/linka.type-backend/internal/config"
	"github.com/linkasu/linka.type-backend/internal/httpapi"
	"github.com/linkasu/linka.type-backend/internal/httpmiddleware"
	"github.com/linkasu/linka.type-backend/internal/jwt"
	"github.com/linkasu/linka.type-backend/internal/models"
	"github.com/linkasu/linka.type-backend/internal/service"
	"github.com/linkasu/linka.type-backend/internal/store"
	"github.com/linkasu/linka.type-backend/internal/userctx"
)

// API wires HTTP handlers for core-api.
type API struct {
	svc        *service.Service
	auth       auth.Verifier
	fbAuth     *fbauth.Client
	jwtManager *jwt.Manager
	config     config.Config
	httpClient *http.Client
}

// New builds the core API router.
func New(svc *service.Service, verifier auth.Verifier, fbAuth *fbauth.Client, jwtManager *jwt.Manager, cfg config.Config) http.Handler {
	api := &API{
		svc:        svc,
		auth:       verifier,
		fbAuth:     fbAuth,
		jwtManager: jwtManager,
		config:     cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	r := chi.NewRouter()
	r.Use(corsMiddleware)
	r.Use(httpmiddleware.RequestID)

	r.Get("/", serveWebFile("index.html", "text/html; charset=utf-8"))
	r.Get("/client.md", serveWebFile("client.md", "text/markdown; charset=utf-8"))
	r.Get("/AGENTS.md", serveWebFile("AGENTS.md", "text/markdown; charset=utf-8"))
	r.Get("/admin/", serveWebFile("admin/index.html", "text/html; charset=utf-8"))
	r.Get("/healthz", healthHandler)
	r.Handle("/assets/*", assetsHandler())

	r.Route("/v1", func(r chi.Router) {
		r.Post("/auth", api.authToken)
		r.Post("/auth/register", api.authRegister)
		r.Post("/auth/refresh", api.authRefresh)
		r.Post("/auth/logout", api.authLogout)
		r.Group(func(r chi.Router) {
			r.Use(httpmiddleware.Auth(verifier))
			r.Get("/categories", api.listCategories)
			r.Post("/categories", api.createCategory)
			r.Patch("/categories/{id}", api.patchCategory)
			r.Delete("/categories/{id}", api.deleteCategory)

			r.Get("/categories/{id}/statements", api.listStatements)
			r.Post("/statements", api.createStatement)
			r.Patch("/statements/{id}", api.patchStatement)
			r.Delete("/statements/{id}", api.deleteStatement)

			r.Get("/user/state", api.getUserState)
			r.Put("/user/state", api.putUserState)
			r.Get("/quickes", api.getQuickes)
			r.Put("/quickes", api.putQuickes)

			r.Get("/global/categories", api.listGlobalCategories)
			r.Get("/global/categories/{id}/statements", api.listGlobalStatements)
			r.Post("/global/import", api.importGlobal)

			r.Get("/factory/questions", api.listFactoryQuestions)
			r.Post("/onboarding/phrases", api.onboardingPhrases)

			r.Post("/user/delete", api.deleteUser)

			if cfg.TTS.ProxyEnabled {
				r.Get("/voices", api.proxyVoices)
				r.MethodFunc(http.MethodPost, "/tts", api.proxyTTS)
				r.MethodFunc(http.MethodGet, "/tts", api.proxyTTS)
			}

			r.Get("/predictor", api.predictorComplete)

			r.Route("/dialog", func(r chi.Router) {
				r.Get("/chats", api.listDialogChats)
				r.Post("/chats", api.createDialogChat)
				r.Delete("/chats/{id}", api.deleteDialogChat)
				r.Get("/chats/{id}/messages", api.listDialogMessages)
				r.Post("/chats/{id}/messages", api.createDialogMessage)
				r.Get("/suggestions", api.listDialogSuggestions)
				r.Post("/suggestions/apply", api.applyDialogSuggestions)
				r.Post("/suggestions/dismiss", api.dismissDialogSuggestions)
			})
		})

		r.Route("/admin", func(r chi.Router) {
			r.Use(httpmiddleware.Auth(verifier))
			r.Use(httpmiddleware.Admin(svc))
			r.Get("/stats", api.adminStats)
			r.Get("/admins", api.adminListAdmins)
			r.Post("/admins", api.adminAddAdmin)
			r.Delete("/admins/{user_id}", api.adminRemoveAdmin)
			r.Get("/client-keys", api.adminListClientKeys)
			r.Post("/client-keys", api.adminCreateClientKey)
			r.Delete("/client-keys/{key_hash}", api.adminRevokeClientKey)
			r.Get("/global/categories", api.adminListGlobalCategories)
			r.Post("/global/categories", api.adminCreateGlobalCategory)
			r.Patch("/global/categories/{id}", api.adminUpdateGlobalCategory)
			r.Delete("/global/categories/{id}", api.adminDeleteGlobalCategory)
			r.Get("/factory/questions", api.adminListFactoryQuestions)
			r.Post("/factory/questions", api.adminCreateFactoryQuestion)
			r.Patch("/factory/questions/{id}", api.adminUpdateFactoryQuestion)
			r.Delete("/factory/questions/{id}", api.adminDeleteFactoryQuestion)
		})
	})

	return r
}

func (api *API) listCategories(w http.ResponseWriter, r *http.Request) {
	user := mustUser(w, r)
	if user.UID == "" {
		return
	}
	categories, err := api.svc.ListCategories(r.Context(), user.UID)
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "categories_failed", err.Error())
		return
	}
	if categories == nil {
		categories = []models.Category{}
	}
	httpapi.WriteJSON(w, http.StatusOK, categories)
}

func (api *API) createCategory(w http.ResponseWriter, r *http.Request) {
	user := mustUser(w, r)
	if user.UID == "" {
		return
	}
	var req struct {
		ID      string `json:"id"`
		Label   string `json:"label"`
		Created int64  `json:"created"`
		Default *bool  `json:"default"`
		AIUse   *bool  `json:"aiUse"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if strings.TrimSpace(req.Label) == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_label", "label is required")
		return
	}

	category, err := api.svc.CreateCategory(r.Context(), user.UID, service.CategoryInput{
		ID:      req.ID,
		Label:   req.Label,
		Created: req.Created,
		Default: req.Default,
		AIUse:   req.AIUse != nil && *req.AIUse,
	})
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "create_category_failed", err.Error())
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, category)
}

func (api *API) patchCategory(w http.ResponseWriter, r *http.Request) {
	user := mustUser(w, r)
	if user.UID == "" {
		return
	}
	categoryID := chi.URLParam(r, "id")
	if categoryID == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_id", "category id is required")
		return
	}

	var req struct {
		Label   *string `json:"label"`
		Default *bool   `json:"default"`
		AIUse   *bool   `json:"aiUse"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if req.Label == nil && req.Default == nil && req.AIUse == nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_payload", "label, default, or aiUse is required")
		return
	}

	category, err := api.svc.UpdateCategory(r.Context(), user.UID, categoryID, service.CategoryPatch{
		Label:   req.Label,
		Default: req.Default,
		AIUse:   req.AIUse,
	})
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "update_category_failed", err.Error())
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, category)
}

func (api *API) deleteCategory(w http.ResponseWriter, r *http.Request) {
	user := mustUser(w, r)
	if user.UID == "" {
		return
	}
	categoryID := chi.URLParam(r, "id")
	if categoryID == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_id", "category id is required")
		return
	}

	if err := api.svc.DeleteCategory(r.Context(), user.UID, categoryID); err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "delete_category_failed", err.Error())
		return
	}
	writeStatusOK(w)
}

func (api *API) listStatements(w http.ResponseWriter, r *http.Request) {
	user := mustUser(w, r)
	if user.UID == "" {
		return
	}
	categoryID := chi.URLParam(r, "id")
	if categoryID == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_id", "category id is required")
		return
	}

	statements, err := api.svc.ListStatements(r.Context(), user.UID, categoryID)
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "statements_failed", err.Error())
		return
	}
	if statements == nil {
		statements = []models.Statement{}
	}
	httpapi.WriteJSON(w, http.StatusOK, statements)
}

func (api *API) createStatement(w http.ResponseWriter, r *http.Request) {
	user := mustUser(w, r)
	if user.UID == "" {
		return
	}
	var req struct {
		ID         string                  `json:"id"`
		CategoryID string                  `json:"categoryId"`
		Text       string                  `json:"text"`
		Created    int64                   `json:"created"`
		Questions  []service.QuestionInput `json:"questions"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	if len(req.Questions) > 0 {
		if _, err := api.svc.OnboardingPhrases(r.Context(), user.UID, req.Questions); err != nil {
			httpapi.WriteError(w, http.StatusInternalServerError, "onboarding_failed", err.Error())
			return
		}
		writeStatusOK(w)
		return
	}

	if strings.TrimSpace(req.CategoryID) == "" || strings.TrimSpace(req.Text) == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_payload", "categoryId and text are required")
		return
	}

	statement, err := api.svc.CreateStatement(r.Context(), user.UID, service.StatementInput{
		ID:         req.ID,
		CategoryID: req.CategoryID,
		Text:       req.Text,
		Created:    req.Created,
	})
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "create_statement_failed", err.Error())
		return
	}

	httpapi.WriteJSON(w, http.StatusOK, map[string]any{
		"status":    "ok",
		"statement": statement,
	})
}

func (api *API) patchStatement(w http.ResponseWriter, r *http.Request) {
	user := mustUser(w, r)
	if user.UID == "" {
		return
	}
	statementID := chi.URLParam(r, "id")
	if statementID == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_id", "statement id is required")
		return
	}

	var req struct {
		Text *string `json:"text"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if req.Text == nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_payload", "text is required")
		return
	}

	statement, err := api.svc.UpdateStatement(r.Context(), user.UID, statementID, service.StatementPatch{Text: req.Text})
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "update_statement_failed", err.Error())
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, statement)
}

func (api *API) deleteStatement(w http.ResponseWriter, r *http.Request) {
	user := mustUser(w, r)
	if user.UID == "" {
		return
	}
	statementID := chi.URLParam(r, "id")
	if statementID == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_id", "statement id is required")
		return
	}

	if err := api.svc.DeleteStatement(r.Context(), user.UID, statementID); err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "delete_statement_failed", err.Error())
		return
	}
	writeStatusOK(w)
}

func (api *API) getUserState(w http.ResponseWriter, r *http.Request) {
	user := mustUser(w, r)
	if user.UID == "" {
		return
	}
	state, err := api.svc.GetUserState(r.Context(), user.UID)
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "state_failed", err.Error())
		return
	}
	if state.Quickes == nil {
		state.Quickes = []string{}
	}
	if state.Preferences == nil {
		state.Preferences = map[string]any{}
	}
	httpapi.WriteJSON(w, http.StatusOK, state)
}

func (api *API) putUserState(w http.ResponseWriter, r *http.Request) {
	user := mustUser(w, r)
	if user.UID == "" {
		return
	}
	var raw map[string]json.RawMessage
	if err := decodeJSON(w, r, &raw); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	var patch service.UserStatePatch
	if value, ok := raw["inited"]; ok {
		if err := json.Unmarshal(value, &patch.Inited); err != nil {
			httpapi.WriteError(w, http.StatusBadRequest, "invalid_inited", err.Error())
			return
		}
	}
	if value, ok := raw["quickes"]; ok {
		patch.QuickesSet = true
		if err := json.Unmarshal(value, &patch.Quickes); err != nil {
			httpapi.WriteError(w, http.StatusBadRequest, "invalid_quickes", err.Error())
			return
		}
	}
	if value, ok := raw["preferences"]; ok {
		patch.PreferencesSet = true
		if err := json.Unmarshal(value, &patch.Preferences); err != nil {
			httpapi.WriteError(w, http.StatusBadRequest, "invalid_preferences", err.Error())
			return
		}
	}
	if patch.Inited == nil && !patch.QuickesSet && !patch.PreferencesSet {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_payload", "inited, quickes, or preferences required")
		return
	}

	state, err := api.svc.UpdateUserState(r.Context(), user.UID, patch)
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "state_update_failed", err.Error())
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, state)
}

func (api *API) getQuickes(w http.ResponseWriter, r *http.Request) {
	user := mustUser(w, r)
	if user.UID == "" {
		return
	}
	state, err := api.svc.GetUserState(r.Context(), user.UID)
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "quickes_failed", err.Error())
		return
	}
	if state.Quickes == nil {
		state.Quickes = []string{}
	}
	httpapi.WriteJSON(w, http.StatusOK, state.Quickes)
}

func (api *API) putQuickes(w http.ResponseWriter, r *http.Request) {
	user := mustUser(w, r)
	if user.UID == "" {
		return
	}
	var req struct {
		Quickes []string `json:"quickes"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if len(req.Quickes) == 0 {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_payload", "quickes are required")
		return
	}

	quickes, err := api.svc.SetQuickes(r.Context(), user.UID, req.Quickes)
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "quickes_update_failed", err.Error())
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, quickes)
}

func (api *API) listGlobalCategories(w http.ResponseWriter, r *http.Request) {
	user := mustUser(w, r)
	if user.UID == "" {
		return
	}
	includeStatements := r.URL.Query().Get("include_statements") == "true"
	categories, err := api.svc.ListGlobalCategories(r.Context(), includeStatements)
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "global_categories_failed", err.Error())
		return
	}
	if categories == nil {
		categories = []models.GlobalCategory{}
	}
	httpapi.WriteJSON(w, http.StatusOK, categories)
}

func (api *API) listGlobalStatements(w http.ResponseWriter, r *http.Request) {
	user := mustUser(w, r)
	if user.UID == "" {
		return
	}
	categoryID := chi.URLParam(r, "id")
	if categoryID == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_id", "category id is required")
		return
	}
	statements, err := api.svc.ListGlobalStatements(r.Context(), categoryID)
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "global_statements_failed", err.Error())
		return
	}
	if statements == nil {
		statements = []models.Statement{}
	}
	httpapi.WriteJSON(w, http.StatusOK, statements)
}

func (api *API) importGlobal(w http.ResponseWriter, r *http.Request) {
	user := mustUser(w, r)
	if user.UID == "" {
		return
	}
	var req struct {
		CategoryID string `json:"category_id"`
		Force      bool   `json:"force"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if strings.TrimSpace(req.CategoryID) == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_payload", "category_id is required")
		return
	}

	status, err := api.svc.ImportGlobalCategory(r.Context(), user.UID, req.CategoryID, req.Force)
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "import_failed", err.Error())
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, map[string]any{"status": status})
}

func (api *API) listFactoryQuestions(w http.ResponseWriter, r *http.Request) {
	user := mustUser(w, r)
	if user.UID == "" {
		return
	}
	questions, err := api.svc.ListFactoryQuestions(r.Context())
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "questions_failed", err.Error())
		return
	}
	if questions == nil {
		questions = []models.FactoryQuestion{}
	}
	httpapi.WriteJSON(w, http.StatusOK, questions)
}

func (api *API) onboardingPhrases(w http.ResponseWriter, r *http.Request) {
	user := mustUser(w, r)
	if user.UID == "" {
		return
	}
	var req struct {
		Questions []service.QuestionInput `json:"questions"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if len(req.Questions) == 0 {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_payload", "questions are required")
		return
	}

	if _, err := api.svc.OnboardingPhrases(r.Context(), user.UID, req.Questions); err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "onboarding_failed", err.Error())
		return
	}
	writeStatusOK(w)
}

func (api *API) deleteUser(w http.ResponseWriter, r *http.Request) {
	user := mustUser(w, r)
	if user.UID == "" {
		return
	}
	var req struct {
		DeleteFirebase bool `json:"delete_firebase"`
	}
	_ = decodeJSON(w, r, &req)

	if err := api.svc.DeleteUser(r.Context(), user.UID, req.DeleteFirebase); err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "delete_failed", err.Error())
		return
	}
	if req.DeleteFirebase && api.fbAuth != nil {
		_ = api.fbAuth.DeleteUser(r.Context(), user.UID)
	}
	writeStatusOK(w)
}

func (api *API) authToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_payload", "email and password are required")
		return
	}
	apiKey := strings.TrimSpace(api.config.Firebase.APIKey)
	if apiKey == "" {
		httpapi.WriteError(w, http.StatusServiceUnavailable, "auth_unavailable", "firebase api key not configured")
		return
	}

	payload := firebaseSignInRequest{
		Email:             req.Email,
		Password:          req.Password,
		ReturnSecureToken: true,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "auth_failed", "failed to build auth request")
		return
	}

	endpoint := "https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=" + url.QueryEscape(apiKey)
	authReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "auth_failed", err.Error())
		return
	}
	authReq.Header.Set("Content-Type", "application/json")

	resp, err := api.httpClient.Do(authReq)
	if err != nil {
		httpapi.WriteError(w, http.StatusBadGateway, "auth_failed", "firebase auth request failed")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		message := firebaseErrorMessage(resp.Body)
		status, code, msg := firebaseAuthError(message)
		httpapi.WriteError(w, status, code, msg)
		return
	}

	var authResp firebaseSignInResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		httpapi.WriteError(w, http.StatusBadGateway, "auth_failed", "invalid auth response")
		return
	}
	if authResp.IDToken == "" || authResp.LocalID == "" {
		httpapi.WriteError(w, http.StatusBadGateway, "auth_failed", "missing token in auth response")
		return
	}

	userPayload := map[string]string{
		"id":    authResp.LocalID,
		"email": authResp.Email,
	}

	if api.jwtManager != nil {
		tokenPair, err := api.jwtManager.GenerateTokenPair(authResp.LocalID, authResp.Email)
		if err != nil {
			httpapi.WriteError(w, http.StatusInternalServerError, "token_failed", "failed to generate tokens")
			return
		}
		api.setRefreshTokenCookie(w, tokenPair.RefreshToken)
		httpapi.WriteJSON(w, http.StatusOK, map[string]any{
			"token": tokenPair.AccessToken,
			"user":  userPayload,
		})
		return
	}

	httpapi.WriteJSON(w, http.StatusOK, map[string]any{
		"token": authResp.IDToken,
		"user":  userPayload,
	})
}

func (api *API) authRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_payload", "email and password are required")
		return
	}
	apiKey := strings.TrimSpace(api.config.Firebase.APIKey)
	if apiKey == "" {
		httpapi.WriteError(w, http.StatusServiceUnavailable, "auth_unavailable", "firebase api key not configured")
		return
	}

	payload := firebaseSignInRequest{
		Email:             req.Email,
		Password:          req.Password,
		ReturnSecureToken: true,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "auth_failed", "failed to build auth request")
		return
	}

	endpoint := "https://identitytoolkit.googleapis.com/v1/accounts:signUp?key=" + url.QueryEscape(apiKey)
	authReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "auth_failed", err.Error())
		return
	}
	authReq.Header.Set("Content-Type", "application/json")

	resp, err := api.httpClient.Do(authReq)
	if err != nil {
		httpapi.WriteError(w, http.StatusBadGateway, "auth_failed", "firebase auth request failed")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		message := firebaseErrorMessage(resp.Body)
		status, code, msg := firebaseAuthError(message)
		httpapi.WriteError(w, status, code, msg)
		return
	}

	var authResp firebaseSignInResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		httpapi.WriteError(w, http.StatusBadGateway, "auth_failed", "invalid auth response")
		return
	}
	if authResp.IDToken == "" || authResp.LocalID == "" {
		httpapi.WriteError(w, http.StatusBadGateway, "auth_failed", "missing token in auth response")
		return
	}

	userPayload := map[string]string{
		"id":    authResp.LocalID,
		"email": authResp.Email,
	}

	if api.jwtManager != nil {
		tokenPair, err := api.jwtManager.GenerateTokenPair(authResp.LocalID, authResp.Email)
		if err != nil {
			httpapi.WriteError(w, http.StatusInternalServerError, "token_failed", "failed to generate tokens")
			return
		}
		api.setRefreshTokenCookie(w, tokenPair.RefreshToken)
		httpapi.WriteJSON(w, http.StatusOK, map[string]any{
			"token": tokenPair.AccessToken,
			"user":  userPayload,
		})
		return
	}

	httpapi.WriteJSON(w, http.StatusOK, map[string]any{
		"token": authResp.IDToken,
		"user":  userPayload,
	})
}

func (api *API) authRefresh(w http.ResponseWriter, r *http.Request) {
	if api.jwtManager == nil {
		httpapi.WriteError(w, http.StatusServiceUnavailable, "jwt_unavailable", "jwt not configured")
		return
	}

	cookie, err := r.Cookie("refresh_token")
	if err != nil || cookie.Value == "" {
		httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing refresh token")
		return
	}

	claims, err := api.jwtManager.ValidateRefreshToken(cookie.Value)
	if err != nil {
		api.clearRefreshTokenCookie(w)
		httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "invalid refresh token")
		return
	}

	accessToken, expiresAt, err := api.jwtManager.GenerateAccessToken(claims.UID, claims.Email)
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "token_failed", "failed to generate access token")
		return
	}

	httpapi.WriteJSON(w, http.StatusOK, map[string]any{
		"token":     accessToken,
		"expiresAt": expiresAt.Unix(),
		"user": map[string]string{
			"id":    claims.UID,
			"email": claims.Email,
		},
	})
}

func (api *API) authLogout(w http.ResponseWriter, r *http.Request) {
	api.clearRefreshTokenCookie(w)
	writeStatusOK(w)
}

func (api *API) setRefreshTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/v1/auth",
		MaxAge:   int(api.jwtManager.RefreshTokenDuration().Seconds()),
		HttpOnly: true,
		Secure:   api.config.JWT.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		Domain:   api.config.JWT.CookieDomain,
	})
}

func (api *API) clearRefreshTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/v1/auth",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   api.config.JWT.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		Domain:   api.config.JWT.CookieDomain,
	})
}

func (api *API) proxyVoices(w http.ResponseWriter, r *http.Request) {
	api.proxyRequest(w, r, api.config.TTS.BaseURL+"/voices")
}

func (api *API) proxyTTS(w http.ResponseWriter, r *http.Request) {
	api.proxyRequest(w, r, api.config.TTS.BaseURL+"/tts")
}

func (api *API) predictorComplete(w http.ResponseWriter, r *http.Request) {
	if api.config.Predictor.APIKey == "" {
		httpapi.WriteError(w, http.StatusServiceUnavailable, "predictor_unavailable", "predictor api key not configured")
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_query", "q parameter is required")
		return
	}

	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "ru"
	}

	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "5"
	}

	predictorURL := "https://predictor.yandex.net/api/v1/predict.json/complete?key=" +
		url.QueryEscape(api.config.Predictor.APIKey) +
		"&q=" + url.QueryEscape(query) +
		"&lang=" + url.QueryEscape(lang) +
		"&limit=" + url.QueryEscape(limit)

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, predictorURL, nil)
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "predictor_failed", err.Error())
		return
	}

	resp, err := api.httpClient.Do(req)
	if err != nil {
		httpapi.WriteError(w, http.StatusBadGateway, "predictor_failed", "failed to call yandex predictor")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		httpapi.WriteError(w, http.StatusBadGateway, "predictor_failed", "yandex predictor returned error")
		return
	}

	var result struct {
		Text      []string `json:"text"`
		Pos       int      `json:"pos"`
		EndOfWord bool     `json:"endOfWord"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		httpapi.WriteError(w, http.StatusBadGateway, "predictor_failed", "invalid predictor response")
		return
	}

	httpapi.WriteJSON(w, http.StatusOK, result)
}

func (api *API) proxyRequest(w http.ResponseWriter, r *http.Request, target string) {
	if r.URL.RawQuery != "" {
		target = target + "?" + r.URL.RawQuery
	}
	req, err := http.NewRequestWithContext(r.Context(), r.Method, target, r.Body)
	if err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "proxy_failed", err.Error())
		return
	}
	req.Header = r.Header.Clone()

	resp, err := api.httpClient.Do(req)
	if err != nil {
		httpapi.WriteError(w, http.StatusBadGateway, "proxy_failed", err.Error())
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func mustUser(w http.ResponseWriter, r *http.Request) auth.User {
	user, ok := userctx.From(r.Context())
	if !ok {
		httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing user context")
		return auth.User{}
	}
	return user
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	body := http.MaxBytesReader(w, r.Body, 2<<20)
	dec := json.NewDecoder(body)
	if err := dec.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}
	return nil
}

func writeStatusOK(w http.ResponseWriter) {
	httpapi.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-Id")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Max-Age", "86400")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

var _ http.Handler = (*chi.Mux)(nil)

type firebaseSignInRequest struct {
	Email             string `json:"email"`
	Password          string `json:"password"`
	ReturnSecureToken bool   `json:"returnSecureToken"`
}

type firebaseSignInResponse struct {
	IDToken      string `json:"idToken"`
	LocalID      string `json:"localId"`
	Email        string `json:"email"`
	RefreshToken string `json:"refreshToken"`
}

type firebaseErrorResponse struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func firebaseErrorMessage(body io.Reader) string {
	payload, err := io.ReadAll(io.LimitReader(body, 1<<20))
	if err != nil || len(payload) == 0 {
		return ""
	}
	var resp firebaseErrorResponse
	if err := json.Unmarshal(payload, &resp); err == nil && resp.Error.Message != "" {
		return resp.Error.Message
	}
	return ""
}

func firebaseAuthError(message string) (int, string, string) {
	switch message {
	case "INVALID_PASSWORD", "EMAIL_NOT_FOUND":
		return http.StatusUnauthorized, "invalid_credentials", "invalid email or password"
	case "EMAIL_EXISTS":
		return http.StatusConflict, "email_exists", "email already registered"
	case "OPERATION_NOT_ALLOWED":
		return http.StatusForbidden, "operation_not_allowed", "registration disabled"
	case "USER_DISABLED":
		return http.StatusForbidden, "user_disabled", "user disabled"
	case "INVALID_EMAIL":
		return http.StatusBadRequest, "invalid_email", "invalid email"
	case "MISSING_EMAIL", "MISSING_PASSWORD":
		return http.StatusBadRequest, "invalid_payload", "email and password are required"
	case "TOO_MANY_ATTEMPTS_TRY_LATER":
		return http.StatusTooManyRequests, "rate_limited", "too many attempts, try later"
	default:
		if strings.HasPrefix(message, "WEAK_PASSWORD") {
			return http.StatusBadRequest, "weak_password", "password is too weak"
		}
		return http.StatusBadGateway, "auth_failed", "firebase auth failed"
	}
}

func (api *API) adminStats(w http.ResponseWriter, r *http.Request) {
	window := r.URL.Query().Get("window")
	if window == "" {
		window = "24h"
	}
	duration, err := time.ParseDuration(window)
	if err != nil {
		duration = 24 * time.Hour
	}
	since := time.Now().Add(-duration)

	stats, err := api.svc.AdminStats(r.Context(), since)
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "stats_failed", err.Error())
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, stats)
}

func (api *API) adminListAdmins(w http.ResponseWriter, r *http.Request) {
	admins, err := api.svc.ListAdmins(r.Context())
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "admins_failed", err.Error())
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, map[string]interface{}{"items": admins})
}

func (api *API) adminAddAdmin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"user_id"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if strings.TrimSpace(req.UserID) == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_payload", "user_id is required")
		return
	}

	if err := api.svc.AddAdmin(r.Context(), req.UserID); err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "add_admin_failed", err.Error())
		return
	}
	writeStatusOK(w)
}

func (api *API) adminRemoveAdmin(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	if userID == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_id", "user_id is required")
		return
	}

	if err := api.svc.RemoveAdmin(r.Context(), userID); err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "remove_admin_failed", err.Error())
		return
	}
	writeStatusOK(w)
}

func (api *API) adminListClientKeys(w http.ResponseWriter, r *http.Request) {
	keys, err := api.svc.ListClientKeys(r.Context())
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "keys_failed", err.Error())
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, map[string]interface{}{"items": keys})
}

func (api *API) adminCreateClientKey(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ClientID string `json:"client_id"`
		Status   string `json:"status"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if strings.TrimSpace(req.ClientID) == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_payload", "client_id is required")
		return
	}
	if req.Status == "" {
		req.Status = "active"
	}

	keyPlain, keyHash, err := generateClientKey()
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "key_generation_failed", err.Error())
		return
	}

	key := store.ClientKey{
		KeyHash:   keyHash,
		ClientID:  req.ClientID,
		Status:    req.Status,
		CreatedAt: time.Now().UnixMilli(),
	}

	if err := api.svc.CreateClientKey(r.Context(), key); err != nil {
		log.Printf("CreateClientKey error: %v", err)
		httpapi.WriteError(w, http.StatusInternalServerError, "key_create_failed", fmt.Sprintf("failed to create client key: %v", err))
		return
	}

	httpapi.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"api_key":    keyPlain,
		"key_hash":   keyHash,
		"client_id":  req.ClientID,
		"status":     req.Status,
		"created_at": key.CreatedAt,
	})
}

func (api *API) adminRevokeClientKey(w http.ResponseWriter, r *http.Request) {
	keyHash := chi.URLParam(r, "key_hash")
	if keyHash == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_hash", "key_hash is required")
		return
	}

	if err := api.svc.RevokeClientKey(r.Context(), keyHash); err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "key_revoke_failed", err.Error())
		return
	}
	writeStatusOK(w)
}

func (api *API) adminListGlobalCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := api.svc.ListGlobalCategories(r.Context(), true)
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "global_categories_failed", err.Error())
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, categories)
}

func (api *API) adminCreateGlobalCategory(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID      string `json:"id"`
		Label   string `json:"label"`
		Default *bool  `json:"default"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if strings.TrimSpace(req.Label) == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_label", "label is required")
		return
	}

	category, err := api.svc.CreateGlobalCategory(r.Context(), service.GlobalCategoryInput{
		ID:      req.ID,
		Label:   req.Label,
		Default: req.Default,
	})
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "create_global_category_failed", err.Error())
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, category)
}

func (api *API) adminUpdateGlobalCategory(w http.ResponseWriter, r *http.Request) {
	categoryID := chi.URLParam(r, "id")
	if categoryID == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_id", "category id is required")
		return
	}

	var req struct {
		Label   *string `json:"label"`
		Default *bool   `json:"default"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	category, err := api.svc.UpdateGlobalCategory(r.Context(), categoryID, service.GlobalCategoryPatch{
		Label:   req.Label,
		Default: req.Default,
	})
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "update_global_category_failed", err.Error())
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, category)
}

func (api *API) adminDeleteGlobalCategory(w http.ResponseWriter, r *http.Request) {
	categoryID := chi.URLParam(r, "id")
	if categoryID == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_id", "category id is required")
		return
	}

	if err := api.svc.DeleteGlobalCategory(r.Context(), categoryID); err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "delete_global_category_failed", err.Error())
		return
	}
	writeStatusOK(w)
}

func (api *API) adminListFactoryQuestions(w http.ResponseWriter, r *http.Request) {
	questions, err := api.svc.ListFactoryQuestions(r.Context())
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "questions_failed", err.Error())
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, questions)
}

func (api *API) adminCreateFactoryQuestion(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID         string   `json:"id"`
		Label      string   `json:"label"`
		Phrases    []string `json:"phrases"`
		Category   string   `json:"category"`
		Type       string   `json:"type"`
		OrderIndex int      `json:"order_index"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if strings.TrimSpace(req.Label) == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_label", "label is required")
		return
	}

	question, err := api.svc.CreateFactoryQuestion(r.Context(), service.FactoryQuestionInput{
		ID:         req.ID,
		Label:      req.Label,
		Phrases:    req.Phrases,
		Category:   req.Category,
		Type:       req.Type,
		OrderIndex: req.OrderIndex,
	})
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "create_question_failed", err.Error())
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, question)
}

func (api *API) adminUpdateFactoryQuestion(w http.ResponseWriter, r *http.Request) {
	questionID := chi.URLParam(r, "id")
	if questionID == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_id", "question id is required")
		return
	}

	var req struct {
		Label      *string  `json:"label"`
		Phrases    []string `json:"phrases"`
		Category   *string  `json:"category"`
		Type       *string  `json:"type"`
		OrderIndex *int     `json:"order_index"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	question, err := api.svc.UpdateFactoryQuestion(r.Context(), questionID, service.FactoryQuestionPatch{
		Label:      req.Label,
		Phrases:    req.Phrases,
		Category:   req.Category,
		Type:       req.Type,
		OrderIndex: req.OrderIndex,
	})
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "update_question_failed", err.Error())
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, question)
}

func (api *API) adminDeleteFactoryQuestion(w http.ResponseWriter, r *http.Request) {
	questionID := chi.URLParam(r, "id")
	if questionID == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_id", "question id is required")
		return
	}

	if err := api.svc.DeleteFactoryQuestion(r.Context(), questionID); err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "delete_question_failed", err.Error())
		return
	}
	writeStatusOK(w)
}

func generateClientKey() (string, string, error) {
	buf := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		return "", "", err
	}
	keyPlain := "ltk_" + hex.EncodeToString(buf)
	hash := sha256.Sum256([]byte(keyPlain))
	keyHash := hex.EncodeToString(hash[:])
	return keyPlain, keyHash, nil
}
