package server

import (
	"bug_tracker/internal/auth"
	"bug_tracker/internal/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": msg,
	})
}

func writeJSONErrorCode(w http.ResponseWriter, status int, code string, msg string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": msg,
		"code":  code,
	})
}

func HandleHealthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

func validatePasswordStrength(password string) (bool, string) {
	if len(password) < 8 {
		return false, "password_too_short"
	}
	var hasUpper, hasLower, hasDigit bool
	for _, ch := range password {
		if ch >= 'A' && ch <= 'Z' {
			hasUpper = true
		}
		if ch >= 'a' && ch <= 'z' {
			hasLower = true
		}
		if ch >= '0' && ch <= '9' {
			hasDigit = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit {
		return false, "password_too_weak"
	}
	return true, ""
}

func HandleCreateUser(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var user sql.User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			slog.Warn("create user: invalid json", "error", err)
			writeJSONError(w, http.StatusBadRequest, "invalid json")
			return
		}

		if user.Email == "" || !strings.Contains(user.Email, "@") {
			slog.Warn("create user: invalid email", "error", "missing @ or string is empty")
			writeJSONError(w, http.StatusBadRequest, "invalid email")
			return
		}

		if user.Password == "" {
			slog.Warn("create user: empty password", "error", "string is empty")
			writeJSONErrorCode(w, http.StatusBadRequest, "password_required", "empty password")
			return
		}
		if ok, code := validatePasswordStrength(user.Password); !ok {
			writeJSONErrorCode(w, http.StatusBadRequest, code, "password must be at least 8 chars and include upper, lower, digit")
			return
		}

		createdUserId, role, err := svc.Register(r.Context(), user)
		if err != nil {
			if err.Error() == "user already exists" {
				writeJSONErrorCode(w, http.StatusConflict, "user_exists", "user already exist")
				return
			}
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}

		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			writeJSONError(w, http.StatusInternalServerError, "JWT_SECRET is not set")
			return
		}
		version, verErr := svc.GetUserJWTVersion(r.Context(), createdUserId)
		if verErr != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to read jwt version")
			return
		}
		token, tokenErr := auth.GenerateJWT(jwtSecret, createdUserId, user.Email, version, 24*time.Hour)
		if tokenErr != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to generate jwt")
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"id": createdUserId, "token": token, "role": role})
	}
}

func HandleGetIdUser(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var user sql.User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			slog.Warn("login user: invalid json", "error", err)
			writeJSONError(w, http.StatusBadRequest, "invalid json")
			return
		}

		if user.Email == "" || !strings.Contains(user.Email, "@") {
			slog.Warn("login user: invalid email", "error", "missing @ or string is empty")
			writeJSONError(w, http.StatusBadRequest, "invalid email")
			return
		}

		if user.Password == "" {
			slog.Warn("login user: empty password", "error", "string is empty")
			writeJSONErrorCode(w, http.StatusBadRequest, "password_required", "empty password")
			return
		}

		loginedUserId, role, err := svc.Login(r.Context(), user)
		if err != nil {
			writeJSONErrorCode(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
			return
		}

		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			writeJSONError(w, http.StatusInternalServerError, "JWT_SECRET is not set")
			return
		}
		version, verErr := svc.GetUserJWTVersion(r.Context(), loginedUserId)
		if verErr != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to read jwt version")
			return
		}
		token, tokenErr := auth.GenerateJWT(jwtSecret, loginedUserId, user.Email, version, 24*time.Hour)
		if tokenErr != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to generate jwt")
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{"id": loginedUserId, "token": token, "role": role})
	}
}

func HandleGetMe(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		u, err := svc.GetUserAuthByID(r.Context(), userID)
		if err != nil || u == nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to fetch profile")
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"id":    u.Id,
			"email": u.Email,
			"role":  u.Role,
		})
	}
}

