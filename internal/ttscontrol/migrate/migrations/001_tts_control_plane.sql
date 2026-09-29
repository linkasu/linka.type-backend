CREATE TABLE tts_installations (
  id text PRIMARY KEY,
  token_hash bytea NOT NULL UNIQUE,
  ip_prefix_hash bytea NOT NULL,
  subject text,
  kind text NOT NULL CHECK (kind IN ('anonymous', 'authenticated')),
  issued_at timestamptz NOT NULL,
  expires_at timestamptz NOT NULL
);
CREATE INDEX tts_installations_ip_issued_at_idx ON tts_installations (ip_prefix_hash, issued_at);

CREATE TABLE tts_token_revocations (
  token_hash bytea PRIMARY KEY,
  revoked_at timestamptz NOT NULL,
  reason text
);

CREATE TABLE tts_quota_daily (
  quota_key text NOT NULL,
  quota_day date NOT NULL,
  chunks_used integer NOT NULL CHECK (chunks_used >= 0),
  chunks_limit integer NOT NULL CHECK (chunks_limit > 0),
  updated_at timestamptz NOT NULL,
  PRIMARY KEY (quota_key, quota_day)
);

CREATE TABLE tts_quota_overrides (
  quota_key text PRIMARY KEY,
  daily_chunks integer,
  max_chunks integer,
  expires_at timestamptz,
  created_at timestamptz NOT NULL
);

CREATE TABLE tts_security_audit (
  id bigserial PRIMARY KEY,
  event text NOT NULL,
  installation_id text,
  ip_prefix_hash bytea,
  metadata jsonb,
  created_at timestamptz NOT NULL
);

CREATE TABLE tts_idempotency (
  scope text NOT NULL,
  idempotency_key text NOT NULL,
  response bytea NOT NULL,
  created_at timestamptz NOT NULL,
  expires_at timestamptz NOT NULL,
  PRIMARY KEY (scope, idempotency_key)
);
