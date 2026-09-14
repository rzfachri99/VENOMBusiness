package health

import (
	"context"
	"database/sql"
	"github.com/venombusiness/venombusiness/apps/api/internal/httpx"
	"net/http"
	"time"
)

type Handler struct {
	DB     *sql.DB
	Driver string
}

func (h Handler) Health(w http.ResponseWriter, r *http.Request) {
	httpx.JSON(w, 200, map[string]string{"status": "ok"})
}
func (h Handler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := h.DB.PingContext(ctx); err != nil {
		httpx.Error(w, 503, "database unavailable")
		return
	}
	var version int
	_ = h.DB.QueryRowContext(ctx, `SELECT COALESCE(MAX(version),0) FROM schema_migrations`).Scan(&version)
	st := h.DB.Stats()
	httpx.JSON(w, 200, map[string]any{"status": "ready", "database": map[string]any{"driver": h.Driver, "migration_version": version, "open_connections": st.OpenConnections, "in_use": st.InUse, "idle": st.Idle}})
}
