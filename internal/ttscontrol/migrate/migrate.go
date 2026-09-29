package migrate

import (
	"context"
	"crypto/sha256"
	"embed"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const ledgerTable = "tts_schema_migrations"

//go:embed migrations/*.sql
var files embed.FS

type Migration struct {
	Version  int
	Name     string
	SQL      string
	Checksum string
}

func Embedded() ([]Migration, error) {
	entries, err := files.ReadDir("migrations")
	if err != nil {
		return nil, err
	}
	var migrations []Migration
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		parts := strings.SplitN(entry.Name(), "_", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid migration name %q", entry.Name())
		}
		version, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid migration name %q: %w", entry.Name(), err)
		}
		body, err := files.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return nil, err
		}
		checksum := sha256.Sum256(body)
		migrations = append(migrations, Migration{Version: version, Name: entry.Name(), SQL: string(body), Checksum: fmt.Sprintf("%x", checksum[:])})
	}
	sort.Slice(migrations, func(i, j int) bool { return migrations[i].Version < migrations[j].Version })
	for i, migration := range migrations {
		if i > 0 && migration.Version == migrations[i-1].Version {
			return nil, fmt.Errorf("duplicate migration version %d", migration.Version)
		}
	}
	return migrations, nil
}

func Plan(migrations []Migration, applied map[int]string) ([]Migration, error) {
	plan := make([]Migration, 0, len(migrations))
	known := make(map[int]struct{}, len(migrations))
	for _, migration := range migrations {
		known[migration.Version] = struct{}{}
		if checksum, ok := applied[migration.Version]; ok {
			if checksum != migration.Checksum {
				return nil, fmt.Errorf("migration %d checksum mismatch", migration.Version)
			}
			continue
		}
		plan = append(plan, migration)
	}
	for version := range applied {
		if _, ok := known[version]; !ok {
			return nil, fmt.Errorf("unknown applied migration version %d", version)
		}
	}
	return plan, nil
}

func Run(ctx context.Context, pool *pgxpool.Pool) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock(hashtext('linka-tts-schema-migrations'))"); err != nil {
		return err
	}
	defer conn.Exec(context.Background(), "SELECT pg_advisory_unlock(hashtext('linka-tts-schema-migrations'))")
	if _, err := conn.Exec(ctx, "CREATE TABLE IF NOT EXISTS "+ledgerTable+" (version integer PRIMARY KEY, checksum text NOT NULL, applied_at timestamptz NOT NULL)"); err != nil {
		return err
	}
	migrations, err := Embedded()
	if err != nil {
		return err
	}
	applied, err := appliedMigrations(ctx, conn)
	if err != nil {
		return err
	}
	plan, err := Plan(migrations, applied)
	if err != nil {
		return err
	}
	for _, migration := range plan {
		tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, migration.SQL); err == nil {
			_, err = tx.Exec(ctx, "INSERT INTO "+ledgerTable+" (version, checksum, applied_at) VALUES ($1, $2, $3)", migration.Version, migration.Checksum, time.Now().UTC())
		}
		if err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("apply migration %d: %w", migration.Version, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %d: %w", migration.Version, err)
		}
	}
	return nil
}

func Check(ctx context.Context, pool *pgxpool.Pool) error {
	migrations, err := Embedded()
	if err != nil {
		return err
	}
	applied, err := appliedMigrations(ctx, pool)
	if err != nil {
		return fmt.Errorf("read TTS schema ledger: %w", err)
	}
	_, err = Plan(migrations, applied)
	if err != nil {
		return err
	}
	if len(applied) != len(migrations) {
		return errors.New("TTS schema is not current; run cmd/tts-migrate manually")
	}
	return nil
}

type queryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func appliedMigrations(ctx context.Context, q queryer) (map[int]string, error) {
	rows, err := q.Query(ctx, "SELECT version, checksum FROM "+ledgerTable)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	applied := map[int]string{}
	for rows.Next() {
		var version int
		var checksum string
		if err := rows.Scan(&version, &checksum); err != nil {
			return nil, err
		}
		applied[version] = checksum
	}
	return applied, rows.Err()
}
