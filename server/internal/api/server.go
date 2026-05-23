package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"server/internal/api/routes"
	"server/internal/config"
	"server/internal/database"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type RouteGroup interface {
	RegisterRoutes(router chi.Router)
}

func newRouter(groups ...RouteGroup) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	router.Route("/api/v1", func(r chi.Router) {
		for _, group := range groups {
			group.RegisterRoutes(r)
		}
	})

	return router
}

type APIServer struct {
	config config.APIConfig
	db     database.Database
}

func NewAPIServer(cfg config.APIConfig, db database.Database) *APIServer {
	return &APIServer{
		config: cfg,
		db:     db,
	}
}

func (s *APIServer) Start(ctx context.Context) error {
	r := newRouter(
		routes.NewARPEventsRoutes(s.db),
	)

	srv := &http.Server{
		Addr:    s.config.Port,
		Handler: r,
	}

	errChan := make(chan error, 1)
	go func() {
		log.Printf("REST API listening at %s", s.config.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("Shutting down REST API...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errChan:
		return err
	}
}
