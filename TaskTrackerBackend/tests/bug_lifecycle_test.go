package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	bugserver "bug_tracker/internal/server"
)

// TestBugLifecycle_PassAndAccept проверяет полный жизненный цикл бага:
// создатель назначает исполнителя → исполнитель сдаёт → создатель принимает.
func TestBugLifecycle_PassAndAccept(t *testing.T) {
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

	// Регистрируем создателя и исполнителя
	resp, raw := doJSON(t, client, srv.URL+"/users", http.MethodPost, apiUserReq{Email: "creator@test.local", Password: "pw"})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register creator: %d %s", resp.StatusCode, raw)
	}
	creatorID := mustDecode[map[string]int](t, raw)["id"]

	resp, raw = doJSON(t, client, srv.URL+"/users", http.MethodPost, apiUserReq{Email: "assignee@test.local", Password: "pw"})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register assignee: %d %s", resp.StatusCode, raw)
	}
	assigneeID := mustDecode[map[string]int](t, raw)["id"]

	// Создаём задачу и баг
	resp, raw = doJSON(t, client, srv.URL+"/tasks", http.MethodPost, apiTaskReq{Title: "t", Description: "d", OwnerId: creatorID})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create task: %d %s", resp.StatusCode, raw)
	}
	tasks := mustDecode[[]map[string]any](t, raw)
	taskID := int(tasks[0]["id"].(float64))

	resp, raw = doJSON(t, client, srv.URL+"/bugs/"+itoa(taskID), http.MethodPost, apiBugReq{
		Severity: "Low", Priority: "Low", OS: "Win", Status: "Open",
		VersionProduct: "1.0", Description: "d", CreatedBy: creatorID,
		PlaybackDescription: "steps", ExpectedResult: "exp", ActualResult: "act",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create bug: %d %s", resp.StatusCode, raw)
	}
	bugs := mustDecode[[]bugResp](t, raw)
	bugID := bugs[0].Id

	basePatch := map[string]any{
		"severity": "Low", "priority": "Low", "os": "Win", "status": "Open",
		"version_product": "1.0", "description": "d", "created_by": creatorID,
		"playback_description": "steps", "expected_result": "exp", "actual_result": "act",
		"task_id": taskID,
	}

	// Шаг 1: создатель назначает исполнителя
	patch1 := copyMap(basePatch)
	patch1["assigned_to_email"] = "assignee@test.local"
	resp, raw = doJSON(t, client, srv.URL+"/bugs/"+itoa(bugID), http.MethodPatch, patch1)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("assign bug: expected 200, got %d %s", resp.StatusCode, raw)
	}
	assigned := mustDecode[[]bugResp](t, raw)
	if assigned[0].AssignedTo == nil || *assigned[0].AssignedTo != assigneeID {
		t.Fatalf("expected assigned_to=%d, got %v", assigneeID, assigned[0].AssignedTo)
	}

	// Шаг 2: исполнитель сдаёт баг (passed_by = assigneeID)
	now := time.Now()
	patch2 := copyMap(basePatch)
	patch2["assigned_to"] = assigneeID
	patch2["passed_by"] = assigneeID
	patch2["passed_time"] = now.Format(time.RFC3339)
	resp, raw = doJSON(t, client, srv.URL+"/bugs/"+itoa(bugID), http.MethodPatch, patch2)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("pass bug: expected 200, got %d %s", resp.StatusCode, raw)
	}
	passed := mustDecode[[]bugResp](t, raw)
	if passed[0].PassedBy == nil || *passed[0].PassedBy != assigneeID {
		t.Fatalf("expected passed_by=%d, got %v", assigneeID, passed[0].PassedBy)
	}
	if passed[0].PassedTime == nil {
		t.Fatalf("expected passed_time not null")
	}

	// Шаг 3: создатель принимает баг (accepted_by = creatorID)
	patch3 := copyMap(basePatch)
	patch3["assigned_to"] = assigneeID
	patch3["passed_by"] = assigneeID
	patch3["passed_time"] = now.Format(time.RFC3339)
	patch3["accepted_by"] = creatorID
	patch3["accepted_time"] = now.Format(time.RFC3339)
	resp, raw = doJSON(t, client, srv.URL+"/bugs/"+itoa(bugID), http.MethodPatch, patch3)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("accept bug: expected 200, got %d %s", resp.StatusCode, raw)
	}
	accepted := mustDecode[[]bugResp](t, raw)
	if accepted[0].AcceptedBy == nil || *accepted[0].AcceptedBy != creatorID {
		t.Fatalf("expected accepted_by=%d, got %v", creatorID, accepted[0].AcceptedBy)
	}
	if accepted[0].AcceptedTime == nil {
		t.Fatalf("expected accepted_time not null")
	}
}

func copyMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
