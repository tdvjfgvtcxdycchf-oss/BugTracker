package service

import (
	"bug_tracker/internal/sql"
	"context"
	"errors"
	"log/slog"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
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

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, "", err
	}
	user.Password = string(passwordHash)

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

	// Prefer bcrypt; keep backward compatibility for legacy plaintext users.
	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(user.Password)); err != nil {
		if existingUser.Password != user.Password {
			slog.Warn("login failed: the password is incorrect")
			return 0, "", errors.New("password incorrect")
		}
		// Legacy plaintext password matched: upgrade to bcrypt hash.
		if hash, hashErr := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost); hashErr == nil {
			_ = sql.UpdateUserPasswordHash(ctx, t.conn, existingUser.Id, string(hash))
		}
	}

	slog.Info("user logined successfully", "id", existingUser.Id, "role", existingUser.Role)
	return existingUser.Id, existingUser.Role, nil
}

func (t *TaskTrackerService) GetOtherUsersEmails(ctx context.Context, excludeId int) ([]string, error) {
	return sql.GetOtherUsersEmails(ctx, t.conn, excludeId)
}

func (t *TaskTrackerService) GetTasksByProjectForUser(ctx context.Context, userID, projectID int) ([]sql.Task, error) {
	return sql.GetTasksByProjectForUser(ctx, t.conn, userID, projectID)
}