func HandleChangeMyEmail(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		var body struct {
			NewEmail        string `json:"new_email"`
			CurrentPassword string `json:"current_password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid body")
			return
		}
		if !strings.Contains(body.NewEmail, "@") || body.CurrentPassword == "" {
			writeJSONError(w, http.StatusBadRequest, "new_email and current_password are required")
			return
		}
		u, err := svc.GetUserAuthByID(r.Context(), userID)
		if err != nil || u == nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to fetch user")
			return
		}
		if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(body.CurrentPassword)) != nil && u.Password != body.CurrentPassword {
			writeJSONError(w, http.StatusUnauthorized, "invalid password")
			return
		}
		if err := svc.UpdateUserEmail(r.Context(), userID, strings.TrimSpace(body.NewEmail)); err != nil {
			writeJSONError(w, http.StatusConflict, "email already in use")
			return
		}
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}
}

func HandleChangeMyPassword(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		var body struct {
			CurrentPassword string `json:"current_password"`
			NewPassword     string `json:"new_password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid body")
			return
		}
		if body.CurrentPassword == "" {
			writeJSONErrorCode(w, http.StatusBadRequest, "password_required", "current_password is required")
			return
		}
		if ok, code := validatePasswordStrength(body.NewPassword); !ok {
			writeJSONErrorCode(w, http.StatusBadRequest, code, "new_password must be at least 8 chars and include upper, lower, digit")
			return
		}
		u, err := svc.GetUserAuthByID(r.Context(), userID)
		if err != nil || u == nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to fetch user")
			return
		}
		if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(body.CurrentPassword)) != nil && u.Password != body.CurrentPassword {
			writeJSONErrorCode(w, http.StatusUnauthorized, "invalid_password", "invalid password")
			return
		}
		if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(body.NewPassword)) == nil || u.Password == body.NewPassword {
			writeJSONErrorCode(w, http.StatusBadRequest, "password_reused", "new password must be different from current password")
			return
		}
		newHash, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to hash password")
			return
		}
		if err := svc.UpdateUserPasswordHash(r.Context(), userID, string(newHash)); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to update password")
			return
		}
		_, _ = svc.IncrementUserJWTVersion(r.Context(), userID)
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}
}

func HandleLogoutAllSessions(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		if _, err := svc.IncrementUserJWTVersion(r.Context(), userID); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to logout sessions")
			return
		}
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}
}

func HandleGetOtherEmails(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		vars := mux.Vars(r)
		excludeStr := vars["id"]

		excludeId, err := strconv.Atoi(excludeStr)
		if err != nil {
			slog.Error("invalid id in path", "id", excludeStr)
			writeJSONError(w, http.StatusBadRequest, "invalid user ID format")
			return
		}

		emails, err := svc.GetOtherUsersEmails(r.Context(), excludeId)
		if err != nil {
			slog.Error("failed to fetch emails", "error", err)
			writeJSONError(w, http.StatusInternalServerError, "could not fetch emails")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(emails)
	}
}

// ── Orgs / Projects ───────────────────────────────────────────────────────────

func HandleGetOrgs(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		orgs, err := svc.GetUserOrgs(r.Context(), userID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "could not fetch orgs")
			return
		}
		if orgs == nil {
			w.Write([]byte("[]"))
			return
		}
		json.NewEncoder(w).Encode(orgs)
	}
}

func HandleCreateOrg(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		var body struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Name) == "" {
			writeJSONError(w, http.StatusBadRequest, "name is required")
			return
		}
		id, err := svc.CreateOrg(r.Context(), strings.TrimSpace(body.Name), userID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to create org")
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int{"id": id})
	}
}

func HandleAddOrgMember(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		orgID, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil || orgID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid org id")
			return
		}
		var body struct {
			Email string `json:"email"`
			Role  string `json:"role"` // owner|admin|member
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || !strings.Contains(body.Email, "@") {
			writeJSONError(w, http.StatusBadRequest, "valid email is required")
			return
		}
		okAdd, reason, err := svc.AddUserToOrgByEmail(r.Context(), orgID, body.Email, body.Role, userID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to add member")
			return
		}
		if !okAdd {
			if reason == "user_not_found" {
				writeJSONError(w, http.StatusNotFound, "user_not_found")
				return
			}
			writeJSONError(w, http.StatusForbidden, "not_allowed")
			return
		}
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}
}

