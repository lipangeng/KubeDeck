package api

import (
	"bytes"
	"encoding/json"
	"errors"
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
		User  struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("expected login JSON body, got error: %v", err)
	}
	if loginBody.User.ID == "" {
		t.Fatalf("expected login user id")
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
		User  struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("expected login JSON body, got error: %v", err)
	}
	if loginBody.User.ID == "" {
		t.Fatalf("expected login user id")
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

func TestRoutePolicyRequiresSessionForIAMUsers(t *testing.T) {
	resetAuthSessions()
	resetAuditWriter()
	router := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/iam/users", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusUnauthorized, resp.Code, resp.Body.String())
	}
}

func TestRoutePolicyRequiresPermissionForAuditEvents(t *testing.T) {
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
		User  struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("expected login JSON body, got error: %v", err)
	}
	if loginBody.User.ID == "" {
		t.Fatalf("expected login user id")
	}
	if loginBody.Token == "" {
		t.Fatalf("expected token in login response")
	}

	req := httptest.NewRequest(http.MethodGet, "/api/audit/events", nil)
	req.Header.Set("Authorization", "Bearer "+loginBody.Token)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusForbidden, resp.Code, resp.Body.String())
	}
	if !strings.Contains(resp.Body.String(), "permission_denied") {
		t.Fatalf("expected permission_denied body, got %s", resp.Body.String())
	}
}

func TestRoutePolicyAllowsAuditReadViaMembershipGroupPermission(t *testing.T) {
	resetAuthSessions()
	resetGroups()
	resetMemberships()
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
		User  struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("expected login JSON body, got error: %v", err)
	}
	if loginBody.User.ID == "" {
		t.Fatalf("expected login user id")
	}

	iamGroupsMu.Lock()
	iamGroups["grp-audit"] = iamGroup{
		ID:          "grp-audit",
		TenantID:    "tenant-dev",
		Name:        "audit-reader",
		Permissions: []string{"audit:read"},
	}
	iamGroupsMu.Unlock()
	iamMembershipsMu.Lock()
	membershipID := "mbr-" + loginBody.User.ID + "-tenant-dev"
	iamMemberships[membershipID] = iamMembership{
		ID:            membershipID,
		TenantID:      "tenant-dev",
		UserID:        loginBody.User.ID,
		UserLabel:     "viewer",
		GroupIDs:      []string{"grp-audit"},
		EffectiveFrom: time.Now().UTC().Add(-1 * time.Hour),
	}
	iamMembershipsMu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/api/audit/events", nil)
	req.Header.Set("Authorization", "Bearer "+loginBody.Token)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestIAMGroupsSeedsDefaultTenantGroups(t *testing.T) {
	resetAuthSessions()
	resetGroups()
	resetMemberships()
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

	req := httptest.NewRequest(http.MethodGet, "/api/iam/groups", nil)
	req.Header.Set("Authorization", "Bearer "+loginBody.Token)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, resp.Code, resp.Body.String())
	}
	var payload struct {
		Groups []iamGroup `json:"groups"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON response, got error: %v", err)
	}
	if !containsGroupName(payload.Groups, "tenant-owner") {
		t.Fatalf("expected tenant-owner group in payload, got %v", payload.Groups)
	}
	if !containsGroupName(payload.Groups, "tenant-admin") {
		t.Fatalf("expected tenant-admin group in payload, got %v", payload.Groups)
	}
	if !containsGroupName(payload.Groups, "tenant-viewer") {
		t.Fatalf("expected tenant-viewer group in payload, got %v", payload.Groups)
	}
}

func containsGroupName(groups []iamGroup, name string) bool {
	for _, group := range groups {
		if group.Name == name {
			return true
		}
	}
	return false
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(target)) {
			return true
		}
	}
	return false
}

func TestAuthLoginMeSwitchLogoutFlow(t *testing.T) {
	resetAuthSessions()
	resetMemberships()
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
		Token          string `json:"token"`
		ActiveTenantID string `json:"active_tenant_id"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("unmarshal login response: %v", err)
	}
	if loginBody.Token == "" {
		t.Fatalf("expected token in login response")
	}
	if loginBody.ActiveTenantID != "tenant-dev" {
		t.Fatalf("expected active tenant tenant-dev, got %q", loginBody.ActiveTenantID)
	}

	meReq := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	meResp := httptest.NewRecorder()
	router.ServeHTTP(meResp, meReq)
	if meResp.Code != http.StatusOK {
		t.Fatalf("expected me status 200, got %d body=%s", meResp.Code, meResp.Body.String())
	}

	switchPayload := []byte(`{"tenant_code":"staging"}`)
	switchReq := httptest.NewRequest(http.MethodPost, "/api/auth/switch-tenant", bytes.NewReader(switchPayload))
	switchReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	switchResp := httptest.NewRecorder()
	router.ServeHTTP(switchResp, switchReq)
	if switchResp.Code != http.StatusForbidden {
		t.Fatalf("expected switch status 403, got %d body=%s", switchResp.Code, switchResp.Body.String())
	}
	var switchBody struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(switchResp.Body.Bytes(), &switchBody); err != nil {
		t.Fatalf("unmarshal switch response: %v", err)
	}
	if switchBody.Error != "tenant_not_found" {
		t.Fatalf("expected tenant_not_found, got %q", switchBody.Error)
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

func TestAuthLoginDeniedWhenUserHasNoTenantMembership(t *testing.T) {
	resetAuthSessions()
	resetMemberships()
	resetAuditWriter()
	t.Setenv("KUBEDECK_ENV", "production")
	t.Setenv("KUBEDECK_LOCAL_AUTH_ENABLED", "true")
	t.Setenv("KUBEDECK_LOCAL_AUTH_PASSWORD", "pw")
	router := NewRouter()

	iamMembershipsMu.Lock()
	iamMemberships["mbr-other-tenant-dev"] = iamMembership{
		ID:            "mbr-other-tenant-dev",
		TenantID:      "tenant-dev",
		UserID:        "other-user",
		UserLabel:     "other",
		EffectiveFrom: time.Now().UTC().Add(-1 * time.Hour),
	}
	iamMembershipsMu.Unlock()

	loginPayload := []byte(`{"username":"alice","password":"pw","tenant_code":"dev"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginPayload))
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)
	if loginResp.Code != http.StatusForbidden {
		t.Fatalf("expected login 403 without membership, got %d body=%s", loginResp.Code, loginResp.Body.String())
	}
	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal login body: %v", err)
	}
	if body.Error != "tenant_not_found" {
		t.Fatalf("expected tenant_not_found, got %q", body.Error)
	}
}

func TestAuthLoginRejectsInvalidTenantCode(t *testing.T) {
	resetAuthSessions()
	resetMemberships()
	resetAuditWriter()
	router := NewRouter()

	loginPayload := []byte(`{"username":"alice","password":"pw","tenant_code":"bad code!"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginPayload))
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)
	if loginResp.Code != http.StatusBadRequest {
		t.Fatalf("expected login 400 for invalid tenant code, got %d body=%s", loginResp.Code, loginResp.Body.String())
	}
	if !strings.Contains(loginResp.Body.String(), "invalid_tenant_code") {
		t.Fatalf("expected invalid_tenant_code error, got %s", loginResp.Body.String())
	}
}

