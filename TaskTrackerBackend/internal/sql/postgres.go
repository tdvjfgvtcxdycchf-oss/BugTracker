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
	Role     string `json:"role"` // qa | developer | pm | admin
}

type BugComment struct {
	Id        int       `json:"id"`
	BugId     int       `json:"bug_id"`
	UserId    int       `json:"user_id"`
	UserEmail string    `json:"user_email"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

type AuditEntry struct {
	Id        int       `json:"id"`
	BugId     int       `json:"bug_id"`
	UserId    int       `json:"user_id"`
	UserEmail string    `json:"user_email"`
	Field     string    `json:"field"`
	OldValue  string    `json:"old_value"`
	NewValue  string    `json:"new_value"`
	ChangedAt time.Time `json:"changed_at"`
}

type BugRelation struct {
	Id           int    `json:"id"`
	FromBugId    int    `json:"from_bug_id"`
	ToBugId      int    `json:"to_bug_id"`
	RelationType string `json:"relation_type"` // duplicate | blocks | related
}

type BugTag struct {
	BugId int    `json:"bug_id"`
	Tag   string `json:"tag"`
}

type BugTemplate struct {
	Id                  int    `json:"id"`
	Name                string `json:"name"`
	Severity            string `json:"severity"`
	Priority            string `json:"priority"`
	OS                  string `json:"os"`
	VersionProduct      string `json:"version_product"`
	Description         string `json:"description"`
	PlaybackDescription string `json:"playback_description"`
	ExpectedResult      string `json:"expected_result"`
	ActualResult        string `json:"actual_result"`
	CreatedBy           int    `json:"created_by"`
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

	CreatedBy   int       `json:"created_by"`
	CreatedTime time.Time `json:"created_time"`

	AssignedTo   *int       `json:"assigned_to,omitempty"`
	AssignedTime *time.Time `json:"assigned_time,omitempty"`

	PassedBy   *int       `json:"passed_by,omitempty"`
	PassedTime *time.Time `json:"passed_time,omitempty"`

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
	role := user.Role
	if role == "" {
		role = "qa"
	}
	sqlQuery := `
		INSERT INTO "User" (email, password, role)
		VALUES ($1, $2, $3)
		RETURNING id_pk;
	`
	var id int
	err := conn.QueryRow(ctx, sqlQuery, user.Email, user.Password, role).Scan(&id)
	if err != nil {
		slog.Error("create user failed", "error", err, "email", user.Email)
		return 0, err
	}
	slog.Info("user created", "id", id, "email", user.Email, "role", role)
	return id, nil
}

func GetByEmail(ctx context.Context, conn *pgxpool.Pool, requestUser User) (*User, error) {
	sqlQuery := `
		SELECT id_pk, email, password, role FROM "User" WHERE email = $1
	`

	var user User
	err := conn.QueryRow(ctx, sqlQuery, requestUser.Email).Scan(&user.Id, &user.Email, &user.Password, &user.Role)
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

func SaveBugPhoto(ctx context.Context, conn *pgxpool.Pool, bugID int, data []byte) error {
	_, err := conn.Exec(ctx, `UPDATE Bug SET photo = $1 WHERE id_pk = $2`, data, bugID)
	if err != nil {
		slog.Error("failed to save bug photo", "bug_id", bugID, "error", err)
	}
	return err
}

func GetBugPhoto(ctx context.Context, conn *pgxpool.Pool, bugID int) ([]byte, error) {
	var data []byte
	err := conn.QueryRow(ctx, `SELECT photo FROM Bug WHERE id_pk = $1`, bugID).Scan(&data)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		slog.Error("failed to get bug photo", "bug_id", bugID, "error", err)
		return nil, err
	}
	return data, nil
}

// ── User role ────────────────────────────────────────────────────────────────

func GetUserRole(ctx context.Context, conn *pgxpool.Pool, userID int) (string, error) {
	var role string
	err := conn.QueryRow(ctx, `SELECT role FROM "User" WHERE id_pk = $1`, userID).Scan(&role)
	if err != nil {
		return "", err
	}
	return role, nil
}

func CreateUserWithRole(ctx context.Context, conn *pgxpool.Pool, user User) (int, error) {
	sqlQuery := `
		INSERT INTO "User" (email, password, role)
		VALUES ($1, $2, $3)
		RETURNING id_pk;
	`
	var id int
	err := conn.QueryRow(ctx, sqlQuery, user.Email, user.Password, user.Role).Scan(&id)
	if err != nil {
		slog.Error("create user with role failed", "error", err, "email", user.Email)
		return 0, err
	}
	slog.Info("user created", "id", id, "email", user.Email, "role", user.Role)
	return id, nil
}

func GetUserByID(ctx context.Context, conn *pgxpool.Pool, userID int) (*User, error) {
	var u User
	err := conn.QueryRow(ctx,
		`SELECT id_pk, email, role FROM "User" WHERE id_pk = $1`, userID,
	).Scan(&u.Id, &u.Email, &u.Role)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// ── Comments ─────────────────────────────────────────────────────────────────

func AddBugComment(ctx context.Context, conn *pgxpool.Pool, bugID, userID int, body string) (*BugComment, error) {
	sqlQuery := `
		INSERT INTO bug_comment (bug_id_fk, user_id_fk, body)
		VALUES ($1, $2, $3)
		RETURNING id_pk, created_at;
	`
	var c BugComment
	c.BugId = bugID
	c.UserId = userID
	c.Body = body
	err := conn.QueryRow(ctx, sqlQuery, bugID, userID, body).Scan(&c.Id, &c.CreatedAt)
	if err != nil {
		slog.Error("add bug comment failed", "error", err)
		return nil, err
	}
	return &c, nil
}

func GetBugComments(ctx context.Context, conn *pgxpool.Pool, bugID int) ([]BugComment, error) {
	rows, err := conn.Query(ctx, `
		SELECT c.id_pk, c.bug_id_fk, c.user_id_fk, u.email, c.body, c.created_at
		FROM bug_comment c
		JOIN "User" u ON u.id_pk = c.user_id_fk
		WHERE c.bug_id_fk = $1
		ORDER BY c.created_at
	`, bugID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []BugComment
	for rows.Next() {
		var c BugComment
		if err := rows.Scan(&c.Id, &c.BugId, &c.UserId, &c.UserEmail, &c.Body, &c.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

// ── Audit log ─────────────────────────────────────────────────────────────────

func AddAuditEntry(ctx context.Context, conn *pgxpool.Pool, bugID, userID int, field, oldVal, newVal string) error {
	_, err := conn.Exec(ctx, `
		INSERT INTO audit_log (bug_id_fk, user_id_fk, field, old_value, new_value)
		VALUES ($1, $2, $3, $4, $5)
	`, bugID, userID, field, oldVal, newVal)
	if err != nil {
		slog.Error("add audit entry failed", "error", err)
	}
	return err
}

func GetAuditLog(ctx context.Context, conn *pgxpool.Pool, bugID int) ([]AuditEntry, error) {
	rows, err := conn.Query(ctx, `
		SELECT a.id_pk, a.bug_id_fk, a.user_id_fk, u.email, a.field,
		       COALESCE(a.old_value,''), COALESCE(a.new_value,''), a.changed_at
		FROM audit_log a
		JOIN "User" u ON u.id_pk = a.user_id_fk
		WHERE a.bug_id_fk = $1
		ORDER BY a.changed_at
	`, bugID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []AuditEntry
	for rows.Next() {
		var e AuditEntry
		if err := rows.Scan(&e.Id, &e.BugId, &e.UserId, &e.UserEmail, &e.Field, &e.OldValue, &e.NewValue, &e.ChangedAt); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, rows.Err()
}

// ── Relations ────────────────────────────────────────────────────────────────

func AddBugRelation(ctx context.Context, conn *pgxpool.Pool, fromID, toID int, relType string) error {
	_, err := conn.Exec(ctx, `
		INSERT INTO bug_relation (from_bug_id_fk, to_bug_id_fk, relation_type)
		VALUES ($1, $2, $3)
		ON CONFLICT (from_bug_id_fk, to_bug_id_fk, relation_type) DO NOTHING
	`, fromID, toID, relType)
	if err != nil {
		slog.Error("add bug relation failed", "error", err)
	}
	return err
}

func DeleteBugRelation(ctx context.Context, conn *pgxpool.Pool, relationID int) error {
	_, err := conn.Exec(ctx, `DELETE FROM bug_relation WHERE id_pk = $1`, relationID)
	return err
}

func GetBugRelations(ctx context.Context, conn *pgxpool.Pool, bugID int) ([]BugRelation, error) {
	rows, err := conn.Query(ctx, `
		SELECT id_pk, from_bug_id_fk, to_bug_id_fk, relation_type
		FROM bug_relation
		WHERE from_bug_id_fk = $1 OR to_bug_id_fk = $1
	`, bugID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []BugRelation
	for rows.Next() {
		var r BugRelation
		if err := rows.Scan(&r.Id, &r.FromBugId, &r.ToBugId, &r.RelationType); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

// ── Tags ─────────────────────────────────────────────────────────────────────

func SetBugTags(ctx context.Context, conn *pgxpool.Pool, bugID int, tags []string) error {
	_, err := conn.Exec(ctx, `DELETE FROM bug_tag WHERE bug_id_fk = $1`, bugID)
	if err != nil {
		return err
	}
	for _, tag := range tags {
		if tag == "" {
			continue
		}
		_, err := conn.Exec(ctx, `
			INSERT INTO bug_tag (bug_id_fk, tag) VALUES ($1, $2)
			ON CONFLICT (bug_id_fk, tag) DO NOTHING
		`, bugID, tag)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetBugTags(ctx context.Context, conn *pgxpool.Pool, bugID int) ([]string, error) {
	rows, err := conn.Query(ctx, `SELECT tag FROM bug_tag WHERE bug_id_fk = $1 ORDER BY tag`, bugID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

// ── Templates ────────────────────────────────────────────────────────────────

func CreateTemplate(ctx context.Context, conn *pgxpool.Pool, t BugTemplate) (int, error) {
	var id int
	err := conn.QueryRow(ctx, `
		INSERT INTO bug_template (name, severity, priority, os, version_product,
			description, playback_description, expected_result, actual_result, created_by_fk)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id_pk
	`, t.Name, t.Severity, t.Priority, t.OS, t.VersionProduct,
		t.Description, t.PlaybackDescription, t.ExpectedResult, t.ActualResult, t.CreatedBy,
	).Scan(&id)
	return id, err
}

func GetTemplates(ctx context.Context, conn *pgxpool.Pool) ([]BugTemplate, error) {
	rows, err := conn.Query(ctx, `
		SELECT id_pk, name, severity, priority, os, version_product,
		       description, playback_description, expected_result, actual_result, created_by_fk
		FROM bug_template ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []BugTemplate
	for rows.Next() {
		var t BugTemplate
		if err := rows.Scan(&t.Id, &t.Name, &t.Severity, &t.Priority, &t.OS, &t.VersionProduct,
			&t.Description, &t.PlaybackDescription, &t.ExpectedResult, &t.ActualResult, &t.CreatedBy,
		); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

func DeleteTemplate(ctx context.Context, conn *pgxpool.Pool, id, userID int) (bool, error) {
	var deleted int
	err := conn.QueryRow(ctx,
		`DELETE FROM bug_template WHERE id_pk = $1 AND created_by_fk = $2 RETURNING id_pk`,
		id, userID,
	).Scan(&deleted)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

// ── Analytics (для PM dashboard) ─────────────────────────────────────────────

type BugStats struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

func GetBugStatsByTask(ctx context.Context, conn *pgxpool.Pool, taskID int) ([]BugStats, error) {
	rows, err := conn.Query(ctx, `
		SELECT status, COUNT(*) FROM Bug WHERE task_id_fk = $1 GROUP BY status
	`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []BugStats
	for rows.Next() {
		var s BugStats
		if err := rows.Scan(&s.Status, &s.Count); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

func GetAllBugStats(ctx context.Context, conn *pgxpool.Pool) ([]BugStats, error) {
	rows, err := conn.Query(ctx, `SELECT status, COUNT(*) FROM Bug GROUP BY status`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []BugStats
	for rows.Next() {
		var s BugStats
		if err := rows.Scan(&s.Status, &s.Count); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, rows.Err()
}
