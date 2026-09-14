

CREATE TABLE IF NOT EXISTS document_sequences (
  company_id TEXT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  kind TEXT NOT NULL CHECK(kind IN ('quotation','invoice')),
  next_value INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY(company_id, kind)
);

CREATE TABLE IF NOT EXISTS products (
  id TEXT PRIMARY KEY,
  company_id TEXT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  type TEXT NOT NULL DEFAULT 'service' CHECK(type IN ('product','service')),
  name TEXT NOT NULL,
  sku TEXT,
  description TEXT,
  unit TEXT NOT NULL DEFAULT 'item',
  price_minor INTEGER NOT NULL DEFAULT 0 CHECK(price_minor >= 0),
  tax_rate_bps INTEGER NOT NULL DEFAULT 0 CHECK(tax_rate_bps BETWEEN 0 AND 10000),
  is_active INTEGER NOT NULL DEFAULT 1 CHECK(is_active IN (0,1)),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(company_id, sku)
);
CREATE INDEX IF NOT EXISTS idx_products_company ON products(company_id, is_active, created_at DESC);

CREATE TABLE IF NOT EXISTS quotations (
  id TEXT PRIMARY KEY,
  company_id TEXT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  customer_id TEXT NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
  number TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'draft' CHECK(status IN ('draft','sent','accepted','rejected','expired')),
  issue_date TEXT NOT NULL,
  valid_until TEXT,
  notes TEXT,
  subtotal_minor INTEGER NOT NULL DEFAULT 0,
  discount_minor INTEGER NOT NULL DEFAULT 0,
  tax_minor INTEGER NOT NULL DEFAULT 0,
  total_minor INTEGER NOT NULL DEFAULT 0,
  created_by TEXT REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(company_id, number)
);
CREATE INDEX IF NOT EXISTS idx_quotations_company ON quotations(company_id, created_at DESC);

CREATE TABLE IF NOT EXISTS quotation_items (
  id TEXT PRIMARY KEY,
  quotation_id TEXT NOT NULL REFERENCES quotations(id) ON DELETE CASCADE,
  product_id TEXT REFERENCES products(id) ON DELETE SET NULL,
  name TEXT NOT NULL,
  description TEXT,
  quantity_milli INTEGER NOT NULL CHECK(quantity_milli > 0),
  unit TEXT NOT NULL DEFAULT 'item',
  unit_price_minor INTEGER NOT NULL CHECK(unit_price_minor >= 0),
  tax_rate_bps INTEGER NOT NULL DEFAULT 0 CHECK(tax_rate_bps BETWEEN 0 AND 10000),
  line_subtotal_minor INTEGER NOT NULL,
  line_tax_minor INTEGER NOT NULL,
  line_total_minor INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_quotation_items_doc ON quotation_items(quotation_id);

CREATE TABLE IF NOT EXISTS invoices (
  id TEXT PRIMARY KEY,
  company_id TEXT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  customer_id TEXT NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
  quotation_id TEXT REFERENCES quotations(id) ON DELETE SET NULL,
  number TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'draft' CHECK(status IN ('draft','sent','partial','paid','overdue','void')),
  issue_date TEXT NOT NULL,
  due_date TEXT,
  notes TEXT,
  subtotal_minor INTEGER NOT NULL DEFAULT 0,
  discount_minor INTEGER NOT NULL DEFAULT 0,
  tax_minor INTEGER NOT NULL DEFAULT 0,
  total_minor INTEGER NOT NULL DEFAULT 0,
  paid_minor INTEGER NOT NULL DEFAULT 0,
  created_by TEXT REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(company_id, number)
);
CREATE INDEX IF NOT EXISTS idx_invoices_company ON invoices(company_id, status, created_at DESC);

CREATE TABLE IF NOT EXISTS invoice_items (
  id TEXT PRIMARY KEY,
  invoice_id TEXT NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
  product_id TEXT REFERENCES products(id) ON DELETE SET NULL,
  name TEXT NOT NULL,
  description TEXT,
  quantity_milli INTEGER NOT NULL CHECK(quantity_milli > 0),
  unit TEXT NOT NULL DEFAULT 'item',
  unit_price_minor INTEGER NOT NULL CHECK(unit_price_minor >= 0),
  tax_rate_bps INTEGER NOT NULL DEFAULT 0 CHECK(tax_rate_bps BETWEEN 0 AND 10000),
  line_subtotal_minor INTEGER NOT NULL,
  line_tax_minor INTEGER NOT NULL,
  line_total_minor INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_invoice_items_doc ON invoice_items(invoice_id);

CREATE TABLE IF NOT EXISTS payments (
  id TEXT PRIMARY KEY,
  company_id TEXT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  invoice_id TEXT NOT NULL REFERENCES invoices(id) ON DELETE RESTRICT,
  amount_minor INTEGER NOT NULL CHECK(amount_minor > 0),
  method TEXT NOT NULL CHECK(method IN ('cash','bank_transfer','qris','ewallet','card','other')),
  reference TEXT,
  paid_at TIMESTAMP NOT NULL,
  notes TEXT,
  created_by TEXT REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_payments_company ON payments(company_id, paid_at DESC);
CREATE INDEX IF NOT EXISTS idx_payments_invoice ON payments(invoice_id, paid_at DESC);
