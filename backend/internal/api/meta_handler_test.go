package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"kubedeck/backend/internal/auth"
	"kubedeck/backend/internal/core/audit"
)

type registryResponse struct {
	Cluster       string `json:"cluster"`
	ResourceTypes []struct {
		ID               string `json:"id"`
		Group            string `json:"group"`
		Version          string `json:"version"`
		Kind             string `json:"kind"`
		Plural           string `json:"plural"`
		Namespaced       bool   `json:"namespaced"`
		PreferredVersion string `json:"preferredVersion"`
		Source           string `json:"source"`
	} `json:"resourceTypes"`
}

func resetAuthSessions() {
	authSessionsMu.Lock()
	authSessions = map[string]authSession{}
	authSessionsMu.Unlock()
	resetGroups()
}

func resetInvites() {
	invitesMu.Lock()
	invites = map[string]iamInvite{}
	invitesMu.Unlock()
}

func resetMemberships() {
	iamMembershipsMu.Lock()
	iamMemberships = map[string]iamMembership{}
	iamMembershipsMu.Unlock()
}

func resetGroups() {
	iamGroupsMu.Lock()
	iamGroups = map[string]iamGroup{}
	iamGroupsMu.Unlock()
}

func resetAuditWriter() {
	defaultAuditWriter = audit.NewMemoryWriter()
}

type menusResponse struct {
	Cluster string `json:"cluster"`
	Menus   []struct {
		ID         string `json:"id"`
		Group      string `json:"group"`
		Title      string `json:"title"`
		TargetType string `json:"targetType"`
		TargetRef  string `json:"targetRef"`
		Source     string `json:"source"`
		Order      int    `json:"order"`
		Visible    bool   `json:"visible"`
	} `json:"menus"`
}

func TestRegistryEndpoint(t *testing.T) {
	resetAuditWriter()
	router := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/meta/registry?cluster=dev", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}

	var body registryResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected JSON response, got error: %v", err)
	}

	if body.Cluster != "dev" {
		t.Fatalf("expected cluster dev, got %q", body.Cluster)
	}
	if len(body.ResourceTypes) == 0 {
		t.Fatalf("expected non-empty resourceTypes, body=%s", resp.Body.String())
	}
	if body.ResourceTypes[0].ID == "" || body.ResourceTypes[0].Kind == "" {
		t.Fatalf("expected typed resource fields, body=%s", resp.Body.String())
	}
}

func TestClustersEndpoint(t *testing.T) {
	router := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/meta/clusters", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected JSON response, got error: %v", err)
	}

	if _, ok := body["clusters"]; !ok {
		t.Fatalf("expected response to contain clusters key, body=%s", resp.Body.String())
	}
}

func TestMenusEndpoint(t *testing.T) {
	router := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/meta/menus?cluster=dev", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}

	var body menusResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected JSON response, got error: %v", err)
	}

	if body.Cluster != "dev" {
		t.Fatalf("expected cluster dev, got %q", body.Cluster)
	}
	if len(body.Menus) == 0 {
		t.Fatalf("expected non-empty menus, body=%s", resp.Body.String())
	}
	first := body.Menus[0]
	if first.ID == "" || first.TargetType == "" || first.TargetRef == "" {
		t.Fatalf("expected typed menu fields, body=%s", resp.Body.String())
	}
}

func TestResourceApplyEndpoint(t *testing.T) {
	resetAuthSessions()
	resetAuditWriter()
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
		t.Fatalf("expected login JSON body, got error: %v", err)
	}
	if loginBody.Token == "" {
		t.Fatalf("expected token in login response")
	}

	body := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-ok
---
apiVersion: v1
kind: Service
metadata:
  name: svc-fail
---
kind: Broken
`
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/resources/apply?cluster=dev&defaultNs=kube-system",
		strings.NewReader(body),
	)
	req.Header.Set("Authorization", "Bearer "+loginBody.Token)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}

	var payload struct {
		Status    string `json:"status"`
		Cluster   string `json:"cluster"`
		DefaultNS string `json:"defaultNamespace"`
		Total     int    `json:"total"`
		Succeeded int    `json:"succeeded"`
		Failed    int    `json:"failed"`
		Results   []struct {
			Index     int    `json:"index"`
			Kind      string `json:"kind"`
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
			Status    string `json:"status"`
			Reason    string `json:"reason"`
		} `json:"results"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON response, got error: %v", err)
	}

	if payload.Status != "partial" {
		t.Fatalf("expected status partial, got %q body=%s", payload.Status, resp.Body.String())
	}
	if payload.Cluster != "dev" {
		t.Fatalf("expected cluster dev, got %q", payload.Cluster)
	}
	if payload.DefaultNS != "kube-system" {
		t.Fatalf("expected defaultNamespace kube-system, got %q", payload.DefaultNS)
	}
	if payload.Total != 3 || payload.Succeeded != 1 || payload.Failed != 2 {
		t.Fatalf("unexpected counters total=%d succeeded=%d failed=%d", payload.Total, payload.Succeeded, payload.Failed)
	}
	if len(payload.Results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(payload.Results))
	}
	if payload.Results[0].Namespace != "kube-system" {
		t.Fatalf("expected default namespace injected, got %q", payload.Results[0].Namespace)
	}
	if payload.Results[1].Status != "failed" {
		t.Fatalf("expected second result failed, got %q", payload.Results[1].Status)
	}
	if payload.Results[2].Status != "failed" {
		t.Fatalf("expected third result failed, got %q", payload.Results[2].Status)
	}
}

func TestResourceApplyEndpointRequiresAuth(t *testing.T) {
	resetAuthSessions()
	resetAuditWriter()
	router := NewRouter()

	req := httptest.NewRequest(http.MethodPost, "/api/resources/apply?cluster=dev", strings.NewReader("kind: ConfigMap"))
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, resp.Code)
	}
}

func TestResourceApplyEndpointRejectsViewer(t *testing.T) {
	resetAuthSessions()
	resetAuditWriter()
	router := NewRouter()

	loginPayload := []byte(`{"username":"viewer","password":"pw","tenant_code":"dev"}`)
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
		t.Fatalf("expected login JSON body, got error: %v", err)
	}
	if loginBody.Token == "" {
		t.Fatalf("expected token in login response")
	}

	req := httptest.NewRequest(http.MethodPost, "/api/resources/apply?cluster=dev", strings.NewReader("kind: ConfigMap"))
	req.Header.Set("Authorization", "Bearer "+loginBody.Token)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusForbidden, resp.Code, resp.Body.String())
	}
}

func TestHealthzEndpoint(t *testing.T) {
	router := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}
	if resp.Body.String() != "ok" {
		t.Fatalf("expected body ok, got %q", resp.Body.String())
	}
}

func TestReadyzEndpoint(t *testing.T) {
	router := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/readyz", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}
	if resp.Body.String() != "ok" {
		t.Fatalf("expected body ok, got %q", resp.Body.String())
	}
}

func TestRegistryMethodNotAllowed(t *testing.T) {
	router := NewRouter()

	req := httptest.NewRequest(http.MethodPost, "/api/meta/registry", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, resp.Code)
	}

	if allow := resp.Header().Get("Allow"); allow != http.MethodGet {
		t.Fatalf("expected Allow header %q, got %q", http.MethodGet, allow)
	}

	var body map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected JSON error response, got error: %v", err)
	}

	if _, ok := body["error"]; !ok {
		t.Fatalf("expected response to contain error key, body=%s", resp.Body.String())
	}
}

