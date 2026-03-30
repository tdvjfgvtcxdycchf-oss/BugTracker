package server

import (
	"bug_tracker/internal/sql"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// ── Mock service ────────────────────────────────────────────────────────────

// mockService provides configurable stubs for each service method.
// Only fields used by the handler under test need to be set;
// unset fields return zero values without panicking.
type mockService struct {
	registerFn              func(ctx context.Context, user sql.User) (int, string, error)
	loginFn                 func(ctx context.Context, user sql.User) (int, string, error)
	getOtherUsersEmailsFn   func(ctx context.Context, excludeId int) ([]string, error)
	getUserAuthByIDFn        func(ctx context.Context, userID int) (*sql.User, error)
	getUserJWTVersionFn      func(ctx context.Context, userID int) (int, error)
	incrementUserJWTVersionFn func(ctx context.Context, userID int) (int, error)
	updateUserEmailFn       func(ctx context.Context, userID int, email string) error
	updateUserPasswordHashFn func(ctx context.Context, userID int, hash string) error
	getUserOrgsFn           func(ctx context.Context, userID int) ([]sql.Organization, error)
	createOrgFn             func(ctx context.Context, name string, ownerUserID int) (int, error)
	addUserToOrgByEmailFn   func(ctx context.Context, orgID int, email, role string, actorUserID int) (bool, string, error)
	getOrgMembersFn         func(ctx context.Context, orgID, actorUserID int) ([]sql.OrgMember, error)
	updateOrgMemberRoleFn   func(ctx context.Context, orgID, targetUserID int, role string, actorUserID int) (bool, error)
	removeOrgMemberFn       func(ctx context.Context, orgID, targetUserID, actorUserID int) (bool, error)
	getUserProjectsFn       func(ctx context.Context, userID, orgID int) ([]sql.Project, error)
	createProjectFn         func(ctx context.Context, orgID int, name string, creatorUserID int) (int, error)
	addUserToProjectByEmailFn func(ctx context.Context, projectID int, email, role string, actorUserID int) (bool, string, error)
	getProjectMembersFn     func(ctx context.Context, projectID, actorUserID int) ([]sql.ProjectMember, error)
	updateProjectMemberRoleFn func(ctx context.Context, projectID, targetUserID int, role string, actorUserID int) (bool, error)
	removeProjectMemberFn   func(ctx context.Context, projectID, targetUserID, actorUserID int) (bool, error)
	getTasksByProjectFn     func(ctx context.Context, userID, projectID int) ([]sql.Task, error)
	createTaskFn            func(ctx context.Context, task sql.Task) (int, error)
	deleteTaskFn            func(ctx context.Context, taskID, ownerID int) (bool, error)
	canUserAccessTaskFn     func(ctx context.Context, userID, taskID int) (bool, error)
	getAllBugsFn             func(ctx context.Context, taskID int) ([]sql.Bug, error)
	createBugFn             func(ctx context.Context, bug sql.Bug) error
	updateBugFn             func(ctx context.Context, bug sql.Bug, assignedEmail string) error
	deleteBugFn             func(ctx context.Context, bugID, creatorID int) (bool, error)
	canUserAccessBugFn      func(ctx context.Context, userID, bugID int) (bool, error)
	saveBugPhotoFn          func(ctx context.Context, bugID int, data []byte) error
	getBugPhotoFn           func(ctx context.Context, bugID int) ([]byte, error)
	addBugCommentFn         func(ctx context.Context, bugID, userID int, body string) (*sql.BugComment, error)
	getBugCommentsFn        func(ctx context.Context, bugID int) ([]sql.BugComment, error)
	getAuditLogFn           func(ctx context.Context, bugID int) ([]sql.AuditEntry, error)
	getRelationFromBugIDFn  func(ctx context.Context, relID int) (int, error)
	addBugRelationFn        func(ctx context.Context, fromID, toID int, relType string) error
	deleteBugRelationFn     func(ctx context.Context, relationID int) error
	getBugRelationsFn       func(ctx context.Context, bugID int) ([]sql.BugRelation, error)
	setBugTagsFn            func(ctx context.Context, bugID int, tags []string) error
	getBugTagsFn            func(ctx context.Context, bugID int) ([]string, error)
	getTemplatesFn          func(ctx context.Context) ([]sql.BugTemplate, error)
	createTemplateFn        func(ctx context.Context, tmpl sql.BugTemplate) (int, error)
	deleteTemplateFn        func(ctx context.Context, id, userID int) (bool, error)
	getBugStatsByTaskFn     func(ctx context.Context, taskID int) ([]sql.BugStats, error)
	getAllBugStatsFn         func(ctx context.Context) ([]sql.BugStats, error)
	ensureOrgThreadFn       func(ctx context.Context, orgID, actorUserID int) (int, error)
	ensureProjectThreadFn   func(ctx context.Context, projectID, actorUserID int) (int, error)
	ensureDMThreadByEmailFn func(ctx context.Context, email string, actorUserID int) (int, error)
	getOrgThreadsFn         func(ctx context.Context, orgID, actorUserID int) ([]sql.ChatThread, error)
	getProjectThreadsFn     func(ctx context.Context, projectID, actorUserID int) ([]sql.ChatThread, error)
	getDMThreadsFn          func(ctx context.Context, actorUserID int) ([]sql.ChatThread, error)
	addChatMessageFn        func(ctx context.Context, threadID, actorUserID int, body string) (int, error)
	getChatMessagesFn       func(ctx context.Context, threadID, actorUserID, limit, beforeID int) ([]sql.ChatMessage, error)
	markThreadReadFn        func(ctx context.Context, threadID, actorUserID int) error
	updateChatMessageFn     func(ctx context.Context, messageID, actorUserID int, body string) (bool, error)
	deleteChatMessageFn     func(ctx context.Context, messageID, actorUserID int) (bool, error)
	upsertTypingStateFn     func(ctx context.Context, threadID, actorUserID int, isTyping bool) error
	getTypingUsersFn        func(ctx context.Context, threadID, actorUserID int) ([]string, error)
}

func (m *mockService) Register(ctx context.Context, u sql.User) (int, string, error) {
	if m.registerFn != nil {
		return m.registerFn(ctx, u)
	}
	return 0, "", nil
}
func (m *mockService) Login(ctx context.Context, u sql.User) (int, string, error) {
	if m.loginFn != nil {
		return m.loginFn(ctx, u)
	}
	return 0, "", nil
}
func (m *mockService) GetOtherUsersEmails(ctx context.Context, id int) ([]string, error) {
	if m.getOtherUsersEmailsFn != nil {
		return m.getOtherUsersEmailsFn(ctx, id)
	}
	return nil, nil
}
func (m *mockService) GetUserAuthByID(ctx context.Context, id int) (*sql.User, error) {
	if m.getUserAuthByIDFn != nil {
		return m.getUserAuthByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *mockService) GetUserJWTVersion(ctx context.Context, id int) (int, error) {
	if m.getUserJWTVersionFn != nil {
		return m.getUserJWTVersionFn(ctx, id)
	}
	return 1, nil
}
func (m *mockService) IncrementUserJWTVersion(ctx context.Context, id int) (int, error) {
	if m.incrementUserJWTVersionFn != nil {
		return m.incrementUserJWTVersionFn(ctx, id)
	}
	return 2, nil
}
func (m *mockService) UpdateUserEmail(ctx context.Context, id int, email string) error {
	if m.updateUserEmailFn != nil {
		return m.updateUserEmailFn(ctx, id, email)
	}
	return nil
}
func (m *mockService) UpdateUserPasswordHash(ctx context.Context, id int, hash string) error {
	if m.updateUserPasswordHashFn != nil {
		return m.updateUserPasswordHashFn(ctx, id, hash)
	}
	return nil
}
func (m *mockService) GetUserOrgs(ctx context.Context, id int) ([]sql.Organization, error) {
	if m.getUserOrgsFn != nil {
		return m.getUserOrgsFn(ctx, id)
	}
	return nil, nil
}
func (m *mockService) CreateOrg(ctx context.Context, name string, ownerID int) (int, error) {
	if m.createOrgFn != nil {
		return m.createOrgFn(ctx, name, ownerID)
	}
	return 0, nil
}
func (m *mockService) AddUserToOrgByEmail(ctx context.Context, orgID int, email, role string, actor int) (bool, string, error) {
	if m.addUserToOrgByEmailFn != nil {
		return m.addUserToOrgByEmailFn(ctx, orgID, email, role, actor)
	}
	return true, "", nil
}
func (m *mockService) GetOrgMembers(ctx context.Context, orgID, actor int) ([]sql.OrgMember, error) {
	if m.getOrgMembersFn != nil {
		return m.getOrgMembersFn(ctx, orgID, actor)
	}
	return nil, nil
}
func (m *mockService) UpdateOrgMemberRole(ctx context.Context, orgID, target int, role string, actor int) (bool, error) {
	if m.updateOrgMemberRoleFn != nil {
		return m.updateOrgMemberRoleFn(ctx, orgID, target, role, actor)
	}
	return true, nil
}
func (m *mockService) RemoveOrgMember(ctx context.Context, orgID, target, actor int) (bool, error) {
	if m.removeOrgMemberFn != nil {
		return m.removeOrgMemberFn(ctx, orgID, target, actor)
	}
	return true, nil
}
func (m *mockService) GetUserProjects(ctx context.Context, userID, orgID int) ([]sql.Project, error) {
	if m.getUserProjectsFn != nil {
		return m.getUserProjectsFn(ctx, userID, orgID)
	}
	return nil, nil
}
func (m *mockService) CreateProject(ctx context.Context, orgID int, name string, creator int) (int, error) {
	if m.createProjectFn != nil {
		return m.createProjectFn(ctx, orgID, name, creator)
	}
	return 0, nil
}
func (m *mockService) AddUserToProjectByEmail(ctx context.Context, projectID int, email, role string, actor int) (bool, string, error) {
	if m.addUserToProjectByEmailFn != nil {
		return m.addUserToProjectByEmailFn(ctx, projectID, email, role, actor)
	}
	return true, "", nil
}
func (m *mockService) GetProjectMembers(ctx context.Context, projectID, actor int) ([]sql.ProjectMember, error) {
	if m.getProjectMembersFn != nil {
		return m.getProjectMembersFn(ctx, projectID, actor)
	}
	return nil, nil
}
func (m *mockService) UpdateProjectMemberRole(ctx context.Context, projectID, target int, role string, actor int) (bool, error) {
	if m.updateProjectMemberRoleFn != nil {
		return m.updateProjectMemberRoleFn(ctx, projectID, target, role, actor)
	}
	return true, nil
}
func (m *mockService) RemoveProjectMember(ctx context.Context, projectID, target, actor int) (bool, error) {
	if m.removeProjectMemberFn != nil {
		return m.removeProjectMemberFn(ctx, projectID, target, actor)
	}
	return true, nil
}
func (m *mockService) GetTasksByProjectForUser(ctx context.Context, userID, projectID int) ([]sql.Task, error) {
	if m.getTasksByProjectFn != nil {
		return m.getTasksByProjectFn(ctx, userID, projectID)
	}
	return nil, nil
}
func (m *mockService) CreateTask(ctx context.Context, task sql.Task) (int, error) {
	if m.createTaskFn != nil {
		return m.createTaskFn(ctx, task)
	}
	return 1, nil
}
func (m *mockService) DeleteTask(ctx context.Context, taskID, ownerID int) (bool, error) {
	if m.deleteTaskFn != nil {
		return m.deleteTaskFn(ctx, taskID, ownerID)
	}
	return true, nil
}
func (m *mockService) CanUserAccessTask(ctx context.Context, userID, taskID int) (bool, error) {
	if m.canUserAccessTaskFn != nil {
		return m.canUserAccessTaskFn(ctx, userID, taskID)
	}
	return true, nil
}
func (m *mockService) GetAllBugs(ctx context.Context, taskID int) ([]sql.Bug, error) {
	if m.getAllBugsFn != nil {
		return m.getAllBugsFn(ctx, taskID)
	}
	return []sql.Bug{}, nil
}
func (m *mockService) CreateBug(ctx context.Context, bug sql.Bug) error {
	if m.createBugFn != nil {
		return m.createBugFn(ctx, bug)
	}
	return nil
}
func (m *mockService) UpdateBug(ctx context.Context, bug sql.Bug, assignedEmail string) error {
	if m.updateBugFn != nil {
		return m.updateBugFn(ctx, bug, assignedEmail)
	}
	return nil
}
func (m *mockService) DeleteBug(ctx context.Context, bugID, creatorID int) (bool, error) {
	if m.deleteBugFn != nil {
		return m.deleteBugFn(ctx, bugID, creatorID)
	}
	return true, nil
}
func (m *mockService) CanUserAccessBug(ctx context.Context, userID, bugID int) (bool, error) {
	if m.canUserAccessBugFn != nil {
		return m.canUserAccessBugFn(ctx, userID, bugID)
	}
	return true, nil
}
func (m *mockService) SaveBugPhoto(ctx context.Context, bugID int, data []byte) error {
	if m.saveBugPhotoFn != nil {
		return m.saveBugPhotoFn(ctx, bugID, data)
	}
	return nil
}
func (m *mockService) GetBugPhoto(ctx context.Context, bugID int) ([]byte, error) {
	if m.getBugPhotoFn != nil {
		return m.getBugPhotoFn(ctx, bugID)
	}
	return nil, nil
}
func (m *mockService) AddBugComment(ctx context.Context, bugID, userID int, body string) (*sql.BugComment, error) {
	if m.addBugCommentFn != nil {
		return m.addBugCommentFn(ctx, bugID, userID, body)
	}
	return &sql.BugComment{}, nil
}
func (m *mockService) GetBugComments(ctx context.Context, bugID int) ([]sql.BugComment, error) {
	if m.getBugCommentsFn != nil {
		return m.getBugCommentsFn(ctx, bugID)
	}
	return nil, nil
}
func (m *mockService) GetAuditLog(ctx context.Context, bugID int) ([]sql.AuditEntry, error) {
	if m.getAuditLogFn != nil {
		return m.getAuditLogFn(ctx, bugID)
	}
	return nil, nil
}
func (m *mockService) GetRelationFromBugID(ctx context.Context, relID int) (int, error) {
	if m.getRelationFromBugIDFn != nil {
		return m.getRelationFromBugIDFn(ctx, relID)
	}
	return 0, nil
}
func (m *mockService) AddBugRelation(ctx context.Context, fromID, toID int, relType string) error {
	if m.addBugRelationFn != nil {
		return m.addBugRelationFn(ctx, fromID, toID, relType)
	}
	return nil
}
func (m *mockService) DeleteBugRelation(ctx context.Context, relationID int) error {
	if m.deleteBugRelationFn != nil {
		return m.deleteBugRelationFn(ctx, relationID)
	}
	return nil
}
func (m *mockService) GetBugRelations(ctx context.Context, bugID int) ([]sql.BugRelation, error) {
	if m.getBugRelationsFn != nil {
		return m.getBugRelationsFn(ctx, bugID)
	}
	return nil, nil
}
func (m *mockService) SetBugTags(ctx context.Context, bugID int, tags []string) error {
	if m.setBugTagsFn != nil {
		return m.setBugTagsFn(ctx, bugID, tags)
	}
	return nil
}
func (m *mockService) GetBugTags(ctx context.Context, bugID int) ([]string, error) {
	if m.getBugTagsFn != nil {
		return m.getBugTagsFn(ctx, bugID)
	}
	return []string{}, nil
}
func (m *mockService) GetTemplates(ctx context.Context) ([]sql.BugTemplate, error) {
	if m.getTemplatesFn != nil {
		return m.getTemplatesFn(ctx)
	}
	return nil, nil
}
func (m *mockService) CreateTemplate(ctx context.Context, tmpl sql.BugTemplate) (int, error) {
	if m.createTemplateFn != nil {
		return m.createTemplateFn(ctx, tmpl)
	}
	return 1, nil
}
func (m *mockService) DeleteTemplate(ctx context.Context, id, userID int) (bool, error) {
	if m.deleteTemplateFn != nil {
		return m.deleteTemplateFn(ctx, id, userID)
	}
	return true, nil
}
func (m *mockService) GetBugStatsByTask(ctx context.Context, taskID int) ([]sql.BugStats, error) {
	if m.getBugStatsByTaskFn != nil {
		return m.getBugStatsByTaskFn(ctx, taskID)
	}
	return nil, nil
}
func (m *mockService) GetAllBugStats(ctx context.Context) ([]sql.BugStats, error) {
	if m.getAllBugStatsFn != nil {
		return m.getAllBugStatsFn(ctx)
	}
	return nil, nil
}
func (m *mockService) EnsureOrgThread(ctx context.Context, orgID, actor int) (int, error) {
	if m.ensureOrgThreadFn != nil {
		return m.ensureOrgThreadFn(ctx, orgID, actor)
	}
	return 1, nil
}
func (m *mockService) EnsureProjectThread(ctx context.Context, projectID, actor int) (int, error) {
	if m.ensureProjectThreadFn != nil {
		return m.ensureProjectThreadFn(ctx, projectID, actor)
	}
	return 1, nil
}
func (m *mockService) EnsureDMThreadByEmail(ctx context.Context, email string, actor int) (int, error) {
	if m.ensureDMThreadByEmailFn != nil {
		return m.ensureDMThreadByEmailFn(ctx, email, actor)
	}
	return 1, nil
}
func (m *mockService) GetOrgThreads(ctx context.Context, orgID, actor int) ([]sql.ChatThread, error) {
	if m.getOrgThreadsFn != nil {
		return m.getOrgThreadsFn(ctx, orgID, actor)
	}
	return nil, nil
}
func (m *mockService) GetProjectThreads(ctx context.Context, projectID, actor int) ([]sql.ChatThread, error) {
	if m.getProjectThreadsFn != nil {
		return m.getProjectThreadsFn(ctx, projectID, actor)
	}
	return nil, nil
}
func (m *mockService) GetDMThreads(ctx context.Context, actor int) ([]sql.ChatThread, error) {
	if m.getDMThreadsFn != nil {
		return m.getDMThreadsFn(ctx, actor)
	}
	return nil, nil
}
func (m *mockService) AddChatMessage(ctx context.Context, threadID, actor int, body string) (int, error) {
	if m.addChatMessageFn != nil {
		return m.addChatMessageFn(ctx, threadID, actor, body)
	}
	return 1, nil
}
func (m *mockService) GetChatMessages(ctx context.Context, threadID, actor, limit, beforeID int) ([]sql.ChatMessage, error) {
	if m.getChatMessagesFn != nil {
		return m.getChatMessagesFn(ctx, threadID, actor, limit, beforeID)
	}
	return nil, nil
}
func (m *mockService) MarkThreadRead(ctx context.Context, threadID, actor int) error {
	if m.markThreadReadFn != nil {
		return m.markThreadReadFn(ctx, threadID, actor)
	}
	return nil
}
func (m *mockService) UpdateChatMessage(ctx context.Context, msgID, actor int, body string) (bool, error) {
	if m.updateChatMessageFn != nil {
		return m.updateChatMessageFn(ctx, msgID, actor, body)
	}
	return true, nil
}
func (m *mockService) DeleteChatMessage(ctx context.Context, msgID, actor int) (bool, error) {
	if m.deleteChatMessageFn != nil {
		return m.deleteChatMessageFn(ctx, msgID, actor)
	}
	return true, nil
}
func (m *mockService) UpsertTypingState(ctx context.Context, threadID, actor int, isTyping bool) error {
	if m.upsertTypingStateFn != nil {
		return m.upsertTypingStateFn(ctx, threadID, actor, isTyping)
	}
	return nil
}
func (m *mockService) GetTypingUsers(ctx context.Context, threadID, actor int) ([]string, error) {
	if m.getTypingUsersFn != nil {
		return m.getTypingUsersFn(ctx, threadID, actor)
	}
	return []string{}, nil
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func withUserID(r *http.Request, userID int) *http.Request {
	ctx := context.WithValue(r.Context(), userIDCtxKey, userID)
	return r.WithContext(ctx)
}

func jsonBody(t *testing.T, v any) *bytes.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return bytes.NewReader(b)
}

// ── HandleHealthz ─────────────────────────────────────────────────────────

func TestHandleHealthz(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	HandleHealthz()(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rr.Code)
	}
	var resp map[string]string
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Errorf("want status=ok, got %q", resp["status"])
	}
}

