package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	bugserver "bug_tracker/internal/server"
)

func TestRegister_DuplicateEmail(t *testing.T) {
	pool := mustPool(t)
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := resetSchema(ctx, pool); err != nil {
		t.Fatalf("resetSchema: %v", err)
	}

	srv := httptest.NewServer(bugserver.NewRouter(ctx, pool))
	defer srv.Close()
	client := srv.Client()

	user := apiUserReq{Email: "dup@test.local", Password: "pw"}

	resp, _ := doJSON(t, client, srv.URL+"/users", http.MethodPost, user)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("first register: expected 201, got %d", resp.StatusCode)
	}

	resp, _ = doJSON(t, client, srv.URL+"/users", http.MethodPost, user)
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("duplicate register: expected 409, got %d", resp.StatusCode)
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	pool := mustPool(t)
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := resetSchema(ctx, pool); err != nil {
		t.Fatalf("resetSchema: %v", err)
	}

	srv := httptest.NewServer(bugserver.NewRouter(ctx, pool))
	defer srv.Close()
	client := srv.Client()

	resp, _ := doJSON(t, client, srv.URL+"/users", http.MethodPost, apiUserReq{Email: "notanemail", Password: "pw"})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid email: expected 400, got %d", resp.StatusCode)
	}
}

func TestRegister_EmptyPassword(t *testing.T) {
	pool := mustPool(t)
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := resetSchema(ctx, pool); err != nil {
		t.Fatalf("resetSchema: %v", err)
	}

	srv := httptest.NewServer(bugserver.NewRouter(ctx, pool))
	defer srv.Close()
	client := srv.Client()

	resp, _ := doJSON(t, client, srv.URL+"/users", http.MethodPost, apiUserReq{Email: "valid@test.local", Password: ""})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("empty password: expected 400, got %d", resp.StatusCode)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	pool := mustPool(t)
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := resetSchema(ctx, pool); err != nil {
		t.Fatalf("resetSchema: %v", err)
	}

	srv := httptest.NewServer(bugserver.NewRouter(ctx, pool))
	defer srv.Close()
	client := srv.Client()

	user := apiUserReq{Email: "user@test.local", Password: "correct"}
	resp, _ := doJSON(t, client, srv.URL+"/users", http.MethodPost, user)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register: expected 201, got %d", resp.StatusCode)
	}

	resp, _ = doJSON(t, client, srv.URL+"/login", http.MethodPost, apiUserReq{Email: "user@test.local", Password: "wrong"})
	if resp.StatusCode == http.StatusOK {
		t.Fatalf("wrong password: expected non-200, got 200")
	}
}

func TestLogin_NonexistentUser(t *testing.T) {
	pool := mustPool(t)
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := resetSchema(ctx, pool); err != nil {
		t.Fatalf("resetSchema: %v", err)
	}

	srv := httptest.NewServer(bugserver.NewRouter(ctx, pool))
	defer srv.Close()
	client := srv.Client()

	resp, _ := doJSON(t, client, srv.URL+"/login", http.MethodPost, apiUserReq{Email: "nobody@test.local", Password: "pw"})
	if resp.StatusCode == http.StatusOK {
		t.Fatalf("nonexistent user login: expected non-200, got 200")
	}
}