func HandleGetProjects(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		orgIDStr := r.URL.Query().Get("org_id")
		orgID, err := strconv.Atoi(orgIDStr)
		if err != nil || orgID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "org_id is required")
			return
		}
		projects, err := svc.GetUserProjects(r.Context(), userID, orgID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "could not fetch projects")
			return
		}
		if projects == nil {
			w.Write([]byte("[]"))
			return
		}
		json.NewEncoder(w).Encode(projects)
	}
}

func HandleCreateProject(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		var body struct {
			OrgId int    `json:"org_id"`
			Name  string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.OrgId <= 0 || strings.TrimSpace(body.Name) == "" {
			writeJSONError(w, http.StatusBadRequest, "org_id and name are required")
			return
		}
		id, err := svc.CreateProject(r.Context(), body.OrgId, strings.TrimSpace(body.Name), userID)
		if err != nil {
			writeJSONError(w, http.StatusForbidden, "not allowed")
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int{"id": id})
	}
}

func HandleAddProjectMember(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		projectID, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil || projectID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid project id")
			return
		}
		var body struct {
			Email string `json:"email"`
			Role  string `json:"role"` // pm|dev|qa|viewer
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || !strings.Contains(body.Email, "@") {
			writeJSONError(w, http.StatusBadRequest, "valid email is required")
			return
		}
		okAdd, reason, err := svc.AddUserToProjectByEmail(r.Context(), projectID, body.Email, body.Role, userID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to add member")
			return
		}
		if !okAdd {
			if reason == "user_not_found" {
				writeJSONError(w, http.StatusNotFound, "user_not_found")
				return
			}
			writeJSONError(w, http.StatusForbidden, "not_allowed")
			return
		}
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}
}

func HandleGetOrgMembers(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		actorID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		orgID, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil || orgID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid org id")
			return
		}
		list, err := svc.GetOrgMembers(r.Context(), orgID, actorID)
		if err != nil {
			writeJSONError(w, http.StatusForbidden, "not allowed")
			return
		}
		if list == nil {
			w.Write([]byte("[]"))
			return
		}
		json.NewEncoder(w).Encode(list)
	}
}

func HandleUpdateOrgMemberRole(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		actorID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		orgID, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil || orgID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid org id")
			return
		}
		targetUserID, err := strconv.Atoi(mux.Vars(r)["userId"])
		if err != nil || targetUserID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid user id")
			return
		}
		var body struct {
			Role string `json:"role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Role == "" {
			writeJSONError(w, http.StatusBadRequest, "role is required")
			return
		}
		okUpd, err := svc.UpdateOrgMemberRole(r.Context(), orgID, targetUserID, body.Role, actorID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to update role")
			return
		}
		if !okUpd {
			writeJSONError(w, http.StatusForbidden, "not allowed")
			return
		}
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}
}

func HandleDeleteOrgMember(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		actorID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		orgID, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil || orgID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid org id")
			return
		}
		targetUserID, err := strconv.Atoi(mux.Vars(r)["userId"])
		if err != nil || targetUserID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid user id")
			return
		}
		okDel, err := svc.RemoveOrgMember(r.Context(), orgID, targetUserID, actorID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to delete member")
			return
		}
		if !okDel {
			writeJSONError(w, http.StatusForbidden, "not allowed")
			return
		}
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}
}

func HandleGetProjectMembers(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		actorID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		projectID, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil || projectID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid project id")
			return
		}
		list, err := svc.GetProjectMembers(r.Context(), projectID, actorID)
		if err != nil {
			writeJSONError(w, http.StatusForbidden, "not allowed")
			return
		}
		if list == nil {
			w.Write([]byte("[]"))
			return
		}
		json.NewEncoder(w).Encode(list)
	}
}

func HandleUpdateProjectMemberRole(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		actorID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		projectID, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil || projectID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid project id")
			return
		}
		targetUserID, err := strconv.Atoi(mux.Vars(r)["userId"])
		if err != nil || targetUserID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid user id")
			return
		}
		var body struct {
			Role string `json:"role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Role == "" {
			writeJSONError(w, http.StatusBadRequest, "role is required")
			return
		}
		okUpd, err := svc.UpdateProjectMemberRole(r.Context(), projectID, targetUserID, body.Role, actorID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to update role")
			return
		}
		if !okUpd {
			writeJSONError(w, http.StatusForbidden, "not allowed")
			return
		}
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}
}

