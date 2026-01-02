package httpmiddleware

import (
	"net/http"
	"strings"

	"github.com/linkasu/linka.type-backend/internal/auth"
	"github.com/linkasu/linka.type-backend/internal/httpapi"
	"github.com/linkasu/linka.type-backend/internal/userctx"
)

// Auth verifies a Firebase bearer token and injects the user into context.
func Auth(verifier auth.Verifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r.Header.Get("Authorization"))
			if token == "" {
				httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing bearer token")
				return
			}

			user, err := verifier.Verify(r.Context(), token)
			if err != nil {
				httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "invalid token")
				return
			}

			ctx := userctx.With(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerToken(header string) string {
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