func TestSwitchTenantRejectsInvalidTenantCode(t *testing.T) {
	resetAuthSessions()
	resetMemberships()
	resetAuditWriter()
	router := NewRouter()

	loginPayload := []byte(`{"username":"alice","password":"pw","tenant_code":"dev"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginPayload))
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("expected login 200, got %d body=%s", loginResp.Code, loginResp.Body.String())
	}
	var loginBody struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("unmarshal login response: %v", err)
	}

	switchPayload := []byte(`{"tenant_code":"bad code!"}`)
	switchReq := httptest.NewRequest(http.MethodPost, "/api/auth/switch-tenant", bytes.NewReader(switchPayload))
	switchReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	switchResp := httptest.NewRecorder()
	router.ServeHTTP(switchResp, switchReq)
	if switchResp.Code != http.StatusBadRequest {
		t.Fatalf("expected switch 400 for invalid tenant code, got %d body=%s", switchResp.Code, switchResp.Body.String())
	}
	if !strings.Contains(switchResp.Body.String(), "invalid_tenant_code") {
		t.Fatalf("expected invalid_tenant_code error, got %s", switchResp.Body.String())
	}
}

func TestAuthLoginWritesAttemptedAndOutcomeAuditEvents(t *testing.T) {
	resetAuthSessions()
	resetMemberships()
	resetAuditWriter()
	router := NewRouter()

	failedPayload := []byte(`{"username":"alice","password":"wrong","tenant_code":"dev"}`)
	failedReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(failedPayload))
	failedResp := httptest.NewRecorder()
	router.ServeHTTP(failedResp, failedReq)
	if failedResp.Code != http.StatusUnauthorized {
		t.Fatalf("expected failed login 401, got %d", failedResp.Code)
	}

	successPayload := []byte(`{"username":"alice","password":"pw","tenant_code":"dev"}`)
	successReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(successPayload))
	successResp := httptest.NewRecorder()
	router.ServeHTTP(successResp, successReq)
	if successResp.Code != http.StatusOK {
		t.Fatalf("expected success login 200, got %d body=%s", successResp.Code, successResp.Body.String())
	}

	writer, ok := defaultAuditWriter.(*audit.MemoryWriter)
	if !ok {
		t.Fatalf("expected memory audit writer")
	}
	events := writer.List("")
	var hasAttempted bool
	var hasFailed bool
	var hasSucceeded bool
	for _, event := range events {
		switch event.Action {
		case "auth.login.attempted":
			hasAttempted = true
		case "auth.login.failed":
			hasFailed = true
		case "auth.login.succeeded":
			hasSucceeded = true
		}
	}
	if !hasAttempted || !hasFailed || !hasSucceeded {
		t.Fatalf("expected attempted/failed/succeeded login events, got %+v", events)
	}
}

func TestOAuthURLFlow(t *testing.T) {
	resetAuthSessions()
	resetAuditWriter()
	router := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/auth/oauth/url", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, resp.Code, resp.Body.String())
	}
	var body struct {
		Provider string `json:"provider"`
		State    string `json:"state"`
		AuthURL  string `json:"auth_url"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected oauth url response json, got %v", err)
	}
	if strings.TrimSpace(body.State) == "" {
		t.Fatalf("expected non-empty state")
	}
	if strings.TrimSpace(body.AuthURL) == "" || !strings.Contains(body.AuthURL, "state=") {
		t.Fatalf("expected auth url with state, got %q", body.AuthURL)
	}
}

