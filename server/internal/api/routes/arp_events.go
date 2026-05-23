package routes

import (
	"net/http"

	"server/internal/database"

	"github.com/go-chi/chi/v5"
)

type ARPEventsRoutes struct {
	db database.Database
}

func NewARPEventsRoutes(db database.Database) *ARPEventsRoutes {
	return &ARPEventsRoutes{db: db}
}

func (arpe *ARPEventsRoutes) RegisterRoutes(router chi.Router) {
	router.Get("/arp_events", arpe.getAllARPEvents)
}

func (arpe *ARPEventsRoutes) getAllARPEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"events": []}`))
}