func TestAuthLoginMeSwitchLogoutFlow(t *testing.T) {
	resetAuthSessions()
	resetAuditWriter()
	router := NewRouter()

	loginPayload := []byte(`{"username":"alice","password":"pw","tenant_code":"staging"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginPayload))
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("expected login status 200, got %d body=%s", loginResp.Code, loginResp.Body.String())
	}

	var loginBody struct {
		Token          string `json:"token"`
		ActiveTenantID string `json:"active_tenant_id"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("unmarshal login response: %v", err)
	}
	if loginBody.Token == "" {
		t.Fatalf("expected token in login response")
	}
	if loginBody.ActiveTenantID != "tenant-staging" {
		t.Fatalf("expected active tenant tenant-staging, got %q", loginBody.ActiveTenantID)
	}

	meReq := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	meResp := httptest.NewRecorder()
	router.ServeHTTP(meResp, meReq)
	if meResp.Code != http.StatusOK {
		t.Fatalf("expected me status 200, got %d body=%s", meResp.Code, meResp.Body.String())
	}

	switchPayload := []byte(`{"tenant_code":"prod"}`)
	switchReq := httptest.NewRequest(http.MethodPost, "/api/auth/switch-tenant", bytes.NewReader(switchPayload))
	switchReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	switchResp := httptest.NewRecorder()
	router.ServeHTTP(switchResp, switchReq)
	if switchResp.Code != http.StatusOK {
		t.Fatalf("expected switch status 200, got %d body=%s", switchResp.Code, switchResp.Body.String())
	}
	var switchBody struct {
		ActiveTenantID string `json:"active_tenant_id"`
	}
	if err := json.Unmarshal(switchResp.Body.Bytes(), &switchBody); err != nil {
		t.Fatalf("unmarshal switch response: %v", err)
	}
	if switchBody.ActiveTenantID != "tenant-prod" {
		t.Fatalf("expected switched tenant tenant-prod, got %q", switchBody.ActiveTenantID)
	}

	logoutReq := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	logoutReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	logoutResp := httptest.NewRecorder()
	router.ServeHTTP(logoutResp, logoutReq)
	if logoutResp.Code != http.StatusOK {
		t.Fatalf("expected logout status 200, got %d", logoutResp.Code)
	}

	meAfterLogoutReq := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	meAfterLogoutReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	meAfterLogoutResp := httptest.NewRecorder()
	router.ServeHTTP(meAfterLogoutResp, meAfterLogoutReq)
	if meAfterLogoutResp.Code != http.StatusUnauthorized {
		t.Fatalf("expected me-after-logout status 401, got %d", meAfterLogoutResp.Code)
	}
}

