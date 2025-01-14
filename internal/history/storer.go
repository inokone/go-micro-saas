package history

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/inokone/go-micro-saas/internal/common"
	"github.com/jmoiron/sqlx"
)

// PostgresStorer is the Storer implementation based on pq library.
type PostgresStorer struct {
	db *sqlx.DB
}

// NewPostgresStorer creates a new PostgresStorer instance based on the pq library.
func NewPostgresStorer(db *sqlx.DB) *PostgresStorer {
	return &PostgresStorer{
		db: db,
	}
}

// Store is a method of the PostgresStorer struct. Takes a common.Event as parameter and persists it.
func (s *PostgresStorer) Store(event *common.Event) error {
	data, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("failed to store history event: %w", err)
	}

	values := map[string]interface{}{
		"history_event_id": event.ID,
		"user_id":          event.User,
		"event_type":       event.Type,
		"event_time":       event.Time,
		"event_data":       data,
	}

	query := `INSERT INTO microsaas.history_events(history_event_id, user_id, event_type, event_time, event_data) VALUES (:history_event_id, :user_id, :event_type, :event_time, :event_data)`
	_, err = s.db.NamedExec(query, values)
	if err != nil {
		return fmt.Errorf("failed to store history event: %w", err)
	}
	return nil
}

type raw struct {
	ID   uuid.UUID `json:"id" db:"history_event_id"`
	Type string    `json:"type" db:"event_type"`
	Time time.Time `json:"time" db:"event_time"`
	Data string    `json:"data" db:"event_data"`
	User uuid.UUID `json:"user" db:"user_id"`
}

// List is a method of the `PostgresStorer` struct. Loads all history entries for the User in parameter.
func (s *PostgresStorer) List(user uuid.UUID, limit int) ([]common.Event, error) {
	var (
		res = make([]common.Event, 0)
		raw []raw
	)

	query := `SELECT history_event_id, user_id, event_type, event_time, event_data FROM microsaas.history_events WHERE user_id = $1 order by event_time desc limit $2`
	if err := s.db.Select(&raw, query, user, limit); err != nil {
		return nil, fmt.Errorf("failed to list history events: %w", err)
	}

	for _, e := range raw {
		var event interface{}
		if err := json.Unmarshal([]byte(e.Data), &event); err != nil {
			return nil, err
		}

		res = append(res, common.Event{
			ID:   e.ID,
			Type: e.Type,
			Time: e.Time,
			Data: event,
			User: e.User,
		})

	}
	return res, nil
}
