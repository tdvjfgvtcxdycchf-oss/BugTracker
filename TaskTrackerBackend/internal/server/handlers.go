package server

import (
	"bug_tracker/internal/auth"
	"bug_tracker/internal/service"
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
)

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": msg,
	})
}

func HandleCreateUser(svc *service.TaskTrackerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var user sql.User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			slog.Warn("create user: invalid json", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			writeJSONError(w, http.StatusInternalServerError, "invalid json")
			return
		}

		if user.Email == "" || !strings.Contains(user.Email, "@") {
			slog.Warn("create user: invalid email", "error", "missing @ or string is empty")
			w.WriteHeader(http.StatusBadRequest)
			writeJSONError(w, http.StatusInternalServerError, "invalid email")
			return
		}

		if user.Password == "" {
			slog.Warn("create user: empty password", "error", "string is empty")
			w.WriteHeader(http.StatusBadRequest)
			writeJSONError(w, http.StatusInternalServerError, "empty password")
			return
		}

		createdUserId, role, err := svc.Register(r.Context(), user)
		if err != nil {
			if err.Error() == "user already exists" {
				w.WriteHeader(http.StatusConflict)
				writeJSONError(w, http.StatusInternalServerError, "user already exist")
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
		token, tokenErr := auth.GenerateJWT(jwtSecret, createdUserId, user.Email, 24*time.Hour)
		if tokenErr != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to generate jwt")
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"id": createdUserId, "token": token, "role": role})
	}
}

func HandleGetIdUser(svc *service.TaskTrackerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var user sql.User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			slog.Warn("login user: invalid json", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			writeJSONError(w, http.StatusInternalServerError, "invalid json")
			return
		}

		if user.Email == "" || !strings.Contains(user.Email, "@") {
			slog.Warn("login user: invalid email", "error", "missing @ or string is empty")
			w.WriteHeader(http.StatusBadRequest)
			writeJSONError(w, http.StatusInternalServerError, "invalid email")
			return
		}

		if user.Password == "" {
			slog.Warn("login user: empty password", "error", "string is empty")
			w.WriteHeader(http.StatusBadRequest)
			writeJSONError(w, http.StatusInternalServerError, "iempty password")
			return
		}

		loginedUserId, role, err := svc.Login(r.Context(), user)
		if err != nil {
			if err.Error() == "user not exist" {
				w.WriteHeader(http.StatusUnauthorized)
				writeJSONError(w, http.StatusInternalServerError, "user not exist")
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
		token, tokenErr := auth.GenerateJWT(jwtSecret, loginedUserId, user.Email, 24*time.Hour)
		if tokenErr != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to generate jwt")
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{"id": loginedUserId, "token": token, "role": role})
	}
}

func HandleGetOtherEmails(svc *service.TaskTrackerService) http.HandlerFunc {
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

func HandleGetAllTasks(svc *service.TaskTrackerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		tasks, err := svc.GetAllTasks(r.Context())
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "could not fetch tasks")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tasks)
	}
}

func HandleCreateTask(svc *service.TaskTrackerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var task sql.Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			slog.Warn("create task: invalid json", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			writeJSONError(w, http.StatusInternalServerError, "invalid json")
			return
		}

		if task.Title == "" {
			slog.Warn("task title: empty tittle", "error", "string is empty")
			w.WriteHeader(http.StatusBadRequest)
			writeJSONError(w, http.StatusInternalServerError, "empty title")
			return
		}

		if task.Description == "" {
			slog.Warn("task description: empty description", "error", "string is empty")
			w.WriteHeader(http.StatusBadRequest)
			writeJSONError(w, http.StatusInternalServerError, "empty description")
			return
		}

		if err := svc.CreateTask(r.Context(), task); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "could not fetch tasks")
			return
		}

		tasks, err := svc.GetAllTasks(r.Context())
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "could not fetch tasks")
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(tasks)
	}
}

func HandleFuncGetAllBugs(svc *service.TaskTrackerService) http.HandlerFunc {
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

		bugs, err := svc.GetAllBugs(r.Context(), id)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "could not fetch bugs")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(bugs)
	}
}

func HandleCreateBug(svc *service.TaskTrackerService) http.HandlerFunc {
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

		var bug sql.Bug
		if err := json.NewDecoder(r.Body).Decode(&bug); err != nil {
			slog.Warn("create task: invalid json", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			writeJSONError(w, http.StatusInternalServerError, "invalid json")
			return
		}

		bug.TaskId = id

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

func HandleUpdateBug(svc *service.TaskTrackerService) http.HandlerFunc {
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

func HandleDeleteTask(svc *service.TaskTrackerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		vars := mux.Vars(r)
		idStr := vars["id"]
		taskID, err := strconv.Atoi(idStr)
		if err != nil || taskID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid task id")
			return
		}

		var input struct {
			OwnerId int `json:"owner_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if input.OwnerId <= 0 {
			writeJSONError(w, http.StatusBadRequest, "owner_id is required")
			return
		}

		deleted, err := svc.DeleteTask(r.Context(), taskID, input.OwnerId)
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

func HandleDeleteBug(svc *service.TaskTrackerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		vars := mux.Vars(r)
		idStr := vars["id"]
		bugID, err := strconv.Atoi(idStr)
		if err != nil || bugID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid bug id")
			return
		}

		var input struct {
			CreatedBy int `json:"created_by"`
		}

		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if input.CreatedBy <= 0 {
			writeJSONError(w, http.StatusBadRequest, "created_by is required")
			return
		}

		deleted, err := svc.DeleteBug(r.Context(), bugID, input.CreatedBy)
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

func HandleUploadBugPhoto(svc *service.TaskTrackerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr := vars["id"]
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid bug id")
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

func HandleAddComment(svc *service.TaskTrackerService) http.HandlerFunc {
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

func HandleGetComments(svc *service.TaskTrackerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		bugID, err := strconv.Atoi(vars["id"])
		if err != nil || bugID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid bug id")
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

func HandleGetAuditLog(svc *service.TaskTrackerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		bugID, err := strconv.Atoi(vars["id"])
		if err != nil || bugID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid bug id")
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

func HandleAddRelation(svc *service.TaskTrackerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		bugID, err := strconv.Atoi(vars["id"])
		if err != nil || bugID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid bug id")
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

func HandleGetRelations(svc *service.TaskTrackerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		bugID, err := strconv.Atoi(vars["id"])
		if err != nil || bugID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid bug id")
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

func HandleDeleteRelation(svc *service.TaskTrackerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		relID, err := strconv.Atoi(vars["relId"])
		if err != nil || relID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid relation id")
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

func HandleSetTags(svc *service.TaskTrackerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		bugID, err := strconv.Atoi(vars["id"])
		if err != nil || bugID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid bug id")
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

func HandleGetTags(svc *service.TaskTrackerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		bugID, err := strconv.Atoi(vars["id"])
		if err != nil || bugID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid bug id")
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

func HandleGetTemplates(svc *service.TaskTrackerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
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

func HandleCreateTemplate(svc *service.TaskTrackerService) http.HandlerFunc {
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

func HandleDeleteTemplate(svc *service.TaskTrackerService) http.HandlerFunc {
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

func HandleGetStats(svc *service.TaskTrackerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
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

func HandleGetBugPhoto(svc *service.TaskTrackerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr := vars["id"]
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			http.NotFound(w, r)
			return
		}

		data, err := svc.GetBugPhoto(r.Context(), id)
		if err != nil || len(data) == 0 {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", http.DetectContentType(data))
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}
