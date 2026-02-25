package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func resetGroups() {
	iamGroupsMu.Lock()
	iamGroups = map[string]iamGroup{}
	iamGroupsMu.Unlock()
}

func TestIAMPersistenceAcrossRouterRebuild(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "iam-persistence.sqlite")
	t.Setenv("KUBEDECK_IAM_PERSIST_IN_TEST", "1")
	t.Setenv("KUBEDECK_SQLITE_DSN", dbPath)

	resetIAMPersistenceForTest()
	resetAuthSessions()
	resetInvites()
	resetMemberships()
	resetGroups()

	router := NewRouter()

	loginPayload := []byte(`{"username":"admin","password":"pw","tenant_code":"dev"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginPayload))
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("expected login status %d, got %d", http.StatusOK, loginResp.Code)
	}
	var loginBody struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("expected login response json, got error: %v", err)
	}
	if loginBody.Token == "" {
		t.Fatalf("expected login token")
	}

	createGroupPayload := []byte(`{"name":"ops","description":"ops team"}`)
	createGroupReq := httptest.NewRequest(http.MethodPost, "/api/iam/groups", bytes.NewReader(createGroupPayload))
	createGroupReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	createGroupResp := httptest.NewRecorder()
	router.ServeHTTP(createGroupResp, createGroupReq)
	if createGroupResp.Code != http.StatusCreated {
		t.Fatalf("expected group create status %d, got %d body=%s", http.StatusCreated, createGroupResp.Code, createGroupResp.Body.String())
	}

	resetAuthSessions()
	resetInvites()
	resetMemberships()
	resetGroups()
	resetIAMPersistenceForTest()

	router = NewRouter()
	loginReq2 := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginPayload))
	loginResp2 := httptest.NewRecorder()
	router.ServeHTTP(loginResp2, loginReq2)
	if loginResp2.Code != http.StatusOK {
		t.Fatalf("expected second login status %d, got %d", http.StatusOK, loginResp2.Code)
	}
	var loginBody2 struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(loginResp2.Body.Bytes(), &loginBody2); err != nil {
		t.Fatalf("expected second login response json, got error: %v", err)
	}
	if loginBody2.Token == "" {
		t.Fatalf("expected second login token")
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/iam/groups", nil)
	listReq.Header.Set("Authorization", "Bearer "+loginBody2.Token)
	listResp := httptest.NewRecorder()
	router.ServeHTTP(listResp, listReq)
	if listResp.Code != http.StatusOK {
		t.Fatalf("expected list groups status %d, got %d body=%s", http.StatusOK, listResp.Code, listResp.Body.String())
	}
	var listBody struct {
		Groups []iamGroup `json:"groups"`
	}
	if err := json.Unmarshal(listResp.Body.Bytes(), &listBody); err != nil {
		t.Fatalf("expected list groups json, got error: %v", err)
	}
	found := false
	for _, group := range listBody.Groups {
		if group.Name == "ops" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected persisted group in list, body=%s", listResp.Body.String())
	}

	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("expected sqlite file at %s: %v", dbPath, err)
	}
}
