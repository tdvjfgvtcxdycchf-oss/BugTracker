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

func CanUserAccessProject(ctx context.Context, conn *pgxpool.Pool, userID, projectID int) (bool, error) {
	var ok bool
	err := conn.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM projects p
			WHERE p.id_pk = $1
			  AND (
			    EXISTS (SELECT 1 FROM project_member pm WHERE pm.project_id_fk = p.id_pk AND pm.user_id_fk = $2)
			    OR EXISTS (SELECT 1 FROM org_member om WHERE om.org_id_fk = p.org_id_fk AND om.user_id_fk = $2 AND om.role IN ('owner','admin'))
			  )
		)
	`, projectID, userID).Scan(&ok)
	return ok, err
}

func CanUserAccessTask(ctx context.Context, conn *pgxpool.Pool, userID, taskID int) (bool, error) {
	var projectID int
	err := conn.QueryRow(ctx, `SELECT project_id_fk FROM Task WHERE id_pk = $1`, taskID).Scan(&projectID)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return CanUserAccessProject(ctx, conn, userID, projectID)
}

func CanUserAccessBug(ctx context.Context, conn *pgxpool.Pool, userID, bugID int) (bool, error) {
	var taskID int
	err := conn.QueryRow(ctx, `SELECT task_id_fk FROM Bug WHERE id_pk = $1`, bugID).Scan(&taskID)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return CanUserAccessTask(ctx, conn, userID, taskID)
}

func GetRelationFromBugID(ctx context.Context, conn *pgxpool.Pool, relID int) (int, error) {
	var bugID int
	err := conn.QueryRow(ctx, `SELECT from_bug_id_fk FROM bug_relation WHERE id_pk = $1`, relID).Scan(&bugID)
	if err == pgx.ErrNoRows {
		return 0, nil
	}
	return bugID, err
}

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
	ProjectId   int    `json:"project_id"`
}

type Organization struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Role string `json:"role,omitempty"` // role of current user (when listing)
}

type Project struct {
	Id    int    `json:"id"`
	OrgId int    `json:"org_id"`
	Name  string `json:"name"`
	Role  string `json:"role,omitempty"` // role of current user (when listing)
}

type OrgMember struct {
	UserId int    `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

type ProjectMember struct {
	UserId int    `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
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

func UpdateUserPasswordHash(ctx context.Context, conn *pgxpool.Pool, userID int, passwordHash string) error {
	_, err := conn.Exec(ctx, `UPDATE "User" SET password = $1 WHERE id_pk = $2`, passwordHash, userID)
	return err
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

func GetTasksByProjectForUser(ctx context.Context, conn *pgxpool.Pool, userID, projectID int) ([]Task, error) {
	sqlQuery := `
		SELECT t.id_pk, t.title, t.description, t.owner_id_fk, t.project_id_fk
		FROM Task t
		WHERE t.project_id_fk = $1
		  AND (
		    EXISTS (SELECT 1 FROM project_member pm WHERE pm.project_id_fk = t.project_id_fk AND pm.user_id_fk = $2)
		    OR EXISTS (
		      SELECT 1
		      FROM projects p
		      JOIN org_member om ON om.org_id_fk = p.org_id_fk
		      WHERE p.id_pk = t.project_id_fk AND om.user_id_fk = $2 AND om.role IN ('owner','admin')
		    )
		  )
		ORDER BY t.id_pk;
	`

	rows, err := conn.Query(ctx, sqlQuery, projectID, userID)
	if err != nil {
		slog.Error("database error", "error", err)
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		err := rows.Scan(&t.Id, &t.Title, &t.Description, &t.OwnerId, &t.ProjectId)
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

func CreateTask(ctx context.Context, conn *pgxpool.Pool, task Task) (int, error) {
	sqlQuery := `
		WITH can_write AS (
			SELECT 1
			FROM projects p
			WHERE p.id_pk = $4
			  AND (
			    EXISTS (SELECT 1 FROM project_member pm WHERE pm.project_id_fk = p.id_pk AND pm.user_id_fk = $3)
			    OR EXISTS (SELECT 1 FROM org_member om WHERE om.org_id_fk = p.org_id_fk AND om.user_id_fk = $3 AND om.role IN ('owner','admin'))
			  )
		)
		INSERT INTO Task (title, description, owner_id_fk, project_id_fk)
		SELECT $1, $2, $3, $4
		WHERE EXISTS (SELECT 1 FROM can_write)
		RETURNING id_pk;
	`

	var id int
	err := conn.QueryRow(ctx, sqlQuery, task.Title, task.Description, task.OwnerId, task.ProjectId).Scan(&id)
	if err != nil {
		slog.Error("database error", "error", err)
		return 0, err
	}
	slog.Info("task successfully created", "task id", id)

	return id, nil
}

func GetUserOrgs(ctx context.Context, conn *pgxpool.Pool, userID int) ([]Organization, error) {
	rows, err := conn.Query(ctx, `
		SELECT o.id_pk, o.name, om.role
		FROM organizations o
		JOIN org_member om ON om.org_id_fk = o.id_pk
		WHERE om.user_id_fk = $1
		ORDER BY o.id_pk
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Organization
	for rows.Next() {
		var o Organization
		if err := rows.Scan(&o.Id, &o.Name, &o.Role); err != nil {
			return nil, err
		}
		list = append(list, o)
	}
	return list, rows.Err()
}

