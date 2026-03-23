package server

import (
	"bug_tracker/internal/service"
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(wrapped, r)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"status", wrapped.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func NewRouter(ctx context.Context, conn *pgxpool.Pool) http.Handler {
	svc := service.NewTaskTrackerService(conn)

	router := mux.NewRouter()
	router.Use(loggingMiddleware)

	router.Path("/users").Methods("POST").HandlerFunc(HandleCreateUser(svc))
	router.Path("/users/{id}").Methods("GET").HandlerFunc(HandleGetOtherEmails(svc))
	router.Path("/login").Methods("POST").HandlerFunc(HandleGetIdUser(svc))
	router.Path("/tasks").Methods("POST").HandlerFunc(HandleCreateTask(svc))
	router.Path("/tasks").Methods("GET").HandlerFunc(HandleGetAllTasks(svc))
	router.Path("/bugs/{id}").Methods("POST").HandlerFunc(HandleCreateBug(svc))
	router.Path("/bugs/{id}").Methods("GET").HandlerFunc(HandleFuncGetAllBugs(svc))
	router.Path("/bugs/{id}").Methods("PATCH").HandlerFunc(HandleUpdateBug(svc))
	return router
}
