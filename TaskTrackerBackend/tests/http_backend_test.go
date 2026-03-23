package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	bugserver "bug_tracker/internal/server"
)

type apiUserReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type apiTaskReq struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	OwnerId     int    `json:"owner_id"`
}

type apiBugReq struct {
	TaskId int `json:"task_id,omitempty"`

	Severity string `json:"severity"`
	Priority string `json:"priority"`
	OS       string `json:"os"`
	Status   string `json:"status"`

	VersionProduct string `json:"version_product"`
	Description    string `json:"description"`

	CreatedBy   int    `json:"created_by"`
	CreatedTime string `json:"created_time,omitempty"`

	AssignedTo   *int       `json:"assigned_to,omitempty"`
	AssignedTime *time.Time `json:"assigned_time,omitempty"`

	PlaybackDescription string `json:"playback_description"`
	ExpectedResult      string `json:"expected_result"`
	ActualResult        string `json:"actual_result"`

	PassedBy   *int       `json:"passed_by,omitempty"`
	PassedTime *time.Time `json:"passed_time,omitempty"`

	AcceptedBy   *int       `json:"accepted_by,omitempty"`
	AcceptedTime *time.Time `json:"accepted_time,omitempty"`
}

type bugResp struct {
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

func doJSON(t *testing.T, client *http.Client, url, method string, body any) (*http.Response, json.RawMessage) {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("json.Marshal: %v", err)
		}
		buf = *bytes.NewBuffer(b)
	}

	var reader io.Reader
	if body != nil {
		reader = &buf
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("client.Do: %v", err)
	}
	defer resp.Body.Close()

	var raw json.RawMessage
	dec := json.NewDecoder(resp.Body)
	dec.DisallowUnknownFields()
	_ = dec.Decode(&raw) // for empty bodies

	return resp, raw
}

func mustDecode[T any](t *testing.T, raw json.RawMessage) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatalf("json.Unmarshal: %v, raw=%s", err, string(raw))
	}
	return v
}