func CreateOrg(ctx context.Context, conn *pgxpool.Pool, name string, ownerUserID int) (int, error) {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var orgID int
	if err := tx.QueryRow(ctx, `INSERT INTO organizations (name) VALUES ($1) RETURNING id_pk`, name).Scan(&orgID); err != nil {
		return 0, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO org_member (org_id_fk, user_id_fk, role) VALUES ($1,$2,'owner')`, orgID, ownerUserID)
	if err != nil {
		return 0, err
	}

	// Create a default project inside org so UI has something to select.
	var projectID int
	if err := tx.QueryRow(ctx, `INSERT INTO projects (org_id_fk, name) VALUES ($1,'General') RETURNING id_pk`, orgID).Scan(&projectID); err != nil {
		return 0, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO project_member (project_id_fk, user_id_fk, role) VALUES ($1,$2,'pm')`, projectID, ownerUserID)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return orgID, nil
}

func GetUserOrgRole(ctx context.Context, conn *pgxpool.Pool, userID, orgID int) (string, error) {
	var role string
	err := conn.QueryRow(ctx, `SELECT role FROM org_member WHERE user_id_fk = $1 AND org_id_fk = $2`, userID, orgID).Scan(&role)
	if err == pgx.ErrNoRows {
		return "", nil
	}
	return role, err
}

