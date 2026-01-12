package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/linkasu/linka.type-backend/internal/config"
	"github.com/linkasu/linka.type-backend/internal/ydb"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
)

var schema = []string{
`CREATE TABLE IF NOT EXISTS users (
  user_id Utf8 NOT NULL,
  email Optional<Utf8>,
  created_at Int64 NOT NULL,
  inited Bool NOT NULL,
  preferences JsonDocument,
  deleted_at Optional<Int64>,
  PRIMARY KEY (user_id)
);`,
	`CREATE TABLE IF NOT EXISTS admins (
  user_id Utf8 NOT NULL,
  PRIMARY KEY (user_id)
);`,
`CREATE TABLE IF NOT EXISTS categories (
  user_id Utf8 NOT NULL,
  category_id Utf8 NOT NULL,
  label Utf8 NOT NULL,
  created_at Int64 NOT NULL,
  is_default Optional<Bool>,
  ai_use Optional<Bool>,
  updated_at Int64 NOT NULL,
  deleted_at Optional<Int64>,
  PRIMARY KEY (user_id, category_id)
);`,
	`CREATE TABLE IF NOT EXISTS statements (
  user_id Utf8 NOT NULL,
  category_id Utf8 NOT NULL,
  statement_id Utf8 NOT NULL,
  text Utf8 NOT NULL,
  created_at Int64 NOT NULL,
  updated_at Int64 NOT NULL,
  deleted_at Optional<Int64>,
  PRIMARY KEY (user_id, category_id, statement_id)
);`,
	`CREATE TABLE IF NOT EXISTS quickes (
  user_id Utf8 NOT NULL,
  slot Int64 NOT NULL,
  text Utf8 NOT NULL,
  updated_at Int64 NOT NULL,
  PRIMARY KEY (user_id, slot)
);`,
	`CREATE TABLE IF NOT EXISTS global_categories (
  category_id Utf8 NOT NULL,
  label Utf8 NOT NULL,
  created_at Int64 NOT NULL,
  is_default Optional<Bool>,
  updated_at Int64 NOT NULL,
  deleted_at Optional<Int64>,
  PRIMARY KEY (category_id)
);`,
	`CREATE TABLE IF NOT EXISTS global_statements (
  category_id Utf8 NOT NULL,
  statement_id Utf8 NOT NULL,
  text Utf8 NOT NULL,
  created_at Int64 NOT NULL,
  updated_at Int64 NOT NULL,
  deleted_at Optional<Int64>,
  PRIMARY KEY (category_id, statement_id)
);`,
	`CREATE TABLE IF NOT EXISTS factory_questions (
  question_id Utf8 NOT NULL,
  label Utf8 NOT NULL,
  phrases JsonDocument NOT NULL,
  category Utf8 NOT NULL,
  type Utf8 NOT NULL,
  order_index Int64 NOT NULL,
  PRIMARY KEY (question_id)
);`,
	`CREATE TABLE IF NOT EXISTS changes (
  user_id Utf8 NOT NULL,
  cursor Utf8 NOT NULL,
  entity_type Utf8 NOT NULL,
  entity_id Utf8 NOT NULL,
  op Utf8 NOT NULL,
  payload JsonDocument NOT NULL,
  updated_at Int64 NOT NULL,
  PRIMARY KEY (user_id, cursor)
);`,
	`CREATE TABLE IF NOT EXISTS client_keys (
  key_hash Utf8 NOT NULL,
  client_id Utf8 NOT NULL,
  status Utf8 NOT NULL,
  created_at Int64 NOT NULL,
  revoked_at Optional<Int64>,
  PRIMARY KEY (key_hash)
);`,
	`CREATE TABLE IF NOT EXISTS dialog_chats (
  user_id Utf8 NOT NULL,
  chat_id Utf8 NOT NULL,
  title Utf8 NOT NULL,
  created_at Int64 NOT NULL,
  updated_at Int64 NOT NULL,
  last_message_at Optional<Int64>,
  message_count Optional<Int64>,
  deleted_at Optional<Int64>,
  PRIMARY KEY (user_id, chat_id)
);`,
	`CREATE TABLE IF NOT EXISTS dialog_messages (
  user_id Utf8 NOT NULL,
  chat_id Utf8 NOT NULL,
  message_id Utf8 NOT NULL,
  role Utf8 NOT NULL,
  content Utf8 NOT NULL,
  source Optional<Utf8>,
  created_at Int64 NOT NULL,
  updated_at Int64 NOT NULL,
  deleted_at Optional<Int64>,
  PRIMARY KEY (user_id, chat_id, message_id)
);`,
	`CREATE TABLE IF NOT EXISTS dialog_suggestions (
  user_id Utf8 NOT NULL,
  suggestion_id Utf8 NOT NULL,
  chat_id Optional<Utf8>,
  message_id Optional<Utf8>,
  text Utf8 NOT NULL,
  status Utf8 NOT NULL,
  category_id Optional<Utf8>,
  created_at Int64 NOT NULL,
  updated_at Int64 NOT NULL,
  PRIMARY KEY (user_id, suggestion_id)
);`,
	`CREATE TABLE IF NOT EXISTS dialog_suggestion_jobs (
  job_id Utf8 NOT NULL,
  user_id Utf8 NOT NULL,
  chat_id Utf8 NOT NULL,
  message_id Utf8 NOT NULL,
  status Utf8 NOT NULL,
  attempts Int64 NOT NULL,
  last_error Optional<Utf8>,
  created_at Int64 NOT NULL,
  updated_at Int64 NOT NULL,
  PRIMARY KEY (job_id)
);`,
	`CREATE TABLE IF NOT EXISTS usage_limits (
  user_id Utf8 NOT NULL,
  month Utf8 NOT NULL,
  inference_count Int64 NOT NULL,
  max_limit Int64 NOT NULL,
  updated_at Int64 NOT NULL,
  PRIMARY KEY (user_id, month)
);`,
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config load: %v\n", err)
		os.Exit(1)
	}

	client, err := ydb.New(ctx, cfg.YDB)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ydb init: %v\n", err)
		os.Exit(1)
	}
	defer client.Close(ctx)

	for _, stmt := range schema {
		query := withPrefix(client.Database(), stmt)
		err := client.Table().Do(ctx, func(ctx context.Context, sess table.Session) error {
			return sess.ExecuteSchemeQuery(ctx, query)
		}, table.WithIdempotent())
		if err != nil {
			fmt.Fprintf(os.Stderr, "schema apply failed: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("schema applied")
}

func withPrefix(database, stmt string) string {
	return fmt.Sprintf("PRAGMA TablePathPrefix(\"%s\");\n%s", database, stmt)
}