// ── validatePasswordStrength ──────────────────────────────────────────────

func TestValidatePasswordStrength(t *testing.T) {
	cases := []struct {
		password string
		wantOK   bool
		wantCode string
	}{
		{"Abcdef1", false, "password_too_short"},   // 7 chars
		{"abcdefgh", false, "password_too_weak"},   // no upper, no digit
		{"ABCDEFGH", false, "password_too_weak"},   // no lower, no digit
		{"Abcdefgh", false, "password_too_weak"},   // no digit
		{"Abcdef12", true, ""},
		{"A1bcdefg", true, ""},
		{"Longpassword99!", true, ""},
	}
	for _, tc := range cases {
		ok, code := validatePasswordStrength(tc.password)
		if ok != tc.wantOK {
			t.Errorf("password=%q: want ok=%v, got %v", tc.password, tc.wantOK, ok)
		}
		if code != tc.wantCode {
			t.Errorf("password=%q: want code=%q, got %q", tc.password, tc.wantCode, code)
		}
	}
}

// ── HandleCreateUser ──────────────────────────────────────────────────────

func TestHandleCreateUser_InvalidJSON(t *testing.T) {
	svc := &mockService{}
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader("not-json"))
	rr := httptest.NewRecorder()
	HandleCreateUser(svc)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rr.Code)
	}
}

