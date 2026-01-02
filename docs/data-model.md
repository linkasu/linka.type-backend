# Data Model (YDB)

## Conventions
- Time fields are stored as epoch milliseconds (int64).
- `deleted_at` is nullable; soft-deleted records are filtered from reads.
- IDs preserve existing Firebase IDs (16-char or push IDs).

## Tables
### users
- PK: `user_id`
- Fields: `email`, `created_at`, `inited`, `deleted_at`

### admins
- PK: `user_id`

### categories
- PK: (`user_id`, `category_id`)
- Fields: `label`, `created_at`, `is_default`, `updated_at`, `deleted_at`

### statements
- PK: (`user_id`, `category_id`, `statement_id`)
- Fields: `text`, `created_at`, `updated_at`, `deleted_at`

### quickes
- PK: (`user_id`, `slot`)
- Fields: `text`, `updated_at`
- `slot` range is 0-5.

### global_categories
- PK: `category_id`
- Fields: `label`, `created_at`, `is_default`, `updated_at`, `deleted_at`

### global_statements
- PK: (`category_id`, `statement_id`)
- Fields: `text`, `created_at`, `updated_at`, `deleted_at`

### factory_questions
- PK: `question_id`
- Fields: `label`, `phrases` (JSON list), `category`, `type`, `order_index`

### changes
- PK: (`user_id`, `cursor`)
- Fields: `entity_type`, `entity_id`, `op`, `payload` (JSON), `updated_at`
- `cursor` is an opaque, monotonically sortable value (ULID or counter).
