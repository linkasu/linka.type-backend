package coreapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/linkasu/linka.type-backend/internal/dialoghelper"
	"github.com/linkasu/linka.type-backend/internal/httpapi"
	"github.com/linkasu/linka.type-backend/internal/models"
	"github.com/linkasu/linka.type-backend/internal/service"
)

type dialogMessagePayload struct {
	Role               string `json:"role"`
	Content            string `json:"content"`
	Source             string `json:"source"`
	Created            int64  `json:"created"`
	IncludeSuggestions bool   `json:"includeSuggestions"`
}

func (api *API) listDialogChats(w http.ResponseWriter, r *http.Request) {
	user := mustUser(w, r)
	if user.UID == "" {
		return
	}
	chats, err := api.svc.ListDialogChats(r.Context(), user.UID)
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "dialog_chats_failed", err.Error())
		return
	}
	if chats == nil {
		chats = []models.DialogChat{}
	}
	httpapi.WriteJSON(w, http.StatusOK, chats)
}

func (api *API) createDialogChat(w http.ResponseWriter, r *http.Request) {
	user := mustUser(w, r)
	if user.UID == "" {
		return
	}
	var req struct {
		Title string `json:"title"`
	}
	if err := decodeJSON(w, r, &req); err != nil && !errors.Is(err, io.EOF) {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	chat, err := api.svc.CreateDialogChat(r.Context(), user.UID, service.DialogChatInput{
		Title: req.Title,
	})
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "create_chat_failed", err.Error())
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, chat)
}

func (api *API) deleteDialogChat(w http.ResponseWriter, r *http.Request) {
	user := mustUser(w, r)
	if user.UID == "" {
		return
	}
	chatID := chi.URLParam(r, "id")
	if chatID == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_id", "chat id is required")
		return
	}
	if err := api.svc.DeleteDialogChat(r.Context(), user.UID, chatID); err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "delete_chat_failed", err.Error())
		return
	}
	writeStatusOK(w)
}

func (api *API) listDialogMessages(w http.ResponseWriter, r *http.Request) {
	user := mustUser(w, r)
	if user.UID == "" {
		return
	}
	chatID := chi.URLParam(r, "id")
	if chatID == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_id", "chat id is required")
		return
	}
	limit := parseIntParam(r.URL.Query().Get("limit"), 200)
	before := parseInt64Param(r.URL.Query().Get("before"), 0)

	messages, err := api.svc.ListDialogMessages(r.Context(), user.UID, chatID, limit, before)
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "dialog_messages_failed", err.Error())
		return
	}
	if messages == nil {
		messages = []models.DialogMessage{}
	}
	httpapi.WriteJSON(w, http.StatusOK, messages)
}

func (api *API) createDialogMessage(w http.ResponseWriter, r *http.Request) {
	user := mustUser(w, r)
	if user.UID == "" {
		return
	}
	chatID := chi.URLParam(r, "id")
	if chatID == "" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_id", "chat id is required")
		return
	}

	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "multipart/form-data") {
		payload, audio, err := api.readDialogMultipart(w, r)
		if err != nil {
			httpapi.WriteError(w, http.StatusBadRequest, "invalid_payload", err.Error())
			return
		}
		result, err := api.svc.CreateDialogMessage(r.Context(), user.UID, chatID, service.DialogMessageInput{
			Role:               payload.Role,
			Content:            payload.Content,
			Source:             payload.Source,
			Created:            payload.Created,
			IncludeSuggestions: payload.IncludeSuggestions,
		}, audio)
		if err != nil {
			httpapi.WriteError(w, http.StatusInternalServerError, "create_dialog_message_failed", err.Error())
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, result)
		return
	}

	var req dialogMessagePayload
	if err := decodeJSON(w, r, &req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	result, err := api.svc.CreateDialogMessage(r.Context(), user.UID, chatID, service.DialogMessageInput{
		Role:               req.Role,
		Content:            req.Content,
		Source:             req.Source,
		Created:            req.Created,
		IncludeSuggestions: req.IncludeSuggestions,
	}, nil)
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "create_dialog_message_failed", err.Error())
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, result)
}

