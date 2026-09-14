package customer

import (
	"context"
	"database/sql"
)

type Customer struct {
	ID        string `json:"id"`
	CompanyID string `json:"company_id"`
	Name      string `json:"name"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
	TaxID     string `json:"tax_id,omitempty"`
	Address   string `json:"address,omitempty"`
	Notes     string `json:"notes,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Repository struct{ DB *sql.DB }

func (r Repository) List(ctx context.Context, companyID string) ([]Customer, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id,company_id,name,COALESCE(email,''),COALESCE(phone,''),COALESCE(tax_id,''),COALESCE(address,''),COALESCE(notes,''),created_at,updated_at FROM customers WHERE company_id=? ORDER BY created_at DESC LIMIT 500`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Customer{}
	for rows.Next() {
		var c Customer
		if err := rows.Scan(&c.ID, &c.CompanyID, &c.Name, &c.Email, &c.Phone, &c.TaxID, &c.Address, &c.Notes, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r Repository) Create(ctx context.Context, c Customer) error {
	_, err := r.DB.ExecContext(ctx, `INSERT INTO customers(id,company_id,name,email,phone,tax_id,address,notes) VALUES(?,?,?,?,?,?,?,?)`, c.ID, c.CompanyID, c.Name, null(c.Email), null(c.Phone), null(c.TaxID), null(c.Address), null(c.Notes))
	return err
}

func (r Repository) Update(ctx context.Context, c Customer) (bool, error) {
	res, err := r.DB.ExecContext(ctx, `UPDATE customers SET name=?,email=?,phone=?,tax_id=?,address=?,notes=?,updated_at=CURRENT_TIMESTAMP WHERE id=? AND company_id=?`, c.Name, null(c.Email), null(c.Phone), null(c.TaxID), null(c.Address), null(c.Notes), c.ID, c.CompanyID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func (r Repository) Delete(ctx context.Context, companyID, customerID string) (bool, error) {
	res, err := r.DB.ExecContext(ctx, `DELETE FROM customers WHERE id=? AND company_id=?`, customerID, companyID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func (r Repository) Audit(ctx context.Context, auditID, companyID, actorUserID, action, customerID string) {
	_, _ = r.DB.ExecContext(ctx, `INSERT INTO audit_logs(id,company_id,actor_user_id,action,entity_type,entity_id) VALUES(?,?,?,?,?,?)`, auditID, companyID, actorUserID, action, "customer", customerID)
}

func null(s string) any {
	if s == "" {
		return nil
	}
	return s
}
