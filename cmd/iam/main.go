// Command iam serves the personal website (a statically built Astro site) and
// its like API as a single binary.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"golang.org/x/time/rate"

	"github.com/moriT958/iam/internal/like"
	"github.com/moriT958/iam/internal/server"
	"github.com/moriT958/iam/web"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("fatal", slog.Any("error", err))
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	site, err := web.Dist()
	if err != nil {
		return err
	}

	slugs, err := like.LoadSlugs(site, "posts.json")
	if err != nil {
		return err
	}

	dbPath := filepath.Join(getenv("IAM_DATA_DIR", "data"), "iam.db")
	store, err := like.OpenSQLite(dbPath)
	if err != nil {
		return err
	}
	defer func() { _ = store.Close() }()
	logger.Info("database ready", slog.String("path", dbPath), slog.Int("posts", len(slugs)))

	// 10 likes at once per visitor, then one every 3 seconds.
	limiter := like.NewLimiter(rate.Every(3*time.Second), 10)

	addr := ":" + getenv("PORT", "8080")
	srv := &http.Server{
		Addr: addr,
		Handler: server.New(server.Config{
			Site:   site,
			Likes:  like.NewHandler(store, slugs, limiter, logger),
			Logger: logger,
		}),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("listening", slog.String("addr", addr))
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