func (api *API) listDialogSuggestions(w http.ResponseWriter, r *http.Request) {
	user := mustUser(w, r)
	if user.UID == "" {
		return
	}
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "pending"
	}
	limit := parseIntParam(r.URL.Query().Get("limit"), 200)

	suggestions, err := api.svc.ListDialogSuggestions(r.Context(), user.UID, status, limit)
	if err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "dialog_suggestions_failed", err.Error())
		return
	}
	if suggestions == nil {
		suggestions = []models.DialogSuggestion{}
	}
	httpapi.WriteJSON(w, http.StatusOK, suggestions)
}

func (api *API) applyDialogSuggestions(w http.ResponseWriter, r *http.Request) {
	user := mustUser(w, r)
	if user.UID == "" {
		return
	}
	var req struct {
		Items []service.DialogSuggestionApplyItem `json:"items"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	result, err := api.svc.ApplyDialogSuggestions(r.Context(), user.UID, req.Items)
	if err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "apply_suggestions_failed", err.Error())
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, result)
}

func (api *API) dismissDialogSuggestions(w http.ResponseWriter, r *http.Request) {
	user := mustUser(w, r)
	if user.UID == "" {
		return
	}
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if err := api.svc.DismissDialogSuggestions(r.Context(), user.UID, req.IDs); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "dismiss_suggestions_failed", err.Error())
		return
	}
	writeStatusOK(w)
}

func (api *API) readDialogMultipart(w http.ResponseWriter, r *http.Request) (dialogMessagePayload, *dialoghelper.AudioPayload, error) {
	maxBytes := api.config.Dialog.MaxAudioBytes
	if maxBytes <= 0 {
		maxBytes = 8 * 1024 * 1024
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	if err := r.ParseMultipartForm(maxBytes); err != nil {
		return dialogMessagePayload{}, nil, err
	}

	payloadField := r.FormValue("payload")
	if payloadField == "" {
		payloadField = r.FormValue("metadata")
	}
	if payloadField == "" {
		payloadField = r.FormValue("json")
	}
	if payloadField == "" {
		return dialogMessagePayload{}, nil, errors.New("payload is required")
	}

	var payload dialogMessagePayload
	if err := json.Unmarshal([]byte(payloadField), &payload); err != nil {
		return dialogMessagePayload{}, nil, fmt.Errorf("invalid payload json: %w", err)
	}
	if strings.TrimSpace(payload.Role) == "" {
		payload.Role = "speaker"
	}

	file, header, err := findMultipartAudio(r)
	if err != nil {
		return dialogMessagePayload{}, nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return dialogMessagePayload{}, nil, errors.New("failed to read audio")
	}

	contentType := header.Header.Get("Content-Type")
	return payload, &dialoghelper.AudioPayload{
		Data:        data,
		Filename:    header.Filename,
		ContentType: contentType,
	}, nil
}

func findMultipartAudio(r *http.Request) (multipart.File, *multipart.FileHeader, error) {
	if r.MultipartForm == nil {
		return nil, nil, errors.New("multipart form is required")
	}
	for _, field := range []string{"audio", "file"} {
		if headers := r.MultipartForm.File[field]; len(headers) > 0 {
			file, err := headers[0].Open()
			if err != nil {
				return nil, nil, errors.New("failed to open audio")
			}
			return file, headers[0], nil
		}
	}
	for _, headers := range r.MultipartForm.File {
		if len(headers) == 0 {
			continue
		}
		file, err := headers[0].Open()
		if err != nil {
			return nil, nil, errors.New("failed to open audio")
		}
		return file, headers[0], nil
	}
	return nil, nil, errors.New("audio file is required")
}

func parseIntParam(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func parseInt64Param(raw string, fallback int64) int64 {
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}
