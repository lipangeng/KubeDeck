package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	router := NewRouter()

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
		InviteLink: "/accept-invite?token=expired-token",
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
