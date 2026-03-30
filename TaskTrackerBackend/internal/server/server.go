package server

import (
	"bug_tracker/internal/service"
	"context"
	"log/slog"
	"net/http"
	"os"
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

	jwtSecret := os.Getenv("JWT_SECRET")
	requireAuth := func(h http.Handler) http.Handler {
		return jwtAuthMiddleware(jwtSecret, h)
	}

	// публичные — регистрация и логин
	router.Path("/users").Methods("POST").HandlerFunc(HandleCreateUser(svc))
	router.Path("/login").Methods("POST").HandlerFunc(HandleGetIdUser(svc))

	// защищённые
	router.Path("/users/{id}").Methods("GET").Handler(requireAuth(HandleGetOtherEmails(svc)))

	router.Path("/tasks").Methods("POST").Handler(requireAuth(HandleCreateTask(svc)))
	router.Path("/tasks").Methods("GET").Handler(requireAuth(HandleGetAllTasks(svc)))
	router.Path("/tasks/{id}").Methods("DELETE").Handler(requireAuth(HandleDeleteTask(svc)))

	router.Path("/bugs/{id}").Methods("POST").Handler(requireAuth(HandleCreateBug(svc)))
	router.Path("/bugs/{id}").Methods("GET").Handler(requireAuth(HandleFuncGetAllBugs(svc)))
	router.Path("/bugs/{id}").Methods("PATCH").Handler(requireAuth(HandleUpdateBug(svc)))
	router.Path("/bugs/{id}").Methods("DELETE").Handler(requireAuth(HandleDeleteBug(svc)))

	router.Path("/bugs/{id}/photo").Methods("POST").Handler(requireAuth(HandleUploadBugPhoto(svc)))
	// GET фото оставляем публичным — фронт грузит через <img src> без заголовков
	router.Path("/bugs/{id}/photo").Methods("GET").HandlerFunc(HandleGetBugPhoto(svc))

	// Comments
	router.Path("/bugs/{id}/comments").Methods("POST").Handler(requireAuth(HandleAddComment(svc)))
	router.Path("/bugs/{id}/comments").Methods("GET").Handler(requireAuth(HandleGetComments(svc)))

	// Audit log
	router.Path("/bugs/{id}/audit").Methods("GET").Handler(requireAuth(HandleGetAuditLog(svc)))

	// Relations
	router.Path("/bugs/{id}/relations").Methods("POST").Handler(requireAuth(HandleAddRelation(svc)))
	router.Path("/bugs/{id}/relations").Methods("GET").Handler(requireAuth(HandleGetRelations(svc)))
	router.Path("/relations/{relId}").Methods("DELETE").Handler(requireAuth(HandleDeleteRelation(svc)))

	// Tags
	router.Path("/bugs/{id}/tags").Methods("PUT").Handler(requireAuth(HandleSetTags(svc)))
	router.Path("/bugs/{id}/tags").Methods("GET").Handler(requireAuth(HandleGetTags(svc)))

	// Templates
	router.Path("/templates").Methods("GET").Handler(requireAuth(HandleGetTemplates(svc)))
	router.Path("/templates").Methods("POST").Handler(requireAuth(HandleCreateTemplate(svc)))
	router.Path("/templates/{id}").Methods("DELETE").Handler(requireAuth(HandleDeleteTemplate(svc)))

	// Analytics
	router.Path("/stats").Methods("GET").Handler(requireAuth(HandleGetStats(svc)))
	router.Path("/stats/tasks/{id}").Methods("GET").Handler(requireAuth(HandleGetStats(svc)))

	if jwtSecret == "" {
		slog.Warn("JWT_SECRET is not set — all protected routes will return 500")
	}

	return router
}
