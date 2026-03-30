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
	authLimiter := newRateLimiter(15, time.Minute)

	jwtSecret := os.Getenv("JWT_SECRET")
	requireAuth := func(h http.Handler) http.Handler {
		return jwtAuthMiddleware(jwtSecret, conn, h)
	}

	// публичные — регистрация и логин
	router.Path("/healthz").Methods("GET").HandlerFunc(HandleHealthz())
	router.Path("/users").Methods("POST").Handler(withRateLimit(authLimiter, "signup", HandleCreateUser(svc)))
	router.Path("/login").Methods("POST").Handler(withRateLimit(authLimiter, "login", HandleGetIdUser(svc)))

	// защищённые
	router.Path("/me").Methods("GET").Handler(requireAuth(HandleGetMe(svc)))
	router.Path("/me/email").Methods("PATCH").Handler(requireAuth(HandleChangeMyEmail(svc)))
	router.Path("/me/password").Methods("PATCH").Handler(requireAuth(HandleChangeMyPassword(svc)))
	router.Path("/me/logout-all").Methods("POST").Handler(requireAuth(HandleLogoutAllSessions(svc)))

	router.Path("/users/{id}").Methods("GET").Handler(requireAuth(HandleGetOtherEmails(svc)))

	// Orgs / projects
	router.Path("/orgs").Methods("GET").Handler(requireAuth(HandleGetOrgs(svc)))
	router.Path("/orgs").Methods("POST").Handler(requireAuth(HandleCreateOrg(svc)))
	router.Path("/orgs/{id}/members").Methods("POST").Handler(requireAuth(HandleAddOrgMember(svc)))
	router.Path("/orgs/{id}/members").Methods("GET").Handler(requireAuth(HandleGetOrgMembers(svc)))
	router.Path("/orgs/{id}/members/{userId}").Methods("PATCH").Handler(requireAuth(HandleUpdateOrgMemberRole(svc)))
	router.Path("/orgs/{id}/members/{userId}").Methods("DELETE").Handler(requireAuth(HandleDeleteOrgMember(svc)))

	router.Path("/projects").Methods("GET").Handler(requireAuth(HandleGetProjects(svc)))
	router.Path("/projects").Methods("POST").Handler(requireAuth(HandleCreateProject(svc)))
	router.Path("/projects/{id}/members").Methods("POST").Handler(requireAuth(HandleAddProjectMember(svc)))
	router.Path("/projects/{id}/members").Methods("GET").Handler(requireAuth(HandleGetProjectMembers(svc)))
	router.Path("/projects/{id}/members/{userId}").Methods("PATCH").Handler(requireAuth(HandleUpdateProjectMemberRole(svc)))
	router.Path("/projects/{id}/members/{userId}").Methods("DELETE").Handler(requireAuth(HandleDeleteProjectMember(svc)))

	// Tasks are scoped by project_id
	router.Path("/tasks").Methods("POST").Handler(requireAuth(HandleCreateTask(svc)))
	router.Path("/tasks").Methods("GET").Handler(requireAuth(HandleGetAllTasks(svc)))
	router.Path("/tasks/{id}").Methods("DELETE").Handler(requireAuth(HandleDeleteTask(svc)))

	router.Path("/bugs/{id}").Methods("POST").Handler(requireAuth(HandleCreateBug(svc)))
	router.Path("/bugs/{id}").Methods("GET").Handler(requireAuth(HandleFuncGetAllBugs(svc)))
	router.Path("/bugs/{id}").Methods("PATCH").Handler(requireAuth(HandleUpdateBug(svc)))
	router.Path("/bugs/{id}").Methods("DELETE").Handler(requireAuth(HandleDeleteBug(svc)))

	router.Path("/bugs/{id}/photo").Methods("POST").Handler(requireAuth(HandleUploadBugPhoto(svc)))
	router.Path("/bugs/{id}/photo").Methods("GET").Handler(requireAuth(HandleGetBugPhoto(svc)))

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

	// Chat
	router.Path("/chat/threads").Methods("POST").Handler(requireAuth(HandleEnsureChatThread(svc)))
	router.Path("/chat/threads").Methods("GET").Handler(requireAuth(HandleGetChatThreads(svc)))
	router.Path("/chat/threads/{id}/messages").Methods("GET").Handler(requireAuth(HandleGetChatMessages(svc)))
	router.Path("/chat/threads/{id}/messages").Methods("POST").Handler(requireAuth(HandleAddChatMessage(svc)))
	router.Path("/chat/messages/{id}").Methods("PATCH").Handler(requireAuth(HandleUpdateChatMessage(svc)))
	router.Path("/chat/messages/{id}").Methods("DELETE").Handler(requireAuth(HandleDeleteChatMessage(svc)))
	router.Path("/chat/threads/{id}/read").Methods("POST").Handler(requireAuth(HandleMarkChatThreadRead(svc)))
	router.Path("/chat/threads/{id}/typing").Methods("POST").Handler(requireAuth(HandleSetTypingState(svc)))
	router.Path("/chat/threads/{id}/typing").Methods("GET").Handler(requireAuth(HandleGetTypingState(svc)))

	if jwtSecret == "" {
		slog.Warn("JWT_SECRET is not set — all protected routes will return 500")
	}

	return router
}
