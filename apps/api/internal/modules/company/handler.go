package company

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/venombusiness/venombusiness/apps/api/internal/httpx"
	"github.com/venombusiness/venombusiness/apps/api/internal/modules/auth"
	"github.com/venombusiness/venombusiness/apps/api/internal/platform/id"
)

type Handler struct{ DB *sql.DB }

type setupRequest struct {
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Timezone string `json:"timezone"`
	Currency string `json:"currency"`
}

var slugRE = regexp.MustCompile(`[^a-z0-9]+`)

func (h Handler) Setup(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.PrincipalFromContext(r.Context())
	if p.Company != nil {
		httpx.Error(w, http.StatusConflict, "business workspace already exists")
		return
	}
	var in setupRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpx.Error(w, 400, "invalid JSON payload")
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	if len(in.Name) < 2 || len(in.Name) > 120 {
		httpx.Error(w, 422, "business name must be 2-120 characters")
		return
	}
	in.Slug = normalizeSlug(in.Slug)
	if in.Slug == "" {
		in.Slug = normalizeSlug(in.Name)
	}
	if len(in.Slug) < 2 || len(in.Slug) > 80 {
		httpx.Error(w, 422, "invalid business slug")
		return
	}
	if in.Timezone == "" {
		in.Timezone = "Asia/Jakarta"
	}
	if in.Currency == "" {
		in.Currency = "IDR"
	}
	in.Currency = strings.ToUpper(strings.TrimSpace(in.Currency))
	if len(in.Currency) != 3 {
		httpx.Error(w, 422, "currency must use a 3-letter ISO code")
		return
	}
	companyID, err := id.New()
	if err != nil {
		httpx.Error(w, 500, "could not create workspace")
		return
	}
	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		httpx.Error(w, 500, "could not create workspace")
		return
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(r.Context(), `INSERT INTO companies(id,name,slug,timezone,currency) VALUES(?,?,?,?,?)`, companyID, in.Name, in.Slug, in.Timezone, in.Currency); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			httpx.Error(w, 409, "business slug is already used")
			return
		}
		httpx.Error(w, 500, "could not create workspace")
		return
	}
	if _, err = tx.ExecContext(r.Context(), `INSERT INTO company_members(company_id,user_id,role) VALUES(?,?,'owner')`, companyID, p.User.ID); err != nil {
		httpx.Error(w, 500, "could not assign owner")
		return
	}
	auditID, _ := id.New()
	_, _ = tx.ExecContext(r.Context(), `INSERT INTO audit_logs(id,company_id,actor_user_id,action,entity_type,entity_id) VALUES(?,?,?,?,?,?)`, auditID, companyID, p.User.ID, "company.created", "company", companyID)
	if err = tx.Commit(); err != nil {
		httpx.Error(w, 500, "could not create workspace")
		return
	}
	httpx.JSON(w, 201, map[string]any{"data": auth.CompanyContext{ID: companyID, Name: in.Name, Slug: in.Slug, Timezone: in.Timezone, Currency: in.Currency, Role: "owner"}})
}

func normalizeSlug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugRE.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

type profileRequest struct {
	LegalName         string `json:"legal_name"`
	TaxID             string `json:"tax_id"`
	Email             string `json:"email"`
	Phone             string `json:"phone"`
	Address           string `json:"address"`
	BankName          string `json:"bank_name"`
	BankAccountName   string `json:"bank_account_name"`
	BankAccountNumber string `json:"bank_account_number"`
	QRISText          string `json:"qris_text"`
}

func (h Handler) Profile(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.PrincipalFromContext(r.Context())
	var x profileRequest
	err := h.DB.QueryRowContext(r.Context(), `SELECT COALESCE(legal_name,''),COALESCE(tax_id,''),COALESCE(email,''),COALESCE(phone,''),COALESCE(address,''),COALESCE(bank_name,''),COALESCE(bank_account_name,''),COALESCE(bank_account_number,''),COALESCE(qris_text,'') FROM companies WHERE id=?`, p.Company.ID).Scan(&x.LegalName, &x.TaxID, &x.Email, &x.Phone, &x.Address, &x.BankName, &x.BankAccountName, &x.BankAccountNumber, &x.QRISText)
	if err != nil {
		httpx.Error(w, 500, "could not load business profile")
		return
	}
	httpx.JSON(w, 200, map[string]any{"data": x})
}
func (h Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.PrincipalFromContext(r.Context())
	var x profileRequest
	if json.NewDecoder(r.Body).Decode(&x) != nil {
		httpx.Error(w, 400, "invalid JSON payload")
		return
	}
	_, err := h.DB.ExecContext(r.Context(), `UPDATE companies SET legal_name=NULLIF(?,''),tax_id=NULLIF(?,''),email=NULLIF(?,''),phone=NULLIF(?,''),address=NULLIF(?,''),bank_name=NULLIF(?,''),bank_account_name=NULLIF(?,''),bank_account_number=NULLIF(?,''),qris_text=NULLIF(?,''),updated_at=CURRENT_TIMESTAMP WHERE id=?`, strings.TrimSpace(x.LegalName), strings.TrimSpace(x.TaxID), strings.TrimSpace(x.Email), strings.TrimSpace(x.Phone), strings.TrimSpace(x.Address), strings.TrimSpace(x.BankName), strings.TrimSpace(x.BankAccountName), strings.TrimSpace(x.BankAccountNumber), strings.TrimSpace(x.QRISText), p.Company.ID)
	if err != nil {
		httpx.Error(w, 500, "could not save business profile")
		return
	}
	w.WriteHeader(204)
}
