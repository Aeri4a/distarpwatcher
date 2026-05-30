package routes

import (
	"encoding/json"
	"net/http"
	"strconv"

	"server/internal/database"

	"github.com/go-chi/chi/v5"
)

type NotificationChannelsRoutes struct {
	db database.Database
}

func NewNotificationChannelsRoutes(db database.Database) *NotificationChannelsRoutes {
	return &NotificationChannelsRoutes{db: db}
}

func (ncr *NotificationChannelsRoutes) RegisterRoutes(router chi.Router) {
	router.Route("/notification_channels", func(r chi.Router) {
		r.Get("/", ncr.getAll)
		r.Post("/", ncr.create)
		r.Put("/{id}", ncr.update)
		r.Delete("/{id}", ncr.delete)
	})
}

func (ncr *NotificationChannelsRoutes) getAll(w http.ResponseWriter, r *http.Request) {
	channels, err := ncr.db.GetAllNotificationChannels(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(channels)
}

func (ncr *NotificationChannelsRoutes) create(w http.ResponseWriter, r *http.Request) {
	var ch database.NotificationChannel
	if err := json.NewDecoder(r.Body).Decode(&ch); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if ch.Name == "" || ch.Type == "" || ch.Target == "" {
		http.Error(w, "missing required fields (name, type, target)", http.StatusBadRequest)
		return
	}
	
	if ch.MinSeverity == "" {
		ch.MinSeverity = "INFO" // default
	}

	id, err := ncr.db.CreateNotificationChannel(r.Context(), &ch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ch.ID = id
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ch)
}

func (ncr *NotificationChannelsRoutes) update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var ch database.NotificationChannel
	if err := json.NewDecoder(r.Body).Decode(&ch); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := ncr.db.UpdateNotificationChannel(r.Context(), id, &ch); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ch.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ch)
}

func (ncr *NotificationChannelsRoutes) delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := ncr.db.DeleteNotificationChannel(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