func TestAuthMeReloadsSessionFromPersistenceWhenCacheMiss(t *testing.T) {
	resetIAMPersistenceForTest()
	t.Cleanup(func() {
		resetIAMPersistenceForTest()
	})
	t.Setenv("KUBEDECK_IAM_PERSIST_IN_TEST", "1")
	t.Setenv("KUBEDECK_SQLITE_DSN", filepath.Join(t.TempDir(), "session-reload.sqlite"))

	resetAuthSessions()
	resetAuditWriter()
	router := NewRouter()

	loginPayload := []byte(`{"username":"alice","password":"pw","tenant_code":"dev"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginPayload))
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("expected login status 200, got %d body=%s", loginResp.Code, loginResp.Body.String())
	}
	var loginBody struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("unmarshal login response: %v", err)
	}
	if loginBody.Token == "" {
		t.Fatalf("expected token in login response")
	}

	authSessionsMu.Lock()
	authSessions = map[string]authSession{}
	authSessionsMu.Unlock()

	meReq := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	meResp := httptest.NewRecorder()
	router.ServeHTTP(meResp, meReq)
	if meResp.Code != http.StatusOK {
		t.Fatalf("expected me status 200 after persistence reload, got %d body=%s", meResp.Code, meResp.Body.String())
	}
}

func TestIAMGroupsReloadFromPersistenceWhenCacheMiss(t *testing.T) {
	resetIAMPersistenceForTest()
	t.Cleanup(func() {
		resetIAMPersistenceForTest()
	})
	t.Setenv("KUBEDECK_IAM_PERSIST_IN_TEST", "1")
	t.Setenv("KUBEDECK_SQLITE_DSN", filepath.Join(t.TempDir(), "groups-reload.sqlite"))

	resetAuthSessions()
	resetGroups()
	resetAuditWriter()
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
		t.Fatalf("expected login JSON body, got error: %v", err)
	}

	createReq := httptest.NewRequest(http.MethodPost, "/api/iam/groups", bytes.NewReader([]byte(`{"name":"ops","description":"Ops Team"}`)))
	createReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	createResp := httptest.NewRecorder()
	router.ServeHTTP(createResp, createReq)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d body=%s", http.StatusCreated, createResp.Code, createResp.Body.String())
	}

	resetGroups()

	listReq := httptest.NewRequest(http.MethodGet, "/api/iam/groups", nil)
	listReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	listResp := httptest.NewRecorder()
	router.ServeHTTP(listResp, listReq)
	if listResp.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d body=%s", http.StatusOK, listResp.Code, listResp.Body.String())
	}
	var listBody struct {
		Groups []iamGroup `json:"groups"`
	}
	if err := json.Unmarshal(listResp.Body.Bytes(), &listBody); err != nil {
		t.Fatalf("expected list JSON body, got error: %v", err)
	}
	if len(listBody.Groups) == 0 || listBody.Groups[0].Name != "ops" {
		t.Fatalf("expected persisted groups after reload, body=%s", listResp.Body.String())
	}
}

func TestIAMTenantMembersReloadFromPersistenceWhenCacheMiss(t *testing.T) {
	resetIAMPersistenceForTest()
	t.Cleanup(func() {
		resetIAMPersistenceForTest()
	})
	t.Setenv("KUBEDECK_IAM_PERSIST_IN_TEST", "1")
	t.Setenv("KUBEDECK_SQLITE_DSN", filepath.Join(t.TempDir(), "members-reload.sqlite"))

	resetAuthSessions()
	resetMemberships()
	resetAuditWriter()
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
		t.Fatalf("expected login JSON body, got error: %v", err)
	}

	createMemberReq := httptest.NewRequest(http.MethodPost, "/api/iam/tenants/tenant-dev/members", bytes.NewReader([]byte(`{"user_id":"u-9","user_label":"user-nine","effective_from":"2026-02-25T00:00:00Z"}`)))
	createMemberReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	createMemberResp := httptest.NewRecorder()
	router.ServeHTTP(createMemberResp, createMemberReq)
	if createMemberResp.Code != http.StatusCreated {
		t.Fatalf("expected member create status %d, got %d body=%s", http.StatusCreated, createMemberResp.Code, createMemberResp.Body.String())
	}

	resetMemberships()

	listMembersReq := httptest.NewRequest(http.MethodGet, "/api/iam/tenants/tenant-dev/members", nil)
	listMembersReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	listMembersResp := httptest.NewRecorder()
	router.ServeHTTP(listMembersResp, listMembersReq)
	if listMembersResp.Code != http.StatusOK {
		t.Fatalf("expected list members status %d, got %d body=%s", http.StatusOK, listMembersResp.Code, listMembersResp.Body.String())
	}
	var membersBody struct {
		Members []iamMembership `json:"members"`
	}
	if err := json.Unmarshal(listMembersResp.Body.Bytes(), &membersBody); err != nil {
		t.Fatalf("expected members JSON body, got error: %v", err)
	}
	if len(membersBody.Members) == 0 || membersBody.Members[0].UserID != "u-9" {
		t.Fatalf("expected persisted tenant members after reload, body=%s", listMembersResp.Body.String())
	}
}

func TestIAMUsersReloadFromPersistenceWhenCacheMiss(t *testing.T) {
	resetIAMPersistenceForTest()
	t.Cleanup(func() {
		resetIAMPersistenceForTest()
	})
	t.Setenv("KUBEDECK_IAM_PERSIST_IN_TEST", "1")
	t.Setenv("KUBEDECK_SQLITE_DSN", filepath.Join(t.TempDir(), "users-reload.sqlite"))

	resetAuthSessions()
	resetMemberships()
	resetAuditWriter()
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
		t.Fatalf("expected login JSON body, got error: %v", err)
	}

	createMemberReq := httptest.NewRequest(http.MethodPost, "/api/iam/tenants/tenant-dev/members", bytes.NewReader([]byte(`{"user_id":"u-7","user_label":"user-seven","effective_from":"2026-02-25T00:00:00Z"}`)))
	createMemberReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	createMemberResp := httptest.NewRecorder()
	router.ServeHTTP(createMemberResp, createMemberReq)
	if createMemberResp.Code != http.StatusCreated {
		t.Fatalf("expected member create status %d, got %d body=%s", http.StatusCreated, createMemberResp.Code, createMemberResp.Body.String())
	}

	resetMemberships()

	usersReq := httptest.NewRequest(http.MethodGet, "/api/iam/users", nil)
	usersReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	usersResp := httptest.NewRecorder()
	router.ServeHTTP(usersResp, usersReq)
	if usersResp.Code != http.StatusOK {
		t.Fatalf("expected users list status %d, got %d body=%s", http.StatusOK, usersResp.Code, usersResp.Body.String())
	}
	var usersBody struct {
		Users []iamUser `json:"users"`
	}
	if err := json.Unmarshal(usersResp.Body.Bytes(), &usersBody); err != nil {
		t.Fatalf("expected users JSON body, got error: %v", err)
	}
	found := false
	for _, user := range usersBody.Users {
		if user.ID == "u-7" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected persisted user/membership after reload, body=%s", usersResp.Body.String())
	}
}

func TestIAMInvitesReloadFromPersistenceWhenCacheMiss(t *testing.T) {
	resetIAMPersistenceForTest()
	t.Cleanup(func() {
		resetIAMPersistenceForTest()
	})
	t.Setenv("KUBEDECK_IAM_PERSIST_IN_TEST", "1")
	t.Setenv("KUBEDECK_SQLITE_DSN", filepath.Join(t.TempDir(), "invites-reload.sqlite"))

	resetAuthSessions()
	resetInvites()
	resetAuditWriter()
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
		t.Fatalf("expected login JSON body, got error: %v", err)
	}

	createInviteReq := httptest.NewRequest(http.MethodPost, "/api/iam/invites", bytes.NewReader([]byte(`{"email":"user@example.com","role_hint":"member","expires_in_hours":2}`)))
	createInviteReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	createInviteResp := httptest.NewRecorder()
	router.ServeHTTP(createInviteResp, createInviteReq)
	if createInviteResp.Code != http.StatusCreated {
		t.Fatalf("expected create invite status %d, got %d body=%s", http.StatusCreated, createInviteResp.Code, createInviteResp.Body.String())
	}

	resetInvites()

	listReq := httptest.NewRequest(http.MethodGet, "/api/iam/invites", nil)
	listReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	listResp := httptest.NewRecorder()
	router.ServeHTTP(listResp, listReq)
	if listResp.Code != http.StatusOK {
		t.Fatalf("expected list invites status %d, got %d body=%s", http.StatusOK, listResp.Code, listResp.Body.String())
	}
	var listBody struct {
		Invites []iamInvite `json:"invites"`
	}
	if err := json.Unmarshal(listResp.Body.Bytes(), &listBody); err != nil {
		t.Fatalf("expected invites JSON body, got error: %v", err)
	}
	if len(listBody.Invites) == 0 || listBody.Invites[0].InviteeEmail != "user@example.com" {
		t.Fatalf("expected persisted invite after reload, body=%s", listResp.Body.String())
	}
}

func TestAuthAcceptInviteReloadsFromPersistenceWhenCacheMiss(t *testing.T) {
	resetIAMPersistenceForTest()
	t.Cleanup(func() {
		resetIAMPersistenceForTest()
	})
	t.Setenv("KUBEDECK_IAM_PERSIST_IN_TEST", "1")
	t.Setenv("KUBEDECK_SQLITE_DSN", filepath.Join(t.TempDir(), "accept-invite-reload.sqlite"))

	resetAuthSessions()
	resetInvites()
	resetAuditWriter()
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
		t.Fatalf("expected login JSON body, got error: %v", err)
	}

	createInviteReq := httptest.NewRequest(http.MethodPost, "/api/iam/invites", bytes.NewReader([]byte(`{"email":"user2@example.com","role_hint":"member","expires_in_hours":2}`)))
	createInviteReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	createInviteResp := httptest.NewRecorder()
	router.ServeHTTP(createInviteResp, createInviteReq)
	if createInviteResp.Code != http.StatusCreated {
		t.Fatalf("expected create invite status %d, got %d body=%s", http.StatusCreated, createInviteResp.Code, createInviteResp.Body.String())
	}
	var inviteBody struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(createInviteResp.Body.Bytes(), &inviteBody); err != nil {
		t.Fatalf("expected invite JSON body, got error: %v", err)
	}
	if inviteBody.Token == "" {
		t.Fatalf("expected invite token")
	}

	resetInvites()

	acceptReq := httptest.NewRequest(http.MethodPost, "/api/auth/accept-invite", bytes.NewReader([]byte(`{"token":"`+inviteBody.Token+`","username":"new-user","password":"pw"}`)))
	acceptResp := httptest.NewRecorder()
	router.ServeHTTP(acceptResp, acceptReq)
	if acceptResp.Code != http.StatusOK {
		t.Fatalf("expected accept invite status %d, got %d body=%s", http.StatusOK, acceptResp.Code, acceptResp.Body.String())
	}
}

func TestIAMRevokeInviteReloadsFromPersistenceWhenCacheMiss(t *testing.T) {
	resetIAMPersistenceForTest()
	t.Cleanup(func() {
		resetIAMPersistenceForTest()
	})
	t.Setenv("KUBEDECK_IAM_PERSIST_IN_TEST", "1")
	t.Setenv("KUBEDECK_SQLITE_DSN", filepath.Join(t.TempDir(), "revoke-invite-reload.sqlite"))

	resetAuthSessions()
	resetInvites()
	resetAuditWriter()
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
		t.Fatalf("expected login JSON body, got error: %v", err)
	}

	createInviteReq := httptest.NewRequest(http.MethodPost, "/api/iam/invites", bytes.NewReader([]byte(`{"email":"user3@example.com","role_hint":"member","expires_in_hours":2}`)))
	createInviteReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	createInviteResp := httptest.NewRecorder()
	router.ServeHTTP(createInviteResp, createInviteReq)
	if createInviteResp.Code != http.StatusCreated {
		t.Fatalf("expected create invite status %d, got %d body=%s", http.StatusCreated, createInviteResp.Code, createInviteResp.Body.String())
	}
	var inviteBody struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(createInviteResp.Body.Bytes(), &inviteBody); err != nil {
		t.Fatalf("expected invite JSON body, got error: %v", err)
	}
	if inviteBody.ID == "" {
		t.Fatalf("expected invite id")
	}

	resetInvites()

	revokeReq := httptest.NewRequest(http.MethodDelete, "/api/iam/invites/"+inviteBody.ID, nil)
	revokeReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	revokeResp := httptest.NewRecorder()
	router.ServeHTTP(revokeResp, revokeReq)
	if revokeResp.Code != http.StatusOK {
		t.Fatalf("expected revoke invite status %d, got %d body=%s", http.StatusOK, revokeResp.Code, revokeResp.Body.String())
	}
}

func TestIAMGroupPatchReloadFromPersistenceWhenCacheMiss(t *testing.T) {
	resetIAMPersistenceForTest()
	t.Cleanup(func() {
		resetIAMPersistenceForTest()
	})
	t.Setenv("KUBEDECK_IAM_PERSIST_IN_TEST", "1")
	t.Setenv("KUBEDECK_SQLITE_DSN", filepath.Join(t.TempDir(), "group-patch-reload.sqlite"))

	resetAuthSessions()
	resetGroups()
	resetAuditWriter()
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
		t.Fatalf("expected login JSON body, got error: %v", err)
	}

	createReq := httptest.NewRequest(http.MethodPost, "/api/iam/groups", bytes.NewReader([]byte(`{"name":"ops","description":"Ops Team"}`)))
	createReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	createResp := httptest.NewRecorder()
	router.ServeHTTP(createResp, createReq)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected create group status %d, got %d body=%s", http.StatusCreated, createResp.Code, createResp.Body.String())
	}
	var created iamGroup
	if err := json.Unmarshal(createResp.Body.Bytes(), &created); err != nil {
		t.Fatalf("expected create group JSON, got error: %v", err)
	}

	resetGroups()

	patchReq := httptest.NewRequest(http.MethodPatch, "/api/iam/groups/"+created.ID, bytes.NewReader([]byte(`{"description":"patched"}`)))
	patchReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	patchResp := httptest.NewRecorder()
	router.ServeHTTP(patchResp, patchReq)
	if patchResp.Code != http.StatusOK {
		t.Fatalf("expected patch status %d, got %d body=%s", http.StatusOK, patchResp.Code, patchResp.Body.String())
	}
}

func TestIAMMembershipValidityReloadFromPersistenceWhenCacheMiss(t *testing.T) {
	resetIAMPersistenceForTest()
	t.Cleanup(func() {
		resetIAMPersistenceForTest()
	})
	t.Setenv("KUBEDECK_IAM_PERSIST_IN_TEST", "1")
	t.Setenv("KUBEDECK_SQLITE_DSN", filepath.Join(t.TempDir(), "membership-validity-reload.sqlite"))

	resetAuthSessions()
	resetMemberships()
	resetAuditWriter()
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
		t.Fatalf("expected login JSON body, got error: %v", err)
	}

	createMemberReq := httptest.NewRequest(http.MethodPost, "/api/iam/tenants/tenant-dev/members", bytes.NewReader([]byte(`{"user_id":"u-99","user_label":"user99","effective_from":"2026-02-25T00:00:00Z"}`)))
	createMemberReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	createMemberResp := httptest.NewRecorder()
	router.ServeHTTP(createMemberResp, createMemberReq)
	if createMemberResp.Code != http.StatusCreated {
		t.Fatalf("expected create member status %d, got %d body=%s", http.StatusCreated, createMemberResp.Code, createMemberResp.Body.String())
	}
	var createdMember iamMembership
	if err := json.Unmarshal(createMemberResp.Body.Bytes(), &createdMember); err != nil {
		t.Fatalf("expected create member JSON, got error: %v", err)
	}

	resetMemberships()

	validityReq := httptest.NewRequest(http.MethodPut, "/api/iam/memberships/"+createdMember.ID+"/validity", bytes.NewReader([]byte(`{"effective_from":"2026-02-25T00:00:00Z","effective_until":"2026-12-31T00:00:00Z"}`)))
	validityReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	validityResp := httptest.NewRecorder()
	router.ServeHTTP(validityResp, validityReq)
	if validityResp.Code != http.StatusOK {
		t.Fatalf("expected validity replace status %d, got %d body=%s", http.StatusOK, validityResp.Code, validityResp.Body.String())
	}
}

func TestAuthLoginTenantCodeDeniedWhenUnknown(t *testing.T) {
	resetAuthSessions()
	resetAuditWriter()
	router := NewRouter()

	loginPayload := []byte(`{"username":"alice","password":"pw","tenant_code":"unknown"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginPayload))
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)
	if loginResp.Code != http.StatusForbidden {
		t.Fatalf("expected login status 403, got %d body=%s", loginResp.Code, loginResp.Body.String())
	}
}

