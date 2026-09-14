package report

import (
	"database/sql"
	"github.com/venombusiness/venombusiness/apps/api/internal/httpx"
	"github.com/venombusiness/venombusiness/apps/api/internal/modules/auth"
	"net/http"
)

type Handler struct{ DB *sql.DB }

func (h Handler) Summary(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.PrincipalFromContext(r.Context())
	var invoiced, paid, expenses int64
	var open int
	_ = h.DB.QueryRowContext(r.Context(), `SELECT COALESCE(SUM(total_minor),0) FROM invoices WHERE company_id=? AND status!='void'`, p.Company.ID).Scan(&invoiced)
	_ = h.DB.QueryRowContext(r.Context(), `SELECT COALESCE(SUM(amount_minor),0) FROM payments WHERE company_id=?`, p.Company.ID).Scan(&paid)
	_ = h.DB.QueryRowContext(r.Context(), `SELECT COALESCE(SUM(amount_minor),0) FROM expenses WHERE company_id=?`, p.Company.ID).Scan(&expenses)
	_ = h.DB.QueryRowContext(r.Context(), `SELECT COUNT(1) FROM invoices WHERE company_id=? AND status NOT IN ('paid','void')`, p.Company.ID).Scan(&open)
	httpx.JSON(w, 200, map[string]any{"data": map[string]any{"invoiced_minor": invoiced, "collected_minor": paid, "expense_minor": expenses, "cash_result_minor": paid - expenses, "receivable_minor": invoiced - paid, "open_invoices": open}})
}