func HandleDeleteProjectMember(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		actorID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		projectID, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil || projectID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid project id")
			return
		}
		targetUserID, err := strconv.Atoi(mux.Vars(r)["userId"])
		if err != nil || targetUserID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid user id")
			return
		}
		okDel, err := svc.RemoveProjectMember(r.Context(), projectID, targetUserID, actorID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to delete member")
			return
		}
		if !okDel {
			writeJSONError(w, http.StatusForbidden, "not allowed")
			return
		}
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}
}

func HandleGetAllTasks(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		projectIDStr := r.URL.Query().Get("project_id")
		projectID, err := strconv.Atoi(projectIDStr)
		if err != nil || projectID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "project_id is required")
			return
		}

		tasks, err := svc.GetTasksByProjectForUser(r.Context(), userID, projectID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "could not fetch tasks")
			return
		}
		if tasks == nil {
			w.Write([]byte("[]"))
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tasks)
	}
}

func HandleCreateTask(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var task sql.Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			slog.Warn("create task: invalid json", "error", err)
			writeJSONError(w, http.StatusBadRequest, "invalid json")
			return
		}

		if task.Title == "" {
			slog.Warn("task title: empty title", "error", "string is empty")
			writeJSONError(w, http.StatusBadRequest, "empty title")
			return
		}

		if task.Description == "" {
			slog.Warn("task description: empty description", "error", "string is empty")
			writeJSONError(w, http.StatusBadRequest, "empty description")
			return
		}

		if task.ProjectId <= 0 {
			writeJSONError(w, http.StatusBadRequest, "project_id is required")
			return
		}
		task.OwnerId = userID

		_, err := svc.CreateTask(r.Context(), task)
		if err != nil {
			writeJSONError(w, http.StatusForbidden, "not allowed")
			return
		}

		tasks, err := svc.GetTasksByProjectForUser(r.Context(), userID, task.ProjectId)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "could not fetch tasks")
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(tasks)
	}
}

func HandleFuncGetAllBugs(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		vars := mux.Vars(r)
		idStr := vars["id"]
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			slog.Warn("get bugs: invalid id", "id", idStr)
			writeJSONError(w, http.StatusBadRequest, "id must be positive integer")
			return
		}

		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		can, err := svc.CanUserAccessTask(r.Context(), userID, id)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to authorize")
			return
		}
		if !can {
			writeJSONError(w, http.StatusForbidden, "not allowed")
			return
		}

		bugs, err := svc.GetAllBugs(r.Context(), id)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "could not fetch bugs")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(bugs)
	}
}

func HandleCreateBug(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		vars := mux.Vars(r)
		idStr := vars["id"]
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			slog.Warn("get bugs: invalid id", "id", idStr)
			writeJSONError(w, http.StatusBadRequest, "id must be positive integer")
			return
		}

		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		can, err := svc.CanUserAccessTask(r.Context(), userID, id)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to authorize")
			return
		}
		if !can {
			writeJSONError(w, http.StatusForbidden, "not allowed")
			return
		}

		var bug sql.Bug
		if err := json.NewDecoder(r.Body).Decode(&bug); err != nil {
			slog.Warn("create bug: invalid json", "error", err)
			writeJSONError(w, http.StatusBadRequest, "invalid json")
			return
		}

		bug.TaskId = id
		if bug.CreatedBy == 0 {
			bug.CreatedBy = userID
		}

		if err := svc.CreateBug(r.Context(), bug); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "could not fetch bugs")
			return
		}

		bugs, err := svc.GetAllBugs(r.Context(), id)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "could not fetch bugs")
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(bugs)
	}
}

