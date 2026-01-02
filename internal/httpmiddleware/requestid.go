package httpmiddleware

import (
	"net/http"

	"github.com/linkasu/linka.type-backend/internal/requestid"
)

// RequestID ensures every request has a request ID and injects it into context.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get(requestid.Header)
		if rid == "" {
			rid = requestid.New()
		}
		w.Header().Set(requestid.Header, rid)
		ctx := requestid.WithContext(r.Context(), rid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
