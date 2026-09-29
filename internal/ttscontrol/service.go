package ttscontrol

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/netip"
	"strconv"
	"strings"
	"time"
)

type Service struct {
	store       Store
	counters    CounterCache
	signingKey  []byte
	previousKey []byte
	ipHashKey   []byte
	now         func() time.Time
	anonDaily   int
	authDaily   int
	anonMax     int
	authMax     int
}

type ServiceConfig struct {
	SigningKey         string
	PreviousKey        string
	IPHashKey          string
	AnonymousDaily     int
	AuthenticatedDaily int
	AnonymousMax       int
	AuthenticatedMax   int
}

func NewService(store Store, counters CounterCache, cfg ServiceConfig) (*Service, error) {
	if store == nil || len(cfg.SigningKey) < 32 || len(cfg.IPHashKey) < 32 {
		return nil, fmt.Errorf("tts control service requires storage and 32-byte signing and IP hash keys")
	}
	return &Service{
		store: store, counters: counters, signingKey: []byte(cfg.SigningKey), previousKey: []byte(cfg.PreviousKey), ipHashKey: []byte(cfg.IPHashKey), now: time.Now,
		anonDaily: cfg.AnonymousDaily, authDaily: cfg.AuthenticatedDaily, anonMax: cfg.AnonymousMax, authMax: cfg.AuthenticatedMax,
	}, nil
}

func (s *Service) IssueInstallationToken(ctx context.Context, ip netip.Addr, subject string) (InstallationToken, error) {
	kind := InstallationAnonymous
	if subject != "" {
		kind = InstallationAuthenticated
	}
	now := s.now().UTC()
	id, err := randomHex(16)
	if err != nil {
		return InstallationToken{}, err
	}
	nonce, err := randomHex(16)
	if err != nil {
		return InstallationToken{}, err
	}
	expiresAt := now.Add(TokenTTL)
	token := s.signToken(id, expiresAt, nonce)
	tokenHash := sha256.Sum256([]byte(token))
	installation := Installation{ID: id, TokenHash: tokenHash[:], IPPrefixHash: s.HashIPPrefix(ip), Subject: subject, Kind: kind, IssuedAt: now, ExpiresAt: expiresAt}
	if err := s.store.CreateInstallation(ctx, installation); err != nil {
		return InstallationToken{}, err
	}
	_ = s.store.AppendSecurityAudit(ctx, SecurityAudit{Event: "installation_issued", InstallationID: id, IPPrefixHash: installation.IPPrefixHash, CreatedAt: now})
	return InstallationToken{Token: token, Installation: installation}, nil
}

func (s *Service) VerifyInstallationToken(ctx context.Context, token string) (VerifiedInstallation, error) {
	id, expiresAt, valid := s.parseAndVerify(token)
	if !valid || !expiresAt.After(s.now()) {
		return VerifiedInstallation{}, ErrTokenInvalid
	}
	hash := sha256.Sum256([]byte(token))
	installation, err := s.store.FindInstallationByTokenHash(ctx, hash[:])
	if err != nil || installation.ID != id || !installation.ExpiresAt.After(s.now()) {
		return VerifiedInstallation{}, ErrTokenInvalid
	}
	revoked, err := s.store.IsTokenRevoked(ctx, hash[:])
	if err != nil || revoked {
		return VerifiedInstallation{}, ErrTokenInvalid
	}
	return VerifiedInstallation{InstallationID: installation.ID, Subject: installation.Subject, Kind: installation.Kind, ExpiresAt: installation.ExpiresAt}, nil
}

func (s *Service) ReserveChunks(ctx context.Context, verified VerifiedInstallation, chunks int) (DailyQuota, error) {
	if chunks <= 0 || chunks > s.maxChunks(verified.Kind) {
		return DailyQuota{}, ErrQuotaExceeded
	}
	quota, err := s.store.ReserveDailyQuota(ctx, "installation:"+verified.InstallationID, s.now().UTC(), s.dailyLimit(verified.Kind), chunks)
	if err != nil {
		return DailyQuota{}, err
	}
	if s.counters != nil {
		_ = s.counters.SetDaily(ctx, quota)
	}
	return quota, nil
}

func (s *Service) HashIPPrefix(ip netip.Addr) []byte {
	prefix := normalizedIPPrefix(ip)
	mac := hmac.New(sha256.New, s.ipHashKey)
	_, _ = mac.Write([]byte(prefix))
	return mac.Sum(nil)
}

func (s *Service) dailyLimit(kind InstallationKind) int {
	if kind == InstallationAuthenticated {
		return s.authDaily
	}
	return s.anonDaily
}

func (s *Service) maxChunks(kind InstallationKind) int {
	if kind == InstallationAuthenticated {
		return s.authMax
	}
	return s.anonMax
}

func (s *Service) signToken(id string, expiresAt time.Time, nonce string) string {
	payload := strings.Join([]string{"v1", id, strconv.FormatInt(expiresAt.Unix(), 10), nonce}, ".")
	mac := hmac.New(sha256.New, s.signingKey)
	_, _ = mac.Write([]byte(payload))
	return payload + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (s *Service) parseAndVerify(token string) (string, time.Time, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 5 || parts[0] != "v1" {
		return "", time.Time{}, false
	}
	expiresUnix, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return "", time.Time{}, false
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[4])
	if err != nil {
		return "", time.Time{}, false
	}
	payload := strings.Join(parts[:4], ".")
	if !validMAC(payload, signature, s.signingKey) && (len(s.previousKey) == 0 || !validMAC(payload, signature, s.previousKey)) {
		return "", time.Time{}, false
	}
	return parts[1], time.Unix(expiresUnix, 0).UTC(), true
}

func normalizedIPPrefix(ip netip.Addr) string {
	if !ip.IsValid() {
		return "invalid"
	}
	if ip.Is4() {
		return netip.PrefixFrom(ip, 24).Masked().String()
	}
	return netip.PrefixFrom(ip, 56).Masked().String()
}

func validMAC(payload string, signature, key []byte) bool {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(payload))
	return hmac.Equal(signature, mac.Sum(nil))
}

func randomHex(bytes int) (string, error) {
	buf := make([]byte, bytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
