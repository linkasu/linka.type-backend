# Migration and Compatibility

## Principles
- Keep Firebase RTDB as source of truth until parity is proven.
- Preserve all existing IDs and `created` timestamps.
- All writes are idempotent and include `updated_at`.

## Dual-write
- Every mutation in core-api writes to YDB first, then mirrors to Firebase RTDB using the same IDs.
- If Firebase write fails, the request returns an error and logs a retryable event.

## Read-through seeding
- Reads check YDB first.
- If missing, core-api pulls from Firebase, writes to YDB, and returns the data.

## Sync worker
- Streams Firebase RTDB changes and applies them to YDB.
- Backfills missing `updated_at` during sync.
- Keeps `admins`, `global`, and `factory/questions` in sync.

## Feature flag rollout
- A user-cohort flag controls read source:
  - `firebase_only`: read from Firebase, write to both.
  - `ydb_primary`: read from YDB, fallback to Firebase.
- Flag is stored per user in YDB or a config table.

## Conflict resolution
- Last-write-wins using `updated_at` (server time).
- If timestamps are equal, prefer YDB or last event arrival.
