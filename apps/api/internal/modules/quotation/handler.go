package quotation

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/venombusiness/venombusiness/apps/api/internal/httpx"
	"github.com/venombusiness/venombusiness/apps/api/internal/modules/auth"
	"github.com/venombusiness/venombusiness/apps/api/internal/platform/id"
)

type Handler struct{ DB *sql.DB }
type ItemInput struct {
	ProductID      string `json:"product_id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	QuantityMilli  int64  `json:"quantity_milli"`
	Unit           string `json:"unit"`
	UnitPriceMinor int64  `json:"unit_price_minor"`
	TaxRateBPS     int    `json:"tax_rate_bps"`
}
type Input struct {
	CustomerID    string      `json:"customer_id"`
	IssueDate     string      `json:"issue_date"`
	ValidUntil    string      `json:"valid_until"`
	DiscountMinor int64       `json:"discount_minor"`
	Notes         string      `json:"notes"`
	Items         []ItemInput `json:"items"`
}
type Quotation struct {
	ID            string `json:"id"`
	Number        string `json:"number"`
	Status        string `json:"status"`
	CustomerID    string `json:"customer_id"`
	CustomerName  string `json:"customer_name"`
	IssueDate     string `json:"issue_date"`
	ValidUntil    string `json:"valid_until,omitempty"`
	SubtotalMinor int64  `json:"subtotal_minor"`
	DiscountMinor int64  `json:"discount_minor"`
	TaxMinor      int64  `json:"tax_minor"`
	TotalMinor    int64  `json:"total_minor"`
	Notes         string `json:"notes,omitempty"`
}

func (h Handler) List(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.PrincipalFromContext(r.Context())
	rows, err := h.DB.QueryContext(r.Context(), `SELECT q.id,q.number,q.status,q.customer_id,c.name,q.issue_date,COALESCE(q.valid_until,''),q.subtotal_minor,q.discount_minor,q.tax_minor,q.total_minor,COALESCE(q.notes,'') FROM quotations q JOIN customers c ON c.id=q.customer_id WHERE q.company_id=? ORDER BY q.created_at DESC`, p.Company.ID)
	if err != nil {
		httpx.Error(w, 500, "could not load quotations")
		return
	}
	defer rows.Close()
	out := []Quotation{}
	for rows.Next() {
		var q Quotation
		if rows.Scan(&q.ID, &q.Number, &q.Status, &q.CustomerID, &q.CustomerName, &q.IssueDate, &q.ValidUntil, &q.SubtotalMinor, &q.DiscountMinor, &q.TaxMinor, &q.TotalMinor, &q.Notes) != nil {
			httpx.Error(w, 500, "could not load quotations")
			return
		}
		out = append(out, q)
	}
	httpx.JSON(w, 200, map[string]any{"data": out})
}
func (h Handler) Create(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.PrincipalFromContext(r.Context())
	var in Input
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		httpx.Error(w, 400, "invalid JSON payload")
		return
	}
	if in.IssueDate == "" {
		in.IssueDate = time.Now().Format("2006-01-02")
	}
	if len(in.Items) == 0 || in.DiscountMinor < 0 {
		httpx.Error(w, 422, "quotation requires at least one item")
		return
	}
	var owns int
	if h.DB.QueryRowContext(r.Context(), `SELECT COUNT(1) FROM customers WHERE id=? AND company_id=?`, in.CustomerID, p.Company.ID).Scan(&owns) != nil || owns == 0 {
		httpx.Error(w, 422, "invalid customer")
		return
	}
	qid, _ := id.New()
	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		httpx.Error(w, 500, "could not create quotation")
		return
	}
	defer tx.Rollback()
	number, err := nextNumber(tx, p.Company.ID, "QTN")
	if err != nil {
		httpx.Error(w, 500, "could not allocate quotation number")
		return
	}
	subtotal, tax, total, items, ok := calculate(in.Items, in.DiscountMinor)
	if !ok {
		httpx.Error(w, 422, "invalid quotation item")
		return
	}
	_, err = tx.ExecContext(r.Context(), `INSERT INTO quotations(id,company_id,customer_id,number,issue_date,valid_until,notes,subtotal_minor,discount_minor,tax_minor,total_minor,created_by) VALUES(?,?,?,?,?,NULLIF(?,''),NULLIF(?,''),?,?,?,?,?)`, qid, p.Company.ID, in.CustomerID, number, in.IssueDate, in.ValidUntil, strings.TrimSpace(in.Notes), subtotal, in.DiscountMinor, tax, total, p.User.ID)
	if err != nil {
		httpx.Error(w, 500, "could not create quotation")
		return
	}
	for _, it := range items {
		iid, _ := id.New()
		_, err = tx.ExecContext(r.Context(), `INSERT INTO quotation_items(id,quotation_id,product_id,name,description,quantity_milli,unit,unit_price_minor,tax_rate_bps,line_subtotal_minor,line_tax_minor,line_total_minor) VALUES(?,?,NULLIF(?,''),?,NULLIF(?,''),?,?,?,?,?,?,?)`, iid, qid, it.ProductID, it.Name, it.Description, it.QuantityMilli, it.Unit, it.UnitPriceMinor, it.TaxRateBPS, it.sub, it.tax, it.total)
		if err != nil {
			httpx.Error(w, 500, "could not save quotation items")
			return
		}
	}
	if err := tx.Commit(); err != nil {
		httpx.Error(w, 500, "could not create quotation")
		return
	}
	audit(h.DB, r, p, "quotation.created", qid)
	httpx.JSON(w, 201, map[string]any{"data": map[string]any{"id": qid, "number": number}})
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
	allowed := map[string]bool{"draft": true, "sent": true, "accepted": true, "rejected": true, "expired": true}
	if !allowed[in.Status] {
		httpx.Error(w, 422, "invalid quotation status")
		return
	}
	res, err := h.DB.ExecContext(r.Context(), `UPDATE quotations SET status=?,updated_at=CURRENT_TIMESTAMP WHERE id=? AND company_id=?`, in.Status, r.PathValue("id"), p.Company.ID)
	if err != nil {
		httpx.Error(w, 500, "could not update status")
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		httpx.Error(w, 404, "quotation not found")
		return
	}
	audit(h.DB, r, p, "quotation.status_changed", r.PathValue("id"))
	w.WriteHeader(204)
}

type calcItem struct {
	ItemInput
	sub, tax, total int64
}

func calculate(items []ItemInput, discount int64) (int64, int64, int64, []calcItem, bool) {
	var subtotal, tax int64
	out := make([]calcItem, 0, len(items))
	for _, it := range items {
		it.Name = strings.TrimSpace(it.Name)
		it.Unit = strings.TrimSpace(it.Unit)
		if it.Unit == "" {
			it.Unit = "item"
		}
		if it.Name == "" || it.QuantityMilli <= 0 || it.UnitPriceMinor < 0 || it.TaxRateBPS < 0 || it.TaxRateBPS > 10000 {
			return 0, 0, 0, nil, false
		}
		sub := (it.UnitPriceMinor*it.QuantityMilli + 500) / 1000
		lineTax := (sub*int64(it.TaxRateBPS) + 5000) / 10000
		subtotal += sub
		tax += lineTax
		out = append(out, calcItem{it, sub, lineTax, sub + lineTax})
	}
	if discount > subtotal+tax {
		return 0, 0, 0, nil, false
	}
	return subtotal, tax, subtotal + tax - discount, out, true
}
func nextNumber(tx *sql.Tx, company, prefix string) (string, error) {
	if _, err := tx.Exec(`INSERT INTO document_sequences(company_id,kind,next_value) VALUES(?, 'quotation', 0) ON CONFLICT(company_id,kind) DO NOTHING`, company); err != nil {
		return "", err
	}
	if _, err := tx.Exec(`UPDATE document_sequences SET next_value=next_value+1 WHERE company_id=? AND kind='quotation'`, company); err != nil {
		return "", err
	}
	var n int
	if err := tx.QueryRow(`SELECT next_value FROM document_sequences WHERE company_id=? AND kind='quotation'`, company).Scan(&n); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s-%04d", prefix, time.Now().Format("200601"), n), nil
}
func audit(db *sql.DB, r *http.Request, p auth.Principal, action, idv string) {
	aid, _ := id.New()
	_, _ = db.ExecContext(r.Context(), `INSERT INTO audit_logs(id,company_id,actor_user_id,action,entity_type,entity_id) VALUES(?,?,?,?,?,?)`, aid, p.Company.ID, p.User.ID, action, "quotation", idv)
}
