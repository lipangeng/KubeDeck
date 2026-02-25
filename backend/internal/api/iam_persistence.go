package api

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"kubedeck/backend/internal/storage"
)

var (
	iamPersistenceOnce sync.Once
	iamPersistenceRepo storage.IAMPersistence
	iamPersistenceErr  error
	iamPersistDriver   string
	iamPersistDSN      string
	iamPersistDisabled bool
)

type PersistenceConfig struct {
	Driver   string
	DSN      string
	Disabled bool
}

func ConfigurePersistence(cfg PersistenceConfig) {
	iamPersistDriver = strings.TrimSpace(cfg.Driver)
	iamPersistDSN = strings.TrimSpace(cfg.DSN)
	iamPersistDisabled = cfg.Disabled
	resetIAMPersistenceForTest()
}

func resetIAMPersistenceForTest() {
	if iamPersistenceRepo != nil {
		_ = iamPersistenceRepo.Close()
	}
	iamPersistenceRepo = nil
	iamPersistenceErr = nil
	iamPersistenceOnce = sync.Once{}
}

func ensureIAMPersistence() {
	iamPersistenceOnce.Do(func() {
		if !iamPersistenceEnabled() {
			return
		}
		driver := iamPersistDriver
		if driver == "" {
			driver = strings.TrimSpace(os.Getenv("KUBEDECK_DB_DRIVER"))
		}
		dsn := iamPersistDSN
		if dsn == "" {
			dsn = strings.TrimSpace(os.Getenv("KUBEDECK_SQLITE_DSN"))
		}
		repo, err := storage.NewIAMPersistence(driver, dsn)
		if err != nil {
			iamPersistenceErr = err
			return
		}
		iamPersistenceRepo = repo
		if err := loadIAMPersistentState(repo); err != nil {
			iamPersistenceErr = err
			return
		}
	})
	if iamPersistenceErr != nil {
		log.Printf("iam persistence disabled due to init error: %v", iamPersistenceErr)
	}
}

func iamPersistenceEnabled() bool {
	if iamPersistDisabled {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("KUBEDECK_IAM_PERSIST")), "0") {
		return false
	}
	argv0 := os.Args[0]
	if strings.HasSuffix(argv0, ".test") && !strings.EqualFold(strings.TrimSpace(os.Getenv("KUBEDECK_IAM_PERSIST_IN_TEST")), "1") {
		return false
	}
	return true
}

func loadIAMPersistentState(repo storage.IAMPersistence) error {
	snapshot, err := repo.Load()
	if err != nil {
		return err
	}

	loadGroups := map[string]iamGroup{}
	for _, item := range snapshot.Groups {
		loadGroups[item.ID] = iamGroup{
			ID:          item.ID,
			TenantID:    item.TenantID,
			Name:        item.Name,
			Description: item.Description,
			Permissions: append([]string{}, item.Permissions...),
		}
	}

	loadMemberships := map[string]iamMembership{}
	for _, item := range snapshot.Memberships {
		loadMemberships[item.ID] = iamMembership{
			ID:             item.ID,
			TenantID:       item.TenantID,
			UserID:         item.UserID,
			UserLabel:      item.UserLabel,
			GroupIDs:       append([]string{}, item.GroupIDs...),
			EffectiveFrom:  item.EffectiveFrom,
			EffectiveUntil: item.EffectiveUntil,
		}
	}

	loadInvites := map[string]iamInvite{}
	for _, item := range snapshot.Invites {
		loadInvites[item.Token] = iamInvite{
			Token:        "",
			ID:           item.ID,
			TenantID:     item.TenantID,
			TenantCode:   item.TenantCode,
			InviteeEmail: item.InviteeEmail,
			InviteePhone: item.InviteePhone,
			RoleHint:     item.RoleHint,
			InviteLink:   item.InviteLink,
			CreatedAt:    item.CreatedAt,
			ExpiresAt:    item.ExpiresAt,
			Status:       item.Status,
		}
	}

	loadSessions, err := decodeAuthSessions(snapshot.Sessions)
	if err != nil {
		return err
	}

	iamGroupsMu.Lock()
	iamGroups = loadGroups
	iamGroupsMu.Unlock()

	iamMembershipsMu.Lock()
	iamMemberships = loadMemberships
	iamMembershipsMu.Unlock()

	invitesMu.Lock()
	invites = loadInvites
	invitesMu.Unlock()

	authSessionsMu.Lock()
	authSessions = loadSessions
	authSessionsMu.Unlock()

	return nil
}

func reloadAuthSessionsFromPersistence() error {
	ensureIAMPersistence()
	if iamPersistenceRepo == nil {
		return nil
	}
	snapshot, err := iamPersistenceRepo.Load()
	if err != nil {
		return err
	}
	loadSessions, err := decodeAuthSessions(snapshot.Sessions)
	if err != nil {
		return err
	}
	authSessionsMu.Lock()
	authSessions = loadSessions
	authSessionsMu.Unlock()
	return nil
}