func HandleUpdateBug(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var input struct {
			sql.Bug
			AssignedToEmail string `json:"assigned_to_email"`
		}

		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			slog.Error("failed to decode bug body", "error", err)
			writeJSONError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		vars := mux.Vars(r)
		bugID, _ := strconv.Atoi(vars["id"])
		input.Bug.Id = bugID

		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		can, err := svc.CanUserAccessBug(r.Context(), userID, bugID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to authorize")
			return
		}
		if !can {
			writeJSONError(w, http.StatusForbidden, "not allowed")
			return
		}

		if err := svc.UpdateBug(r.Context(), input.Bug, input.AssignedToEmail); err != nil {
			slog.Error("failed to update bug", "error", err)
			writeJSONError(w, http.StatusInternalServerError, "failed to update bug")
			return
		}

		updatedBugs, err := svc.GetAllBugs(r.Context(), input.Bug.TaskId)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to fetch list")
			return
		}

		json.NewEncoder(w).Encode(updatedBugs)
	}
}

func HandleDeleteTask(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		vars := mux.Vars(r)
		idStr := vars["id"]
		taskID, err := strconv.Atoi(idStr)
		if err != nil || taskID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid task id")
			return
		}
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		deleted, err := svc.DeleteTask(r.Context(), taskID, userID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to delete task")
			return
		}

		if !deleted {
			writeJSONError(w, http.StatusForbidden, "task not found or not allowed")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"deleted": true})
	}
}

func HandleDeleteBug(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		vars := mux.Vars(r)
		idStr := vars["id"]
		bugID, err := strconv.Atoi(idStr)
		if err != nil || bugID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid bug id")
			return
		}

		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		can, err := svc.CanUserAccessBug(r.Context(), userID, bugID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to authorize")
			return
		}
		if !can {
			writeJSONError(w, http.StatusForbidden, "not allowed")
			return
		}

		deleted, err := svc.DeleteBug(r.Context(), bugID, userID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to delete bug")
			return
		}

		if !deleted {
			writeJSONError(w, http.StatusForbidden, "bug not found or not allowed")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"deleted": true})
	}
}

func HandleUploadBugPhoto(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr := vars["id"]
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid bug id")
			return
		}

		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		can, err := svc.CanUserAccessBug(r.Context(), userID, id)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to authorize")
			return
		}
		if !can {
			writeJSONError(w, http.StatusForbidden, "not allowed")
			return
		}

		if err := r.ParseMultipartForm(10 << 20); err != nil {
			writeJSONError(w, http.StatusBadRequest, "file too large")
			return
		}

		file, _, err := r.FormFile("photo")
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "no photo provided")
			return
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			slog.Error("failed to read photo file", "error", err)
			writeJSONError(w, http.StatusInternalServerError, "failed to read file")
			return
		}

		if err := svc.SaveBugPhoto(r.Context(), id, data); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to save photo")
			return
		}

		slog.Info("bug photo saved to db", "bug_id", id)
		w.WriteHeader(http.StatusCreated)
	}
}

// ── Comments ─────────────────────────────────────────────────────────────────

func HandleAddComment(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		bugID, err := strconv.Atoi(vars["id"])
		if err != nil || bugID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid bug id")
			return
		}
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		can, err := svc.CanUserAccessBug(r.Context(), userID, bugID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to authorize")
			return
		}
		if !can {
			writeJSONError(w, http.StatusForbidden, "not allowed")
			return
		}
		var body struct {
			Body string `json:"body"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Body == "" {
			writeJSONError(w, http.StatusBadRequest, "body is required")
			return
		}
		comment, err := svc.AddBugComment(r.Context(), bugID, userID, body.Body)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to add comment")
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(comment)
	}
}

func HandleGetComments(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		bugID, err := strconv.Atoi(vars["id"])
		if err != nil || bugID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid bug id")
			return
		}
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		can, err := svc.CanUserAccessBug(r.Context(), userID, bugID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to authorize")
			return
		}
		if !can {
			writeJSONError(w, http.StatusForbidden, "not allowed")
			return
		}
		comments, err := svc.GetBugComments(r.Context(), bugID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to fetch comments")
			return
		}
		if comments == nil {
			w.Write([]byte("[]"))
			return
		}
		json.NewEncoder(w).Encode(comments)
	}
}

// ── Audit log ─────────────────────────────────────────────────────────────────

func HandleGetAuditLog(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		bugID, err := strconv.Atoi(vars["id"])
		if err != nil || bugID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid bug id")
			return
		}
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		can, err := svc.CanUserAccessBug(r.Context(), userID, bugID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to authorize")
			return
		}
		if !can {
			writeJSONError(w, http.StatusForbidden, "not allowed")
			return
		}
		log, err := svc.GetAuditLog(r.Context(), bugID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to fetch audit log")
			return
		}
		if log == nil {
			w.Write([]byte("[]"))
			return
		}
		json.NewEncoder(w).Encode(log)
	}
}

// ── Relations ────────────────────────────────────────────────────────────────

func HandleAddRelation(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		bugID, err := strconv.Atoi(vars["id"])
		if err != nil || bugID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid bug id")
			return
		}
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		can, err := svc.CanUserAccessBug(r.Context(), userID, bugID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to authorize")
			return
		}
		if !can {
			writeJSONError(w, http.StatusForbidden, "not allowed")
			return
		}
		var body struct {
			ToBugId      int    `json:"to_bug_id"`
			RelationType string `json:"relation_type"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid body")
			return
		}
		if body.ToBugId <= 0 || body.RelationType == "" {
			writeJSONError(w, http.StatusBadRequest, "to_bug_id and relation_type are required")
			return
		}
		if err := svc.AddBugRelation(r.Context(), bugID, body.ToBugId, body.RelationType); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to add relation")
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}
}

