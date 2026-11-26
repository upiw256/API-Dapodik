package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"dapodik-readonly-api/internal/config"
	"dapodik-readonly-api/internal/database"
	"dapodik-readonly-api/internal/middleware"
	"dapodik-readonly-api/internal/modules/pesertadidik"
	"dapodik-readonly-api/internal/modules/ptk"
	"dapodik-readonly-api/internal/modules/referensi"
	"dapodik-readonly-api/internal/modules/rombel"
	"dapodik-readonly-api/internal/modules/sekolah"
	"dapodik-readonly-api/internal/period"
	"dapodik-readonly-api/internal/response"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if err := runServiceOrConsole(); err != nil {
		log.Fatal(err)
	}
}

func runAPI(serverCtx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	startupCtx, cancelStartup := context.WithTimeout(context.Background(), cfg.DBConnectTimeout)
	pool, err := database.Open(startupCtx, cfg.DBDSN, cfg.DBConnectTimeout, cfg.DBMinConns, cfg.DBMaxConns)
	cancelStartup()
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()
	periodCtx, cancelPeriod := context.WithTimeout(context.Background(), cfg.DBConnectTimeout)
	selectedPeriod, err := period.Resolve(periodCtx, pool, cfg.Semester)
	cancelPeriod()
	if err != nil {
		return fmt.Errorf("resolve configured semester: %w", err)
	}
	log.Printf("Using semester %q (ID %s, academic year %d)", selectedPeriod.Label, selectedPeriod.SemesterID, selectedPeriod.TahunAjaranID)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", healthHandler(pool))
	referensiRepository := referensi.NewPostgresRepository(pool, selectedPeriod)
	referensiService := referensi.NewService(referensiRepository)
	referensi.NewHandler(referensiService).RegisterRoutes(mux)
	sekolahRepository := sekolah.NewPostgresRepository(pool, selectedPeriod, cfg.NPSN)
	sekolahService := sekolah.NewService(sekolahRepository)
	sekolah.NewHandler(sekolahService).RegisterRoutes(mux)
	ptkRepository := ptk.NewPostgresRepository(pool, selectedPeriod)
	ptkService := ptk.NewService(ptkRepository)
	ptk.NewHandler(ptkService).RegisterRoutes(mux)
	pdRepository := pesertadidik.NewPostgresRepository(pool, selectedPeriod)
	pdService := pesertadidik.NewService(pdRepository)
	pesertadidik.NewHandler(pdService).RegisterRoutes(mux)
	rombelRepository := rombel.NewPostgresRepository(pool, selectedPeriod)
	rombelService := rombel.NewService(rombelRepository)
	rombel.NewHandler(rombelService).RegisterRoutes(mux)

	var handler http.Handler = mux
	handler = middleware.RateLimit(5, 10*time.Minute)(handler)
	handler = middleware.APIKey(cfg.APIKey)(handler)
	handler = middleware.ReadOnly(handler)
	handler = middleware.RequestLog(handler)

	server := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      handler,
		ReadTimeout:  cfg.HTTPReadTimeout,
		WriteTimeout: cfg.HTTPWriteTimeout,
		IdleTimeout:  cfg.HTTPIdleTimeout,
	}

	serveErr := make(chan error, 1)
	go func() {
		log.Printf("API listening on :%s", cfg.AppPort)
		serveErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("HTTP server: %w", err)
		}
		return nil
	case <-serverCtx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("HTTP server shutdown: %w", err)
		}
		if err := <-serveErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("HTTP server: %w", err)
		}
		return nil
	}
}

func healthHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			response.WriteInternalErrorStatus(w, r, http.StatusServiceUnavailable, "database unavailable", err)
			return
		}
		response.WriteSuccess(w, http.StatusOK, "healthy", map[string]string{"database": "connected"}, nil)
	}
}
