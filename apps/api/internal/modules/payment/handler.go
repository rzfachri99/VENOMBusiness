package payment

import (
	"database/sql"
	"encoding/json"
	"github.com/venombusiness/venombusiness/apps/api/internal/httpx"
	"github.com/venombusiness/venombusiness/apps/api/internal/modules/auth"
	"github.com/venombusiness/venombusiness/apps/api/internal/platform/id"
	"net/http"
	"strings"
	"time"
)

type Handler struct{ DB *sql.DB }
type input struct {
	InvoiceID   string `json:"invoice_id"`
	AmountMinor int64  `json:"amount_minor"`
	Method      string `json:"method"`
	Reference   string `json:"reference"`
	PaidAt      string `json:"paid_at"`
	Notes       string `json:"notes"`
}
type Payment struct {
	ID            string `json:"id"`
	InvoiceID     string `json:"invoice_id"`
	InvoiceNumber string `json:"invoice_number"`
	AmountMinor   int64  `json:"amount_minor"`
	Method        string `json:"method"`
	Reference     string `json:"reference,omitempty"`
	PaidAt        string `json:"paid_at"`
}

func (h Handler) List(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.PrincipalFromContext(r.Context())
	rows, err := h.DB.QueryContext(r.Context(), `SELECT p.id,p.invoice_id,i.number,p.amount_minor,p.method,COALESCE(p.reference,''),p.paid_at FROM payments p JOIN invoices i ON i.id=p.invoice_id WHERE p.company_id=? ORDER BY p.paid_at DESC,p.created_at DESC`, p.Company.ID)
	if err != nil {
		httpx.Error(w, 500, "could not load payments")
		return
	}
	defer rows.Close()
	out := []Payment{}
	for rows.Next() {
		var x Payment
		if rows.Scan(&x.ID, &x.InvoiceID, &x.InvoiceNumber, &x.AmountMinor, &x.Method, &x.Reference, &x.PaidAt) != nil {
			httpx.Error(w, 500, "could not load payments")
			return
		}
		out = append(out, x)
	}
	httpx.JSON(w, 200, map[string]any{"data": out})
}
func (h Handler) Create(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.PrincipalFromContext(r.Context())
	var in input
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		httpx.Error(w, 400, "invalid JSON payload")
		return
	}
	in.Method = strings.ToLower(strings.TrimSpace(in.Method))
	allowed := map[string]bool{"cash": true, "bank_transfer": true, "qris": true, "ewallet": true, "card": true, "other": true}
	if in.AmountMinor <= 0 || !allowed[in.Method] {
		httpx.Error(w, 422, "invalid payment")
		return
	}
	if in.PaidAt == "" {
		in.PaidAt = time.Now().Format("2006-01-02")
	}
	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		httpx.Error(w, 500, "could not record payment")
		return
	}
	defer tx.Rollback()
	var total, paid int64
	var status string
	err = tx.QueryRowContext(r.Context(), `SELECT total_minor,paid_minor,status FROM invoices WHERE id=? AND company_id=?`, in.InvoiceID, p.Company.ID).Scan(&total, &paid, &status)
	if err == sql.ErrNoRows {
		httpx.Error(w, 404, "invoice not found")
		return
	}
	if err != nil {
		httpx.Error(w, 500, "could not load invoice")
		return
	}
	if status == "void" {
		httpx.Error(w, 409, "void invoice cannot receive payments")
		return
	}
	if paid+in.AmountMinor > total {
		httpx.Error(w, 422, "payment exceeds outstanding balance")
		return
	}
	pid, _ := id.New()
	_, err = tx.ExecContext(r.Context(), `INSERT INTO payments(id,company_id,invoice_id,amount_minor,method,reference,paid_at,notes,created_by) VALUES(?,?,?,?,?,NULLIF(?,''),?,NULLIF(?,''),?)`, pid, p.Company.ID, in.InvoiceID, in.AmountMinor, in.Method, strings.TrimSpace(in.Reference), in.PaidAt, strings.TrimSpace(in.Notes), p.User.ID)
	if err != nil {
		httpx.Error(w, 500, "could not record payment")
		return
	}
	newPaid := paid + in.AmountMinor
	newStatus := "partial"
	if newPaid == total {
		newStatus = "paid"
	}
	_, err = tx.ExecContext(r.Context(), `UPDATE invoices SET paid_minor=?,status=?,updated_at=CURRENT_TIMESTAMP WHERE id=? AND company_id=?`, newPaid, newStatus, in.InvoiceID, p.Company.ID)
	if err != nil {
		httpx.Error(w, 500, "could not update invoice")
		return
	}
	if err := tx.Commit(); err != nil {
		httpx.Error(w, 500, "could not record payment")
		return
	}
	aid, _ := id.New()
	_, _ = h.DB.ExecContext(r.Context(), `INSERT INTO audit_logs(id,company_id,actor_user_id,action,entity_type,entity_id) VALUES(?,?,?,?,?,?)`, aid, p.Company.ID, p.User.ID, "payment.created", "payment", pid)
	httpx.JSON(w, 201, map[string]any{"data": map[string]any{"id": pid}})
}
