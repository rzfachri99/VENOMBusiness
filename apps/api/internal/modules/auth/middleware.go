package auth

import (
	"context"
	"net/http"

	"github.com/venombusiness/venombusiness/apps/api/internal/httpx"
	"github.com/venombusiness/venombusiness/apps/api/internal/platform/session"
)

type principalKey struct{}

type Middleware struct{ Repo Repository }

func (m Middleware) RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(CookieName)
		if err != nil || cookie.Value == "" {
			httpx.Error(w, http.StatusUnauthorized, "authentication required")
			return
		}
		userID, err := m.Repo.UserIDBySessionHash(r.Context(), session.Digest(cookie.Value))
		if err != nil {
			httpx.Error(w, http.StatusUnauthorized, "invalid or expired session")
			return
		}
		user, err := m.Repo.UserByID(r.Context(), userID)
		if err != nil {
			httpx.Error(w, http.StatusUnauthorized, "invalid account")
			return
		}
		company, err := m.Repo.CompanyForUser(r.Context(), userID)
		if err != nil {
			httpx.Error(w, 500, "could not resolve workspace")
			return
		}
		ctx := context.WithValue(r.Context(), principalKey{}, Principal{User: user, Company: company})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m Middleware) RequireCompany(next http.Handler) http.Handler {
	return m.RequireUser(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, _ := PrincipalFromContext(r.Context())
		if p.Company == nil {
			httpx.Error(w, http.StatusPreconditionRequired, "business setup required")
			return
		}
		next.ServeHTTP(w, r)
	}))
}

func (m Middleware) RequireRoles(roles ...string) func(http.Handler) http.Handler {
	allowed := map[string]bool{}
	for _, role := range roles {
		allowed[role] = true
	}
	return func(next http.Handler) http.Handler {
		return m.RequireCompany(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, _ := PrincipalFromContext(r.Context())
			if p.Company == nil || !allowed[p.Company.Role] {
				httpx.Error(w, http.StatusForbidden, "permission denied")
				return
			}
			next.ServeHTTP(w, r)
		}))
	}
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(principalKey{}).(Principal)
	return p, ok
}