func TestSessionExpiredAfterLoginIsDeniedOnProtectedAPIs(t *testing.T) {
	resetAuthSessions()
	resetAuditWriter()
	router := NewRouter()

	loginPayload := []byte(`{"username":"alice","password":"pw","tenant_code":"dev"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginPayload))
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("expected login status 200, got %d body=%s", loginResp.Code, loginResp.Body.String())
	}

	var loginBody struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("unmarshal login response: %v", err)
	}
	if loginBody.Token == "" {
		t.Fatalf("expected token in login response")
	}

	authSessionsMu.Lock()
	s := authSessions[loginBody.Token]
	s.User.Memberships = []auth.TenantMembership{
		{
			TenantID:      s.ActiveTenantID,
			UserID:        s.User.ID,
			EffectiveFrom: time.Now().UTC().Add(-48 * time.Hour),
			EffectiveUntil: func() *time.Time {
				v := time.Now().UTC().Add(-1 * time.Hour)
				return &v
			}(),
		},
	}
	authSessions[loginBody.Token] = s
	authSessionsMu.Unlock()

	meReq := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	meResp := httptest.NewRecorder()
	router.ServeHTTP(meResp, meReq)
	if meResp.Code != http.StatusForbidden {
		t.Fatalf("expected me status 403 after membership expired, got %d body=%s", meResp.Code, meResp.Body.String())
	}
	if !strings.Contains(meResp.Body.String(), "membership_expired") {
		t.Fatalf("expected membership_expired body, got %s", meResp.Body.String())
	}

	iamReq := httptest.NewRequest(http.MethodGet, "/api/iam/groups", nil)
	iamReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	iamResp := httptest.NewRecorder()
	router.ServeHTTP(iamResp, iamReq)
	if iamResp.Code != http.StatusUnauthorized {
		t.Fatalf("expected iam status 401 after expired session cleanup, got %d body=%s", iamResp.Code, iamResp.Body.String())
	}
}

func TestIAMGroupManagementFlow(t *testing.T) {
	resetAuthSessions()
	resetMemberships()
	iamGroupsMu.Lock()
	iamGroups = map[string]iamGroup{}
	iamGroupsMu.Unlock()
	resetAuditWriter()
	router := NewRouter()

	loginPayload := []byte(`{"username":"admin","password":"pw","tenant_code":"dev"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginPayload))
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("expected login 200, got %d", loginResp.Code)
	}
	var loginBody struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("unmarshal login: %v", err)
	}

	createPayload := []byte(`{"name":"platform-admins","description":"Platform admins"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/iam/groups", bytes.NewReader(createPayload))
	createReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	createResp := httptest.NewRecorder()
	router.ServeHTTP(createResp, createReq)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected create group 201, got %d body=%s", createResp.Code, createResp.Body.String())
	}

	var created iamGroup
	if err := json.Unmarshal(createResp.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal created group: %v", err)
	}

	permPayload := []byte(`{"permissions":["iam:read","iam:write"]}`)
	permReq := httptest.NewRequest(http.MethodPut, "/api/iam/groups/"+created.ID+"/permissions", bytes.NewReader(permPayload))
	permReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	permResp := httptest.NewRecorder()
	router.ServeHTTP(permResp, permReq)
	if permResp.Code != http.StatusOK {
		t.Fatalf("expected replace permissions 200, got %d body=%s", permResp.Code, permResp.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/iam/groups", nil)
	listReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	listResp := httptest.NewRecorder()
	router.ServeHTTP(listResp, listReq)
	if listResp.Code != http.StatusOK {
		t.Fatalf("expected list groups 200, got %d", listResp.Code)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/iam/groups/"+created.ID, nil)
	deleteReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	deleteResp := httptest.NewRecorder()
	router.ServeHTTP(deleteResp, deleteReq)
	if deleteResp.Code != http.StatusOK {
		t.Fatalf("expected delete group 200, got %d", deleteResp.Code)
	}
}

func TestIAMGroupCreateDuplicateRejected(t *testing.T) {
	resetAuthSessions()
	resetAuditWriter()
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
		t.Fatalf("expected login JSON body, got error: %v", err)
	}

	firstCreateReq := httptest.NewRequest(http.MethodPost, "/api/iam/groups", bytes.NewReader([]byte(`{"name":"ops","description":"Ops Team"}`)))
	firstCreateReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	firstCreateResp := httptest.NewRecorder()
	router.ServeHTTP(firstCreateResp, firstCreateReq)
	if firstCreateResp.Code != http.StatusCreated {
		t.Fatalf("expected first create status %d, got %d body=%s", http.StatusCreated, firstCreateResp.Code, firstCreateResp.Body.String())
	}

	secondCreateReq := httptest.NewRequest(http.MethodPost, "/api/iam/groups", bytes.NewReader([]byte(`{"name":"OPS","description":"Duplicate Name"}`)))
	secondCreateReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	secondCreateResp := httptest.NewRecorder()
	router.ServeHTTP(secondCreateResp, secondCreateReq)
	if secondCreateResp.Code != http.StatusConflict {
		t.Fatalf("expected duplicate create status %d, got %d body=%s", http.StatusConflict, secondCreateResp.Code, secondCreateResp.Body.String())
	}
}

func TestMembershipListAndReplaceGroupsFlow(t *testing.T) {
	resetAuthSessions()
	resetMemberships()
	resetAuditWriter()
	router := NewRouter()

	loginPayload := []byte(`{"username":"admin","password":"pw","tenant_code":"dev"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginPayload))
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("expected login 200, got %d", loginResp.Code)
	}
	var loginBody struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("unmarshal login: %v", err)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/iam/memberships", nil)
	listReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	listResp := httptest.NewRecorder()
	router.ServeHTTP(listResp, listReq)
	if listResp.Code != http.StatusOK {
		t.Fatalf("expected list memberships 200, got %d body=%s", listResp.Code, listResp.Body.String())
	}
	var listBody struct {
		Memberships []iamMembership `json:"memberships"`
	}
	if err := json.Unmarshal(listResp.Body.Bytes(), &listBody); err != nil {
		t.Fatalf("unmarshal list memberships: %v", err)
	}
	if len(listBody.Memberships) == 0 {
		t.Fatalf("expected memberships in response")
	}

	createAdminsPayload := []byte(`{"name":"admins","description":"admins group"}`)
	createAdminsReq := httptest.NewRequest(http.MethodPost, "/api/iam/groups", bytes.NewReader(createAdminsPayload))
	createAdminsReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	createAdminsResp := httptest.NewRecorder()
	router.ServeHTTP(createAdminsResp, createAdminsReq)
	if createAdminsResp.Code != http.StatusCreated {
		t.Fatalf("expected create admins group 201, got %d body=%s", createAdminsResp.Code, createAdminsResp.Body.String())
	}

	createDevopsPayload := []byte(`{"name":"devops","description":"devops group"}`)
	createDevopsReq := httptest.NewRequest(http.MethodPost, "/api/iam/groups", bytes.NewReader(createDevopsPayload))
	createDevopsReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	createDevopsResp := httptest.NewRecorder()
	router.ServeHTTP(createDevopsResp, createDevopsReq)
	if createDevopsResp.Code != http.StatusCreated {
		t.Fatalf("expected create devops group 201, got %d body=%s", createDevopsResp.Code, createDevopsResp.Body.String())
	}

	replacePayload := []byte(`{"group_ids":["grp-admins","grp-devops"]}`)
	replaceReq := httptest.NewRequest(http.MethodPut, "/api/iam/memberships/"+listBody.Memberships[0].ID+"/groups", bytes.NewReader(replacePayload))
	replaceReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	replaceResp := httptest.NewRecorder()
	router.ServeHTTP(replaceResp, replaceReq)
	if replaceResp.Code != http.StatusOK {
		t.Fatalf("expected replace membership groups 200, got %d body=%s", replaceResp.Code, replaceResp.Body.String())
	}
	var updated iamMembership
	if err := json.Unmarshal(replaceResp.Body.Bytes(), &updated); err != nil {
		t.Fatalf("unmarshal updated membership: %v", err)
	}
	if len(updated.GroupIDs) != 2 {
		t.Fatalf("expected 2 group ids, got %d", len(updated.GroupIDs))
	}

	validityPayload := []byte(`{"effective_from":"2026-02-25T00:00:00Z","effective_until":"2026-12-31T00:00:00Z"}`)
	validityReq := httptest.NewRequest(http.MethodPut, "/api/iam/memberships/"+listBody.Memberships[0].ID+"/validity", bytes.NewReader(validityPayload))
	validityReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	validityResp := httptest.NewRecorder()
	router.ServeHTTP(validityResp, validityReq)
	if validityResp.Code != http.StatusOK {
		t.Fatalf("expected replace membership validity 200, got %d body=%s", validityResp.Code, validityResp.Body.String())
	}
	var validityUpdated iamMembership
	if err := json.Unmarshal(validityResp.Body.Bytes(), &validityUpdated); err != nil {
		t.Fatalf("unmarshal validity membership: %v", err)
	}
	if validityUpdated.EffectiveUntil == nil {
		t.Fatalf("expected effective_until in updated membership")
	}
}