func HandleGetRelations(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		bugID, err := strconv.Atoi(vars["id"])
		if err != nil || bugID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid bug id")
			return
		}
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		can, err := svc.CanUserAccessBug(r.Context(), userID, bugID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to authorize")
			return
		}
		if !can {
			writeJSONError(w, http.StatusForbidden, "not allowed")
			return
		}
		relations, err := svc.GetBugRelations(r.Context(), bugID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to fetch relations")
			return
		}
		if relations == nil {
			w.Write([]byte("[]"))
			return
		}
		json.NewEncoder(w).Encode(relations)
	}
}

func HandleDeleteRelation(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		relID, err := strconv.Atoi(vars["relId"])
		if err != nil || relID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid relation id")
			return
		}
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		// authorize via relation -> bug -> task -> project
		bugID, err := svc.GetRelationFromBugID(r.Context(), relID)
		if err != nil || bugID <= 0 {
			writeJSONError(w, http.StatusNotFound, "not found")
			return
		}
		can, err := svc.CanUserAccessBug(r.Context(), userID, bugID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to authorize")
			return
		}
		if !can {
			writeJSONError(w, http.StatusForbidden, "not allowed")
			return
		}
		if err := svc.DeleteBugRelation(r.Context(), relID); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to delete relation")
			return
		}
		json.NewEncoder(w).Encode(map[string]bool{"deleted": true})
	}
}

// ── Tags ─────────────────────────────────────────────────────────────────────

