package tests

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	bugserver "bug_tracker/internal/server"
)

func TestBugPhoto_UploadAndServe(t *testing.T) {
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

	// Создаём пользователя, задачу, баг
	resp, raw := doJSON(t, client, srv.URL+"/users", http.MethodPost, apiUserReq{Email: "u@test.local", Password: "pw"})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register: %d %s", resp.StatusCode, raw)
	}
	userID := mustDecode[map[string]int](t, raw)["id"]

	resp, raw = doJSON(t, client, srv.URL+"/tasks", http.MethodPost, apiTaskReq{Title: "t", Description: "d", OwnerId: userID})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create task: %d %s", resp.StatusCode, raw)
	}
	tasks := mustDecode[[]map[string]any](t, raw)
	taskID := int(tasks[0]["id"].(float64))

	resp, raw = doJSON(t, client, srv.URL+"/bugs/"+itoa(taskID), http.MethodPost, apiBugReq{
		Severity: "Low", Priority: "Low", OS: "Win", Status: "Open",
		VersionProduct: "1.0", Description: "d", CreatedBy: userID,
		PlaybackDescription: "steps", ExpectedResult: "exp", ActualResult: "act",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create bug: %d %s", resp.StatusCode, raw)
	}
	bugs := mustDecode[[]bugResp](t, raw)
	bugID := bugs[0].Id

	// GET фото до загрузки — должен вернуть 404
	resp, err := client.Get(srv.URL + "/bugs/" + itoa(bugID) + "/photo")
	if err != nil {
		t.Fatalf("GET photo before upload: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("GET photo before upload: expected 404, got %d", resp.StatusCode)
	}

	// POST фото
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile("photo", "test.png")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	// минимальный 1x1 PNG
	pngBytes := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
		0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41,
		0x54, 0x08, 0xD7, 0x63, 0xF8, 0xCF, 0xC0, 0x00,
		0x00, 0x00, 0x02, 0x00, 0x01, 0xE2, 0x21, 0xBC,
		0x33, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E,
		0x44, 0xAE, 0x42, 0x60, 0x82,
	}
	if _, err := fw.Write(pngBytes); err != nil {
		t.Fatalf("write png: %v", err)
	}
	mw.Close()

	req, err := http.NewRequest(http.MethodPost, srv.URL+"/bugs/"+itoa(bugID)+"/photo", &buf)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("POST photo: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST photo: expected 201, got %d", resp.StatusCode)
	}

	// GET фото после загрузки — должен вернуть 200 и данные
	resp, err = client.Get(srv.URL + "/bugs/" + itoa(bugID) + "/photo")
	if err != nil {
		t.Fatalf("GET photo after upload: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET photo after upload: expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if len(body) == 0 {
		t.Fatal("GET photo: expected non-empty body")
	}
}
