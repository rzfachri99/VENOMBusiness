package expense

import (
	"database/sql"
	"encoding/json"
	"github.com/venombusiness/venombusiness/apps/api/internal/httpx"
	"github.com/venombusiness/venombusiness/apps/api/internal/modules/auth"
	"github.com/venombusiness/venombusiness/apps/api/internal/platform/id"
	"net/http"
	"strings"
)

type Handler struct{ DB *sql.DB }
type Expense struct {
	ID          string `json:"id"`
	Category    string `json:"category"`
	Date        string `json:"expense_date"`
	Vendor      string `json:"vendor"`
	Description string `json:"description"`
	Amount      int64  `json:"amount_minor"`
	Method      string `json:"payment_method"`
	Reference   string `json:"reference"`
}
type Input struct {
	Category    string `json:"category"`
	Date        string `json:"expense_date"`
	Vendor      string `json:"vendor"`
	Description string `json:"description"`
	Amount      int64  `json:"amount_minor"`
	Method      string `json:"payment_method"`
	Reference   string `json:"reference"`
}

func (h Handler) List(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.PrincipalFromContext(r.Context())
	rows, err := h.DB.QueryContext(r.Context(), `SELECT e.id,COALESCE(c.name,''),e.expense_date,COALESCE(e.vendor,''),e.description,e.amount_minor,e.payment_method,COALESCE(e.reference,'') FROM expenses e LEFT JOIN expense_categories c ON c.id=e.category_id WHERE e.company_id=? ORDER BY e.expense_date DESC,e.created_at DESC`, p.Company.ID)
	if err != nil {
		httpx.Error(w, 500, "could not load expenses")
		return
	}
	defer rows.Close()
	out := []Expense{}
	for rows.Next() {
		var x Expense
		if rows.Scan(&x.ID, &x.Category, &x.Date, &x.Vendor, &x.Description, &x.Amount, &x.Method, &x.Reference) != nil {
			httpx.Error(w, 500, "could not load expenses")
			return
		}
		out = append(out, x)
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
	in.Category = strings.TrimSpace(in.Category)
	in.Description = strings.TrimSpace(in.Description)
	if in.Date == "" || in.Description == "" || in.Amount <= 0 {
		httpx.Error(w, 422, "date, description and positive amount are required")
		return
	}
	allowed := map[string]bool{"cash": true, "bank_transfer": true, "qris": true, "ewallet": true, "card": true, "other": true}
	if !allowed[in.Method] {
		in.Method = "bank_transfer"
	}
	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		httpx.Error(w, 500, "could not create expense")
		return
	}
	defer tx.Rollback()
	var cat any = nil
	if in.Category != "" {
		var cid string
		err = tx.QueryRowContext(r.Context(), `SELECT id FROM expense_categories WHERE company_id=? AND name=?`, p.Company.ID, in.Category).Scan(&cid)
		if err == sql.ErrNoRows {
			cid, _ = id.New()
			_, err = tx.ExecContext(r.Context(), `INSERT INTO expense_categories(id,company_id,name) VALUES(?,?,?)`, cid, p.Company.ID, in.Category)
		}
		if err != nil && err != sql.ErrNoRows {
			httpx.Error(w, 500, "could not save expense category")
			return
		}
		cat = cid
	}
	eid, _ := id.New()
	_, err = tx.ExecContext(r.Context(), `INSERT INTO expenses(id,company_id,category_id,expense_date,vendor,description,amount_minor,payment_method,reference,created_by) VALUES(?,?,?, ?,NULLIF(?,''),?,?,?,NULLIF(?,''),?)`, eid, p.Company.ID, cat, in.Date, strings.TrimSpace(in.Vendor), in.Description, in.Amount, in.Method, strings.TrimSpace(in.Reference), p.User.ID)
	if err != nil {
		httpx.Error(w, 500, "could not create expense")
		return
	}
	aid, _ := id.New()
	_, _ = tx.ExecContext(r.Context(), `INSERT INTO audit_logs(id,company_id,actor_user_id,action,entity_type,entity_id) VALUES(?,?,?,?,?,?)`, aid, p.Company.ID, p.User.ID, "expense.created", "expense", eid)
	if tx.Commit() != nil {
		httpx.Error(w, 500, "could not create expense")
		return
	}
	httpx.JSON(w, 201, map[string]any{"data": map[string]string{"id": eid}})
}
func (h Handler) Delete(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.PrincipalFromContext(r.Context())
	res, err := h.DB.ExecContext(r.Context(), `DELETE FROM expenses WHERE id=? AND company_id=?`, r.PathValue("id"), p.Company.ID)
	if err != nil {
		httpx.Error(w, 500, "could not delete expense")
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		httpx.Error(w, 404, "expense not found")
		return
	}
	w.WriteHeader(204)
}
