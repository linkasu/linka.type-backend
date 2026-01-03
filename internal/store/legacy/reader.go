package legacy

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"firebase.google.com/go/v4/db"
	"github.com/linkasu/linka.type-backend/internal/models"
	"github.com/linkasu/linka.type-backend/internal/store"
)

// Reader reads data from Firebase RTDB.
type Reader struct {
	db *db.Client
}

// NewReader creates a legacy reader.
func NewReader(client *db.Client) (*Reader, error) {
	if client == nil {
		return nil, fmt.Errorf("firebase db client is nil")
	}
	return &Reader{db: client}, nil
}

type firebaseCategory struct {
	ID         string                       `json:"id"`
	Label      string                       `json:"label"`
	Created    int64                        `json:"created"`
	Default    *bool                        `json:"default,omitempty"`
	Statements map[string]firebaseStatement `json:"statements,omitempty"`
}

type firebaseStatement struct {
	ID         string `json:"id"`
	CategoryID string `json:"categoryId"`
	Text       string `json:"text"`
	Created    int64  `json:"created"`
}

type firebaseQuestion struct {
	Label      string   `json:"label"`
	Phrases    []string `json:"phrases"`
	Category   string   `json:"category"`
	Type       string   `json:"type"`
	OrderIndex int      `json:"order_index"`
}

// FetchUserData returns categories and statements for a user.
func (r *Reader) FetchUserData(ctx context.Context, userID string) ([]models.Category, []models.Statement, error) {
	ref := r.db.NewRef(fmt.Sprintf("users/%s/Category", userID))
	var raw map[string]firebaseCategory
	if err := ref.Get(ctx, &raw); err != nil {
		return nil, nil, err
	}
	if raw == nil {
		return nil, nil, nil
	}

	now := time.Now().UnixMilli()
	categories := make([]models.Category, 0, len(raw))
	statements := make([]models.Statement, 0)

	for key, cat := range raw {
		catID := cat.ID
		if catID == "" {
			catID = key
		}
		created := cat.Created
		if created == 0 {
			created = now
		}
		categories = append(categories, models.Category{
			ID:        catID,
			Label:     cat.Label,
			Created:   created,
			Default:   cat.Default,
			UpdatedAt: created,
		})

		for stmtKey, stmt := range cat.Statements {
			stmtID := stmt.ID
			if stmtID == "" {
				stmtID = stmtKey
			}
			stmtCatID := stmt.CategoryID
			if stmtCatID == "" {
				stmtCatID = catID
			}
			stmtCreated := stmt.Created
			if stmtCreated == 0 {
				stmtCreated = now
			}
			statements = append(statements, models.Statement{
				ID:         stmtID,
				CategoryID: stmtCatID,
				Text:       stmt.Text,
				Created:    stmtCreated,
				UpdatedAt:  stmtCreated,
			})
		}
	}

	return categories, statements, nil
}

// GetUserState returns inited flag and quick phrases for a user.
func (r *Reader) GetUserState(ctx context.Context, userID string) (models.UserState, error) {
	ref := r.db.NewRef(fmt.Sprintf("users/%s", userID))
	var raw map[string]any
	if err := ref.Get(ctx, &raw); err != nil {
		return models.UserState{}, err
	}
	if raw == nil {
		return models.UserState{}, nil
	}

	state := models.UserState{}
	if inited, ok := raw["inited"].(bool); ok {
		state.Inited = inited
	}
	state.Quickes = parseQuickes(raw["quickes"])

	return state, nil
}

// ListGlobalCategories returns global categories with nested statements.
func (r *Reader) ListGlobalCategories(ctx context.Context) ([]models.GlobalCategory, error) {
	ref := r.db.NewRef("global/Category")
	var raw map[string]firebaseCategory
	if err := ref.Get(ctx, &raw); err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, nil
	}

	now := time.Now().UnixMilli()
	categories := make([]models.GlobalCategory, 0, len(raw))

	for key, cat := range raw {
		catID := cat.ID
		if catID == "" {
			catID = key
		}
		created := cat.Created
		if created == 0 {
			created = now
		}
		global := models.GlobalCategory{
			ID:        catID,
			Label:     cat.Label,
			Created:   created,
			Default:   cat.Default,
			UpdatedAt: created,
		}

		for stmtKey, stmt := range cat.Statements {
			stmtID := stmt.ID
			if stmtID == "" {
				stmtID = stmtKey
			}
			stmtCatID := stmt.CategoryID
			if stmtCatID == "" {
				stmtCatID = catID
			}
			stmtCreated := stmt.Created
			if stmtCreated == 0 {
				stmtCreated = now
			}
			global.Statements = append(global.Statements, models.Statement{
				ID:         stmtID,
				CategoryID: stmtCatID,
				Text:       stmt.Text,
				Created:    stmtCreated,
				UpdatedAt:  stmtCreated,
			})
		}

		categories = append(categories, global)
	}

	return categories, nil
}

