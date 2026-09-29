package pgstore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/linkasu/linka.type-backend/internal/ttscontrol"
)

var ErrNotFound = errors.New("tts control record not found")

type Store struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) CreateInstallation(ctx context.Context, installation ttscontrol.Installation) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	day := installation.IssuedAt.UTC().Format("2006-01-02")
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtext($1))", fmt.Sprintf("tts-install:%x:%s", installation.IPPrefixHash, day)); err != nil {
		return err
	}
	var issued int
	if err := tx.QueryRow(ctx, "SELECT count(*) FROM tts_installations WHERE ip_prefix_hash = $1 AND issued_at >= date_trunc('day', $2::timestamptz) AND issued_at < date_trunc('day', $2::timestamptz) + interval '1 day'", installation.IPPrefixHash, installation.IssuedAt).Scan(&issued); err != nil {
		return err
	}
	if issued >= 3 {
		return ttscontrol.ErrIssuanceLimit
	}
	_, err = tx.Exec(ctx, "INSERT INTO tts_installations (id, token_hash, ip_prefix_hash, subject, kind, issued_at, expires_at) VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6, $7)", installation.ID, installation.TokenHash, installation.IPPrefixHash, installation.Subject, installation.Kind, installation.IssuedAt, installation.ExpiresAt)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) FindInstallationByTokenHash(ctx context.Context, tokenHash []byte) (ttscontrol.Installation, error) {
	var installation ttscontrol.Installation
	err := s.pool.QueryRow(ctx, "SELECT id, token_hash, ip_prefix_hash, COALESCE(subject, ''), kind, issued_at, expires_at FROM tts_installations WHERE token_hash = $1", tokenHash).Scan(&installation.ID, &installation.TokenHash, &installation.IPPrefixHash, &installation.Subject, &installation.Kind, &installation.IssuedAt, &installation.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ttscontrol.Installation{}, ErrNotFound
	}
	return installation, err
}

func (s *Store) RevokeToken(ctx context.Context, tokenHash []byte, reason string) error {
	_, err := s.pool.Exec(ctx, "INSERT INTO tts_token_revocations (token_hash, revoked_at, reason) VALUES ($1, now(), NULLIF($2, '')) ON CONFLICT (token_hash) DO NOTHING", tokenHash, reason)
	return err
}

func (s *Store) IsTokenRevoked(ctx context.Context, tokenHash []byte) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM tts_token_revocations WHERE token_hash = $1)", tokenHash).Scan(&exists)
	return exists, err
}

func (s *Store) ReserveDailyQuota(ctx context.Context, key string, day time.Time, limit, chunks int) (ttscontrol.DailyQuota, error) {
	var quota ttscontrol.DailyQuota
	err := s.pool.QueryRow(ctx, `INSERT INTO tts_quota_daily (quota_key, quota_day, chunks_used, chunks_limit, updated_at)
VALUES ($1, $2::date, $3, $4, now())
ON CONFLICT (quota_key, quota_day) DO UPDATE SET
 chunks_used = tts_quota_daily.chunks_used + EXCLUDED.chunks_used,
 chunks_limit = EXCLUDED.chunks_limit,
 updated_at = now()
WHERE tts_quota_daily.chunks_used + EXCLUDED.chunks_used <= EXCLUDED.chunks_limit
RETURNING quota_key, quota_day, chunks_used, chunks_limit, updated_at`, key, day.UTC(), chunks, limit).Scan(&quota.Key, &quota.Day, &quota.Used, &quota.Limit, &quota.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ttscontrol.DailyQuota{}, ttscontrol.ErrQuotaExceeded
	}
	return quota, err
}

func (s *Store) AppendSecurityAudit(ctx context.Context, audit ttscontrol.SecurityAudit) error {
	_, err := s.pool.Exec(ctx, "INSERT INTO tts_security_audit (event, installation_id, ip_prefix_hash, metadata, created_at) VALUES ($1, NULLIF($2, ''), $3, NULLIF($4, '')::jsonb, $5)", audit.Event, audit.InstallationID, audit.IPPrefixHash, string(audit.Metadata), audit.CreatedAt)
	return err
}

func (s *Store) GetIdempotency(ctx context.Context, scope, key string) (ttscontrol.IdempotencyRecord, error) {
	var record ttscontrol.IdempotencyRecord
	err := s.pool.QueryRow(ctx, "SELECT idempotency_key, scope, response, created_at, expires_at FROM tts_idempotency WHERE scope = $1 AND idempotency_key = $2 AND expires_at > now()", scope, key).Scan(&record.Key, &record.Scope, &record.Response, &record.CreatedAt, &record.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ttscontrol.IdempotencyRecord{}, ErrNotFound
	}
	return record, err
}

func (s *Store) PutIdempotency(ctx context.Context, record ttscontrol.IdempotencyRecord) error {
	_, err := s.pool.Exec(ctx, "INSERT INTO tts_idempotency (idempotency_key, scope, response, created_at, expires_at) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (scope, idempotency_key) DO NOTHING", record.Key, record.Scope, record.Response, record.CreatedAt, record.ExpiresAt)
	return err
}