func HandleSetTags(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		bugID, err := strconv.Atoi(vars["id"])
		if err != nil || bugID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid bug id")
			return
		}
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		can, err := svc.CanUserAccessBug(r.Context(), userID, bugID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to authorize")
			return
		}
		if !can {
			writeJSONError(w, http.StatusForbidden, "not allowed")
			return
		}
		var body struct {
			Tags []string `json:"tags"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid body")
			return
		}
		if err := svc.SetBugTags(r.Context(), bugID, body.Tags); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to set tags")
			return
		}
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}
}

func HandleGetTags(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		bugID, err := strconv.Atoi(vars["id"])
		if err != nil || bugID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid bug id")
			return
		}
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		can, err := svc.CanUserAccessBug(r.Context(), userID, bugID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to authorize")
			return
		}
		if !can {
			writeJSONError(w, http.StatusForbidden, "not allowed")
			return
		}
		tags, err := svc.GetBugTags(r.Context(), bugID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to get tags")
			return
		}
		if tags == nil {
			tags = []string{}
		}
		json.NewEncoder(w).Encode(tags)
	}
}

// ── Templates ─────────────────────────────────────────────────────────────────

func HandleGetTemplates(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// TODO: scope templates by org/project in a follow-up. For now require auth only.
		templates, err := svc.GetTemplates(r.Context())
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to fetch templates")
			return
		}
		if templates == nil {
			w.Write([]byte("[]"))
			return
		}
		json.NewEncoder(w).Encode(templates)
	}
}

func HandleCreateTemplate(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		var tmpl sql.BugTemplate
		if err := json.NewDecoder(r.Body).Decode(&tmpl); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid body")
			return
		}
		if tmpl.Name == "" {
			writeJSONError(w, http.StatusBadRequest, "name is required")
			return
		}
		tmpl.CreatedBy = userID
		id, err := svc.CreateTemplate(r.Context(), tmpl)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to create template")
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int{"id": id})
	}
}

func HandleDeleteTemplate(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		tmplID, err := strconv.Atoi(vars["id"])
		if err != nil || tmplID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid template id")
			return
		}
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		deleted, err := svc.DeleteTemplate(r.Context(), tmplID, userID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to delete template")
			return
		}
		if !deleted {
			writeJSONError(w, http.StatusForbidden, "not found or not allowed")
			return
		}
		json.NewEncoder(w).Encode(map[string]bool{"deleted": true})
	}
}

// ── Analytics ─────────────────────────────────────────────────────────────────

func HandleGetStats(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// TODO: scope stats by org/project in a follow-up. For now require auth only.
		vars := mux.Vars(r)
		taskIDStr := vars["id"]

		if taskIDStr != "" {
			taskID, err := strconv.Atoi(taskIDStr)
			if err != nil || taskID <= 0 {
				writeJSONError(w, http.StatusBadRequest, "invalid task id")
				return
			}
			stats, err := svc.GetBugStatsByTask(r.Context(), taskID)
			if err != nil {
				writeJSONError(w, http.StatusInternalServerError, "failed to get stats")
				return
			}
			json.NewEncoder(w).Encode(stats)
			return
		}

		stats, err := svc.GetAllBugStats(r.Context())
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to get stats")
			return
		}
		json.NewEncoder(w).Encode(stats)
	}
}

func HandleGetBugPhoto(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		vars := mux.Vars(r)
		idStr := vars["id"]
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid bug id")
			return
		}

		can, err := svc.CanUserAccessBug(r.Context(), userID, id)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to authorize")
			return
		}
		if !can {
			writeJSONError(w, http.StatusForbidden, "not allowed")
			return
		}

		data, err := svc.GetBugPhoto(r.Context(), id)
		if err != nil || len(data) == 0 {
			writeJSONError(w, http.StatusNotFound, "not found")
			return
		}

		w.Header().Set("Content-Type", http.DetectContentType(data))
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

// ── Chat ──────────────────────────────────────────────────────────────────────

func HandleEnsureChatThread(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var body struct {
			Scope     string `json:"scope"` // org|project|dm
			OrgId     int    `json:"org_id"`
			ProjectId int    `json:"project_id"`
			Email     string `json:"email"` // for dm
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid body")
			return
		}

		var (
			threadID int
			err      error
		)
		switch body.Scope {
		case "org":
			threadID, err = svc.EnsureOrgThread(r.Context(), body.OrgId, userID)
		case "project":
			threadID, err = svc.EnsureProjectThread(r.Context(), body.ProjectId, userID)
		case "dm":
			threadID, err = svc.EnsureDMThreadByEmail(r.Context(), body.Email, userID)
		default:
			writeJSONError(w, http.StatusBadRequest, "invalid scope")
			return
		}
		if err != nil || threadID <= 0 {
			writeJSONError(w, http.StatusForbidden, "not allowed")
			return
		}
		json.NewEncoder(w).Encode(map[string]int{"id": threadID})
	}
}

func HandleGetChatThreads(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		scope := r.URL.Query().Get("scope")
		switch scope {
		case "org":
			orgID, _ := strconv.Atoi(r.URL.Query().Get("org_id"))
			if orgID <= 0 {
				writeJSONError(w, http.StatusBadRequest, "org_id is required")
				return
			}
			list, err := svc.GetOrgThreads(r.Context(), orgID, userID)
			if err != nil {
				writeJSONError(w, http.StatusForbidden, "not allowed")
				return
			}
			if list == nil {
				w.Write([]byte("[]"))
				return
			}
			json.NewEncoder(w).Encode(list)
		case "project":
			projectID, _ := strconv.Atoi(r.URL.Query().Get("project_id"))
			if projectID <= 0 {
				writeJSONError(w, http.StatusBadRequest, "project_id is required")
				return
			}
			list, err := svc.GetProjectThreads(r.Context(), projectID, userID)
			if err != nil {
				writeJSONError(w, http.StatusForbidden, "not allowed")
				return
			}
			if list == nil {
				w.Write([]byte("[]"))
				return
			}
			json.NewEncoder(w).Encode(list)
		case "dm":
			list, err := svc.GetDMThreads(r.Context(), userID)
			if err != nil {
				writeJSONError(w, http.StatusInternalServerError, "failed to fetch threads")
				return
			}
			if list == nil {
				w.Write([]byte("[]"))
				return
			}
			json.NewEncoder(w).Encode(list)
		default:
			writeJSONError(w, http.StatusBadRequest, "scope is required")
		}
	}
}

func HandleGetChatMessages(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		threadID, _ := strconv.Atoi(mux.Vars(r)["id"])
		if threadID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid thread id")
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		beforeID, _ := strconv.Atoi(r.URL.Query().Get("before_id"))
		msgs, err := svc.GetChatMessages(r.Context(), threadID, userID, limit, beforeID)
		if err != nil {
			writeJSONError(w, http.StatusForbidden, "not allowed")
			return
		}
		if msgs == nil {
			w.Write([]byte("[]"))
			return
		}
		json.NewEncoder(w).Encode(msgs)
	}
}

func HandleUpdateChatMessage(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		messageID, _ := strconv.Atoi(mux.Vars(r)["id"])
		if messageID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid message id")
			return
		}
		var body struct {
			Body string `json:"body"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Body) == "" {
			writeJSONError(w, http.StatusBadRequest, "body is required")
			return
		}
		okUpd, err := svc.UpdateChatMessage(r.Context(), messageID, userID, strings.TrimSpace(body.Body))
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to update")
			return
		}
		if !okUpd {
			writeJSONError(w, http.StatusForbidden, "not allowed")
			return
		}
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}
}

