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
	Password int    `json:"password"`
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
