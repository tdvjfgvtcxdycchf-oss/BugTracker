package server

import (
	"bug_tracker/internal/service"
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
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

func NewRouter(ctx context.Context, conn *pgx.Conn) http.Handler {
	svc := service.NewTaskTrackerService(conn)

	router := mux.NewRouter()
	router.Use(loggingMiddleware)

	router.Path("/users").Methods("POST").HandlerFunc(HandleCreateUser(ctx, svc))
	router.Path("/login").Methods("POST").HandlerFunc(HandleGetIdUser(ctx, svc))
	router.Path("/tasks").Methods("POST").HandlerFunc(HandleCreateTask(ctx, svc))
	router.Path("/tasks").Methods("GET").HandlerFunc(HandleGetAllTasks(ctx, svc))
	router.Path("/bugs/{id}").Methods("POST").HandlerFunc(HandleCreateBug(ctx, svc))
	router.Path("/bugs/{id}").Methods("GET").HandlerFunc(HandleFuncGetAllBugs(ctx, svc))
	router.Path("/bugs/{id}").Methods("PATCH").HandlerFunc(HandleUpdateBug(ctx, svc))
	return router
}
