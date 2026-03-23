package tests

import (
	"context"
	"testing"
	"time"

	"bug_tracker/internal/sql"
)

func TestChangeBug_AssignSetsAssignedFields(t *testing.T) {
	pool := mustPool(t)
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := resetSchema(ctx, pool); err != nil {
		t.Fatalf("resetSchema: %v", err)
	}

	creatorID, err := createUser(ctx, pool, "creator@test.local", "pw")
	if err != nil {
		t.Fatalf("createUser creator: %v", err)
	}
	assigneeID, err := createUser(ctx, pool, "assignee@test.local", "pw")
	if err != nil {
		t.Fatalf("createUser assignee: %v", err)
	}

	taskID, err := createTask(ctx, pool, "task", creatorID)
	if err != nil {
		t.Fatalf("createTask: %v", err)
	}

	bugID, err := createBug(ctx, pool, taskID, creatorID)
	if err != nil {
		t.Fatalf("createBug: %v", err)
	}

	beforeAssignedTo, beforeAssignedTime, err := getBugAssigned(ctx, pool, bugID)
	if err == nil && beforeAssignedTo != nil {
		t.Fatalf("expected assigned_to_fk to be NULL before, got %v", *beforeAssignedTo)
	}
	if err == nil && beforeAssignedTime != nil {
		t.Fatalf("expected assigned_time to be NULL before")
	}

	bug := buildBugForUpdate(bugID, taskID, creatorID)
	if err := sql.ChangeBug(ctx, pool, bug, "assignee@test.local"); err != nil {
		t.Fatalf("ChangeBug: %v", err)
	}

	afterAssignedTo, afterAssignedTime, err := getBugAssigned(ctx, pool, bugID)
	if err != nil {
		t.Fatalf("getBugAssigned after: %v", err)
	}
	if afterAssignedTo == nil || *afterAssignedTo != assigneeID {
		t.Fatalf("expected assigned_to_fk=%d, got %v", assigneeID, afterAssignedTo)
	}
	if afterAssignedTime == nil {
		t.Fatalf("expected assigned_time to be not NULL")
	}
}

func TestDeleteTaskAndDeleteBug_AuthorizationByOwnerAndCreator(t *testing.T) {
	pool := mustPool(t)
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := resetSchema(ctx, pool); err != nil {
		t.Fatalf("resetSchema: %v", err)
	}

	ownerID, err := createUser(ctx, pool, "owner@test.local", "pw")
	if err != nil {
		t.Fatalf("createUser owner: %v", err)
	}
	otherID, err := createUser(ctx, pool, "other@test.local", "pw")
	if err != nil {
		t.Fatalf("createUser other: %v", err)
	}

	taskID, err := createTask(ctx, pool, "task", ownerID)
	if err != nil {
		t.Fatalf("createTask: %v", err)
	}

	// Task delete: only owner can delete.
	deleted, err := sql.DeleteTask(ctx, pool, taskID, otherID)
	if err != nil {
		t.Fatalf("DeleteTask wrong owner returned error: %v", err)
	}
	if deleted {
		t.Fatalf("expected DeleteTask to be forbidden for other owner")
	}

	deleted, err = sql.DeleteTask(ctx, pool, taskID, ownerID)
	if err != nil {
		t.Fatalf("DeleteTask owner returned error: %v", err)
	}
	if !deleted {
		t.Fatalf("expected DeleteTask to succeed for owner")
	}

	taskID, err = createTask(ctx, pool, "task2", ownerID)
	if err != nil {
		t.Fatalf("createTask2: %v", err)
	}

	bugID, err := createBug(ctx, pool, taskID, ownerID)
	if err != nil {
		t.Fatalf("createBug: %v", err)
	}

	deletedBug, err := sql.DeleteBug(ctx, pool, bugID, otherID)
	if err != nil {
		t.Fatalf("DeleteBug wrong creator returned error: %v", err)
	}
	if deletedBug {
		t.Fatalf("expected DeleteBug to be forbidden for other creator")
	}

	deletedBug, err = sql.DeleteBug(ctx, pool, bugID, ownerID)
	if err != nil {
		t.Fatalf("DeleteBug owner returned error: %v", err)
	}
	if !deletedBug {
		t.Fatalf("expected DeleteBug to succeed for creator")
	}
}
