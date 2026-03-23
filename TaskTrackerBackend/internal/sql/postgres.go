package sql

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

func CreateConnection(ctx context.Context) (*pgxpool.Pool, error) {
	slog.Debug("loading .env")

	if err := godotenv.Load(); err != nil {
		slog.Warn("failed to load .env, using environment", "error", err)
	} else {
		slog.Debug("loaded .env")
	}

	slog.Info("connecting to postgres pool")

	pool, err := pgxpool.New(ctx, os.Getenv("POSTGRES_URL"))

	if err != nil {
		slog.Error("postgres pool connect failed", "error", err)
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		slog.Error("postgres ping failed", "error", err)
		return nil, err
	}

	return pool, nil
}

func CreateUser(ctx context.Context, conn *pgxpool.Pool, user User) (int, error) {
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

func GetByEmail(ctx context.Context, conn *pgxpool.Pool, requestUser User) (*User, error) {
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

func GetOtherUsersEmails(ctx context.Context, conn *pgxpool.Pool, excludeId int) ([]string, error) {
	sqlQuery := `
        SELECT email 
        FROM "User"
        WHERE id_pk != $1;
    `

	rows, err := conn.Query(ctx, sqlQuery, excludeId)
	if err != nil {
		slog.Error("database error while fetching emails", "error", err)
		return nil, err
	}
	defer rows.Close()

	var emails []string
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			return nil, err
		}
		emails = append(emails, email)
	}

	if err := rows.Err(); err != nil {
		slog.Error("failed to get emails", "error", err)
		return nil, err
	}

	return emails, nil
}

func GetAllTasks(ctx context.Context, conn *pgxpool.Pool) ([]Task, error) {
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

func CreateTask(ctx context.Context, conn *pgxpool.Pool, task Task) error {
	sqlQuery := `
		INSERT INTO Task (title, description, owner_id_fk)
		VALUES ($1, $2, $3)
		RETURNING id_pk;
	`

	id, err := conn.Exec(ctx, sqlQuery,
		task.Title,
		task.Description,
		task.OwnerId,
	)
	if err != nil {
		slog.Error("database error", "error", err)
		return err
	}
	slog.Info("task successfully created", "task id", id)

	return nil
}

func GetBugsByTaskId(ctx context.Context, conn *pgxpool.Pool, taskId int) ([]Bug, error) {
	sqlQuery := `
		SELECT id_pk, task_id_fk, severity, priority,
		os, status, version_product, description,
		created_by_fk, created_time, assigned_to_fk, assigned_time,
		passed_by_fk, passed_time, accepted_by_fk, accepted_time,
		playback_description, expected_result, actual_result FROM Bug
		WHERE task_id_fk = $1
	`

	rows, err := conn.Query(ctx, sqlQuery, taskId)
	if err != nil {
		slog.Error("database error", "error", err)
		return nil, err
	}
	defer rows.Close()

	bugs := make([]Bug, 0)
	for rows.Next() {
		var b Bug
		err := rows.Scan(&b.Id, &b.TaskId, &b.Severity, &b.Priority,
			&b.OS, &b.Status, &b.VersionProduct, &b.Description,
			&b.CreatedBy, &b.CreatedTime, &b.AssignedTo, &b.AssignedTime,
			&b.PassedBy, &b.PassedTime, &b.AcceptedBy, &b.AcceptedTime,
			&b.PlaybackDescription, &b.ExpectedResult, &b.ActualResult,
		)
		if err != nil {
			return nil, err
		}
		bugs = append(bugs, b)
	}

	if err = rows.Err(); err != nil {
		slog.Error("failed to get bugs", "error", err)
		return nil, err
	}
	slog.Info("bugs successfully received")

	return bugs, nil
}

func CreateBug(ctx context.Context, conn *pgxpool.Pool, bug Bug) error {
	sqlQuery := `
        INSERT INTO Bug (task_id_fk, severity, priority, os, 
            status, version_product, description, created_by_fk, 
            playback_description, expected_result, actual_result
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id_pk;
    `

	id, err := conn.Exec(ctx, sqlQuery,
		bug.TaskId,
		bug.Severity,
		bug.Priority,
		bug.OS,
		bug.Status,
		bug.VersionProduct,
		bug.Description,
		bug.CreatedBy,
		bug.PlaybackDescription,
		bug.ExpectedResult,
		bug.ActualResult,
	)

	if err != nil {
		slog.Error("database error while creating bug", "error", err, "task_id", bug.TaskId)
		return err
	}

	slog.Info("bug successfully created", "bug id", id)
	return nil
}

func ChangeBug(ctx context.Context, conn *pgxpool.Pool, bug Bug, assignedEmail string) error {
	if assignedEmail != "" && assignedEmail != "—" {
		var newID int
		err := conn.QueryRow(ctx, `SELECT id_pk FROM "User" WHERE email = $1`, assignedEmail).Scan(&newID)
		if err == nil {
			bug.AssignedTo = &newID
			now := time.Now()
			bug.AssignedTime = &now
		} else if err != pgx.ErrNoRows {
			slog.Error("error finding user by email", "email", assignedEmail, "error", err)
		}
	}

	sqlQuery := `
        UPDATE Bug 
        SET severity = $1, priority = $2, os = $3, status = $4, 
            version_product = $5, description = $6, playback_description = $7, expected_result = $8, 
            actual_result = $9, assigned_to_fk = $10, assigned_time = $11, passed_by_fk = $12, 
            passed_time = $13, accepted_by_fk = $14, accepted_time = $15
        WHERE id_pk = $16;
    `

	_, err := conn.Exec(ctx, sqlQuery,
		bug.Severity, bug.Priority, bug.OS, bug.Status,
		bug.VersionProduct, bug.Description, bug.PlaybackDescription,
		bug.ExpectedResult, bug.ActualResult,
		bug.AssignedTo, bug.AssignedTime,
		bug.PassedBy, bug.PassedTime, bug.AcceptedBy, bug.AcceptedTime,
		bug.Id,
	)

	if err != nil {
		slog.Error("UPDATE Bug failed", "error", err, "bug_id", bug.Id)
		return err
	}

	slog.Info("bug updated successfully", "id", bug.Id)

	return err
}

func DeleteTask(ctx context.Context, conn *pgxpool.Pool, taskID int, ownerID int) (bool, error) {
	var deletedID int
	err := conn.QueryRow(
		ctx,
		`DELETE FROM Task
		 WHERE id_pk = $1 AND owner_id_fk = $2
		 RETURNING id_pk`,
		taskID,
		ownerID,
	).Scan(&deletedID)

	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func DeleteBug(ctx context.Context, conn *pgxpool.Pool, bugID int, creatorID int) (bool, error) {
	var deletedID int
	err := conn.QueryRow(
		ctx,
		`DELETE FROM Bug
		 WHERE id_pk = $1 AND created_by_fk = $2
		 RETURNING id_pk`,
		bugID,
		creatorID,
	).Scan(&deletedID)

	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}
