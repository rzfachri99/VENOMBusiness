package catalog

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/venombusiness/venombusiness/apps/api/internal/httpx"
	"github.com/venombusiness/venombusiness/apps/api/internal/modules/auth"
	"github.com/venombusiness/venombusiness/apps/api/internal/platform/id"
)

type Handler struct{ DB *sql.DB }
type Product struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	SKU         string `json:"sku,omitempty"`
	Description string `json:"description,omitempty"`
	Unit        string `json:"unit"`
	PriceMinor  int64  `json:"price_minor"`
	TaxRateBPS  int    `json:"tax_rate_bps"`
	IsActive    bool   `json:"is_active"`
}
type input struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	SKU         string `json:"sku"`
	Description string `json:"description"`
	Unit        string `json:"unit"`
	PriceMinor  int64  `json:"price_minor"`
	TaxRateBPS  int    `json:"tax_rate_bps"`
	IsActive    *bool  `json:"is_active"`
}

func (h Handler) List(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.PrincipalFromContext(r.Context())
	rows, err := h.DB.QueryContext(r.Context(), `SELECT id,type,name,COALESCE(sku,''),COALESCE(description,''),unit,price_minor,tax_rate_bps,is_active FROM products WHERE company_id=? ORDER BY is_active DESC, created_at DESC`, p.Company.ID)
	if err != nil {
		httpx.Error(w, 500, "could not load products")
		return
	}
	defer rows.Close()
	out := []Product{}
	for rows.Next() {
		var x Product
		var active int
		if err := rows.Scan(&x.ID, &x.Type, &x.Name, &x.SKU, &x.Description, &x.Unit, &x.PriceMinor, &x.TaxRateBPS, &active); err != nil {
			httpx.Error(w, 500, "could not load products")
			return
		}
		x.IsActive = active == 1
		out = append(out, x)
	}
	httpx.JSON(w, 200, map[string]any{"data": out})
}
func (h Handler) Create(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.PrincipalFromContext(r.Context())
	in, ok := decode(w, r)
	if !ok {
		return
	}
	pid, _ := id.New()
	active := 1
	if in.IsActive != nil && !*in.IsActive {
		active = 0
	}
	_, err := h.DB.ExecContext(r.Context(), `INSERT INTO products(id,company_id,type,name,sku,description,unit,price_minor,tax_rate_bps,is_active) VALUES(?,?,?,?,NULLIF(?,''),NULLIF(?,''),?,?,?,?)`, pid, p.Company.ID, in.Type, in.Name, in.SKU, in.Description, in.Unit, in.PriceMinor, in.TaxRateBPS, active)
	if err != nil {
		httpx.Error(w, 409, "SKU already exists or product could not be created")
		return
	}
	audit(h.DB, r, p, "product.created", pid)
	httpx.JSON(w, 201, map[string]any{"data": map[string]any{"id": pid}})
}
func (h Handler) Update(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.PrincipalFromContext(r.Context())
	in, ok := decode(w, r)
	if !ok {
		return
	}
	active := 1
	if in.IsActive != nil && !*in.IsActive {
		active = 0
	}
	res, err := h.DB.ExecContext(r.Context(), `UPDATE products SET type=?,name=?,sku=NULLIF(?,''),description=NULLIF(?,''),unit=?,price_minor=?,tax_rate_bps=?,is_active=?,updated_at=CURRENT_TIMESTAMP WHERE id=? AND company_id=?`, in.Type, in.Name, in.SKU, in.Description, in.Unit, in.PriceMinor, in.TaxRateBPS, active, r.PathValue("id"), p.Company.ID)
	if err != nil {
		httpx.Error(w, 409, "SKU already exists or product could not be updated")
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		httpx.Error(w, 404, "product not found")
		return
	}
	audit(h.DB, r, p, "product.updated", r.PathValue("id"))
	w.WriteHeader(204)
}
func (h Handler) Delete(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.PrincipalFromContext(r.Context())
	res, err := h.DB.ExecContext(r.Context(), `DELETE FROM products WHERE id=? AND company_id=?`, r.PathValue("id"), p.Company.ID)
	if err != nil {
		httpx.Error(w, 409, "product is already used by a document; deactivate it instead")
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		httpx.Error(w, 404, "product not found")
		return
	}
	audit(h.DB, r, p, "product.deleted", r.PathValue("id"))
	w.WriteHeader(204)
}
func decode(w http.ResponseWriter, r *http.Request) (input, bool) {
	var in input
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		httpx.Error(w, 400, "invalid JSON payload")
		return in, false
	}
	in.Type = strings.ToLower(strings.TrimSpace(in.Type))
	if in.Type == "" {
		in.Type = "service"
	}
	in.Name = strings.TrimSpace(in.Name)
	in.SKU = strings.TrimSpace(in.SKU)
	in.Description = strings.TrimSpace(in.Description)
	in.Unit = strings.TrimSpace(in.Unit)
	if in.Unit == "" {
		in.Unit = "item"
	}
	if (in.Type != "product" && in.Type != "service") || len(in.Name) < 2 || len(in.Name) > 140 || in.PriceMinor < 0 || in.TaxRateBPS < 0 || in.TaxRateBPS > 10000 {
		httpx.Error(w, 422, "invalid product fields")
		return in, false
	}
	return in, true
}
func audit(db *sql.DB, r *http.Request, p auth.Principal, action, entityID string) {
	aid, _ := id.New()
	_, _ = db.ExecContext(r.Context(), `INSERT INTO audit_logs(id,company_id,actor_user_id,action,entity_type,entity_id) VALUES(?,?,?,?,?,?)`, aid, p.Company.ID, p.User.ID, action, "product", entityID)
}
