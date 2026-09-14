package customer

import (
	"encoding/json"
	"net/http"
	"net/mail"
	"strings"

	"github.com/venombusiness/venombusiness/apps/api/internal/httpx"
	"github.com/venombusiness/venombusiness/apps/api/internal/modules/auth"
	"github.com/venombusiness/venombusiness/apps/api/internal/platform/id"
)

type Handler struct{ Repo Repository }

type input struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	TaxID   string `json:"tax_id"`
	Address string `json:"address"`
	Notes   string `json:"notes"`
}

func (h Handler) List(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.PrincipalFromContext(r.Context())
	items, err := h.Repo.List(r.Context(), p.Company.ID)
	if err != nil {
		httpx.Error(w, 500, "could not load customers")
		return
	}
	httpx.JSON(w, 200, map[string]any{"data": items})
}

func (h Handler) Create(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.PrincipalFromContext(r.Context())
	in, ok := decodeInput(w, r)
	if !ok {
		return
	}
	cid, err := id.New()
	if err != nil {
		httpx.Error(w, 500, "could not create customer")
		return
	}
	c := Customer{ID: cid, CompanyID: p.Company.ID, Name: in.Name, Email: in.Email, Phone: in.Phone, TaxID: in.TaxID, Address: in.Address, Notes: in.Notes}
	if err := h.Repo.Create(r.Context(), c); err != nil {
		httpx.Error(w, 500, "could not create customer")
		return
	}
	auditID, _ := id.New()
	h.Repo.Audit(r.Context(), auditID, p.Company.ID, p.User.ID, "customer.created", c.ID)
	httpx.JSON(w, 201, map[string]any{"data": c})
}

func (h Handler) Update(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.PrincipalFromContext(r.Context())
	in, ok := decodeInput(w, r)
	if !ok {
		return
	}
	c := Customer{ID: r.PathValue("id"), CompanyID: p.Company.ID, Name: in.Name, Email: in.Email, Phone: in.Phone, TaxID: in.TaxID, Address: in.Address, Notes: in.Notes}
	found, err := h.Repo.Update(r.Context(), c)
	if err != nil {
		httpx.Error(w, 500, "could not update customer")
		return
	}
	if !found {
		httpx.Error(w, 404, "customer not found")
		return
	}
	auditID, _ := id.New()
	h.Repo.Audit(r.Context(), auditID, p.Company.ID, p.User.ID, "customer.updated", c.ID)
	httpx.JSON(w, 200, map[string]any{"data": c})
}

func (h Handler) Delete(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.PrincipalFromContext(r.Context())
	found, err := h.Repo.Delete(r.Context(), p.Company.ID, r.PathValue("id"))
	if err != nil {
		httpx.Error(w, 500, "could not delete customer")
		return
	}
	if !found {
		httpx.Error(w, 404, "customer not found")
		return
	}
	auditID, _ := id.New()
	h.Repo.Audit(r.Context(), auditID, p.Company.ID, p.User.ID, "customer.deleted", r.PathValue("id"))
	w.WriteHeader(http.StatusNoContent)
}

func decodeInput(w http.ResponseWriter, r *http.Request) (input, bool) {
	var in input
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpx.Error(w, 400, "invalid JSON payload")
		return in, false
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	in.Phone = strings.TrimSpace(in.Phone)
	in.TaxID = strings.TrimSpace(in.TaxID)
	in.Address = strings.TrimSpace(in.Address)
	in.Notes = strings.TrimSpace(in.Notes)
	if len(in.Name) < 2 || len(in.Name) > 120 {
		httpx.Error(w, 422, "customer name must be 2-120 characters")
		return in, false
	}
	if in.Email != "" {
		if _, err := mail.ParseAddress(in.Email); err != nil {
			httpx.Error(w, 422, "invalid customer email")
			return in, false
		}
	}
	if len(in.Notes) > 2000 || len(in.Address) > 1000 {
		httpx.Error(w, 422, "customer text fields are too long")
		return in, false
	}
	return in, true
}
