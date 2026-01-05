package httpmiddleware

import (
	"net/http"

	"github.com/linkasu/linka.type-backend/internal/httpapi"
	"github.com/linkasu/linka.type-backend/internal/service"
	"github.com/linkasu/linka.type-backend/internal/userctx"
)

// Admin requires the user to be authenticated and have admin status.
func Admin(svc *service.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := userctx.From(r.Context())
			if !ok || user.UID == "" {
				httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
				return
			}

			isAdmin, err := svc.IsAdmin(r.Context(), user.UID)
			if err != nil {
				httpapi.WriteError(w, http.StatusInternalServerError, "admin_check_failed", "failed to check admin status")
				return
			}
			if !isAdmin {
				httpapi.WriteError(w, http.StatusForbidden, "forbidden", "admin access required")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