func TestMembershipReplaceGroupsRejectsUnknownGroupIDs(t *testing.T) {
	resetAuthSessions()
	resetMemberships()
	resetAuditWriter()
	router := NewRouter()

	loginPayload := []byte(`{"username":"admin","password":"pw","tenant_code":"dev"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginPayload))
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("expected login 200, got %d", loginResp.Code)
	}
	var loginBody struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("unmarshal login: %v", err)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/iam/memberships", nil)
	listReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	listResp := httptest.NewRecorder()
	router.ServeHTTP(listResp, listReq)
	if listResp.Code != http.StatusOK {
		t.Fatalf("expected list memberships 200, got %d", listResp.Code)
	}
	var listBody struct {
		Memberships []iamMembership `json:"memberships"`
	}
	if err := json.Unmarshal(listResp.Body.Bytes(), &listBody); err != nil {
		t.Fatalf("unmarshal list memberships: %v", err)
	}
	if len(listBody.Memberships) == 0 {
		t.Fatalf("expected memberships in response")
	}

	replacePayload := []byte(`{"group_ids":["grp-does-not-exist"]}`)
	replaceReq := httptest.NewRequest(http.MethodPut, "/api/iam/memberships/"+listBody.Memberships[0].ID+"/groups", bytes.NewReader(replacePayload))
	replaceReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	replaceResp := httptest.NewRecorder()
	router.ServeHTTP(replaceResp, replaceReq)
	if replaceResp.Code != http.StatusBadRequest {
		t.Fatalf("expected replace membership groups 400, got %d body=%s", replaceResp.Code, replaceResp.Body.String())
	}
}

func TestGroupDeleteCleansMembershipGroupRefs(t *testing.T) {
	resetAuthSessions()
	resetMemberships()
	resetAuditWriter()
	router := NewRouter()

	loginPayload := []byte(`{"username":"admin","password":"pw","tenant_code":"dev"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginPayload))
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("expected login 200, got %d", loginResp.Code)
	}
	var loginBody struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("unmarshal login: %v", err)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/iam/memberships", nil)
	listReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	listResp := httptest.NewRecorder()
	router.ServeHTTP(listResp, listReq)
	if listResp.Code != http.StatusOK {
		t.Fatalf("expected list memberships 200, got %d", listResp.Code)
	}
	var listBody struct {
		Memberships []iamMembership `json:"memberships"`
	}
	if err := json.Unmarshal(listResp.Body.Bytes(), &listBody); err != nil {
		t.Fatalf("unmarshal list memberships: %v", err)
	}
	if len(listBody.Memberships) == 0 {
		t.Fatalf("expected memberships in response")
	}

	createPayload := []byte(`{"name":"admins","description":"admins group"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/iam/groups", bytes.NewReader(createPayload))
	createReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	createResp := httptest.NewRecorder()
	router.ServeHTTP(createResp, createReq)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected create group 201, got %d body=%s", createResp.Code, createResp.Body.String())
	}
	var created iamGroup
	if err := json.Unmarshal(createResp.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal created group: %v", err)
	}

	replacePayload := []byte(`{"group_ids":["` + created.ID + `"]}`)
	replaceReq := httptest.NewRequest(http.MethodPut, "/api/iam/memberships/"+listBody.Memberships[0].ID+"/groups", bytes.NewReader(replacePayload))
	replaceReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	replaceResp := httptest.NewRecorder()
	router.ServeHTTP(replaceResp, replaceReq)
	if replaceResp.Code != http.StatusOK {
		t.Fatalf("expected replace membership groups 200, got %d body=%s", replaceResp.Code, replaceResp.Body.String())
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/iam/groups/"+created.ID, nil)
	deleteReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	deleteResp := httptest.NewRecorder()
	router.ServeHTTP(deleteResp, deleteReq)
	if deleteResp.Code != http.StatusOK {
		t.Fatalf("expected delete group 200, got %d body=%s", deleteResp.Code, deleteResp.Body.String())
	}

	listAfterReq := httptest.NewRequest(http.MethodGet, "/api/iam/memberships", nil)
	listAfterReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	listAfterResp := httptest.NewRecorder()
	router.ServeHTTP(listAfterResp, listAfterReq)
	if listAfterResp.Code != http.StatusOK {
		t.Fatalf("expected list memberships after delete 200, got %d", listAfterResp.Code)
	}
	var listAfterBody struct {
		Memberships []iamMembership `json:"memberships"`
	}
	if err := json.Unmarshal(listAfterResp.Body.Bytes(), &listAfterBody); err != nil {
		t.Fatalf("unmarshal list memberships after delete: %v", err)
	}
	if len(listAfterBody.Memberships) == 0 {
		t.Fatalf("expected memberships after delete")
	}
	if len(listAfterBody.Memberships[0].GroupIDs) != 0 {
		t.Fatalf("expected membership group ids cleaned after group delete, got %v", listAfterBody.Memberships[0].GroupIDs)
	}
}

func TestIAMGroupWriteDeniedForViewer(t *testing.T) {
	resetAuthSessions()
	resetAuditWriter()
	router := NewRouter()

	loginPayload := []byte(`{"username":"viewer","password":"pw","tenant_code":"dev"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginPayload))
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("expected login 200, got %d", loginResp.Code)
	}
	var loginBody struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("unmarshal login: %v", err)
	}

	createPayload := []byte(`{"name":"readonly-group"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/iam/groups", bytes.NewReader(createPayload))
	createReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	createResp := httptest.NewRecorder()
	router.ServeHTTP(createResp, createReq)
	if createResp.Code != http.StatusForbidden {
		t.Fatalf("expected create group 403 for viewer, got %d", createResp.Code)
	}
}

func TestIAMUsersListFlow(t *testing.T) {
	resetAuthSessions()
	resetMemberships()
	resetAuditWriter()
	router := NewRouter()

	loginPayload := []byte(`{"username":"admin","password":"pw","tenant_code":"dev"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginPayload))
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("expected login 200, got %d", loginResp.Code)
	}
	var loginBody struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("unmarshal login: %v", err)
	}

	usersReq := httptest.NewRequest(http.MethodGet, "/api/iam/users", nil)
	usersReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	usersResp := httptest.NewRecorder()
	router.ServeHTTP(usersResp, usersReq)
	if usersResp.Code != http.StatusOK {
		t.Fatalf("expected list users 200, got %d body=%s", usersResp.Code, usersResp.Body.String())
	}
	var payload struct {
		Users []iamUser `json:"users"`
	}
	if err := json.Unmarshal(usersResp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal users response: %v", err)
	}
	if len(payload.Users) == 0 {
		t.Fatalf("expected non-empty users list")
	}
	if payload.Users[0].TenantID != "tenant-dev" {
		t.Fatalf("expected tenant-dev user, got %q", payload.Users[0].TenantID)
	}
}

