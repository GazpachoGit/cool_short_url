package eventsender

import (
	"context"
	"log/slog"
	"short-url/internal/http-server/model/domain"
	"short-url/internal/storage"
	"short-url/internal/storage/sqlite"
	"time"
)

type Sender struct {
	storage *sqlite.Storage
	log     *slog.Logger
}

func New(storage *sqlite.Storage, log *slog.Logger) *Sender {
	return &Sender{
		storage: storage,
		log:     log,
	}
}

func (s *Sender) StartProcessEvents(ctx context.Context, handelPeriod time.Duration) {
	const op = "event-sender.StartProcessEvents"
	log := s.log.With(slog.String("op", op))

	ticker := time.NewTicker(handelPeriod)

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Info("context done, stopping event sender")
				return
			case <-ticker.C:
			}
			ev, err := s.storage.GetNewEvent()
			if err != nil {
				if err == storage.ErrEventNotFound {
					log.Debug("no new events found, waiting for next tick")
					continue
				}
				log.Error("error getting new event", slog.Any("error", err))
				continue
			}

			s.stubSendEventMessage(ev)

			if err := s.storage.MarkEventAsDone(ev.ID); err != nil {
				log.Error("error marking event as done", slog.Any("error", err), slog.Int("event_id", ev.ID))
				continue
			}
		}
	}()
}

func (s *Sender) stubSendEventMessage(event domain.Event) {
	const op = "event-sender.StubSendEventMessage"

	log := s.log.With(slog.String("op", op))

	log.Info("Sending event message", slog.Any("event", event))
}