func HandleDeleteChatMessage(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		messageID, _ := strconv.Atoi(mux.Vars(r)["id"])
		if messageID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid message id")
			return
		}
		okDel, err := svc.DeleteChatMessage(r.Context(), messageID, userID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to delete")
			return
		}
		if !okDel {
			writeJSONError(w, http.StatusForbidden, "not allowed")
			return
		}
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}
}

func HandleAddChatMessage(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		threadID, _ := strconv.Atoi(mux.Vars(r)["id"])
		if threadID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid thread id")
			return
		}
		var body struct {
			Body string `json:"body"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Body) == "" {
			writeJSONError(w, http.StatusBadRequest, "body is required")
			return
		}
		id, err := svc.AddChatMessage(r.Context(), threadID, userID, strings.TrimSpace(body.Body))
		if err != nil {
			writeJSONError(w, http.StatusForbidden, "not allowed")
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int{"id": id})
	}
}

func HandleMarkChatThreadRead(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		threadID, _ := strconv.Atoi(mux.Vars(r)["id"])
		if threadID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid thread id")
			return
		}
		if err := svc.MarkThreadRead(r.Context(), threadID, userID); err != nil {
			writeJSONError(w, http.StatusForbidden, "not allowed")
			return
		}
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}
}

func HandleSetTypingState(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		threadID, _ := strconv.Atoi(mux.Vars(r)["id"])
		if threadID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid thread id")
			return
		}
		var body struct {
			IsTyping bool `json:"is_typing"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid body")
			return
		}
		if err := svc.UpsertTypingState(r.Context(), threadID, userID, body.IsTyping); err != nil {
			writeJSONError(w, http.StatusForbidden, "not allowed")
			return
		}
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}
}

func HandleGetTypingState(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		userID, ok := GetUserIDFromContext(r.Context())
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		threadID, _ := strconv.Atoi(mux.Vars(r)["id"])
		if threadID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid thread id")
			return
		}
		users, err := svc.GetTypingUsers(r.Context(), threadID, userID)
		if err != nil {
			writeJSONError(w, http.StatusForbidden, "not allowed")
			return
		}
		if users == nil {
			users = []string{}
		}
		json.NewEncoder(w).Encode(users)
	}
}