func TestIAMTenantsAndTenantMembersFlow(t *testing.T) {
	resetAuthSessions()
	resetMemberships()
	resetAuditWriter()
	router := NewRouter()

	loginPayload := []byte(`{"username":"admin","password":"pw","tenant_code":"dev"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginPayload))
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("expected login 200, got %d", loginResp.Code)
	}
	var loginBody struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("unmarshal login: %v", err)
	}

	tenantsReq := httptest.NewRequest(http.MethodGet, "/api/iam/tenants", nil)
	tenantsReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	tenantsResp := httptest.NewRecorder()
	router.ServeHTTP(tenantsResp, tenantsReq)
	if tenantsResp.Code != http.StatusOK {
		t.Fatalf("expected tenants 200, got %d body=%s", tenantsResp.Code, tenantsResp.Body.String())
	}

	createMemberPayload := []byte(`{"user_id":"u-member-1","user_label":"member1","effective_from":"2026-02-25T00:00:00Z"}`)
	createMemberReq := httptest.NewRequest(http.MethodPost, "/api/iam/tenants/tenant-dev/members", bytes.NewReader(createMemberPayload))
	createMemberReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	createMemberResp := httptest.NewRecorder()
	router.ServeHTTP(createMemberResp, createMemberReq)
	if createMemberResp.Code != http.StatusCreated {
		t.Fatalf("expected create tenant member 201, got %d body=%s", createMemberResp.Code, createMemberResp.Body.String())
	}
	var created iamMembership
	if err := json.Unmarshal(createMemberResp.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal created member: %v", err)
	}

	listMembersReq := httptest.NewRequest(http.MethodGet, "/api/iam/tenants/tenant-dev/members", nil)
	listMembersReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	listMembersResp := httptest.NewRecorder()
	router.ServeHTTP(listMembersResp, listMembersReq)
	if listMembersResp.Code != http.StatusOK {
		t.Fatalf("expected list tenant members 200, got %d body=%s", listMembersResp.Code, listMembersResp.Body.String())
	}

	deleteMemberReq := httptest.NewRequest(http.MethodDelete, "/api/iam/tenants/tenant-dev/members?membership_id="+created.ID, nil)
	deleteMemberReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	deleteMemberResp := httptest.NewRecorder()
	router.ServeHTTP(deleteMemberResp, deleteMemberReq)
	if deleteMemberResp.Code != http.StatusOK {
		t.Fatalf("expected delete tenant member 200, got %d body=%s", deleteMemberResp.Code, deleteMemberResp.Body.String())
	}
}

func TestIAMTenantMemberDuplicateRejected(t *testing.T) {
	resetAuthSessions()
	resetMemberships()
	resetAuditWriter()
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
		t.Fatalf("expected login JSON body, got error: %v", err)
	}

	memberPayload := []byte(`{"user_id":"u-3","user_label":"user-three","effective_from":"2026-02-25T00:00:00Z"}`)
	firstCreateReq := httptest.NewRequest(http.MethodPost, "/api/iam/tenants/tenant-dev/members", bytes.NewReader(memberPayload))
	firstCreateReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	firstCreateResp := httptest.NewRecorder()
	router.ServeHTTP(firstCreateResp, firstCreateReq)
	if firstCreateResp.Code != http.StatusCreated {
		t.Fatalf("expected first member create status %d, got %d body=%s", http.StatusCreated, firstCreateResp.Code, firstCreateResp.Body.String())
	}

	secondCreateReq := httptest.NewRequest(http.MethodPost, "/api/iam/tenants/tenant-dev/members", bytes.NewReader(memberPayload))
	secondCreateReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	secondCreateResp := httptest.NewRecorder()
	router.ServeHTTP(secondCreateResp, secondCreateReq)
	if secondCreateResp.Code != http.StatusConflict {
		t.Fatalf("expected duplicate member create status %d, got %d body=%s", http.StatusConflict, secondCreateResp.Code, secondCreateResp.Body.String())
	}
}

func TestInviteCreateListAndAcceptFlow(t *testing.T) {
	resetAuthSessions()
	resetInvites()
	resetAuditWriter()
	router := NewRouter()

	loginPayload := []byte(`{"username":"admin","password":"pw","tenant_code":"dev"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginPayload))
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("expected login 200, got %d", loginResp.Code)
	}
	var loginBody struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("unmarshal login: %v", err)
	}

	createInvitePayload := []byte(`{"email":"user@example.com","role_hint":"member","expires_in_hours":2}`)
	createInviteReq := httptest.NewRequest(http.MethodPost, "/api/iam/invites", bytes.NewReader(createInvitePayload))
	createInviteReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	createInviteResp := httptest.NewRecorder()
	router.ServeHTTP(createInviteResp, createInviteReq)
	if createInviteResp.Code != http.StatusCreated {
		t.Fatalf("expected create invite 201, got %d body=%s", createInviteResp.Code, createInviteResp.Body.String())
	}
	var inviteBody iamInvite
	if err := json.Unmarshal(createInviteResp.Body.Bytes(), &inviteBody); err != nil {
		t.Fatalf("unmarshal invite create: %v", err)
	}
	if inviteBody.Token == "" {
		t.Fatalf("expected invite token")
	}
	if !strings.HasPrefix(inviteBody.InviteLink, "#/accept-invite?token=") {
		t.Fatalf("expected hash invite link, got %q", inviteBody.InviteLink)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/iam/invites", nil)
	listReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	listResp := httptest.NewRecorder()
	router.ServeHTTP(listResp, listReq)
	if listResp.Code != http.StatusOK {
		t.Fatalf("expected list invites 200, got %d", listResp.Code)
	}

	acceptPayload := []byte(`{"token":"` + inviteBody.Token + `","username":"new-user","password":"pw"}`)
	acceptReq := httptest.NewRequest(http.MethodPost, "/api/auth/accept-invite", bytes.NewReader(acceptPayload))
	acceptResp := httptest.NewRecorder()
	router.ServeHTTP(acceptResp, acceptReq)
	if acceptResp.Code != http.StatusOK {
		t.Fatalf("expected accept invite 200, got %d body=%s", acceptResp.Code, acceptResp.Body.String())
	}
}