func reloadIAMStateFromPersistence() error {
	ensureIAMPersistence()
	if iamPersistenceRepo == nil {
		return nil
	}
	return loadIAMPersistentState(iamPersistenceRepo)
}

func decodeAuthSessions(records []storage.AuthSessionRecord) (map[string]authSession, error) {
	loadSessions := map[string]authSession{}
	for _, item := range records {
		var session authSession
		if err := json.Unmarshal([]byte(item.UserJSON), &session.User); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(item.AvailableJSON), &session.Available); err != nil {
			return nil, err
		}
		session.Token = item.Token
		session.ActiveTenantID = item.ActiveTenantID
		session.ExpiresAt = item.ExpiresAt.UTC()
		if session.ExpiresAt.IsZero() {
			session.ExpiresAt = time.Now().UTC().Add(authSessionTTL())
		}
		loadSessions[item.Token] = session
	}
	return loadSessions, nil
}

func persistIAMGroups() {
	ensureIAMPersistence()
	if iamPersistenceRepo == nil {
		return
	}
	iamGroupsMu.RLock()
	snapshot := make([]storage.IAMGroupRecord, 0, len(iamGroups))
	for _, item := range iamGroups {
		snapshot = append(snapshot, storage.IAMGroupRecord{
			ID:          item.ID,
			TenantID:    item.TenantID,
			Name:        item.Name,
			Description: item.Description,
			Permissions: append([]string{}, item.Permissions...),
		})
	}
	iamGroupsMu.RUnlock()
	if err := iamPersistenceRepo.ReplaceGroups(snapshot); err != nil {
		log.Printf("persist iam groups failed: %v", err)
	}
}

func persistIAMMemberships() {
	ensureIAMPersistence()
	if iamPersistenceRepo == nil {
		return
	}
	iamMembershipsMu.RLock()
	snapshot := make([]storage.IAMMembershipRecord, 0, len(iamMemberships))
	for _, item := range iamMemberships {
		snapshot = append(snapshot, storage.IAMMembershipRecord{
			ID:             item.ID,
			TenantID:       item.TenantID,
			UserID:         item.UserID,
			UserLabel:      item.UserLabel,
			GroupIDs:       append([]string{}, item.GroupIDs...),
			EffectiveFrom:  item.EffectiveFrom,
			EffectiveUntil: item.EffectiveUntil,
		})
	}
	iamMembershipsMu.RUnlock()
	if err := iamPersistenceRepo.ReplaceMemberships(snapshot); err != nil {
		log.Printf("persist iam memberships failed: %v", err)
	}
}

func persistIAMInvites() {
	ensureIAMPersistence()
	if iamPersistenceRepo == nil {
		return
	}
	invitesMu.RLock()
	snapshot := make([]storage.IAMInviteRecord, 0, len(invites))
	for tokenKey, item := range invites {
		token := strings.TrimSpace(item.Token)
		if token == "" {
			token = tokenKey
		}
		snapshot = append(snapshot, storage.IAMInviteRecord{
			Token:        token,
			ID:           item.ID,
			TenantID:     item.TenantID,
			TenantCode:   item.TenantCode,
			InviteeEmail: item.InviteeEmail,
			InviteePhone: item.InviteePhone,
			RoleHint:     item.RoleHint,
			InviteLink:   item.InviteLink,
			CreatedAt:    item.CreatedAt,
			ExpiresAt:    item.ExpiresAt,
			Status:       item.Status,
		})
	}
	invitesMu.RUnlock()
	if err := iamPersistenceRepo.ReplaceInvites(snapshot); err != nil {
		log.Printf("persist iam invites failed: %v", err)
	}
}

func persistAuthSessions() {
	ensureIAMPersistence()
	if iamPersistenceRepo == nil {
		return
	}
	authSessionsMu.RLock()
	snapshot := make([]storage.AuthSessionRecord, 0, len(authSessions))
	for _, item := range authSessions {
		userJSON, _ := json.Marshal(item.User)
		availableJSON, _ := json.Marshal(item.Available)
		snapshot = append(snapshot, storage.AuthSessionRecord{
			Token:          item.Token,
			UserJSON:       string(userJSON),
			AvailableJSON:  string(availableJSON),
			ActiveTenantID: item.ActiveTenantID,
			ExpiresAt:      item.ExpiresAt.UTC(),
		})
	}
	authSessionsMu.RUnlock()
	if err := iamPersistenceRepo.ReplaceSessions(snapshot); err != nil {
		log.Printf("persist auth sessions failed: %v", err)
	}
}
