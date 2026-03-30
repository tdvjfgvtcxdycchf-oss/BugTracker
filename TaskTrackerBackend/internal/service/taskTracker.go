package service

import (
	"bug_tracker/internal/sql"
	"context"
	"errors"
	"log/slog"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskTrackerService struct {
	conn *pgxpool.Pool
}

func NewTaskTrackerService(conn *pgxpool.Pool) *TaskTrackerService {
	return &TaskTrackerService{conn: conn}
}

func (t *TaskTrackerService) Register(ctx context.Context, user sql.User) (int, string, error) {
	existingUser, err := sql.GetByEmail(ctx, t.conn, user)
	if err != nil {
		return 0, "", err
	}

	if existingUser != nil {
		slog.Warn("registration failed: email already taken", "email", user.Email)
		return 0, "", errors.New("user already exists")
	}

	id, err := sql.CreateUser(ctx, t.conn, user)
	if err != nil {
		return 0, "", err
	}

	role := user.Role
	if role == "" {
		role = "qa"
	}

	slog.Info("user registered successfully", "id", strconv.Itoa(id), "role", role)
	return id, role, nil
}

func (t *TaskTrackerService) Login(ctx context.Context, user sql.User) (int, string, error) {
	existingUser, err := sql.GetByEmail(ctx, t.conn, user)
	if err != nil {
		return 0, "", err
	}

	if existingUser == nil {
		slog.Warn("login failed: there is no user with this email", "email", user.Email)
		return 0, "", errors.New("user not exist")
	}

	if existingUser.Password != user.Password {
		slog.Warn("login failed: the password is incorrect")
		return 0, "", errors.New("password incorrect")
	}

	slog.Info("user logined successfully", "id", existingUser.Id, "role", existingUser.Role)
	return existingUser.Id, existingUser.Role, nil
}

func (t *TaskTrackerService) GetOtherUsersEmails(ctx context.Context, excludeId int) ([]string, error) {
	return sql.GetOtherUsersEmails(ctx, t.conn, excludeId)
}

func (t *TaskTrackerService) GetAllTasks(ctx context.Context) ([]sql.Task, error) {
	return sql.GetAllTasks(ctx, t.conn)
}

func (t *TaskTrackerService) CreateTask(ctx context.Context, task sql.Task) error {
	return sql.CreateTask(ctx, t.conn, task)
}

func (t *TaskTrackerService) GetAllBugs(ctx context.Context, id int) ([]sql.Bug, error) {
	return sql.GetBugsByTaskId(ctx, t.conn, id)
}

func (t *TaskTrackerService) CreateBug(ctx context.Context, bug sql.Bug) error {
	return sql.CreateBug(ctx, t.conn, bug)
}

func (t *TaskTrackerService) UpdateBug(ctx context.Context, bug sql.Bug, assignedEmail string) error {
	return sql.ChangeBug(ctx, t.conn, bug, assignedEmail)
}

func (t *TaskTrackerService) DeleteTask(ctx context.Context, taskID int, ownerID int) (bool, error) {
	return sql.DeleteTask(ctx, t.conn, taskID, ownerID)
}

func (t *TaskTrackerService) DeleteBug(ctx context.Context, bugID int, creatorID int) (bool, error) {
	return sql.DeleteBug(ctx, t.conn, bugID, creatorID)
}

func (t *TaskTrackerService) SaveBugPhoto(ctx context.Context, bugID int, data []byte) error {
	return sql.SaveBugPhoto(ctx, t.conn, bugID, data)
}

func (t *TaskTrackerService) GetBugPhoto(ctx context.Context, bugID int) ([]byte, error) {
	return sql.GetBugPhoto(ctx, t.conn, bugID)
}

// ── Comments ─────────────────────────────────────────────────────────────────

func (t *TaskTrackerService) AddBugComment(ctx context.Context, bugID, userID int, body string) (*sql.BugComment, error) {
	return sql.AddBugComment(ctx, t.conn, bugID, userID, body)
}

func (t *TaskTrackerService) GetBugComments(ctx context.Context, bugID int) ([]sql.BugComment, error) {
	return sql.GetBugComments(ctx, t.conn, bugID)
}

// ── Audit log ─────────────────────────────────────────────────────────────────

func (t *TaskTrackerService) AddAuditEntry(ctx context.Context, bugID, userID int, field, oldVal, newVal string) error {
	return sql.AddAuditEntry(ctx, t.conn, bugID, userID, field, oldVal, newVal)
}

func (t *TaskTrackerService) GetAuditLog(ctx context.Context, bugID int) ([]sql.AuditEntry, error) {
	return sql.GetAuditLog(ctx, t.conn, bugID)
}

// ── Relations ────────────────────────────────────────────────────────────────

func (t *TaskTrackerService) AddBugRelation(ctx context.Context, fromID, toID int, relType string) error {
	return sql.AddBugRelation(ctx, t.conn, fromID, toID, relType)
}

func (t *TaskTrackerService) DeleteBugRelation(ctx context.Context, relationID int) error {
	return sql.DeleteBugRelation(ctx, t.conn, relationID)
}

func (t *TaskTrackerService) GetBugRelations(ctx context.Context, bugID int) ([]sql.BugRelation, error) {
	return sql.GetBugRelations(ctx, t.conn, bugID)
}

// ── Tags ─────────────────────────────────────────────────────────────────────

func (t *TaskTrackerService) SetBugTags(ctx context.Context, bugID int, tags []string) error {
	return sql.SetBugTags(ctx, t.conn, bugID, tags)
}

func (t *TaskTrackerService) GetBugTags(ctx context.Context, bugID int) ([]string, error) {
	return sql.GetBugTags(ctx, t.conn, bugID)
}

// ── Templates ────────────────────────────────────────────────────────────────

func (t *TaskTrackerService) CreateTemplate(ctx context.Context, tmpl sql.BugTemplate) (int, error) {
	return sql.CreateTemplate(ctx, t.conn, tmpl)
}

func (t *TaskTrackerService) GetTemplates(ctx context.Context) ([]sql.BugTemplate, error) {
	return sql.GetTemplates(ctx, t.conn)
}

func (t *TaskTrackerService) DeleteTemplate(ctx context.Context, id, userID int) (bool, error) {
	return sql.DeleteTemplate(ctx, t.conn, id, userID)
}

// ── Analytics ─────────────────────────────────────────────────────────────────

func (t *TaskTrackerService) GetBugStatsByTask(ctx context.Context, taskID int) ([]sql.BugStats, error) {
	return sql.GetBugStatsByTask(ctx, t.conn, taskID)
}

func (t *TaskTrackerService) GetAllBugStats(ctx context.Context) ([]sql.BugStats, error) {
	return sql.GetAllBugStats(ctx, t.conn)
}

// ── User ──────────────────────────────────────────────────────────────────────

func (t *TaskTrackerService) GetUserRole(ctx context.Context, userID int) (string, error) {
	return sql.GetUserRole(ctx, t.conn, userID)
}

func (t *TaskTrackerService) GetUserByID(ctx context.Context, userID int) (*sql.User, error) {
	return sql.GetUserByID(ctx, t.conn, userID)
}