func TestInviteRevokePendingFlow(t *testing.T) {
	resetAuthSessions()
	resetInvites()
	resetAuditWriter()
	router := NewRouter()

	loginPayload := []byte(`{"username":"admin","password":"pw","tenant_code":"dev"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginPayload))
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("expected login 200, got %d", loginResp.Code)
	}
	var loginBody struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("unmarshal login: %v", err)
	}

	createInvitePayload := []byte(`{"email":"user@example.com","role_hint":"member","expires_in_hours":2}`)
	createInviteReq := httptest.NewRequest(http.MethodPost, "/api/iam/invites", bytes.NewReader(createInvitePayload))
	createInviteReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	createInviteResp := httptest.NewRecorder()
	router.ServeHTTP(createInviteResp, createInviteReq)
	if createInviteResp.Code != http.StatusCreated {
		t.Fatalf("expected create invite 201, got %d body=%s", createInviteResp.Code, createInviteResp.Body.String())
	}
	var inviteBody iamInvite
	if err := json.Unmarshal(createInviteResp.Body.Bytes(), &inviteBody); err != nil {
		t.Fatalf("unmarshal invite create: %v", err)
	}

	revokeReq := httptest.NewRequest(http.MethodDelete, "/api/iam/invites/"+inviteBody.ID, nil)
	revokeReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	revokeResp := httptest.NewRecorder()
	router.ServeHTTP(revokeResp, revokeReq)
	if revokeResp.Code != http.StatusOK {
		t.Fatalf("expected revoke invite 200, got %d body=%s", revokeResp.Code, revokeResp.Body.String())
	}
	var revokeBody iamInvite
	if err := json.Unmarshal(revokeResp.Body.Bytes(), &revokeBody); err != nil {
		t.Fatalf("unmarshal revoke invite: %v", err)
	}
	if revokeBody.Status != "revoked" {
		t.Fatalf("expected revoked status, got %q", revokeBody.Status)
	}

	revokeAgainReq := httptest.NewRequest(http.MethodDelete, "/api/iam/invites/"+inviteBody.ID, nil)
	revokeAgainReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	revokeAgainResp := httptest.NewRecorder()
	router.ServeHTTP(revokeAgainResp, revokeAgainReq)
	if revokeAgainResp.Code != http.StatusConflict {
		t.Fatalf("expected second revoke 409, got %d body=%s", revokeAgainResp.Code, revokeAgainResp.Body.String())
	}
}

func TestInviteListOrderedByCreatedAtDesc(t *testing.T) {
	resetAuthSessions()
	resetInvites()
	resetAuditWriter()
	router := NewRouter()

	loginPayload := []byte(`{"username":"admin","password":"pw","tenant_code":"dev"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginPayload))
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("expected login 200, got %d", loginResp.Code)
	}
	var loginBody struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("unmarshal login: %v", err)
	}

	createInvite := func(email string) iamInvite {
		payload := []byte(`{"email":"` + email + `","role_hint":"member","expires_in_hours":2}`)
		req := httptest.NewRequest(http.MethodPost, "/api/iam/invites", bytes.NewReader(payload))
		req.Header.Set("Authorization", "Bearer "+loginBody.Token)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if resp.Code != http.StatusCreated {
			t.Fatalf("expected create invite 201, got %d body=%s", resp.Code, resp.Body.String())
		}
		var invite iamInvite
		if err := json.Unmarshal(resp.Body.Bytes(), &invite); err != nil {
			t.Fatalf("unmarshal invite: %v", err)
		}
		return invite
	}

	first := createInvite("first@example.com")
	second := createInvite("second@example.com")

	invitesMu.Lock()
	firstRecord := invites[first.Token]
	secondRecord := invites[second.Token]
	firstRecord.CreatedAt = time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	secondRecord.CreatedAt = time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)
	invites[first.Token] = firstRecord
	invites[second.Token] = secondRecord
	invitesMu.Unlock()

	listReq := httptest.NewRequest(http.MethodGet, "/api/iam/invites", nil)
	listReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	listResp := httptest.NewRecorder()
	router.ServeHTTP(listResp, listReq)
	if listResp.Code != http.StatusOK {
		t.Fatalf("expected list invites 200, got %d body=%s", listResp.Code, listResp.Body.String())
	}
	var listBody struct {
		Invites []iamInvite `json:"invites"`
	}
	if err := json.Unmarshal(listResp.Body.Bytes(), &listBody); err != nil {
		t.Fatalf("unmarshal list invites: %v", err)
	}
	if len(listBody.Invites) < 2 {
		t.Fatalf("expected at least 2 invites, got %d", len(listBody.Invites))
	}
	if listBody.Invites[0].ID != second.ID {
		t.Fatalf("expected latest invite first, got %q then %q", listBody.Invites[0].ID, listBody.Invites[1].ID)
	}
}