func TestHandleCreateUser_InvalidEmail(t *testing.T) {
	svc := &mockService{}
	body := jsonBody(t, map[string]string{"email": "notanemail", "password": "Password1"})
	req := httptest.NewRequest(http.MethodPost, "/users", body)
	rr := httptest.NewRecorder()
	HandleCreateUser(svc)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rr.Code)
	}
}

func TestHandleCreateUser_EmptyPassword(t *testing.T) {
	svc := &mockService{}
	body := jsonBody(t, map[string]string{"email": "a@b.com", "password": ""})
	req := httptest.NewRequest(http.MethodPost, "/users", body)
	rr := httptest.NewRecorder()
	HandleCreateUser(svc)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rr.Code)
	}
}

func TestHandleCreateUser_WeakPassword(t *testing.T) {
	svc := &mockService{}
	body := jsonBody(t, map[string]string{"email": "a@b.com", "password": "weakpass"})
	req := httptest.NewRequest(http.MethodPost, "/users", body)
	rr := httptest.NewRecorder()
	HandleCreateUser(svc)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rr.Code)
	}
	var resp map[string]string
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["code"] != "password_too_weak" {
		t.Errorf("want code=password_too_weak, got %q", resp["code"])
	}
}

func TestHandleCreateUser_Success(t *testing.T) {
	os.Setenv("JWT_SECRET", "testsecret")
	defer os.Unsetenv("JWT_SECRET")

	svc := &mockService{
		registerFn: func(_ context.Context, _ sql.User) (int, string, error) {
			return 42, "qa", nil
		},
		getUserJWTVersionFn: func(_ context.Context, _ int) (int, error) {
			return 1, nil
		},
	}
	body := jsonBody(t, map[string]string{"email": "new@user.com", "password": "Password1"})
	req := httptest.NewRequest(http.MethodPost, "/users", body)
	rr := httptest.NewRecorder()
	HandleCreateUser(svc)(rr, req)
	if rr.Code != http.StatusCreated {
		t.Errorf("want 201, got %d — body: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["token"] == nil || resp["token"] == "" {
		t.Error("expected non-empty token in response")
	}
}

// ── HandleGetIdUser (login) ───────────────────────────────────────────────

func TestHandleGetIdUser_InvalidJSON(t *testing.T) {
	svc := &mockService{}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("{bad"))
	rr := httptest.NewRecorder()
	HandleGetIdUser(svc)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rr.Code)
	}
}

