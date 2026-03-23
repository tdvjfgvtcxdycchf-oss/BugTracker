package server

import (
	"bug_tracker/internal/service"
	"bug_tracker/internal/sql"
	"encoding/json"
	"log/slog"
	"net/http"
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
