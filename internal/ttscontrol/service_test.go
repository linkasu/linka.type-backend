package ttscontrol

import (
	"context"
	"crypto/sha256"
	"errors"
	"net/netip"
	"sync"
	"testing"
	"time"
)

type memoryStore struct {
	mu            sync.Mutex
	installations map[[32]byte]Installation
	revoked       map[[32]byte]bool
	quotas        map[string]DailyQuota
	findErr       error
	reserveCalls  int
}

func newMemoryStore() *memoryStore {
	return &memoryStore{installations: map[[32]byte]Installation{}, revoked: map[[32]byte]bool{}, quotas: map[string]DailyQuota{}}
}

func (s *memoryStore) CreateInstallation(_ context.Context, installation Installation) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	issued := 0
	for _, existing := range s.installations {
		if string(existing.IPPrefixHash) == string(installation.IPPrefixHash) && existing.IssuedAt.UTC().Format("2006-01-02") == installation.IssuedAt.UTC().Format("2006-01-02") {
			issued++
		}
	}
	if issued >= 3 {
		return ErrIssuanceLimit
	}
	s.installations[sha256.Sum256(installation.TokenHash)] = installation
	return nil
}

func (s *memoryStore) FindInstallationByTokenHash(_ context.Context, tokenHash []byte) (Installation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.findErr != nil {
		return Installation{}, s.findErr
	}
	installation, ok := s.installations[sha256.Sum256(tokenHash)]
	if !ok {
		return Installation{}, ErrTokenInvalid
	}
	return installation, nil
}

func TestVerifyInstallationTokenPropagatesStorageFailure(t *testing.T) {
	store := newMemoryStore()
	service := testService(t, store)
	service.now = func() time.Time { return time.Date(2026, 9, 29, 12, 0, 0, 0, time.UTC) }
	issued, err := service.IssueInstallationToken(context.Background(), mustAddr(t, "203.0.113.55"), "")
	if err != nil {
		t.Fatal(err)
	}
	store.findErr = errors.New("postgres unavailable")
	if _, err := service.VerifyInstallationToken(context.Background(), issued.Token); err == nil || errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("storage error = %v", err)
	}
}

func (s *memoryStore) RevokeToken(_ context.Context, tokenHash []byte, _ string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.revoked[sha256.Sum256(tokenHash)] = true
	return nil
}
func (s *memoryStore) IsTokenRevoked(_ context.Context, tokenHash []byte) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.revoked[sha256.Sum256(tokenHash)], nil
}

func (s *memoryStore) ReserveDailyQuota(_ context.Context, key string, day time.Time, limit, chunks int) (DailyQuota, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reserveCalls++
	mapKey := key + day.UTC().Format("2006-01-02")
	quota := s.quotas[mapKey]
	if quota.Used+chunks > limit {
		return DailyQuota{}, ErrQuotaExceeded
	}
	quota = DailyQuota{Key: key, Day: day.UTC().Truncate(24 * time.Hour), Used: quota.Used + chunks, Limit: limit, UpdatedAt: time.Now()}
	s.quotas[mapKey] = quota
	return quota, nil
}

func (*memoryStore) AppendSecurityAudit(context.Context, SecurityAudit) error { return nil }
func (*memoryStore) GetIdempotency(context.Context, string, string) (IdempotencyRecord, error) {
	return IdempotencyRecord{}, ErrTokenInvalid
}
func (*memoryStore) PutIdempotency(context.Context, IdempotencyRecord) error { return nil }

type failingCounter struct{ calls int }

func (c *failingCounter) SetDaily(context.Context, DailyQuota) error {
	c.calls++
	return errors.New("redis unavailable")
}

func testService(t *testing.T, store *memoryStore) *Service {
	t.Helper()
	service, err := NewService(store, nil, ServiceConfig{SigningKey: "signing-key-is-at-least-thirty-two-bytes", IPHashKey: "ip-hash-key-is-at-least-thirty-two-bytes", AnonymousDaily: 30, AuthenticatedDaily: 200, AnonymousMax: 5, AuthenticatedMax: 21})
	if err != nil {
		t.Fatal(err)
	}
	return service
}

func TestInstallationTokenUsesHashAndRevocation(t *testing.T) {
	store := newMemoryStore()
	service := testService(t, store)
	service.now = func() time.Time { return time.Date(2026, 9, 29, 12, 0, 0, 0, time.UTC) }
	issued, err := service.IssueInstallationToken(context.Background(), mustAddr(t, "203.0.113.55"), "")
	if err != nil {
		t.Fatal(err)
	}
	if string(issued.Installation.TokenHash) == issued.Token {
		t.Fatal("plaintext token was stored")
	}
	verified, err := service.VerifyInstallationToken(context.Background(), issued.Token)
	if err != nil || verified.InstallationID != issued.Installation.ID {
		t.Fatalf("verify = %#v, %v", verified, err)
	}
	if err := store.RevokeToken(context.Background(), issued.Installation.TokenHash, "test"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.VerifyInstallationToken(context.Background(), issued.Token); err != ErrTokenInvalid {
		t.Fatalf("revoked token error = %v", err)
	}
}

func TestIssueInstallationTokenLimitsPrefixToThreePerDay(t *testing.T) {
	store := newMemoryStore()
	service := testService(t, store)
	service.now = func() time.Time { return time.Date(2026, 9, 29, 12, 0, 0, 0, time.UTC) }
	for i := 0; i < 3; i++ {
		if _, err := service.IssueInstallationToken(context.Background(), mustAddr(t, "2001:db8:1:2::1"), ""); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := service.IssueInstallationToken(context.Background(), mustAddr(t, "2001:db8:1:2::99"), ""); err != ErrIssuanceLimit {
		t.Fatalf("fourth issue error = %v", err)
	}
}

func TestReserveChunksIsAtomic(t *testing.T) {
	store := newMemoryStore()
	service := testService(t, store)
	service.now = func() time.Time { return time.Date(2026, 9, 29, 12, 0, 0, 0, time.UTC) }
	verified := VerifiedInstallation{InstallationID: "installation", Kind: InstallationAnonymous}
	var wg sync.WaitGroup
	var accepted int
	var mu sync.Mutex
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := service.ReserveChunks(context.Background(), verified, 1); err == nil {
				mu.Lock()
				accepted++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if accepted != 30 {
		t.Fatalf("accepted = %d, want 30", accepted)
	}
}

func TestReserveChunksSucceedsWhenRedisMirrorFails(t *testing.T) {
	store := newMemoryStore()
	counter := &failingCounter{}
	service, err := NewService(store, counter, ServiceConfig{SigningKey: "signing-key-is-at-least-thirty-two-bytes", IPHashKey: "ip-hash-key-is-at-least-thirty-two-bytes", AnonymousDaily: 30, AuthenticatedDaily: 200, AnonymousMax: 5, AuthenticatedMax: 21})
	if err != nil {
		t.Fatal(err)
	}
	quota, err := service.ReserveChunks(context.Background(), VerifiedInstallation{InstallationID: "installation", Kind: InstallationAnonymous}, 1)
	if err != nil || quota.Used != 1 {
		t.Fatalf("quota = %#v, err = %v", quota, err)
	}
	if store.reserveCalls != 1 || counter.calls != 1 {
		t.Fatalf("PG reservations = %d, Redis writes = %d", store.reserveCalls, counter.calls)
	}
}

func mustAddr(t *testing.T, raw string) netip.Addr {
	t.Helper()
	addr, err := netip.ParseAddr(raw)
	if err != nil {
		t.Fatal(err)
	}
	return addr
}