func GetUserProjects(ctx context.Context, conn *pgxpool.Pool, userID, orgID int) ([]Project, error) {
	// If org admin/owner -> all projects in org; else only membership projects.
	role, err := GetUserOrgRole(ctx, conn, userID, orgID)
	if err != nil {
		return nil, err
	}
	if role == "owner" || role == "admin" {
		rows, err := conn.Query(ctx, `SELECT id_pk, org_id_fk, name FROM projects WHERE org_id_fk = $1 ORDER BY id_pk`, orgID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var list []Project
		for rows.Next() {
			var p Project
			if err := rows.Scan(&p.Id, &p.OrgId, &p.Name); err != nil {
				return nil, err
			}
			list = append(list, p)
		}
		return list, rows.Err()
	}

	rows, err := conn.Query(ctx, `
		SELECT p.id_pk, p.org_id_fk, p.name, pm.role
		FROM projects p
		JOIN project_member pm ON pm.project_id_fk = p.id_pk
		WHERE p.org_id_fk = $1 AND pm.user_id_fk = $2
		ORDER BY p.id_pk
	`, orgID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.Id, &p.OrgId, &p.Name, &p.Role); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

func CreateProject(ctx context.Context, conn *pgxpool.Pool, orgID int, name string, creatorUserID int) (int, error) {
	role, err := GetUserOrgRole(ctx, conn, creatorUserID, orgID)
	if err != nil {
		return 0, err
	}
	if role != "owner" && role != "admin" {
		return 0, pgx.ErrNoRows
	}
	var id int
	if err := conn.QueryRow(ctx, `INSERT INTO projects (org_id_fk, name) VALUES ($1,$2) RETURNING id_pk`, orgID, name).Scan(&id); err != nil {
		return 0, err
	}
	// Make creator a pm in this project.
	_, _ = conn.Exec(ctx, `INSERT INTO project_member (project_id_fk, user_id_fk, role) VALUES ($1,$2,'pm') ON CONFLICT DO NOTHING`, id, creatorUserID)
	return id, nil
}

func AddUserToOrgByEmail(ctx context.Context, conn *pgxpool.Pool, orgID int, email string, role string, actorUserID int) (bool, string, error) {
	actorRole, err := GetUserOrgRole(ctx, conn, actorUserID, orgID)
	if err != nil {
		return false, "", err
	}
	if actorRole != "owner" && actorRole != "admin" {
		return false, "not_allowed", nil
	}
	var userID int
	if err := conn.QueryRow(ctx, `SELECT id_pk FROM "User" WHERE email = $1`, email).Scan(&userID); err != nil {
		if err == pgx.ErrNoRows {
			return false, "user_not_found", nil
		}
		return false, "", err
	}
	if role == "" {
		role = "member"
	}
	ct, err := conn.Exec(ctx, `INSERT INTO org_member (org_id_fk, user_id_fk, role) VALUES ($1,$2,$3) ON CONFLICT (org_id_fk,user_id_fk) DO UPDATE SET role = EXCLUDED.role`, orgID, userID, role)
	if err != nil {
		return false, "", err
	}
	return ct.RowsAffected() > 0, "", nil
}

func AddUserToProjectByEmail(ctx context.Context, conn *pgxpool.Pool, projectID int, email string, role string, actorUserID int) (bool, string, error) {
	// check actor is org admin/owner of project's org
	var orgID int
	if err := conn.QueryRow(ctx, `SELECT org_id_fk FROM projects WHERE id_pk = $1`, projectID).Scan(&orgID); err != nil {
		return false, "", err
	}
	actorRole, err := GetUserOrgRole(ctx, conn, actorUserID, orgID)
	if err != nil {
		return false, "", err
	}
	if actorRole != "owner" && actorRole != "admin" {
		return false, "not_allowed", nil
	}
	var userID int
	if err := conn.QueryRow(ctx, `SELECT id_pk FROM "User" WHERE email = $1`, email).Scan(&userID); err != nil {
		if err == pgx.ErrNoRows {
			return false, "user_not_found", nil
		}
		return false, "", err
	}
	if role == "" {
		role = "viewer"
	}
	ct, err := conn.Exec(ctx, `INSERT INTO project_member (project_id_fk, user_id_fk, role) VALUES ($1,$2,$3) ON CONFLICT (project_id_fk,user_id_fk) DO UPDATE SET role = EXCLUDED.role`, projectID, userID, role)
	if err != nil {
		return false, "", err
	}
	return ct.RowsAffected() > 0, "", nil
}

func GetOrgMembers(ctx context.Context, conn *pgxpool.Pool, orgID, actorUserID int) ([]OrgMember, error) {
	actorRole, err := GetUserOrgRole(ctx, conn, actorUserID, orgID)
	if err != nil {
		return nil, err
	}
	if actorRole != "owner" && actorRole != "admin" {
		return nil, pgx.ErrNoRows
	}
	rows, err := conn.Query(ctx, `
		SELECT om.user_id_fk, u.email, om.role
		FROM org_member om
		JOIN "User" u ON u.id_pk = om.user_id_fk
		WHERE om.org_id_fk = $1
		ORDER BY u.email
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []OrgMember
	for rows.Next() {
		var m OrgMember
		if err := rows.Scan(&m.UserId, &m.Email, &m.Role); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, rows.Err()
}

func GetProjectMembers(ctx context.Context, conn *pgxpool.Pool, projectID, actorUserID int) ([]ProjectMember, error) {
	var orgID int
	if err := conn.QueryRow(ctx, `SELECT org_id_fk FROM projects WHERE id_pk = $1`, projectID).Scan(&orgID); err != nil {
		return nil, err
	}
	actorRole, err := GetUserOrgRole(ctx, conn, actorUserID, orgID)
	if err != nil {
		return nil, err
	}
	if actorRole != "owner" && actorRole != "admin" {
		return nil, pgx.ErrNoRows
	}
	rows, err := conn.Query(ctx, `
		SELECT pm.user_id_fk, u.email, pm.role
		FROM project_member pm
		JOIN "User" u ON u.id_pk = pm.user_id_fk
		WHERE pm.project_id_fk = $1
		ORDER BY u.email
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []ProjectMember
	for rows.Next() {
		var m ProjectMember
		if err := rows.Scan(&m.UserId, &m.Email, &m.Role); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, rows.Err()
}

func UpdateOrgMemberRole(ctx context.Context, conn *pgxpool.Pool, orgID, targetUserID int, role string, actorUserID int) (bool, error) {
	actorRole, err := GetUserOrgRole(ctx, conn, actorUserID, orgID)
	if err != nil {
		return false, err
	}
	if actorRole != "owner" && actorRole != "admin" {
		return false, nil
	}
	ct, err := conn.Exec(ctx, `UPDATE org_member SET role = $1 WHERE org_id_fk = $2 AND user_id_fk = $3`, role, orgID, targetUserID)
	if err != nil {
		return false, err
	}
	return ct.RowsAffected() > 0, nil
}

func RemoveOrgMember(ctx context.Context, conn *pgxpool.Pool, orgID, targetUserID, actorUserID int) (bool, error) {
	actorRole, err := GetUserOrgRole(ctx, conn, actorUserID, orgID)
	if err != nil {
		return false, err
	}
	if actorRole != "owner" && actorRole != "admin" {
		return false, nil
	}
	// Prevent removing last owner from org.
	var targetRole string
	_ = conn.QueryRow(ctx, `SELECT role FROM org_member WHERE org_id_fk = $1 AND user_id_fk = $2`, orgID, targetUserID).Scan(&targetRole)
	if targetRole == "owner" {
		var owners int
		if err := conn.QueryRow(ctx, `SELECT COUNT(*) FROM org_member WHERE org_id_fk = $1 AND role = 'owner'`, orgID).Scan(&owners); err == nil && owners <= 1 {
			return false, nil
		}
	}
	ct, err := conn.Exec(ctx, `DELETE FROM org_member WHERE org_id_fk = $1 AND user_id_fk = $2`, orgID, targetUserID)
	if err != nil {
		return false, err
	}
	return ct.RowsAffected() > 0, nil
}

func UpdateProjectMemberRole(ctx context.Context, conn *pgxpool.Pool, projectID, targetUserID int, role string, actorUserID int) (bool, error) {
	var orgID int
	if err := conn.QueryRow(ctx, `SELECT org_id_fk FROM projects WHERE id_pk = $1`, projectID).Scan(&orgID); err != nil {
		return false, err
	}
	actorRole, err := GetUserOrgRole(ctx, conn, actorUserID, orgID)
	if err != nil {
		return false, err
	}
	if actorRole != "owner" && actorRole != "admin" {
		return false, nil
	}
	ct, err := conn.Exec(ctx, `UPDATE project_member SET role = $1 WHERE project_id_fk = $2 AND user_id_fk = $3`, role, projectID, targetUserID)
	if err != nil {
		return false, err
	}
	return ct.RowsAffected() > 0, nil
}

func RemoveProjectMember(ctx context.Context, conn *pgxpool.Pool, projectID, targetUserID, actorUserID int) (bool, error) {
	var orgID int
	if err := conn.QueryRow(ctx, `SELECT org_id_fk FROM projects WHERE id_pk = $1`, projectID).Scan(&orgID); err != nil {
		return false, err
	}
	actorRole, err := GetUserOrgRole(ctx, conn, actorUserID, orgID)
	if err != nil {
		return false, err
	}
	if actorRole != "owner" && actorRole != "admin" {
		return false, nil
	}
	ct, err := conn.Exec(ctx, `DELETE FROM project_member WHERE project_id_fk = $1 AND user_id_fk = $2`, projectID, targetUserID)
	if err != nil {
		return false, err
	}
	return ct.RowsAffected() > 0, nil
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

func DeleteTask(ctx context.Context, conn *pgxpool.Pool, taskID int, actorUserID int) (bool, error) {
	var deletedID int
	err := conn.QueryRow(
		ctx,
		`DELETE FROM Task t
		 WHERE t.id_pk = $1
		   AND (
		     t.owner_id_fk = $2
		     OR EXISTS (
		       SELECT 1
		       FROM projects p
		       JOIN org_member om ON om.org_id_fk = p.org_id_fk
		       WHERE p.id_pk = t.project_id_fk AND om.user_id_fk = $2 AND om.role IN ('owner','admin')
		     )
		   )
		 RETURNING id_pk`,
		taskID,
		actorUserID,
	).Scan(&deletedID)

	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func DeleteBug(ctx context.Context, conn *pgxpool.Pool, bugID int, actorUserID int) (bool, error) {
	var deletedID int
	err := conn.QueryRow(
		ctx,
		`DELETE FROM Bug
		 WHERE id_pk = $1
		   AND (
		     created_by_fk = $2
		     OR EXISTS (
		       SELECT 1
		       FROM Bug b
		       JOIN Task t ON t.id_pk = b.task_id_fk
		       JOIN projects p ON p.id_pk = t.project_id_fk
		       JOIN org_member om ON om.org_id_fk = p.org_id_fk
		       WHERE b.id_pk = $1 AND om.user_id_fk = $2 AND om.role IN ('owner','admin')
		     )
		   )
		 RETURNING id_pk`,
		bugID,
		actorUserID,
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

func GetUserAuthByID(ctx context.Context, conn *pgxpool.Pool, userID int) (*User, error) {
	var u User
	err := conn.QueryRow(ctx,
		`SELECT id_pk, email, password, role FROM "User" WHERE id_pk = $1`, userID,
	).Scan(&u.Id, &u.Email, &u.Password, &u.Role)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func UpdateUserEmail(ctx context.Context, conn *pgxpool.Pool, userID int, email string) error {
	_, err := conn.Exec(ctx, `UPDATE "User" SET email = $1 WHERE id_pk = $2`, email, userID)
	return err
}

func GetUserJWTVersion(ctx context.Context, conn *pgxpool.Pool, userID int) (int, error) {
	var v int
	err := conn.QueryRow(ctx, `SELECT jwt_version FROM "User" WHERE id_pk = $1`, userID).Scan(&v)
	if err == pgx.ErrNoRows {
		return 0, nil
	}
	return v, err
}

func IncrementUserJWTVersion(ctx context.Context, conn *pgxpool.Pool, userID int) (int, error) {
	var v int
	err := conn.QueryRow(ctx, `UPDATE "User" SET jwt_version = jwt_version + 1 WHERE id_pk = $1 RETURNING jwt_version`, userID).Scan(&v)
	return v, err
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