func TestBackend_AllEndpointsHappyPath(t *testing.T) {
	pool := mustPool(t)
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := resetSchema(ctx, pool); err != nil {
		t.Fatalf("resetSchema: %v", err)
	}

	srv := httptest.NewServer(bugserver.NewRouter(ctx, pool))
	defer srv.Close()

	client := srv.Client()

	// Register two users
	creatorReq := apiUserReq{Email: "creator@test.local", Password: "pw"}
	assigneeReq := apiUserReq{Email: "assignee@test.local", Password: "pw"}

	resp, raw := doJSON(t, client, srv.URL+"/users", http.MethodPost, creatorReq)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /users expected 201, got %d (%s)", resp.StatusCode, string(raw))
	}
	creatorID := mustDecode[map[string]int](t, raw)["id"]

	resp, raw = doJSON(t, client, srv.URL+"/users", http.MethodPost, assigneeReq)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /users expected 201, got %d (%s)", resp.StatusCode, string(raw))
	}
	assigneeID := mustDecode[map[string]int](t, raw)["id"]

	// Login creator
	resp, raw = doJSON(t, client, srv.URL+"/login", http.MethodPost, creatorReq)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /login expected 200, got %d (%s)", resp.StatusCode, string(raw))
	}
	loginID := mustDecode[map[string]int](t, raw)["id"]
	if loginID != creatorID {
		t.Fatalf("login id mismatch: want=%d got=%d", creatorID, loginID)
	}

	// GET other emails for creator
	resp, raw = doJSON(t, client, srv.URL+"/users/999999", http.MethodGet, nil)
	_ = resp
	// Note: /users/{id} returns all emails except that id, but our handler expects id parsing.
	// We test the happy path by calling with an existing id.
	resp, raw = doJSON(t, client, srv.URL+"/users/"+itoa(creatorID), http.MethodGet, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /users/{id} expected 200, got %d (%s)", resp.StatusCode, string(raw))
	}
	otherEmails := mustDecode[[]string](t, raw)
	if len(otherEmails) != 1 {
		t.Fatalf("expected 1 other email, got %d (%v)", len(otherEmails), otherEmails)
	}

	// Create task (creator is owner)
	taskReq := apiTaskReq{Title: "t1", Description: "d1", OwnerId: creatorID}
	resp, raw = doJSON(t, client, srv.URL+"/tasks", http.MethodPost, taskReq)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /tasks expected 201, got %d (%s)", resp.StatusCode, string(raw))
	}
	tasks := mustDecode[[]map[string]any](t, raw)
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d (%v)", len(tasks), tasks)
	}
	taskID := int(tasks[0]["id"].(float64))

	// Create bug under task
	bug := apiBugReq{
		Severity:            "Low",
		Priority:            "Low",
		OS:                  "Win",
		Status:              "Open",
		VersionProduct:      "1.0.0",
		Description:         "bug desc",
		CreatedBy:           creatorID,
		PlaybackDescription: "steps",
		ExpectedResult:      "expected",
		ActualResult:        "actual",
	}
	resp, raw = doJSON(t, client, srv.URL+"/bugs/"+itoa(taskID), http.MethodPost, bug)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /bugs/{id} expected 201, got %d (%s)", resp.StatusCode, string(raw))
	}
	bugs := mustDecode[[]bugResp](t, raw)
	if len(bugs) != 1 {
		t.Fatalf("expected 1 bug, got %d", len(bugs))
	}
	bugID := bugs[0].Id

	// GET bugs for task
	resp, raw = doJSON(t, client, srv.URL+"/bugs/"+itoa(taskID), http.MethodGet, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /bugs/{id} expected 200, got %d (%s)", resp.StatusCode, string(raw))
	}
	gotBugs := mustDecode[[]bugResp](t, raw)
	if len(gotBugs) != 1 {
		t.Fatalf("expected 1 bug on GET, got %d", len(gotBugs))
	}

	// PATCH bug: assign to assignee by email
	patchBody := bug
	patchBody.Severity = "Medium"
	patchBody.Priority = "High"
	bodyForPatch := map[string]any{
		"severity":             patchBody.Severity,
		"priority":             patchBody.Priority,
		"os":                   patchBody.OS,
		"status":               patchBody.Status,
		"version_product":      patchBody.VersionProduct,
		"description":          patchBody.Description,
		"playback_description": patchBody.PlaybackDescription,
		"expected_result":      patchBody.ExpectedResult,
		"actual_result":        patchBody.ActualResult,
		"assigned_to_email":    "assignee@test.local",
		// created_by fields are not required by ChangeBug UPDATE, but handler decodes sql.Bug.
		"created_by": creatorID,
		"task_id":    taskID,
	}
	resp, raw = doJSON(t, client, srv.URL+"/bugs/"+itoa(bugID), http.MethodPatch, bodyForPatch)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("PATCH /bugs/{id} expected 200, got %d (%s)", resp.StatusCode, string(raw))
	}
	_ = mustDecode[[]bugResp](t, raw)

	// Verify assignment persisted
	resp, raw = doJSON(t, client, srv.URL+"/bugs/"+itoa(taskID), http.MethodGet, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /bugs/{id} after PATCH expected 200, got %d (%s)", resp.StatusCode, string(raw))
	}
	afterBugs := mustDecode[[]bugResp](t, raw)
	if afterBugs[0].AssignedTo == nil || *afterBugs[0].AssignedTo != assigneeID {
		t.Fatalf("expected assigned_to_fk=%d, got %v", assigneeID, afterBugs[0].AssignedTo)
	}
	if afterBugs[0].AssignedTime == nil {
		t.Fatalf("expected assigned_time not null")
	}

	// DELETE bug (wrong creator should forbid)
	resp, raw = doJSON(t, client, srv.URL+"/bugs/"+itoa(bugID), http.MethodDelete, map[string]any{"created_by": 123456})
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("DELETE /bugs expected 403 for wrong creator, got %d (%s)", resp.StatusCode, string(raw))
	}

	// DELETE bug (correct creator)
	resp, raw = doJSON(t, client, srv.URL+"/bugs/"+itoa(bugID), http.MethodDelete, map[string]any{"created_by": creatorID})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("DELETE /bugs expected 200, got %d (%s)", resp.StatusCode, string(raw))
	}

	resp, raw = doJSON(t, client, srv.URL+"/bugs/"+itoa(taskID), http.MethodGet, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /bugs after delete expected 200, got %d (%s)", resp.StatusCode, string(raw))
	}
	afterDeleteBugs := mustDecode[[]bugResp](t, raw)
	if len(afterDeleteBugs) != 0 {
		t.Fatalf("expected 0 bugs after delete, got %d", len(afterDeleteBugs))
	}

	// DELETE task (wrong owner forbidden)
	resp, raw = doJSON(t, client, srv.URL+"/tasks/"+itoa(taskID), http.MethodDelete, map[string]any{"owner_id": assigneeID})
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("DELETE /tasks expected 403 for wrong owner, got %d (%s)", resp.StatusCode, string(raw))
	}

	// DELETE task (correct owner)
	resp, raw = doJSON(t, client, srv.URL+"/tasks/"+itoa(taskID), http.MethodDelete, map[string]any{"owner_id": creatorID})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("DELETE /tasks expected 200, got %d (%s)", resp.StatusCode, string(raw))
	}

	resp, raw = doJSON(t, client, srv.URL+"/tasks", http.MethodGet, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /tasks expected 200, got %d (%s)", resp.StatusCode, string(raw))
	}
	tasksAfter := mustDecode[[]map[string]any](t, raw)
	if len(tasksAfter) != 0 {
		t.Fatalf("expected 0 tasks after delete, got %d", len(tasksAfter))
	}
}

func itoa(v int) string { return jsonNumberString(v) }

func jsonNumberString(v int) string {
	if v == 0 {
		return "0"
	}
	sign := ""
	if v < 0 {
		sign = "-"
		v = -v
	}
	var b [32]byte
	i := len(b)
	for v > 0 {
		i--
		b[i] = byte('0' + v%10)
		v /= 10
	}
	return sign + string(b[i:])
}