// ListFactoryQuestions returns onboarding question templates.
func (r *Reader) ListFactoryQuestions(ctx context.Context) ([]models.FactoryQuestion, error) {
	ref := r.db.NewRef("factory/questions")
	var raw any
	if err := ref.Get(ctx, &raw); err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, nil
	}

	questions := parseFactoryQuestions(raw)
	sort.Slice(questions, func(i, j int) bool {
		return questions[i].OrderIndex < questions[j].OrderIndex
	})
	return questions, nil
}

func parseFactoryQuestions(raw any) []models.FactoryQuestion {
	questions := make([]models.FactoryQuestion, 0)
	switch value := raw.(type) {
	case []any:
		for idx, item := range value {
			q := questionFromAny(item)
			if q.ID == "" {
				q.ID = fmt.Sprintf("%d", idx)
			}
			if q.OrderIndex == 0 {
				q.OrderIndex = idx
			}
			questions = append(questions, q)
		}
	case map[string]any:
		idx := 0
		for key, item := range value {
			q := questionFromAny(item)
			if q.ID == "" {
				q.ID = key
			}
			if q.OrderIndex == 0 {
				q.OrderIndex = idx
			}
			questions = append(questions, q)
			idx++
		}
	}
	return questions
}

func questionFromAny(raw any) models.FactoryQuestion {
	data, ok := raw.(map[string]any)
	if !ok {
		return models.FactoryQuestion{}
	}
	q := models.FactoryQuestion{
		Label:    str(data["label"], ""),
		Category: str(data["category"], ""),
		Type:     str(data["type"], ""),
	}
	if uid := str(data["uid"], ""); uid != "" {
		q.ID = uid
	}
	if id := str(data["id"], ""); id != "" && q.ID == "" {
		q.ID = id
	}

	if rawOrder, ok := data["order_index"]; ok {
		q.OrderIndex = int(int64From(rawOrder, 0))
	}
	if rawOrder, ok := data["orderIndex"]; ok && q.OrderIndex == 0 {
		q.OrderIndex = int(int64From(rawOrder, 0))
	}

	q.Phrases = toStringSlice(data["phrases"])
	return q
}

func toStringSlice(raw any) []string {
	switch value := raw.(type) {
	case []string:
		return value
	case []any:
		out := make([]string, 0, len(value))
		for _, item := range value {
			if strVal, ok := item.(string); ok {
				out = append(out, strVal)
			}
		}
		return out
	default:
		return nil
	}
}

// IsAdmin checks if a user is in the admins list.
func (r *Reader) IsAdmin(ctx context.Context, userID string) (bool, error) {
	ref := r.db.NewRef(fmt.Sprintf("admins/%s", userID))
	var value any
	if err := ref.Get(ctx, &value); err != nil {
		return false, err
	}
	if value == nil {
		return false, nil
	}
	if b, ok := value.(bool); ok {
		return b, nil
	}
	return true, nil
}

func parseQuickes(raw any) []string {
	switch val := raw.(type) {
	case nil:
		return nil
	case []string:
		return val
	case []any:
		out := make([]string, 0, len(val))
		for _, item := range val {
			if str, ok := item.(string); ok {
				out = append(out, str)
			}
		}
		return out
	case map[string]any:
		slots := make([]int, 0, len(val))
		for key := range val {
			idx, err := strconv.Atoi(key)
			if err != nil {
				continue
			}
			slots = append(slots, idx)
		}
		sort.Ints(slots)
		out := make([]string, 0, len(slots))
		for _, idx := range slots {
			rawVal := val[strconv.Itoa(idx)]
			if str, ok := rawVal.(string); ok {
				out = append(out, str)
			}
		}
		return out
	default:
		return nil
	}
}

func str(raw any, fallback string) string {
	if value, ok := raw.(string); ok && value != "" {
		return value
	}
	return fallback
}

func int64From(raw any, fallback int64) int64 {
	switch value := raw.(type) {
	case int64:
		return value
	case int:
		return int64(value)
	case float64:
		return int64(value)
	default:
		return fallback
	}
}

var _ store.LegacyReader = (*Reader)(nil)
