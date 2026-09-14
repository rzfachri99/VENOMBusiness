package auth

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/venombusiness/venombusiness/apps/api/internal/httpx"
	"github.com/venombusiness/venombusiness/apps/api/internal/platform/id"
	"github.com/venombusiness/venombusiness/apps/api/internal/platform/security"
	"github.com/venombusiness/venombusiness/apps/api/internal/platform/session"
)

const CookieName = "venom_session"

type Handler struct {
	Repo         Repository
	CookieSecure bool
	SessionTTL   time.Duration
}

type registerRequest struct {
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h Handler) Register(w http.ResponseWriter, r *http.Request) {
	var in registerRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	in.DisplayName = strings.TrimSpace(in.DisplayName)
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	if len(in.DisplayName) < 2 || len(in.DisplayName) > 80 {
		httpx.Error(w, http.StatusUnprocessableEntity, "display name must be 2-80 characters")
		return
	}
	if _, err := mail.ParseAddress(in.Email); err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, "invalid email address")
		return
	}
	hash, err := security.HashPassword(in.Password)
	if err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	userID, err := id.New()
	if err != nil {
		httpx.Error(w, 500, "could not create account")
		return
	}
	if err := h.Repo.CreateUser(r.Context(), userID, in.Email, in.DisplayName, hash); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			httpx.Error(w, http.StatusConflict, "an account with this email already exists")
			return
		}
		httpx.Error(w, 500, "could not create account")
		return
	}
	if err := h.issueSession(w, r, userID); err != nil {
		httpx.Error(w, 500, "account created but session could not be started")
		return
	}
	httpx.JSON(w, http.StatusCreated, map[string]any{"data": Principal{User: User{ID: userID, Email: in.Email, DisplayName: in.DisplayName}}})
}

func (h Handler) Login(w http.ResponseWriter, r *http.Request) {
	var in loginRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	u, err := h.Repo.UserByEmail(r.Context(), strings.ToLower(strings.TrimSpace(in.Email)))
	if errors.Is(err, sql.ErrNoRows) {
		httpx.Error(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if err != nil || !u.IsActive {
		httpx.Error(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	ok, err := security.VerifyPassword(u.PasswordHash, in.Password)
	if err != nil || !ok {
		httpx.Error(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if err := h.issueSession(w, r, u.ID); err != nil {
		httpx.Error(w, 500, "could not start session")
		return
	}
	company, _ := h.Repo.CompanyForUser(r.Context(), u.ID)
	httpx.JSON(w, 200, map[string]any{"data": Principal{User: u.User, Company: company}})
}

func (h Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(CookieName); err == nil && c.Value != "" {
		_ = h.Repo.DeleteSession(r.Context(), session.Digest(c.Value))
	}
	h.clearCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h Handler) Me(w http.ResponseWriter, r *http.Request) {
	p, ok := PrincipalFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}
	httpx.JSON(w, 200, map[string]any{"data": p})
}

func (h Handler) issueSession(w http.ResponseWriter, r *http.Request, userID string) error {
	plain, digest, err := session.NewToken()
	if err != nil {
		return err
	}
	sessionID, err := id.New()
	if err != nil {
		return err
	}
	ttl := h.SessionTTL
	if ttl <= 0 {
		ttl = 7 * 24 * time.Hour
	}
	expires := time.Now().UTC().Add(ttl)
	if err := h.Repo.CreateSession(r.Context(), sessionID, userID, digest, expires.Format("2006-01-02 15:04:05"), truncate(r.UserAgent(), 255)); err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{Name: CookieName, Value: plain, Path: "/", HttpOnly: true, Secure: h.CookieSecure, SameSite: http.SameSiteLaxMode, MaxAge: int(ttl.Seconds()), Expires: expires})
	return nil
}

func (h Handler) clearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: CookieName, Value: "", Path: "/", HttpOnly: true, Secure: h.CookieSecure, SameSite: http.SameSiteLaxMode, MaxAge: -1, Expires: time.Unix(1, 0)})
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
