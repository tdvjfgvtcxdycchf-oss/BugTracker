package server

import (
	"bug_tracker/internal/service"
	"bug_tracker/internal/sql"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
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
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid json"})
			return
		}

		if user.Email == "" || !strings.Contains(user.Email, "@") {
			slog.Warn("create user: invalid email", "error", "missing @ or string is empty")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid email"})
			return
		}

		if user.Password == "" {
			slog.Warn("create user: empty password", "error", "string is empty")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "empty password"})
			return
		}

		createdUserId, err := svc.Register(ctx, user)
		if err != nil {
			if err.Error() == "user already exists" {
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(map[string]string{"error": "user already exists"})
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
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid json"})
			return
		}

		if user.Email == "" || !strings.Contains(user.Email, "@") {
			slog.Warn("login user: invalid email", "error", "missing @ or string is empty")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid email"})
			return
		}

		if user.Password == "" {
			slog.Warn("login user: empty password", "error", "string is empty")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "empty password"})
			return
		}

		loginedUserId, err := svc.Login(ctx, user)
		if err != nil {
			if err.Error() == "user not exist" {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "user not exist"})
				return
			}
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]int{"id": loginedUserId})
	}
}

func HandlerGetAllTasks(ctx context.Context, svc *service.TaskTrackerService) http.HandlerFunc {
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
