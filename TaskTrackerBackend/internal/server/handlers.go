package server

import (
	"bug_tracker/internal/service"
	"bug_tracker/internal/sql"
	"context"
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

func HandleCreateUser(ctx context.Context, svc *service.TaskTrackerService) http.HandlerFunc {
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

		createdUserId, err := svc.Register(ctx, user)
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

func HandleGetIdUser(ctx context.Context, svc *service.TaskTrackerService) http.HandlerFunc {
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

		loginedUserId, err := svc.Login(ctx, user)
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

func HandleGetAllTasks(ctx context.Context, svc *service.TaskTrackerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		tasks, err := svc.GetAllTasks(ctx)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "could not fetch tasks")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tasks)
	}
}

func HandleCreateTask(ctx context.Context, svc *service.TaskTrackerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var task sql.Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			slog.Warn("login user: invalid json", "error", err)
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

		if err := svc.CreateTask(ctx, task); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "could not fetch tasks")
			return
		}

		tasks, err := svc.GetAllTasks(ctx)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "could not fetch tasks")
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(tasks)
	}
}

func HandleFuncGetAllBugs(ctx context.Context, svc *service.TaskTrackerService) http.HandlerFunc {
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

		bugs, err := svc.GetAllBugs(ctx, id)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "could not fetch bugs")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(bugs)
	}
}
