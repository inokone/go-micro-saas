package history

import (
	"github.com/google/uuid"
	"github.com/inokone/go-micro-saas/internal/common"
)

type Storer interface {
	Store(event *common.Event) error

	List(usr uuid.UUID, limit int) ([]common.Event, error)
}
