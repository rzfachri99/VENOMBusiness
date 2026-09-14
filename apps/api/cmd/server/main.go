package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/venombusiness/venombusiness/apps/api/internal/config"
	"github.com/venombusiness/venombusiness/apps/api/internal/httpx"
	"github.com/venombusiness/venombusiness/apps/api/internal/modules/auth"
	"github.com/venombusiness/venombusiness/apps/api/internal/modules/catalog"
	"github.com/venombusiness/venombusiness/apps/api/internal/modules/company"
	"github.com/venombusiness/venombusiness/apps/api/internal/modules/customer"
	"github.com/venombusiness/venombusiness/apps/api/internal/modules/dashboard"
	"github.com/venombusiness/venombusiness/apps/api/internal/modules/document"
	"github.com/venombusiness/venombusiness/apps/api/internal/modules/expense"
	"github.com/venombusiness/venombusiness/apps/api/internal/modules/health"
	"github.com/venombusiness/venombusiness/apps/api/internal/modules/invoice"
	"github.com/venombusiness/venombusiness/apps/api/internal/modules/payment"
	"github.com/venombusiness/venombusiness/apps/api/internal/modules/quotation"
	"github.com/venombusiness/venombusiness/apps/api/internal/modules/report"
	"github.com/venombusiness/venombusiness/apps/api/internal/platform/database"
	"github.com/venombusiness/venombusiness/apps/api/internal/platform/migrations"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx := context.Background()
	db, err := database.Open(ctx, cfg.DatabaseDriver, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database startup failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	if err := migrations.Up(ctx, db, cfg.DatabaseDriver); err != nil {
		logger.Error("migration failed", "error", err)
		os.Exit(1)
	}

	authRepo := auth.Repository{DB: db}
	authH := auth.Handler{Repo: authRepo, CookieSecure: cfg.CookieSecure, SessionTTL: 7 * 24 * time.Hour}
	authM := auth.Middleware{Repo: authRepo}
	companyH := company.Handler{DB: db}
	customerH := customer.Handler{Repo: customer.Repository{DB: db}}
	dashboardH := dashboard.Handler{DB: db}
	healthH := health.Handler{DB: db, Driver: cfg.DatabaseDriver}
	catalogH := catalog.Handler{DB: db}
	quotationH := quotation.Handler{DB: db}
	invoiceH := invoice.Handler{DB: db}
	paymentH := payment.Handler{DB: db}
	expenseH := expense.Handler{DB: db}
	reportH := report.Handler{DB: db}
	documentH := document.Handler{DB: db}
	authLimiter := httpx.NewRateLimiter(12, time.Minute)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthH.Health)
	mux.HandleFunc("GET /readyz", healthH.Ready)
	mux.Handle("POST /api/v1/auth/register", authLimiter.Middleware(http.HandlerFunc(authH.Register)))
	mux.Handle("POST /api/v1/auth/login", authLimiter.Middleware(http.HandlerFunc(authH.Login)))
	mux.HandleFunc("POST /api/v1/auth/logout", authH.Logout)
	mux.Handle("GET /api/v1/me", authM.RequireUser(http.HandlerFunc(authH.Me)))
	mux.Handle("POST /api/v1/company/setup", authM.RequireUser(http.HandlerFunc(companyH.Setup)))
	mux.Handle("GET /api/v1/dashboard", authM.RequireCompany(http.HandlerFunc(dashboardH.Summary)))
	mux.Handle("GET /api/v1/customers", authM.RequireCompany(http.HandlerFunc(customerH.List)))
	mux.Handle("POST /api/v1/customers", authM.RequireRoles("owner", "admin", "manager", "staff")(http.HandlerFunc(customerH.Create)))
	mux.Handle("PUT /api/v1/customers/{id}", authM.RequireRoles("owner", "admin", "manager", "staff")(http.HandlerFunc(customerH.Update)))
	mux.Handle("DELETE /api/v1/customers/{id}", authM.RequireRoles("owner", "admin", "manager")(http.HandlerFunc(customerH.Delete)))
	mux.Handle("GET /api/v1/products", authM.RequireCompany(http.HandlerFunc(catalogH.List)))
	mux.Handle("POST /api/v1/products", authM.RequireRoles("owner", "admin", "manager", "staff")(http.HandlerFunc(catalogH.Create)))
	mux.Handle("PUT /api/v1/products/{id}", authM.RequireRoles("owner", "admin", "manager", "staff")(http.HandlerFunc(catalogH.Update)))
	mux.Handle("DELETE /api/v1/products/{id}", authM.RequireRoles("owner", "admin", "manager")(http.HandlerFunc(catalogH.Delete)))
	mux.Handle("GET /api/v1/quotations", authM.RequireCompany(http.HandlerFunc(quotationH.List)))
	mux.Handle("POST /api/v1/quotations", authM.RequireRoles("owner", "admin", "manager", "staff")(http.HandlerFunc(quotationH.Create)))
	mux.Handle("PATCH /api/v1/quotations/{id}/status", authM.RequireRoles("owner", "admin", "manager", "staff")(http.HandlerFunc(quotationH.Status)))
	mux.Handle("GET /api/v1/invoices", authM.RequireCompany(http.HandlerFunc(invoiceH.List)))
	mux.Handle("POST /api/v1/quotations/{id}/invoice", authM.RequireRoles("owner", "admin", "manager")(http.HandlerFunc(invoiceH.FromQuotation)))
	mux.Handle("PATCH /api/v1/invoices/{id}/status", authM.RequireRoles("owner", "admin", "manager")(http.HandlerFunc(invoiceH.Status)))
	mux.Handle("GET /api/v1/payments", authM.RequireCompany(http.HandlerFunc(paymentH.List)))
	mux.Handle("POST /api/v1/payments", authM.RequireRoles("owner", "admin", "manager", "staff")(http.HandlerFunc(paymentH.Create)))
	mux.Handle("GET /api/v1/company/profile", authM.RequireCompany(http.HandlerFunc(companyH.Profile)))
	mux.Handle("PUT /api/v1/company/profile", authM.RequireRoles("owner", "admin")(http.HandlerFunc(companyH.UpdateProfile)))
	mux.Handle("GET /api/v1/expenses", authM.RequireCompany(http.HandlerFunc(expenseH.List)))
	mux.Handle("POST /api/v1/expenses", authM.RequireRoles("owner", "admin", "manager", "staff")(http.HandlerFunc(expenseH.Create)))
	mux.Handle("DELETE /api/v1/expenses/{id}", authM.RequireRoles("owner", "admin", "manager")(http.HandlerFunc(expenseH.Delete)))
	mux.Handle("GET /api/v1/reports/finance", authM.RequireCompany(http.HandlerFunc(reportH.Summary)))
	mux.Handle("GET /api/v1/documents/quotations/{id}/pdf", authM.RequireCompany(http.HandlerFunc(documentH.Quotation)))
	mux.Handle("GET /api/v1/documents/invoices/{id}/pdf", authM.RequireCompany(http.HandlerFunc(documentH.Invoice)))
	mux.Handle("GET /api/v1/documents/payments/{id}/pdf", authM.RequireCompany(http.HandlerFunc(documentH.Receipt)))

	handler := httpx.Chain(mux,
		httpx.Recover,
		httpx.SecurityHeaders,
		httpx.SameOriginCSRF(cfg.AppOrigin),
		httpx.LimitBody(1<<20),
		httpx.Timeout(15*time.Second),
	)
	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 20 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 120 * time.Second}

	go func() {
		logger.Info("VENOMBusiness API started", "addr", cfg.HTTPAddr, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	logger.Info("server stopped")
}
