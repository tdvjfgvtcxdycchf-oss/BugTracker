package sql

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

type User struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Task struct {
	Id          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	OwnerId     int    `json:"owner_id"`
}

type Bug struct {
	Id             int    `json:"id"`
	TaskId         int    `json:"task_id"`
	Severity       string `json:"severity"`
	Priority       string `json:"priority"`
	OS             string `json:"os"`
	Status         string `json:"status"`
	VersionProduct string `json:"version_product"`
	Description    string `json:"description"`

	// Кто создал
	CreatedBy   int       `json:"created_by"`
	CreatedTime time.Time `json:"created_time"`

	// Кто закреплён
	AssignedTo   *int       `json:"assigned_to,omitempty"`
	AssignedTime *time.Time `json:"assigned_time,omitempty"`

	// Кто сдал
	PassedBy   *int       `json:"passed_by,omitempty"`
	PassedTime *time.Time `json:"passed_time,omitempty"`

	// Кто принял
	AcceptedBy   *int       `json:"accepted_by,omitempty"`
	AcceptedTime *time.Time `json:"accepted_time,omitempty"`

	PlaybackDescription string `json:"playback_description"`
	ExpectedResult      string `json:"expected_result"`
	ActualResult        string `json:"actual_result"`
}

func CreateConnection(ctx context.Context) (*pgx.Conn, error) {
	slog.Debug("loading .env")

	if err := godotenv.Load(); err != nil {
		slog.Warn("failed to load .env, using environment", "error", err)
	} else {
		slog.Debug("loaded .env")
	}

	slog.Info("connecting to postgres")

	conn, err := pgx.Connect(ctx, os.Getenv("POSTGRES_URL"))

	if err != nil {
		slog.Error("postgres connect failed", "error", err)
		return nil, err
	}

	return conn, nil
}

func CreateUser(ctx context.Context, conn *pgx.Conn, user User) (int, error) {
	sqlQuery := `
		INSERT INTO "User" (email, password)
		VALUES ($1, $2)
		RETURNING id_pk;
	`
	var id int
	err := conn.QueryRow(ctx, sqlQuery,
		user.Email,
		user.Password,
	).Scan(&id)

	if err != nil {
		slog.Error("create user failed", "error", err, "email", user.Email)
		return 0, err
	}
	slog.Info("user created", "id", id, "email", user.Email)

	return id, nil
}

func GetByEmail(ctx context.Context, conn *pgx.Conn, requestUser User) (*User, error) {
	sqlQuery := `
		SELECT id_pk, email, password FROM "User" WHERE email = $1 
	`

	var user User
	err := conn.QueryRow(ctx, sqlQuery, requestUser.Email).Scan(&user.Id, &user.Email, &user.Password)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		slog.Error("database error", "error", err, "email", user.Email)
		return nil, err
	}
	slog.Info("user found", "id", user.Id, "email", user.Email)

	return &user, nil
}

func GetAllTasks(ctx context.Context, conn *pgx.Conn) ([]Task, error) {
	sqlQuery := `
		SELECT id_pk, title, description, owner_id_fk FROM Task 
	`

	rows, err := conn.Query(ctx, sqlQuery)
	if err != nil {
		slog.Error("database error", "error", err)
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		err := rows.Scan(&t.Id, &t.Title, &t.Description, &t.OwnerId)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	if err = rows.Err(); err != nil {
		slog.Error("failed to get tasks", "error", err)
		return nil, err
	}
	slog.Info("tasks successfully received")

	return tasks, nil
}
