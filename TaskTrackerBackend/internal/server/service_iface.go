package server

import (
	"bug_tracker/internal/sql"
	"context"
)

// Service is the interface satisfied by *service.TaskTrackerService.
// Using an interface here decouples HTTP handlers from the concrete service
// and allows injection of test doubles without a real database.
type Service interface {
	// Auth
	Register(ctx context.Context, user sql.User) (int, string, error)
	Login(ctx context.Context, user sql.User) (int, string, error)

	// User profile
	GetOtherUsersEmails(ctx context.Context, excludeId int) ([]string, error)
	GetUserAuthByID(ctx context.Context, userID int) (*sql.User, error)
	GetUserJWTVersion(ctx context.Context, userID int) (int, error)
	IncrementUserJWTVersion(ctx context.Context, userID int) (int, error)
	UpdateUserEmail(ctx context.Context, userID int, email string) error
	UpdateUserPasswordHash(ctx context.Context, userID int, hash string) error

	// Orgs
	GetUserOrgs(ctx context.Context, userID int) ([]sql.Organization, error)
	CreateOrg(ctx context.Context, name string, ownerUserID int) (int, error)
	AddUserToOrgByEmail(ctx context.Context, orgID int, email, role string, actorUserID int) (bool, string, error)
	GetOrgMembers(ctx context.Context, orgID, actorUserID int) ([]sql.OrgMember, error)
	UpdateOrgMemberRole(ctx context.Context, orgID, targetUserID int, role string, actorUserID int) (bool, error)
	RemoveOrgMember(ctx context.Context, orgID, targetUserID, actorUserID int) (bool, error)

	// Projects
	GetUserProjects(ctx context.Context, userID, orgID int) ([]sql.Project, error)
	CreateProject(ctx context.Context, orgID int, name string, creatorUserID int) (int, error)
	AddUserToProjectByEmail(ctx context.Context, projectID int, email, role string, actorUserID int) (bool, string, error)
	GetProjectMembers(ctx context.Context, projectID, actorUserID int) ([]sql.ProjectMember, error)
	UpdateProjectMemberRole(ctx context.Context, projectID, targetUserID int, role string, actorUserID int) (bool, error)
	RemoveProjectMember(ctx context.Context, projectID, targetUserID, actorUserID int) (bool, error)

	// Tasks
	GetTasksByProjectForUser(ctx context.Context, userID, projectID int) ([]sql.Task, error)
	CreateTask(ctx context.Context, task sql.Task) (int, error)
	DeleteTask(ctx context.Context, taskID, ownerID int) (bool, error)
	CanUserAccessTask(ctx context.Context, userID, taskID int) (bool, error)

	// Bugs
	GetAllBugs(ctx context.Context, taskID int) ([]sql.Bug, error)
	CreateBug(ctx context.Context, bug sql.Bug) error
	UpdateBug(ctx context.Context, bug sql.Bug, assignedEmail string) error
	DeleteBug(ctx context.Context, bugID, creatorID int) (bool, error)
	CanUserAccessBug(ctx context.Context, userID, bugID int) (bool, error)
	SaveBugPhoto(ctx context.Context, bugID int, data []byte) error
	GetBugPhoto(ctx context.Context, bugID int) ([]byte, error)

	// Comments
	AddBugComment(ctx context.Context, bugID, userID int, body string) (*sql.BugComment, error)
	GetBugComments(ctx context.Context, bugID int) ([]sql.BugComment, error)

	// Audit log
	GetAuditLog(ctx context.Context, bugID int) ([]sql.AuditEntry, error)

	// Relations
	GetRelationFromBugID(ctx context.Context, relID int) (int, error)
	AddBugRelation(ctx context.Context, fromID, toID int, relType string) error
	DeleteBugRelation(ctx context.Context, relationID int) error
	GetBugRelations(ctx context.Context, bugID int) ([]sql.BugRelation, error)

	// Tags
	SetBugTags(ctx context.Context, bugID int, tags []string) error
	GetBugTags(ctx context.Context, bugID int) ([]string, error)

	// Templates
	GetTemplates(ctx context.Context) ([]sql.BugTemplate, error)
	CreateTemplate(ctx context.Context, tmpl sql.BugTemplate) (int, error)
	DeleteTemplate(ctx context.Context, id, userID int) (bool, error)

	// Analytics
	GetBugStatsByTask(ctx context.Context, taskID int) ([]sql.BugStats, error)
	GetAllBugStats(ctx context.Context) ([]sql.BugStats, error)

	// Chat
	EnsureOrgThread(ctx context.Context, orgID, actorUserID int) (int, error)
	EnsureProjectThread(ctx context.Context, projectID, actorUserID int) (int, error)
	EnsureDMThreadByEmail(ctx context.Context, email string, actorUserID int) (int, error)
	GetOrgThreads(ctx context.Context, orgID, actorUserID int) ([]sql.ChatThread, error)
	GetProjectThreads(ctx context.Context, projectID, actorUserID int) ([]sql.ChatThread, error)
	GetDMThreads(ctx context.Context, actorUserID int) ([]sql.ChatThread, error)
	AddChatMessage(ctx context.Context, threadID, actorUserID int, body string) (int, error)
	GetChatMessages(ctx context.Context, threadID, actorUserID, limit, beforeID int) ([]sql.ChatMessage, error)
	MarkThreadRead(ctx context.Context, threadID, actorUserID int) error
	UpdateChatMessage(ctx context.Context, messageID, actorUserID int, body string) (bool, error)
	DeleteChatMessage(ctx context.Context, messageID, actorUserID int) (bool, error)
	UpsertTypingState(ctx context.Context, threadID, actorUserID int, isTyping bool) error
	GetTypingUsers(ctx context.Context, threadID, actorUserID int) ([]string, error)
}
