package server

import (
	"bug_tracker/internal/service"
	"bug_tracker/internal/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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

		createdUserId, err := svc.Register(r.Context(), user)
		if err != nil {
			if err.Error() == "user already exists" {
				w.WriteHeader(http.StatusConflict)
				writeJSONError(w, http.StatusInternalServerError, "user already exist")
				return
			}
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int{"id": createdUserId})
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

		loginedUserId, err := svc.Login(r.Context(), user)
		if err != nil {
			if err.Error() == "user not exist" {
				w.WriteHeader(http.StatusUnauthorized)
				writeJSONError(w, http.StatusInternalServerError, "user not exist")
				return
			}
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]int{"id": loginedUserId})
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

func HandleUploadBugPhoto(uploadsDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		if err := r.ParseMultipartForm(10 << 20); err != nil {
			writeJSONError(w, http.StatusBadRequest, "file too large")
			return
		}

		file, header, err := r.FormFile("photo")
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "no photo provided")
			return
		}
		defer file.Close()

		// Удаляем старое фото для этого бага
		old, _ := filepath.Glob(filepath.Join(uploadsDir, "bug_"+id+".*"))
		for _, f := range old {
			os.Remove(f)
		}

		ext := filepath.Ext(header.Filename)
		if ext == "" {
			ext = ".jpg"
		}

		dstPath := filepath.Join(uploadsDir, "bug_"+id+ext)
		dst, err := os.Create(dstPath)
		if err != nil {
			slog.Error("failed to create photo file", "error", err)
			writeJSONError(w, http.StatusInternalServerError, "failed to save file")
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			slog.Error("failed to write photo file", "error", err)
			writeJSONError(w, http.StatusInternalServerError, "failed to save file")
			return
		}

		slog.Info("bug photo uploaded", "bug_id", id)
		w.WriteHeader(http.StatusCreated)
	}
}

func HandleGetBugPhoto(uploadsDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		matches, err := filepath.Glob(filepath.Join(uploadsDir, "bug_"+id+".*"))
		if err != nil || len(matches) == 0 {
			http.NotFound(w, r)
			return
		}

		http.ServeFile(w, r, matches[0])
	}
}
