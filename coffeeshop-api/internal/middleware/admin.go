package middleware

import "net/http"

// AdminOnly wraps a handler to require the super_admin role.
// Must be used inside Auth() — it reads the role from context.
func AdminOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		role := GetUserRole(r.Context())
		if role != "super_admin" {
			http.Error(w, `{"error":"forbidden: super_admin role required","code":403}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}
}