func TestInviteAcceptExpired(t *testing.T) {
	resetAuthSessions()
	resetInvites()
	resetAuditWriter()
	router := NewRouter()

	invite := iamInvite{
		ID:         "inv-expired",
		TenantID:   "tenant-dev",
		TenantCode: "dev",
		Token:      "expired-token",
		InviteLink: "#/accept-invite?token=expired-token",
		ExpiresAt:  time.Now().UTC().Add(-1 * time.Minute),
		Status:     "pending",
	}
	invitesMu.Lock()
	invites[invite.Token] = invite
	invitesMu.Unlock()

	acceptPayload := []byte(`{"token":"expired-token","username":"new-user","password":"pw"}`)
	acceptReq := httptest.NewRequest(http.MethodPost, "/api/auth/accept-invite", bytes.NewReader(acceptPayload))
	acceptResp := httptest.NewRecorder()
	router.ServeHTTP(acceptResp, acceptReq)
	if acceptResp.Code != http.StatusGone {
		t.Fatalf("expected accept expired invite 410, got %d", acceptResp.Code)
	}
}

func TestAuditEventsEndpointIncludesAuthEvents(t *testing.T) {
	resetAuthSessions()
	resetAuditWriter()
	router := NewRouter()

	loginPayload := []byte(`{"username":"admin","password":"pw","tenant_code":"dev"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginPayload))
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("expected login 200, got %d", loginResp.Code)
	}
	var loginBody struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("unmarshal login: %v", err)
	}

	eventsReq := httptest.NewRequest(http.MethodGet, "/api/audit/events", nil)
	eventsReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	eventsResp := httptest.NewRecorder()
	router.ServeHTTP(eventsResp, eventsReq)
	if eventsResp.Code != http.StatusOK {
		t.Fatalf("expected audit events 200, got %d body=%s", eventsResp.Code, eventsResp.Body.String())
	}
	var payload struct {
		Events []struct {
			Action string `json:"action"`
		} `json:"events"`
	}
	if err := json.Unmarshal(eventsResp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal audit events: %v", err)
	}
	if len(payload.Events) == 0 {
		t.Fatalf("expected non-empty audit events")
	}
}

func TestAuditEventsEndpointSupportsFilters(t *testing.T) {
	resetAuthSessions()
	resetAuditWriter()
	router := NewRouter()

	loginPayload := []byte(`{"username":"admin","password":"pw","tenant_code":"dev"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginPayload))
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("expected login 200, got %d", loginResp.Code)
	}
	var loginBody struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("unmarshal login: %v", err)
	}

	switchPayload := []byte(`{"tenant_code":"staging"}`)
	switchReq := httptest.NewRequest(http.MethodPost, "/api/auth/switch-tenant", bytes.NewReader(switchPayload))
	switchReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	switchResp := httptest.NewRecorder()
	router.ServeHTTP(switchResp, switchReq)
	if switchResp.Code != http.StatusOK {
		t.Fatalf("expected switch tenant 200, got %d", switchResp.Code)
	}

	eventsReq := httptest.NewRequest(http.MethodGet, "/api/audit/events?action=switch_tenant&result=allowed&limit=1", nil)
	eventsReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	eventsResp := httptest.NewRecorder()
	router.ServeHTTP(eventsResp, eventsReq)
	if eventsResp.Code != http.StatusOK {
		t.Fatalf("expected audit events 200, got %d body=%s", eventsResp.Code, eventsResp.Body.String())
	}
	var payload struct {
		Events []struct {
			Action string `json:"action"`
			Result string `json:"result"`
		} `json:"events"`
	}
	if err := json.Unmarshal(eventsResp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal audit events: %v", err)
	}
	if len(payload.Events) != 1 {
		t.Fatalf("expected exactly one event with limit=1, got %d", len(payload.Events))
	}
	if !strings.Contains(payload.Events[0].Action, "switch_tenant") {
		t.Fatalf("expected switch_tenant action, got %q", payload.Events[0].Action)
	}
	if payload.Events[0].Result != "allowed" {
		t.Fatalf("expected result allowed, got %q", payload.Events[0].Result)
	}
}
