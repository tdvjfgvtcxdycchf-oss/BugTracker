package server

import (
	"bug_tracker/internal/auth"
	"bug_tracker/internal/sql"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ctxKey string

const userIDCtxKey ctxKey = "userID"

func GetUserIDFromContext(ctx context.Context) (int, bool) {
	v := ctx.Value(userIDCtxKey)
	if v == nil {
		return 0, false
	}
	id, ok := v.(int)
	return id, ok
}

func jwtAuthMiddleware(jwtSecret string, conn *pgxpool.Pool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if jwtSecret == "" {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "JWT_SECRET is not set"})
			return
		}

		w.Header().Set("Content-Type", "application/json")

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "missing Authorization header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid Authorization header"})
			return
		}

		claims, err := auth.VerifyJWT(jwtSecret, parts[1])
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid token"})
			return
		}

		userID, err := strconv.Atoi(claims.Sub)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid token sub"})
			return
		}

		ver, err := sql.GetUserJWTVersion(r.Context(), conn, userID)
		if err != nil || ver != claims.Ver {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "token revoked"})
			return
		}

		ctx := context.WithValue(r.Context(), userIDCtxKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