func TestHandleGetIdUser_InvalidEmail(t *testing.T) {
	svc := &mockService{}
	body := jsonBody(t, map[string]string{"email": "noemail", "password": "p"})
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	rr := httptest.NewRecorder()
	HandleGetIdUser(svc)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rr.Code)
	}
}

func TestHandleGetIdUser_EmptyPassword(t *testing.T) {
	svc := &mockService{}
	body := jsonBody(t, map[string]string{"email": "a@b.com", "password": ""})
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	rr := httptest.NewRecorder()
	HandleGetIdUser(svc)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rr.Code)
	}
}

// ── HandleGetMe ───────────────────────────────────────────────────────────

func TestHandleGetMe_NoContext(t *testing.T) {
	svc := &mockService{}
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rr := httptest.NewRecorder()
	HandleGetMe(svc)(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", rr.Code)
	}
}

// ── HandleCreateTask ──────────────────────────────────────────────────────

func TestHandleCreateTask_InvalidJSON(t *testing.T) {
	svc := &mockService{}
	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader("{{"))
	req = withUserID(req, 1)
	rr := httptest.NewRecorder()
	HandleCreateTask(svc)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rr.Code)
	}
}

func TestHandleCreateTask_EmptyTitle(t *testing.T) {
	svc := &mockService{}
	body := jsonBody(t, map[string]any{"title": "", "description": "desc", "project_id": 1})
	req := httptest.NewRequest(http.MethodPost, "/tasks", body)
	req = withUserID(req, 1)
	rr := httptest.NewRecorder()
	HandleCreateTask(svc)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rr.Code)
	}
}

