package document

import (
	"database/sql"
	"fmt"
	"github.com/venombusiness/venombusiness/apps/api/internal/modules/auth"
	pdfx "github.com/venombusiness/venombusiness/apps/api/internal/platform/pdf"
	"net/http"
	"strings"
)

type Handler struct{ DB *sql.DB }

func money(v int64, c string) string { return fmt.Sprintf("%s %d.%02d", c, v/100, v%100) }
func (h Handler) Invoice(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.PrincipalFromContext(r.Context())
	var number, status, cust, issue, due, notes, currency, company, address, email, phone, bank, acctName, acctNo, qris string
	var sub, tax, disc, total, paid int64
	err := h.DB.QueryRowContext(r.Context(), `SELECT i.number,i.status,c.name,i.issue_date,COALESCE(i.due_date,''),COALESCE(i.notes,''),i.subtotal_minor,i.tax_minor,i.discount_minor,i.total_minor,i.paid_minor,co.currency,co.name,COALESCE(co.address,''),COALESCE(co.email,''),COALESCE(co.phone,''),COALESCE(co.bank_name,''),COALESCE(co.bank_account_name,''),COALESCE(co.bank_account_number,''),COALESCE(co.qris_text,'') FROM invoices i JOIN customers c ON c.id=i.customer_id JOIN companies co ON co.id=i.company_id WHERE i.id=? AND i.company_id=?`, r.PathValue("id"), p.Company.ID).Scan(&number, &status, &cust, &issue, &due, &notes, &sub, &tax, &disc, &total, &paid, &currency, &company, &address, &email, &phone, &bank, &acctName, &acctNo, &qris)
	if err != nil {
		http.Error(w, "invoice not found", 404)
		return
	}
	lines := []string{company, address, strings.TrimSpace(email + "  " + phone), "", "Invoice: " + number, "Customer: " + cust, "Issue: " + issue + "   Due: " + due, "Status: " + status, "", "Subtotal: " + money(sub, currency), "Tax: " + money(tax, currency), "Discount: " + money(disc, currency), "Total: " + money(total, currency), "Paid: " + money(paid, currency), "Balance: " + money(total-paid, currency), "", "Payment: " + strings.TrimSpace(bank+" "+acctNo+" "+acctName), "QRIS: " + qris, "Notes: " + notes}
	b := pdfx.Build("INVOICE", lines)
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s.pdf"`, number))
	w.Write(b)
}
func (h Handler) Quotation(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.PrincipalFromContext(r.Context())
	var number, status, cust, issue, valid, notes, currency, company, address string
	var sub, tax, disc, total int64
	err := h.DB.QueryRowContext(r.Context(), `SELECT q.number,q.status,c.name,q.issue_date,COALESCE(q.valid_until,''),COALESCE(q.notes,''),q.subtotal_minor,q.tax_minor,q.discount_minor,q.total_minor,co.currency,co.name,COALESCE(co.address,'') FROM quotations q JOIN customers c ON c.id=q.customer_id JOIN companies co ON co.id=q.company_id WHERE q.id=? AND q.company_id=?`, r.PathValue("id"), p.Company.ID).Scan(&number, &status, &cust, &issue, &valid, &notes, &sub, &tax, &disc, &total, &currency, &company, &address)
	if err != nil {
		http.Error(w, "quotation not found", 404)
		return
	}
	b := pdfx.Build("QUOTATION", []string{company, address, "", "Quotation: " + number, "Customer: " + cust, "Issue: " + issue + "   Valid until: " + valid, "Status: " + status, "", "Subtotal: " + money(sub, currency), "Tax: " + money(tax, currency), "Discount: " + money(disc, currency), "Total: " + money(total, currency), "", "Notes: " + notes})
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s.pdf"`, number))
	w.Write(b)
}
func (h Handler) Receipt(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.PrincipalFromContext(r.Context())
	var idv, inv, method, ref, paidAt, currency, company, cust string
	var amount int64
	err := h.DB.QueryRowContext(r.Context(), `SELECT p.id,i.number,p.method,COALESCE(p.reference,''),p.paid_at,co.currency,co.name,c.name,p.amount_minor FROM payments p JOIN invoices i ON i.id=p.invoice_id JOIN companies co ON co.id=p.company_id JOIN customers c ON c.id=i.customer_id WHERE p.id=? AND p.company_id=?`, r.PathValue("id"), p.Company.ID).Scan(&idv, &inv, &method, &ref, &paidAt, &currency, &company, &cust, &amount)
	if err != nil {
		http.Error(w, "payment not found", 404)
		return
	}
	b := pdfx.Build("PAYMENT RECEIPT", []string{company, "", "Receipt: " + idv, "Invoice: " + inv, "Customer: " + cust, "Paid at: " + paidAt, "Method: " + method, "Reference: " + ref, "", "Amount received: " + money(amount, currency)})
	w.Header().Set("Content-Type", "application/pdf")
	w.Write(b)
}
