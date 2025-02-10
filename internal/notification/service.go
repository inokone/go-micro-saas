package notification

import (
	"context"

	"github.com/inokone/go-micro-saas/internal/common"
	"github.com/inokone/go-micro-saas/internal/mail"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	source chan common.Event
	mailer mail.Mailer
}

func NewService(source chan common.Event, mailer mail.Mailer) *Service {
	return &Service{
		source: source,
		mailer: mailer,
	}
}

func (s *Service) Start(ctx context.Context) {
	log.Info("Notification service starting...")
	go func() {
		for {
			select {
			case msg := <-s.source:
				log.WithField("type", msg.Type).WithField("time", msg.Time).WithField("user", msg.User).Debug("Received notification event.")
				if err := s.Send(&msg); err != nil {
					log.WithError(err).Error("Failed to send notification.")
				}

			case <-ctx.Done():
				close(s.source)
				log.Info("Notification service stopped.")
				return
			}
		}
	}()
}

func (s *Service) Send(event *common.Event) error {
	switch event.Type {
	default:
		return nil
	}
}