func TestOAuthCallbackFlow(t *testing.T) {
	resetAuthSessions()
	resetAuditWriter()
	router := NewRouter()

	oauthURLReq := httptest.NewRequest(http.MethodGet, "/api/auth/oauth/url", nil)
	oauthURLResp := httptest.NewRecorder()
	router.ServeHTTP(oauthURLResp, oauthURLReq)
	if oauthURLResp.Code != http.StatusOK {
		t.Fatalf("expected oauth url status %d, got %d body=%s", http.StatusOK, oauthURLResp.Code, oauthURLResp.Body.String())
	}
	var oauthURLBody struct {
		State string `json:"state"`
	}
	if err := json.Unmarshal(oauthURLResp.Body.Bytes(), &oauthURLBody); err != nil {
		t.Fatalf("expected oauth url response json, got %v", err)
	}

	payload := []byte(`{"code":"oauth-admin","tenant_code":"dev","state":"` + oauthURLBody.State + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/oauth/callback", bytes.NewReader(payload))
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, resp.Code, resp.Body.String())
	}
	var body struct {
		Token          string `json:"token"`
		ActiveTenantID string `json:"active_tenant_id"`
		User           struct {
			Username string   `json:"username"`
			Roles    []string `json:"roles"`
		} `json:"user"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected oauth callback json, got %v", err)
	}
	if body.Token == "" {
		t.Fatalf("expected token in oauth callback response")
	}
	if body.ActiveTenantID != "tenant-dev" {
		t.Fatalf("expected tenant-dev, got %q", body.ActiveTenantID)
	}
	if body.User.Username != "oauth-admin" {
		t.Fatalf("expected oauth-admin user, got %q", body.User.Username)
	}
}

func TestOAuthCallbackRejectsInvalidCode(t *testing.T) {
	resetAuthSessions()
	resetAuditWriter()
	router := NewRouter()

	oauthURLReq := httptest.NewRequest(http.MethodGet, "/api/auth/oauth/url", nil)
	oauthURLResp := httptest.NewRecorder()
	router.ServeHTTP(oauthURLResp, oauthURLReq)
	var oauthURLBody struct {
		State string `json:"state"`
	}
	if err := json.Unmarshal(oauthURLResp.Body.Bytes(), &oauthURLBody); err != nil {
		t.Fatalf("expected oauth url response json, got %v", err)
	}

	payload := []byte(`{"code":"invalid-oauth-code","tenant_code":"dev","state":"` + oauthURLBody.State + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/oauth/callback", bytes.NewReader(payload))
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusUnauthorized, resp.Code, resp.Body.String())
	}
}