func (t *TaskTrackerService) CreateTask(ctx context.Context, task sql.Task) (int, error) {
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

func (t *TaskTrackerService) CanUserAccessTask(ctx context.Context, userID, taskID int) (bool, error) {
	return sql.CanUserAccessTask(ctx, t.conn, userID, taskID)
}

func (t *TaskTrackerService) CanUserAccessBug(ctx context.Context, userID, bugID int) (bool, error) {
	return sql.CanUserAccessBug(ctx, t.conn, userID, bugID)
}

func (t *TaskTrackerService) GetRelationFromBugID(ctx context.Context, relID int) (int, error) {
	return sql.GetRelationFromBugID(ctx, t.conn, relID)
}

// ── Chat ──────────────────────────────────────────────────────────────────────

func (t *TaskTrackerService) EnsureOrgThread(ctx context.Context, orgID, actorUserID int) (int, error) {
	return sql.EnsureOrgThread(ctx, t.conn, orgID, actorUserID)
}

func (t *TaskTrackerService) EnsureProjectThread(ctx context.Context, projectID, actorUserID int) (int, error) {
	return sql.EnsureProjectThread(ctx, t.conn, projectID, actorUserID)
}

func (t *TaskTrackerService) EnsureDMThreadByEmail(ctx context.Context, email string, actorUserID int) (int, error) {
	return sql.EnsureDMThreadByEmail(ctx, t.conn, email, actorUserID)
}

func (t *TaskTrackerService) GetOrgThread(ctx context.Context, orgID, actorUserID int) (*sql.ChatThread, error) {
	return sql.GetOrgThread(ctx, t.conn, orgID, actorUserID)
}

func (t *TaskTrackerService) GetProjectThread(ctx context.Context, projectID, actorUserID int) (*sql.ChatThread, error) {
	return sql.GetProjectThread(ctx, t.conn, projectID, actorUserID)
}

func (t *TaskTrackerService) GetOrgThreads(ctx context.Context, orgID, actorUserID int) ([]sql.ChatThread, error) {
	return sql.GetOrgThreads(ctx, t.conn, orgID, actorUserID)
}

func (t *TaskTrackerService) GetProjectThreads(ctx context.Context, projectID, actorUserID int) ([]sql.ChatThread, error) {
	return sql.GetProjectThreads(ctx, t.conn, projectID, actorUserID)
}

func (t *TaskTrackerService) GetDMThreads(ctx context.Context, actorUserID int) ([]sql.ChatThread, error) {
	return sql.GetDMThreads(ctx, t.conn, actorUserID)
}

func (t *TaskTrackerService) AddChatMessage(ctx context.Context, threadID, actorUserID int, body string) (int, error) {
	return sql.AddChatMessage(ctx, t.conn, threadID, actorUserID, body)
}

func (t *TaskTrackerService) GetChatMessages(ctx context.Context, threadID, actorUserID int, limit int, beforeID int) ([]sql.ChatMessage, error) {
	return sql.GetChatMessages(ctx, t.conn, threadID, actorUserID, limit, beforeID)
}

func (t *TaskTrackerService) MarkThreadRead(ctx context.Context, threadID, actorUserID int) error {
	return sql.MarkThreadRead(ctx, t.conn, threadID, actorUserID)
}

func (t *TaskTrackerService) UpdateChatMessage(ctx context.Context, messageID, actorUserID int, body string) (bool, error) {
	return sql.UpdateChatMessage(ctx, t.conn, messageID, actorUserID, body)
}

func (t *TaskTrackerService) DeleteChatMessage(ctx context.Context, messageID, actorUserID int) (bool, error) {
	return sql.DeleteChatMessage(ctx, t.conn, messageID, actorUserID)
}

func (t *TaskTrackerService) UpsertTypingState(ctx context.Context, threadID, actorUserID int, isTyping bool) error {
	return sql.UpsertTypingState(ctx, t.conn, threadID, actorUserID, isTyping)
}

func (t *TaskTrackerService) GetTypingUsers(ctx context.Context, threadID, actorUserID int) ([]string, error) {
	return sql.GetTypingUsers(ctx, t.conn, threadID, actorUserID)
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

func (t *TaskTrackerService) GetUserAuthByID(ctx context.Context, userID int) (*sql.User, error) {
	return sql.GetUserAuthByID(ctx, t.conn, userID)
}

func (t *TaskTrackerService) GetUserJWTVersion(ctx context.Context, userID int) (int, error) {
	return sql.GetUserJWTVersion(ctx, t.conn, userID)
}

func (t *TaskTrackerService) IncrementUserJWTVersion(ctx context.Context, userID int) (int, error) {
	return sql.IncrementUserJWTVersion(ctx, t.conn, userID)
}

func (t *TaskTrackerService) UpdateUserEmail(ctx context.Context, userID int, email string) error {
	return sql.UpdateUserEmail(ctx, t.conn, userID, email)
}

func (t *TaskTrackerService) UpdateUserPasswordHash(ctx context.Context, userID int, hash string) error {
	return sql.UpdateUserPasswordHash(ctx, t.conn, userID, hash)
}

// ── Orgs / Projects ───────────────────────────────────────────────────────────

func (t *TaskTrackerService) GetUserOrgs(ctx context.Context, userID int) ([]sql.Organization, error) {
	return sql.GetUserOrgs(ctx, t.conn, userID)
}

func (t *TaskTrackerService) CreateOrg(ctx context.Context, name string, ownerUserID int) (int, error) {
	return sql.CreateOrg(ctx, t.conn, name, ownerUserID)
}

func (t *TaskTrackerService) GetUserProjects(ctx context.Context, userID, orgID int) ([]sql.Project, error) {
	return sql.GetUserProjects(ctx, t.conn, userID, orgID)
}

func (t *TaskTrackerService) CreateProject(ctx context.Context, orgID int, name string, creatorUserID int) (int, error) {
	return sql.CreateProject(ctx, t.conn, orgID, name, creatorUserID)
}

func (t *TaskTrackerService) AddUserToOrgByEmail(ctx context.Context, orgID int, email string, role string, actorUserID int) (bool, string, error) {
	return sql.AddUserToOrgByEmail(ctx, t.conn, orgID, email, role, actorUserID)
}

func (t *TaskTrackerService) AddUserToProjectByEmail(ctx context.Context, projectID int, email string, role string, actorUserID int) (bool, string, error) {
	return sql.AddUserToProjectByEmail(ctx, t.conn, projectID, email, role, actorUserID)
}

func (t *TaskTrackerService) GetOrgMembers(ctx context.Context, orgID, actorUserID int) ([]sql.OrgMember, error) {
	return sql.GetOrgMembers(ctx, t.conn, orgID, actorUserID)
}

func (t *TaskTrackerService) GetProjectMembers(ctx context.Context, projectID, actorUserID int) ([]sql.ProjectMember, error) {
	return sql.GetProjectMembers(ctx, t.conn, projectID, actorUserID)
}

func (t *TaskTrackerService) UpdateOrgMemberRole(ctx context.Context, orgID, targetUserID int, role string, actorUserID int) (bool, error) {
	return sql.UpdateOrgMemberRole(ctx, t.conn, orgID, targetUserID, role, actorUserID)
}

func (t *TaskTrackerService) RemoveOrgMember(ctx context.Context, orgID, targetUserID, actorUserID int) (bool, error) {
	return sql.RemoveOrgMember(ctx, t.conn, orgID, targetUserID, actorUserID)
}

func (t *TaskTrackerService) UpdateProjectMemberRole(ctx context.Context, projectID, targetUserID int, role string, actorUserID int) (bool, error) {
	return sql.UpdateProjectMemberRole(ctx, t.conn, projectID, targetUserID, role, actorUserID)
}

func (t *TaskTrackerService) RemoveProjectMember(ctx context.Context, projectID, targetUserID, actorUserID int) (bool, error) {
	return sql.RemoveProjectMember(ctx, t.conn, projectID, targetUserID, actorUserID)
}
