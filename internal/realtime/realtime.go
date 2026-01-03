package realtime

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/linkasu/linka.type-backend/internal/auth"
	"github.com/linkasu/linka.type-backend/internal/httpapi"
	"github.com/linkasu/linka.type-backend/internal/httpmiddleware"
	"github.com/linkasu/linka.type-backend/internal/models"
	"github.com/linkasu/linka.type-backend/internal/store"
	"github.com/linkasu/linka.type-backend/internal/userctx"
	"nhooyr.io/websocket"
)

// Server handles realtime APIs.
type Server struct {
	store store.Store
}

// New builds the realtime router.
func New(store store.Store, verifier auth.Verifier) http.Handler {
	s := &Server{store: store}

	r := chi.NewRouter()
	r.Use(httpmiddleware.RequestID)
	r.Use(httpmiddleware.Auth(verifier))

	r.Get("/v1/changes", s.longPoll)
	r.Get("/v1/stream", s.stream)

	return r
}

func (s *Server) longPoll(w http.ResponseWriter, r *http.Request) {
	user := mustUser(w, r)
	if user.UID == "" {
		return
	}
	cursor := r.URL.Query().Get("cursor")
	limit := parseLimit(r.URL.Query().Get("limit"), 100)
	timeout := parseTimeout(r.URL.Query().Get("timeout"), 25*time.Second)

	deadline := time.Now().Add(timeout)

	for {
		ctx := r.Context()
		nextCursor, changes, err := s.store.ListChanges(ctx, user.UID, cursor, limit)
		if err != nil {
			httpapi.WriteError(w, http.StatusInternalServerError, "changes_failed", err.Error())
			return
		}
		if len(changes) > 0 {
			writeChanges(w, nextCursor, changes)
			return
		}
		if time.Now().After(deadline) {
			writeChanges(w, cursor, nil)
			return
		}
		select {
		case <-ctx.Done():
			httpapi.WriteError(w, http.StatusRequestTimeout, "timeout", "request canceled")
			return
		case <-time.After(500 * time.Millisecond):
		}
	}
}

func (s *Server) stream(w http.ResponseWriter, r *http.Request) {
	user := mustUser(w, r)
	if user.UID == "" {
		return
	}
	cursor := r.URL.Query().Get("cursor")
	limit := parseLimit(r.URL.Query().Get("limit"), 100)

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: []string{"*"}})
	if err != nil {
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	ctx := r.Context()
	pollInterval := 1 * time.Second
	heartbeatInterval := 25 * time.Second
	lastSend := time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		nextCursor, changes, err := s.store.ListChanges(ctx, user.UID, cursor, limit)
		if err != nil {
			_ = conn.Close(websocket.StatusInternalError, "changes_failed")
			return
		}

		if len(changes) > 0 {
			cursor = nextCursor
			payload := map[string]any{
				"type":    "changes",
				"cursor":  cursor,
				"changes": stripCursor(changes),
			}
			if err := writeWS(ctx, conn, payload); err != nil {
				return
			}
			lastSend = time.Now()
			continue
		}

		if time.Since(lastSend) >= heartbeatInterval {
			payload := map[string]any{
				"type":   "heartbeat",
				"cursor": cursor,
			}
			if err := writeWS(ctx, conn, payload); err != nil {
				return
			}
			lastSend = time.Now()
		}

		timer := time.NewTimer(pollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

func writeWS(ctx context.Context, conn *websocket.Conn, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return conn.Write(ctx, websocket.MessageText, data)
}

func writeChanges(w http.ResponseWriter, cursor string, changes []models.ChangeEvent) {
	resp := map[string]any{
		"cursor":  cursor,
		"changes": stripCursor(changes),
	}
	httpapi.WriteJSON(w, http.StatusOK, resp)
}

func stripCursor(changes []models.ChangeEvent) []models.ChangeEvent {
	out := make([]models.ChangeEvent, 0, len(changes))
	for _, change := range changes {
		change.Cursor = ""
		out = append(out, change)
	}
	return out
}

func mustUser(w http.ResponseWriter, r *http.Request) auth.User {
	user, ok := userctx.From(r.Context())
	if !ok {
		httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing user context")
		return auth.User{}
	}
	return user
}

func parseLimit(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	if value > 500 {
		return 500
	}
	return value
}

func parseTimeout(raw string, fallback time.Duration) time.Duration {
	if raw == "" {
		return fallback
	}
	value, err := time.ParseDuration(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	if value > 60*time.Second {
		return 60 * time.Second
	}
	return value
}