func TestHandleCreateTask_EmptyDescription(t *testing.T) {
	svc := &mockService{}
	body := jsonBody(t, map[string]any{"title": "My task", "description": "", "project_id": 1})
	req := httptest.NewRequest(http.MethodPost, "/tasks", body)
	req = withUserID(req, 1)
	rr := httptest.NewRecorder()
	HandleCreateTask(svc)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rr.Code)
	}
}

func TestHandleCreateTask_MissingProjectID(t *testing.T) {
	svc := &mockService{}
	body := jsonBody(t, map[string]any{"title": "T", "description": "D"})
	req := httptest.NewRequest(http.MethodPost, "/tasks", body)
	req = withUserID(req, 1)
	rr := httptest.NewRecorder()
	HandleCreateTask(svc)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rr.Code)
	}
}

func TestHandleCreateTask_Unauthenticated(t *testing.T) {
	svc := &mockService{}
	body := jsonBody(t, map[string]any{"title": "T", "description": "D", "project_id": 1})
	req := httptest.NewRequest(http.MethodPost, "/tasks", body)
	// No userID in context
	rr := httptest.NewRecorder()
	HandleCreateTask(svc)(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", rr.Code)
	}
}

// ── HandleLogoutAllSessions ───────────────────────────────────────────────

func TestHandleLogoutAllSessions_Unauthenticated(t *testing.T) {
	svc := &mockService{}
	req := httptest.NewRequest(http.MethodPost, "/me/logout-all", nil)
	rr := httptest.NewRecorder()
	HandleLogoutAllSessions(svc)(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", rr.Code)
	}
}

func TestHandleLogoutAllSessions_Success(t *testing.T) {
	svc := &mockService{}
	req := httptest.NewRequest(http.MethodPost, "/me/logout-all", nil)
	req = withUserID(req, 5)
	rr := httptest.NewRecorder()
	HandleLogoutAllSessions(svc)(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rr.Code)
	}
}

// ── HandleGetAllTasks ─────────────────────────────────────────────────────

func TestHandleGetAllTasks_MissingProjectID(t *testing.T) {
	svc := &mockService{}
	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	req = withUserID(req, 1)
	rr := httptest.NewRecorder()
	HandleGetAllTasks(svc)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rr.Code)
	}
}

func TestHandleGetAllTasks_Unauthenticated(t *testing.T) {
	svc := &mockService{}
	req := httptest.NewRequest(http.MethodGet, "/tasks?project_id=1", nil)
	rr := httptest.NewRecorder()
	HandleGetAllTasks(svc)(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", rr.Code)
	}
}