func TestOAuthCallbackRejectsInvalidState(t *testing.T) {
	resetAuthSessions()
	resetAuditWriter()
	router := NewRouter()

	payload := []byte(`{"code":"oauth-admin","tenant_code":"dev","state":"invalid-state"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/oauth/callback", bytes.NewReader(payload))
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusBadRequest, resp.Code, resp.Body.String())
	}
}

func TestOAuthURLReturnsErrorWhenOIDCInitFailsInProduction(t *testing.T) {
	resetAuthSessions()
	resetAuditWriter()
	t.Setenv("KUBEDECK_OAUTH_MODE", "oidc")
	t.Setenv("KUBEDECK_ENV", "production")
	t.Setenv("KUBEDECK_OIDC_ISSUER", "")
	t.Setenv("KUBEDECK_OIDC_CLIENT_ID", "")
	t.Setenv("KUBEDECK_OIDC_CLIENT_SECRET", "")
	t.Setenv("KUBEDECK_OIDC_REDIRECT_URL", "")

	router := NewRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/oauth/url", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusServiceUnavailable, resp.Code, resp.Body.String())
	}

	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected JSON body, got %v", err)
	}
	if body.Error != "oauth_provider_unavailable" {
		t.Fatalf("expected oauth_provider_unavailable, got %q", body.Error)
	}
}

func TestOAuthCallbackReturnsErrorWhenOIDCInitFailsInProduction(t *testing.T) {
	resetAuthSessions()
	resetAuditWriter()
	t.Setenv("KUBEDECK_OAUTH_MODE", "oidc")
	t.Setenv("KUBEDECK_ENV", "production")
	t.Setenv("KUBEDECK_OIDC_ISSUER", "")
	t.Setenv("KUBEDECK_OIDC_CLIENT_ID", "")
	t.Setenv("KUBEDECK_OIDC_CLIENT_SECRET", "")
	t.Setenv("KUBEDECK_OIDC_REDIRECT_URL", "")

	router := NewRouter()
	payload := []byte(`{"code":"oauth-admin","tenant_code":"dev","state":"any-state"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/oauth/callback", bytes.NewReader(payload))
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusServiceUnavailable, resp.Code, resp.Body.String())
	}

	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected JSON body, got %v", err)
	}
	if body.Error != "oauth_provider_unavailable" {
		t.Fatalf("expected oauth_provider_unavailable, got %q", body.Error)
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
	if !containsGroupName(listBody.Groups, "ops") {
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
	foundUser9 := false
	for _, member := range membersBody.Members {
		if member.UserID == "u-9" {
			foundUser9 = true
			break
		}
	}
	if !foundUser9 {
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

func TestIAMRepoFirstCacheMissEndToEnd(t *testing.T) {
	resetIAMPersistenceForTest()
	t.Cleanup(func() {
		resetIAMPersistenceForTest()
	})
	t.Setenv("KUBEDECK_IAM_PERSIST_IN_TEST", "1")
	t.Setenv("KUBEDECK_SQLITE_DSN", filepath.Join(t.TempDir(), "iam-e2e-cache-miss.sqlite"))

	resetAuthSessions()
	resetGroups()
	resetMemberships()
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

	createGroupReq := httptest.NewRequest(http.MethodPost, "/api/iam/groups", bytes.NewReader([]byte(`{"name":"platform-admins","description":"platform group"}`)))
	createGroupReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	createGroupResp := httptest.NewRecorder()
	router.ServeHTTP(createGroupResp, createGroupReq)
	if createGroupResp.Code != http.StatusCreated {
		t.Fatalf("expected create group status %d, got %d body=%s", http.StatusCreated, createGroupResp.Code, createGroupResp.Body.String())
	}
	var createdGroup iamGroup
	if err := json.Unmarshal(createGroupResp.Body.Bytes(), &createdGroup); err != nil {
		t.Fatalf("expected group JSON body, got error: %v", err)
	}

	createMemberReq := httptest.NewRequest(http.MethodPost, "/api/iam/tenants/tenant-dev/members", bytes.NewReader([]byte(`{"user_id":"u-42","user_label":"user-42","effective_from":"2026-02-25T00:00:00Z"}`)))
	createMemberReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	createMemberResp := httptest.NewRecorder()
	router.ServeHTTP(createMemberResp, createMemberReq)
	if createMemberResp.Code != http.StatusCreated {
		t.Fatalf("expected create member status %d, got %d body=%s", http.StatusCreated, createMemberResp.Code, createMemberResp.Body.String())
	}
	var createdMember iamMembership
	if err := json.Unmarshal(createMemberResp.Body.Bytes(), &createdMember); err != nil {
		t.Fatalf("expected membership JSON body, got error: %v", err)
	}

	createInviteReq := httptest.NewRequest(http.MethodPost, "/api/iam/invites", bytes.NewReader([]byte(`{"email":"e2e@example.com","role_hint":"member","expires_in_hours":2}`)))
	createInviteReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	createInviteResp := httptest.NewRecorder()
	router.ServeHTTP(createInviteResp, createInviteReq)
	if createInviteResp.Code != http.StatusCreated {
		t.Fatalf("expected create invite status %d, got %d body=%s", http.StatusCreated, createInviteResp.Code, createInviteResp.Body.String())
	}
	var createdInvite iamInvite
	if err := json.Unmarshal(createInviteResp.Body.Bytes(), &createdInvite); err != nil {
		t.Fatalf("expected invite JSON body, got error: %v", err)
	}

	resetAuthSessions()
	resetGroups()
	resetMemberships()
	resetInvites()

	meReq := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	meResp := httptest.NewRecorder()
	router.ServeHTTP(meResp, meReq)
	if meResp.Code != http.StatusOK {
		t.Fatalf("expected me status %d after cache reset, got %d body=%s", http.StatusOK, meResp.Code, meResp.Body.String())
	}

	groupPatchReq := httptest.NewRequest(http.MethodPatch, "/api/iam/groups/"+createdGroup.ID, bytes.NewReader([]byte(`{"description":"patched-after-reload"}`)))
	groupPatchReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	groupPatchResp := httptest.NewRecorder()
	router.ServeHTTP(groupPatchResp, groupPatchReq)
	if groupPatchResp.Code != http.StatusOK {
		t.Fatalf("expected group patch status %d after cache reset, got %d body=%s", http.StatusOK, groupPatchResp.Code, groupPatchResp.Body.String())
	}

	validityReq := httptest.NewRequest(http.MethodPut, "/api/iam/memberships/"+createdMember.ID+"/validity", bytes.NewReader([]byte(`{"effective_from":"2026-02-25T00:00:00Z","effective_until":"2026-12-31T00:00:00Z"}`)))
	validityReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	validityResp := httptest.NewRecorder()
	router.ServeHTTP(validityResp, validityReq)
	if validityResp.Code != http.StatusOK {
		t.Fatalf("expected validity replace status %d after cache reset, got %d body=%s", http.StatusOK, validityResp.Code, validityResp.Body.String())
	}

	usersReq := httptest.NewRequest(http.MethodGet, "/api/iam/users", nil)
	usersReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	usersResp := httptest.NewRecorder()
	router.ServeHTTP(usersResp, usersReq)
	if usersResp.Code != http.StatusOK {
		t.Fatalf("expected users status %d after cache reset, got %d body=%s", http.StatusOK, usersResp.Code, usersResp.Body.String())
	}

	listInvitesReq := httptest.NewRequest(http.MethodGet, "/api/iam/invites", nil)
	listInvitesReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	listInvitesResp := httptest.NewRecorder()
	router.ServeHTTP(listInvitesResp, listInvitesReq)
	if listInvitesResp.Code != http.StatusOK {
		t.Fatalf("expected invites status %d after cache reset, got %d body=%s", http.StatusOK, listInvitesResp.Code, listInvitesResp.Body.String())
	}

	acceptInviteReq := httptest.NewRequest(http.MethodPost, "/api/auth/accept-invite", bytes.NewReader([]byte(`{"token":"`+createdInvite.Token+`","username":"new-e2e-user","password":"pw"}`)))
	acceptInviteResp := httptest.NewRecorder()
	router.ServeHTTP(acceptInviteResp, acceptInviteReq)
	if acceptInviteResp.Code != http.StatusOK {
		t.Fatalf("expected accept invite status %d after cache reset, got %d body=%s", http.StatusOK, acceptInviteResp.Code, acceptInviteResp.Body.String())
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

func TestIAMGroupSameNameAllowedAcrossTenants(t *testing.T) {
	resetAuthSessions()
	resetMemberships()
	resetGroups()
	resetAuditWriter()
	router := NewRouter()

	userID := auth.LocalUserID("admin")
	now := time.Now().UTC().Add(-1 * time.Hour)
	iamMembershipsMu.Lock()
	iamMemberships["mbr-"+userID+"-tenant-dev"] = iamMembership{
		ID:            "mbr-" + userID + "-tenant-dev",
		TenantID:      "tenant-dev",
		UserID:        userID,
		UserLabel:     "admin",
		EffectiveFrom: now,
	}
	iamMemberships["mbr-"+userID+"-tenant-qa"] = iamMembership{
		ID:            "mbr-" + userID + "-tenant-qa",
		TenantID:      "tenant-qa",
		UserID:        userID,
		UserLabel:     "admin",
		EffectiveFrom: now,
	}
	iamMembershipsMu.Unlock()

	loginPayload := []byte(`{"username":"admin","password":"pw","tenant_code":"dev"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginPayload))
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("expected login 200, got %d body=%s", loginResp.Code, loginResp.Body.String())
	}
	var loginBody struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("unmarshal login: %v", err)
	}

	createDevReq := httptest.NewRequest(http.MethodPost, "/api/iam/groups", bytes.NewReader([]byte(`{"name":"ops","description":"dev ops"}`)))
	createDevReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	createDevResp := httptest.NewRecorder()
	router.ServeHTTP(createDevResp, createDevReq)
	if createDevResp.Code != http.StatusCreated {
		t.Fatalf("expected create dev group 201, got %d body=%s", createDevResp.Code, createDevResp.Body.String())
	}
	var devGroup iamGroup
	if err := json.Unmarshal(createDevResp.Body.Bytes(), &devGroup); err != nil {
		t.Fatalf("unmarshal dev group: %v", err)
	}

	switchReq := httptest.NewRequest(http.MethodPost, "/api/auth/switch-tenant", bytes.NewReader([]byte(`{"tenant_code":"qa"}`)))
	switchReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	switchResp := httptest.NewRecorder()
	router.ServeHTTP(switchResp, switchReq)
	if switchResp.Code != http.StatusOK {
		t.Fatalf("expected switch tenant 200, got %d body=%s", switchResp.Code, switchResp.Body.String())
	}

	createQAReq := httptest.NewRequest(http.MethodPost, "/api/iam/groups", bytes.NewReader([]byte(`{"name":"ops","description":"qa ops"}`)))
	createQAReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	createQAResp := httptest.NewRecorder()
	router.ServeHTTP(createQAResp, createQAReq)
	if createQAResp.Code != http.StatusCreated {
		t.Fatalf("expected create qa group 201, got %d body=%s", createQAResp.Code, createQAResp.Body.String())
	}
	var qaGroup iamGroup
	if err := json.Unmarshal(createQAResp.Body.Bytes(), &qaGroup); err != nil {
		t.Fatalf("unmarshal qa group: %v", err)
	}

	if devGroup.ID == qaGroup.ID {
		t.Fatalf("expected tenant-scoped group ids to differ, got %q", devGroup.ID)
	}
	if !strings.HasPrefix(devGroup.ID, "grp-tenant-dev-") {
		t.Fatalf("expected dev group id prefix, got %q", devGroup.ID)
	}
	if !strings.HasPrefix(qaGroup.ID, "grp-tenant-qa-") {
		t.Fatalf("expected qa group id prefix, got %q", qaGroup.ID)
	}
}

func TestIAMGroupPatchRejectsDuplicateNameInTenant(t *testing.T) {
	resetAuthSessions()
	resetMemberships()
	resetGroups()
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

	createOpsReq := httptest.NewRequest(http.MethodPost, "/api/iam/groups", bytes.NewReader([]byte(`{"name":"ops","description":"ops group"}`)))
	createOpsReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	createOpsResp := httptest.NewRecorder()
	router.ServeHTTP(createOpsResp, createOpsReq)
	if createOpsResp.Code != http.StatusCreated {
		t.Fatalf("expected create ops group 201, got %d body=%s", createOpsResp.Code, createOpsResp.Body.String())
	}
	var opsGroup iamGroup
	if err := json.Unmarshal(createOpsResp.Body.Bytes(), &opsGroup); err != nil {
		t.Fatalf("unmarshal ops group: %v", err)
	}

	createDevopsReq := httptest.NewRequest(http.MethodPost, "/api/iam/groups", bytes.NewReader([]byte(`{"name":"devops","description":"devops group"}`)))
	createDevopsReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	createDevopsResp := httptest.NewRecorder()
	router.ServeHTTP(createDevopsResp, createDevopsReq)
	if createDevopsResp.Code != http.StatusCreated {
		t.Fatalf("expected create devops group 201, got %d body=%s", createDevopsResp.Code, createDevopsResp.Body.String())
	}

	patchReq := httptest.NewRequest(http.MethodPatch, "/api/iam/groups/"+opsGroup.ID, bytes.NewReader([]byte(`{"name":"devops"}`)))
	patchReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	patchResp := httptest.NewRecorder()
	router.ServeHTTP(patchResp, patchReq)
	if patchResp.Code != http.StatusConflict {
		t.Fatalf("expected patch conflict 409, got %d body=%s", patchResp.Code, patchResp.Body.String())
	}
}

func TestGroupIDRebuildRewritesMembershipRefs(t *testing.T) {
	resetGroups()
	resetMemberships()
	resetAuditWriter()

	iamGroupsMu.Lock()
	iamGroups["grp-admins"] = iamGroup{
		ID:       "grp-admins",
		TenantID: "tenant-dev",
		Name:     "Admins",
	}
	iamGroups["grp-devops"] = iamGroup{
		ID:       "grp-devops",
		TenantID: "tenant-dev",
		Name:     "Dev Ops",
	}
	iamGroupsMu.Unlock()

	iamMembershipsMu.Lock()
	iamMemberships["mbr-u1-tenant-dev"] = iamMembership{
		ID:        "mbr-u1-tenant-dev",
		TenantID:  "tenant-dev",
		UserID:    "u-1",
		UserLabel: "u1",
		GroupIDs:  []string{"grp-admins", "grp-devops", "grp-missing"},
	}
	iamMembershipsMu.Unlock()

	if err := rebuildGroupIDsForAllTenants(); err != nil {
		t.Fatalf("rebuildGroupIDsForAllTenants should succeed, got %v", err)
	}

	expectedAdminsID := groupIDForTenantName("tenant-dev", "Admins")
	expectedDevopsID := groupIDForTenantName("tenant-dev", "Dev Ops")

	iamGroupsMu.RLock()
	_, hasOldAdmins := iamGroups["grp-admins"]
	_, hasOldDevops := iamGroups["grp-devops"]
	_, hasNewAdmins := iamGroups[expectedAdminsID]
	_, hasNewDevops := iamGroups[expectedDevopsID]
	iamGroupsMu.RUnlock()
	if hasOldAdmins || hasOldDevops {
		t.Fatalf("expected old group ids removed after rebuild")
	}
	if !hasNewAdmins || !hasNewDevops {
		t.Fatalf("expected new rebuilt group ids present")
	}

	iamMembershipsMu.RLock()
	rebuiltMembership := iamMemberships["mbr-u1-tenant-dev"]
	iamMembershipsMu.RUnlock()
	if len(rebuiltMembership.GroupIDs) != 2 {
		t.Fatalf("expected rewritten group ids filtered to existing groups, got %v", rebuiltMembership.GroupIDs)
	}
	if !contains(rebuiltMembership.GroupIDs, expectedAdminsID) || !contains(rebuiltMembership.GroupIDs, expectedDevopsID) {
		t.Fatalf("expected membership group ids rewritten, got %v", rebuiltMembership.GroupIDs)
	}
}

func TestGroupIDRebuildRejectsCanonicalNameConflict(t *testing.T) {
	resetGroups()
	resetMemberships()
	resetAuditWriter()

	iamGroupsMu.Lock()
	iamGroups["grp-a"] = iamGroup{
		ID:       "grp-a",
		TenantID: "tenant-dev",
		Name:     "Ops Team",
	}
	iamGroups["grp-b"] = iamGroup{
		ID:       "grp-b",
		TenantID: "tenant-dev",
		Name:     "ops   team",
	}
	iamGroupsMu.Unlock()

	err := rebuildGroupIDsForAllTenants()
	if !errors.Is(err, ErrGroupCanonicalNameConflict) {
		t.Fatalf("expected ErrGroupCanonicalNameConflict, got %v", err)
	}

	iamGroupsMu.RLock()
	_, hasGroupA := iamGroups["grp-a"]
	_, hasGroupB := iamGroups["grp-b"]
	iamGroupsMu.RUnlock()
	if !hasGroupA || !hasGroupB {
		t.Fatalf("expected original group ids retained on rebuild conflict")
	}
}

func TestIAMRebuildGroupIDsEndpoint(t *testing.T) {
	resetAuthSessions()
	resetGroups()
	resetMemberships()
	resetAuditWriter()
	router := NewRouter()

	loginPayload := []byte(`{"username":"admin","password":"pw","tenant_code":"dev"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginPayload))
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("expected login 200, got %d body=%s", loginResp.Code, loginResp.Body.String())
	}
	var loginBody struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("unmarshal login: %v", err)
	}

	iamGroupsMu.Lock()
	iamGroups["grp-admins"] = iamGroup{ID: "grp-admins", TenantID: "tenant-dev", Name: "Admins"}
	iamGroups["grp-devops"] = iamGroup{ID: "grp-devops", TenantID: "tenant-dev", Name: "Dev Ops"}
	iamGroupsMu.Unlock()
	iamMembershipsMu.Lock()
	iamMemberships["mbr-u1-tenant-dev"] = iamMembership{
		ID:        "mbr-u1-tenant-dev",
		TenantID:  "tenant-dev",
		UserID:    "u-1",
		UserLabel: "u1",
		GroupIDs:  []string{"grp-admins", "grp-devops"},
	}
	iamMembershipsMu.Unlock()

	rebuildReq := httptest.NewRequest(http.MethodPost, "/api/iam/groups/rebuild-ids", nil)
	rebuildReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	rebuildResp := httptest.NewRecorder()
	router.ServeHTTP(rebuildResp, rebuildReq)
	if rebuildResp.Code != http.StatusOK {
		t.Fatalf("expected rebuild endpoint 200, got %d body=%s", rebuildResp.Code, rebuildResp.Body.String())
	}

	expectedAdminsID := groupIDForTenantName("tenant-dev", "Admins")
	expectedDevopsID := groupIDForTenantName("tenant-dev", "Dev Ops")
	iamMembershipsMu.RLock()
	rebuiltMembership := iamMemberships["mbr-u1-tenant-dev"]
	iamMembershipsMu.RUnlock()
	if !contains(rebuiltMembership.GroupIDs, expectedAdminsID) || !contains(rebuiltMembership.GroupIDs, expectedDevopsID) {
		t.Fatalf("expected membership refs rewritten via endpoint, got %v", rebuiltMembership.GroupIDs)
	}
}

func TestIAMRebuildGroupIDsEndpointRejectsCanonicalConflict(t *testing.T) {
	resetAuthSessions()
	resetGroups()
	resetMemberships()
	resetAuditWriter()
	router := NewRouter()

	loginPayload := []byte(`{"username":"admin","password":"pw","tenant_code":"dev"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginPayload))
	loginResp := httptest.NewRecorder()
	router.ServeHTTP(loginResp, loginReq)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("expected login 200, got %d body=%s", loginResp.Code, loginResp.Body.String())
	}
	var loginBody struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("unmarshal login: %v", err)
	}

	iamGroupsMu.Lock()
	iamGroups["grp-a"] = iamGroup{ID: "grp-a", TenantID: "tenant-dev", Name: "Ops Team"}
	iamGroups["grp-b"] = iamGroup{ID: "grp-b", TenantID: "tenant-dev", Name: "ops   team"}
	iamGroupsMu.Unlock()

	rebuildReq := httptest.NewRequest(http.MethodPost, "/api/iam/groups/rebuild-ids", nil)
	rebuildReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	rebuildResp := httptest.NewRecorder()
	router.ServeHTTP(rebuildResp, rebuildReq)
	if rebuildResp.Code != http.StatusConflict {
		t.Fatalf("expected rebuild endpoint 409, got %d body=%s", rebuildResp.Code, rebuildResp.Body.String())
	}
	if !strings.Contains(rebuildResp.Body.String(), "group_name_conflict") {
		t.Fatalf("expected group_name_conflict error, got %s", rebuildResp.Body.String())
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
	var adminsGroup iamGroup
	if err := json.Unmarshal(createAdminsResp.Body.Bytes(), &adminsGroup); err != nil {
		t.Fatalf("unmarshal admins group: %v", err)
	}

	createDevopsPayload := []byte(`{"name":"devops","description":"devops group"}`)
	createDevopsReq := httptest.NewRequest(http.MethodPost, "/api/iam/groups", bytes.NewReader(createDevopsPayload))
	createDevopsReq.Header.Set("Authorization", "Bearer "+loginBody.Token)
	createDevopsResp := httptest.NewRecorder()
	router.ServeHTTP(createDevopsResp, createDevopsReq)
	if createDevopsResp.Code != http.StatusCreated {
		t.Fatalf("expected create devops group 201, got %d body=%s", createDevopsResp.Code, createDevopsResp.Body.String())
	}
	var devopsGroup iamGroup
	if err := json.Unmarshal(createDevopsResp.Body.Bytes(), &devopsGroup); err != nil {
		t.Fatalf("unmarshal devops group: %v", err)
	}

	replacePayload := []byte(`{"group_ids":["` + adminsGroup.ID + `","` + devopsGroup.ID + `"]}`)
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
	var listBody struct {
		Invites []iamInvite `json:"invites"`
	}
	if err := json.Unmarshal(listResp.Body.Bytes(), &listBody); err != nil {
		t.Fatalf("unmarshal list invites: %v", err)
	}
	if len(listBody.Invites) == 0 {
		t.Fatalf("expected at least one invite in list")
	}
	if listBody.Invites[0].Token != "" {
		t.Fatalf("expected invite token redacted in list, got %q", listBody.Invites[0].Token)
	}
	if listBody.Invites[0].InviteLink != "#/accept-invite" {
		t.Fatalf("expected invite link redacted in list, got %q", listBody.Invites[0].InviteLink)
	}

	acceptPayload := []byte(`{"token":"` + inviteBody.Token + `","username":"new-user","password":"pw"}`)
	acceptReq := httptest.NewRequest(http.MethodPost, "/api/auth/accept-invite", bytes.NewReader(acceptPayload))
	acceptResp := httptest.NewRecorder()
	router.ServeHTTP(acceptResp, acceptReq)
	if acceptResp.Code != http.StatusOK {
		t.Fatalf("expected accept invite 200, got %d body=%s", acceptResp.Code, acceptResp.Body.String())
	}
}

func TestAcceptInviteCreatesMembershipWithRoleHintDefaultViewer(t *testing.T) {
	resetAuthSessions()
	resetInvites()
	resetMemberships()
	resetGroups()
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

	createInvitePayload := []byte(`{"email":"newmember@example.com","role_hint":"member","expires_in_hours":2}`)
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

	acceptPayload := []byte(`{"token":"` + inviteBody.Token + `","username":"new-member","password":"pw"}`)
	acceptReq := httptest.NewRequest(http.MethodPost, "/api/auth/accept-invite", bytes.NewReader(acceptPayload))
	acceptResp := httptest.NewRecorder()
	router.ServeHTTP(acceptResp, acceptReq)
	if acceptResp.Code != http.StatusOK {
		t.Fatalf("expected accept invite 200, got %d body=%s", acceptResp.Code, acceptResp.Body.String())
	}
	var acceptBody struct {
		Membership iamMembership `json:"membership"`
	}
	if err := json.Unmarshal(acceptResp.Body.Bytes(), &acceptBody); err != nil {
		t.Fatalf("unmarshal accept response: %v", err)
	}
	if acceptBody.Membership.UserID != auth.LocalUserID("new-member") {
		t.Fatalf("expected membership user id mapped from username, got %q", acceptBody.Membership.UserID)
	}
	if !contains(acceptBody.Membership.GroupIDs, "grp-tenant-dev-viewer") {
		t.Fatalf("expected default viewer group from member role hint, got %v", acceptBody.Membership.GroupIDs)
	}
}

func TestAcceptInviteMapsAdminRoleHintToTenantAdminGroup(t *testing.T) {
	resetAuthSessions()
	resetInvites()
	resetMemberships()
	resetGroups()
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

	createInvitePayload := []byte(`{"email":"admininvite@example.com","role_hint":"admin","expires_in_hours":2}`)
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

	acceptPayload := []byte(`{"token":"` + inviteBody.Token + `","username":"tenant-admin-user","password":"pw"}`)
	acceptReq := httptest.NewRequest(http.MethodPost, "/api/auth/accept-invite", bytes.NewReader(acceptPayload))
	acceptResp := httptest.NewRecorder()
	router.ServeHTTP(acceptResp, acceptReq)
	if acceptResp.Code != http.StatusOK {
		t.Fatalf("expected accept invite 200, got %d body=%s", acceptResp.Code, acceptResp.Body.String())
	}
	var acceptBody struct {
		Membership iamMembership `json:"membership"`
	}
	if err := json.Unmarshal(acceptResp.Body.Bytes(), &acceptBody); err != nil {
		t.Fatalf("unmarshal accept response: %v", err)
	}
	if !contains(acceptBody.Membership.GroupIDs, "grp-tenant-dev-admin") {
		t.Fatalf("expected tenant admin group from admin role hint, got %v", acceptBody.Membership.GroupIDs)
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

func TestAcceptInviteRequiresUsernameAndPassword(t *testing.T) {
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

	createInvitePayload := []byte(`{"email":"required@example.com","role_hint":"member","expires_in_hours":2}`)
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

	acceptPayload := []byte(`{"token":"` + inviteBody.Token + `","username":"","password":"pw"}`)
	acceptReq := httptest.NewRequest(http.MethodPost, "/api/auth/accept-invite", bytes.NewReader(acceptPayload))
	acceptResp := httptest.NewRecorder()
	router.ServeHTTP(acceptResp, acceptReq)
	if acceptResp.Code != http.StatusBadRequest {
		t.Fatalf("expected accept invite 400, got %d body=%s", acceptResp.Code, acceptResp.Body.String())
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
	firstRecord := invites[hashToken(first.Token)]
	secondRecord := invites[hashToken(second.Token)]
	firstRecord.CreatedAt = time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	secondRecord.CreatedAt = time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)
	invites[hashToken(first.Token)] = firstRecord
	invites[hashToken(second.Token)] = secondRecord
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

	switchPayload := []byte(`{"tenant_code":"dev"}`)
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

func TestSwitchTenantNotFoundWritesDeniedAuditEvent(t *testing.T) {
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
	if switchResp.Code != http.StatusForbidden {
		t.Fatalf("expected switch tenant 403, got %d body=%s", switchResp.Code, switchResp.Body.String())
	}

	eventsReq := httptest.NewRequest(http.MethodGet, "/api/audit/events?action=switch_tenant&result=denied&limit=1", nil)
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
			Reason string `json:"reason"`
		} `json:"events"`
	}
	if err := json.Unmarshal(eventsResp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal audit events: %v", err)
	}
	if len(payload.Events) == 0 {
		t.Fatalf("expected denied switch_tenant event")
	}
	if !strings.Contains(payload.Events[0].Action, "switch_tenant") {
		t.Fatalf("expected switch_tenant action, got %q", payload.Events[0].Action)
	}
	if payload.Events[0].Result != "denied" {
		t.Fatalf("expected denied result, got %q", payload.Events[0].Result)
	}
}
