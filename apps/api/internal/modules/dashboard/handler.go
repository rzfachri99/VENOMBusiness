package dashboard

import (
	"database/sql"
	"net/http"

	"github.com/venombusiness/venombusiness/apps/api/internal/httpx"
	"github.com/venombusiness/venombusiness/apps/api/internal/modules/auth"
)

type Handler struct{ DB *sql.DB }

func (h Handler) Summary(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.PrincipalFromContext(r.Context())
	var customers, products, openInvoices int
	var invoiced, paid int64
	queries := []struct {
		q string
		v any
	}{
		{`SELECT COUNT(1) FROM customers WHERE company_id=?`, &customers},
		{`SELECT COUNT(1) FROM products WHERE company_id=? AND is_active=1`, &products},
		{`SELECT COUNT(1) FROM invoices WHERE company_id=? AND status NOT IN ('paid','void')`, &openInvoices},
		{`SELECT COALESCE(SUM(total_minor),0) FROM invoices WHERE company_id=? AND status!='void'`, &invoiced},
		{`SELECT COALESCE(SUM(amount_minor),0) FROM payments WHERE company_id=?`, &paid},
	}
	for _, q := range queries {
		if err := h.DB.QueryRowContext(r.Context(), q.q, p.Company.ID).Scan(q.v); err != nil {
			httpx.Error(w, 500, "could not load dashboard")
			return
		}
	}
	httpx.JSON(w, 200, map[string]any{"data": map[string]any{
		"customers": customers, "products": products, "open_invoices": openInvoices,
		"invoiced_minor": invoiced, "paid_minor": paid, "outstanding_minor": invoiced - paid,
		"company": p.Company,
		"modules": map[string]string{"customers": "active", "catalog": "active", "quotation": "active", "invoice": "active", "payments": "active", "inventory": "planned"},
	}})
}
