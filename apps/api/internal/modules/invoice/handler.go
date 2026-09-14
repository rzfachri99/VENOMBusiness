package invoice

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/venombusiness/venombusiness/apps/api/internal/httpx"
	"github.com/venombusiness/venombusiness/apps/api/internal/modules/auth"
	"github.com/venombusiness/venombusiness/apps/api/internal/platform/id"
	"net/http"
	"time"
)

type Handler struct{ DB *sql.DB }
type Invoice struct {
	ID           string `json:"id"`
	Number       string `json:"number"`
	Status       string `json:"status"`
	CustomerID   string `json:"customer_id"`
	CustomerName string `json:"customer_name"`
	IssueDate    string `json:"issue_date"`
	DueDate      string `json:"due_date,omitempty"`
	TotalMinor   int64  `json:"total_minor"`
	PaidMinor    int64  `json:"paid_minor"`
	BalanceMinor int64  `json:"balance_minor"`
	QuotationID  string `json:"quotation_id,omitempty"`
}

func (h Handler) List(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.PrincipalFromContext(r.Context())
	rows, err := h.DB.QueryContext(r.Context(), `SELECT i.id,i.number,i.status,i.customer_id,c.name,i.issue_date,COALESCE(i.due_date,''),i.total_minor,i.paid_minor,i.total_minor-i.paid_minor,COALESCE(i.quotation_id,'') FROM invoices i JOIN customers c ON c.id=i.customer_id WHERE i.company_id=? ORDER BY i.created_at DESC`, p.Company.ID)
	if err != nil {
		httpx.Error(w, 500, "could not load invoices")
		return
	}
	defer rows.Close()
	out := []Invoice{}
	for rows.Next() {
		var x Invoice
		if rows.Scan(&x.ID, &x.Number, &x.Status, &x.CustomerID, &x.CustomerName, &x.IssueDate, &x.DueDate, &x.TotalMinor, &x.PaidMinor, &x.BalanceMinor, &x.QuotationID) != nil {
			httpx.Error(w, 500, "could not load invoices")
			return
		}
		out = append(out, x)
	}
	httpx.JSON(w, 200, map[string]any{"data": out})
}
func (h Handler) FromQuotation(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.PrincipalFromContext(r.Context())
	qid := r.PathValue("id")
	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		httpx.Error(w, 500, "could not create invoice")
		return
	}
	defer tx.Rollback()
	var customerID, notes, quoteStatus string
	var subtotal, discount, tax, total int64
	var existing sql.NullString
	err = tx.QueryRowContext(r.Context(), `SELECT q.customer_id,COALESCE(q.notes,''),q.status,q.subtotal_minor,q.discount_minor,q.tax_minor,q.total_minor,(SELECT id FROM invoices WHERE quotation_id=q.id AND company_id=q.company_id LIMIT 1) FROM quotations q WHERE q.id=? AND q.company_id=?`, qid, p.Company.ID).Scan(&customerID, &notes, &quoteStatus, &subtotal, &discount, &tax, &total, &existing)
	if err == sql.ErrNoRows {
		httpx.Error(w, 404, "quotation not found")
		return
	}
	if err != nil {
		httpx.Error(w, 500, "could not load quotation")
		return
	}
	if quoteStatus != "accepted" {
		httpx.Error(w, 409, "quotation must be accepted before invoicing")
		return
	}
	if existing.Valid {
		httpx.Error(w, 409, "quotation already has an invoice")
		return
	}
	iid, _ := id.New()
	number, err := nextNumber(tx, p.Company.ID)
	if err != nil {
		httpx.Error(w, 500, "could not allocate invoice number")
		return
	}
	issue := time.Now()
	due := issue.AddDate(0, 0, 14)
	_, err = tx.ExecContext(r.Context(), `INSERT INTO invoices(id,company_id,customer_id,quotation_id,number,issue_date,due_date,notes,subtotal_minor,discount_minor,tax_minor,total_minor,created_by) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`, iid, p.Company.ID, customerID, qid, number, issue.Format("2006-01-02"), due.Format("2006-01-02"), notes, subtotal, discount, tax, total, p.User.ID)
	if err != nil {
		httpx.Error(w, 500, "could not create invoice")
		return
	}
	rows, err := tx.QueryContext(r.Context(), `SELECT product_id,name,description,quantity_milli,unit,unit_price_minor,tax_rate_bps,line_subtotal_minor,line_tax_minor,line_total_minor FROM quotation_items WHERE quotation_id=?`, qid)
	if err != nil {
		httpx.Error(w, 500, "could not copy invoice items")
		return
	}
	for rows.Next() {
		var productID, description sql.NullString
		var name, unit string
		var qty, price, sub, ltax, ltotal int64
		var rate int
		if rows.Scan(&productID, &name, &description, &qty, &unit, &price, &rate, &sub, &ltax, &ltotal) != nil {
			rows.Close()
			httpx.Error(w, 500, "could not copy invoice items")
			return
		}
		itemID, _ := id.New()
		_, err = tx.ExecContext(r.Context(), `INSERT INTO invoice_items(id,invoice_id,product_id,name,description,quantity_milli,unit,unit_price_minor,tax_rate_bps,line_subtotal_minor,line_tax_minor,line_total_minor) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`, itemID, iid, productID, name, description, qty, unit, price, rate, sub, ltax, ltotal)
		if err != nil {
			rows.Close()
			httpx.Error(w, 500, "could not copy invoice items")
			return
		}
	}
	rows.Close()
	_, _ = tx.ExecContext(r.Context(), `UPDATE quotations SET status='accepted',updated_at=CURRENT_TIMESTAMP WHERE id=?`, qid)
	if err := tx.Commit(); err != nil {
		httpx.Error(w, 500, "could not create invoice")
		return
	}
	audit(h.DB, r, p, "invoice.created", iid)
	httpx.JSON(w, 201, map[string]any{"data": map[string]any{"id": iid, "number": number}})
}
func (h Handler) Status(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.PrincipalFromContext(r.Context())
	var in struct {
		Status string `json:"status"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		httpx.Error(w, 400, "invalid JSON payload")
		return
	}
	allowed := map[string]bool{"draft": true, "sent": true, "overdue": true, "void": true}
	if !allowed[in.Status] {
		httpx.Error(w, 422, "this status is payment-controlled or invalid")
		return
	}
	var paid int64
	if err := h.DB.QueryRowContext(r.Context(), `SELECT paid_minor FROM invoices WHERE id=? AND company_id=?`, r.PathValue("id"), p.Company.ID).Scan(&paid); err != nil {
		httpx.Error(w, 404, "invoice not found")
		return
	}
	if paid > 0 && in.Status == "void" {
		httpx.Error(w, 409, "paid invoice cannot be voided")
		return
	}
	_, err := h.DB.ExecContext(r.Context(), `UPDATE invoices SET status=?,updated_at=CURRENT_TIMESTAMP WHERE id=? AND company_id=?`, in.Status, r.PathValue("id"), p.Company.ID)
	if err != nil {
		httpx.Error(w, 500, "could not update invoice")
		return
	}
	audit(h.DB, r, p, "invoice.status_changed", r.PathValue("id"))
	w.WriteHeader(204)
}
func nextNumber(tx *sql.Tx, company string) (string, error) {
	if _, err := tx.Exec(`INSERT INTO document_sequences(company_id,kind,next_value) VALUES(?, 'invoice', 0) ON CONFLICT(company_id,kind) DO NOTHING`, company); err != nil {
		return "", err
	}
	if _, err := tx.Exec(`UPDATE document_sequences SET next_value=next_value+1 WHERE company_id=? AND kind='invoice'`, company); err != nil {
		return "", err
	}
	var n int
	if err := tx.QueryRow(`SELECT next_value FROM document_sequences WHERE company_id=? AND kind='invoice'`, company).Scan(&n); err != nil {
		return "", err
	}
	return fmt.Sprintf("INV-%s-%04d", time.Now().Format("200601"), n), nil
}
func audit(db *sql.DB, r *http.Request, p auth.Principal, action, idv string) {
	aid, _ := id.New()
	_, _ = db.ExecContext(r.Context(), `INSERT INTO audit_logs(id,company_id,actor_user_id,action,entity_type,entity_id) VALUES(?,?,?,?,?,?)`, aid, p.Company.ID, p.User.ID, action, "invoice", idv)
}
