PRAGMA foreign_keys = ON;

ALTER TABLE companies ADD COLUMN legal_name TEXT;
ALTER TABLE companies ADD COLUMN tax_id TEXT;
ALTER TABLE companies ADD COLUMN email TEXT;
ALTER TABLE companies ADD COLUMN phone TEXT;
ALTER TABLE companies ADD COLUMN address TEXT;
ALTER TABLE companies ADD COLUMN bank_name TEXT;
ALTER TABLE companies ADD COLUMN bank_account_name TEXT;
ALTER TABLE companies ADD COLUMN bank_account_number TEXT;
ALTER TABLE companies ADD COLUMN qris_text TEXT;

CREATE TABLE IF NOT EXISTS expense_categories (
  id TEXT PRIMARY KEY,
  company_id TEXT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(company_id, name)
);

CREATE TABLE IF NOT EXISTS expenses (
  id TEXT PRIMARY KEY,
  company_id TEXT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  category_id TEXT REFERENCES expense_categories(id) ON DELETE SET NULL,
  expense_date TEXT NOT NULL,
  vendor TEXT,
  description TEXT NOT NULL,
  amount_minor INTEGER NOT NULL CHECK(amount_minor > 0),
  payment_method TEXT NOT NULL DEFAULT 'bank_transfer' CHECK(payment_method IN ('cash','bank_transfer','qris','ewallet','card','other')),
  reference TEXT,
  created_by TEXT REFERENCES users(id) ON DELETE SET NULL,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_expenses_company_date ON expenses(company_id, expense_date DESC);
