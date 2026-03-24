package tests

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"bug_tracker/internal/sql"

	"github.com/joho/godotenv"
	"github.com/jackc/pgx/v5/pgxpool"
)

func mustPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("POSTGRES_URL")
	if dsn == "" {
		// Загружаем локальный .env на случай запуска тестов в окружении, где переменные не выставлены.
		// В рабочей директории пакета тестов обычно TaskTrackerBackend/tests,
		// поэтому пробуем и текущий каталог, и папку на уровень выше.
		_ = godotenv.Load(".env")
		_ = godotenv.Load("../.env")
		dsn = os.Getenv("POSTGRES_URL")
	}
	if dsn == "" {
		t.Skip("POSTGRES_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("db ping: %v", err)
	}

	return pool
}

func resetSchema(ctx context.Context, pool *pgxpool.Pool) error {

	_, _ = pool.Exec(ctx, `DROP TABLE IF EXISTS Bug CASCADE;`)
	_, _ = pool.Exec(ctx, `DROP TABLE IF EXISTS Task CASCADE;`)
	_, _ = pool.Exec(ctx, `DROP TABLE IF EXISTS "User" CASCADE;`)

	createUser := `
CREATE TABLE "User" (
    id_pk SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL
);`
	if _, err := pool.Exec(ctx, createUser); err != nil {
		return fmt.Errorf("create User: %w", err)
	}

	createTask := `
CREATE TABLE Task (
    id_pk SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id_fk INTEGER NOT NULL,

    CONSTRAINT fk_task_owner FOREIGN KEY (owner_id_fk) REFERENCES "User"(id_pk) ON DELETE CASCADE
);`
	if _, err := pool.Exec(ctx, createTask); err != nil {
		return fmt.Errorf("create Task: %w", err)
	}

	createBug := `
CREATE TABLE Bug (
    id_pk SERIAL PRIMARY KEY,
    task_id_fk INTEGER NOT NULL,
    severity VARCHAR(50),
    priority VARCHAR(50),
    os VARCHAR(200),
    status VARCHAR(50),
    version_product VARCHAR(50),
    description TEXT,

    -- Кто создал
    created_by_fk INTEGER NOT NULL,
    created_time DATE DEFAULT CURRENT_DATE,

    -- За кем закреплён
    assigned_to_fk INTEGER,
    assigned_time DATE,

    -- Кто сдал
    passed_by_fk INTEGER,
    passed_time DATE,

    -- Кто принял
    accepted_by_fk INTEGER,
    accepted_time DATE,

    playback_description TEXT,
    expected_result TEXT,
    actual_result TEXT,

    CONSTRAINT fk_bug_task FOREIGN KEY (task_id_fk) REFERENCES Task(id_pk) ON DELETE CASCADE,
    CONSTRAINT fk_bug_creator FOREIGN KEY (created_by_fk) REFERENCES "User"(id_pk),
    CONSTRAINT fk_bug_assignee FOREIGN KEY (assigned_to_fk) REFERENCES "User"(id_pk),
    CONSTRAINT fk_bug_tester FOREIGN KEY (passed_by_fk) REFERENCES "User"(id_pk),
    CONSTRAINT fk_bug_approver FOREIGN KEY (accepted_by_fk) REFERENCES "User"(id_pk)
);`
	if _, err := pool.Exec(ctx, createBug); err != nil {
		return fmt.Errorf("create Bug: %w", err)
	}

	return nil
}

func createUser(ctx context.Context, pool *pgxpool.Pool, email, password string) (int, error) {
	var id int
	if err := pool.QueryRow(ctx,
		`INSERT INTO "User"(email, password) VALUES ($1, $2) RETURNING id_pk`,
		email, password,
	).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func createTask(ctx context.Context, pool *pgxpool.Pool, title string, ownerID int) (int, error) {
	var id int
	if err := pool.QueryRow(ctx,
		`INSERT INTO Task(title, description, owner_id_fk) VALUES ($1, $2, $3) RETURNING id_pk`,
		title, "desc", ownerID,
	).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func createBug(ctx context.Context, pool *pgxpool.Pool, taskID int, createdByID int) (int, error) {
	var id int
	if err := pool.QueryRow(ctx,
		`INSERT INTO Bug(task_id_fk, severity, priority, os, status, version_product, description, created_by_fk, playback_description, expected_result, actual_result)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		 RETURNING id_pk`,
		taskID, "Low", "Low", "Win", "Open", "1.0.0", "bug desc", createdByID, "steps", "expected", "actual",
	).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func getBugAssigned(ctx context.Context, pool *pgxpool.Pool, bugID int) (assignedTo *int, assignedTime *time.Time, err error) {
	var assignedToNullable *int
	var assignedTimeNullable *time.Time

	err = pool.QueryRow(ctx,
		`SELECT assigned_to_fk, assigned_time FROM Bug WHERE id_pk = $1`,
		bugID,
	).Scan(&assignedToNullable, &assignedTimeNullable)

	return assignedToNullable, assignedTimeNullable, err
}

func buildBugForUpdate(bugID, taskID, createdByID int) sql.Bug {
	return sql.Bug{
		Id:                  bugID,
		TaskId:              taskID,
		Severity:            "Low",
		Priority:            "Low",
		OS:                  "Win",
		Status:              "Open",
		VersionProduct:      "1.0.0",
		Description:         "bug desc",
		CreatedBy:           createdByID,
		PlaybackDescription: "steps",
		ExpectedResult:      "expected",
		ActualResult:        "actual",
	}
}
