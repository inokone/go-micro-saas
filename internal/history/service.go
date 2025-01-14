package history

import (
	"context"

	"github.com/inokone/go-micro-saas/internal/common"
	log "github.com/sirupsen/logrus"
)

type WriterService struct {
	source chan common.Event
	events Storer
}

func NewService(source chan common.Event, events Storer) *WriterService {
	return &WriterService{
		source: source,
		events: events,
	}
}

func (w *WriterService) Start(ctx context.Context) {
	log.Info("History service starting...")
	go func() {
		for {
			select {
			case msg := <-w.source:
				// Process the message
				log.WithField("type", msg.Type).WithField("time", msg.Time).WithField("user", msg.User).Debug("Received history event.")
				if err := w.events.Store(&msg); err != nil {
					log.WithError(err).Error("Failed to store history event.")
				}

			case <-ctx.Done():
				close(w.source)
				log.Info("History service stopped.")
				return
			}
		}
	}()
}
