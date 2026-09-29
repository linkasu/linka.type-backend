// Package ttscontrol contains the isolated, route-free TTS control-plane foundation.
package ttscontrol

import (
	"context"
	"errors"
	"time"
)

const TokenTTL = 30 * 24 * time.Hour

var (
	ErrIssuanceLimit = errors.New("tts installation issuance limit exceeded")
	ErrQuotaExceeded = errors.New("tts quota exceeded")
	ErrTokenInvalid  = errors.New("tts installation token is invalid")
)

type InstallationKind string

const (
	InstallationAnonymous     InstallationKind = "anonymous"
	InstallationAuthenticated InstallationKind = "authenticated"
)

type Installation struct {
	ID           string
	TokenHash    []byte
	IPPrefixHash []byte
	Subject      string
	Kind         InstallationKind
	IssuedAt     time.Time
	ExpiresAt    time.Time
}

type InstallationToken struct {
	Token        string
	Installation Installation
}

type VerifiedInstallation struct {
	InstallationID string
	Subject        string
	Kind           InstallationKind
	ExpiresAt      time.Time
}

type DailyQuota struct {
	Key       string
	Day       time.Time
	Used      int
	Limit     int
	UpdatedAt time.Time
}

type SecurityAudit struct {
	Event          string
	InstallationID string
	IPPrefixHash   []byte
	Metadata       []byte
	CreatedAt      time.Time
}

type IdempotencyRecord struct {
	Key       string
	Scope     string
	Response  []byte
	CreatedAt time.Time
	ExpiresAt time.Time
}

// Store persists control-plane state. PostgreSQL is the authoritative implementation.
type Store interface {
	CreateInstallation(context.Context, Installation) error
	FindInstallationByTokenHash(context.Context, []byte) (Installation, error)
	RevokeToken(context.Context, []byte, string) error
	IsTokenRevoked(context.Context, []byte) (bool, error)
	ReserveDailyQuota(context.Context, string, time.Time, int, int) (DailyQuota, error)
	AppendSecurityAudit(context.Context, SecurityAudit) error
	GetIdempotency(context.Context, string, string) (IdempotencyRecord, error)
	PutIdempotency(context.Context, IdempotencyRecord) error
}

// CounterCache is disposable hot state; it can always be rebuilt from tts_quota_daily.
type CounterCache interface {
	SetDaily(context.Context, DailyQuota) error
}
