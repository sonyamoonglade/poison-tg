package telegram

import (
	"context"
	"errors"
	"sync"
	"time"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sonyamoonglade/poison-tg/pkg/logger"
	"go.uber.org/zap"
)

var (
	ErrNoHandler     = errors.New("handler not found")
	ErrInvalidUpdate = errors.New("invalid update")
)

type RouteHandler interface {
	Menu(ctx context.Context, m *tg.Message) error
	HandleError(ctx context.Context, err error, m tg.Update)
}

type Router struct {
	h              RouteHandler
	updates        <-chan tg.Update
	shutdown       chan struct{}
	handlerTimeout time.Duration
	wg             *sync.WaitGroup
}

func NewRouter(updates <-chan tg.Update, h RouteHandler, timeout time.Duration) *Router {
	return &Router{
		h:              h,
		updates:        updates,
		handlerTimeout: timeout,
		shutdown:       make(chan struct{}),
		wg:             new(sync.WaitGroup),
	}
}

func (r *Router) Bootstrap() {
	logger.Get().Info("router is listening for updates")
	for {
		select {
		case <-r.shutdown:
			logger.Get().Info("router is shutting down")
			return
		case update, ok := <-r.updates:
			if !ok {
				return
			}

			logger.Get().Debug("new update",
				zap.String("from", update.FromChat().UserName),
				zap.Int64("userID", update.FromChat().ID))

			ctx, cancel := context.WithTimeout(context.Background(), r.handlerTimeout)
			r.wg.Add(1)
			go func() {
				if err := r.mapToHandler(ctx, update); err != nil {
					r.h.HandleError(ctx, err, update)
					logger.Get().Error("error in handler occurred",
						zap.String("from", update.FromChat().UserName),
						zap.Int64("userID", update.FromChat().ID))
				}
				defer cancel()
				defer r.wg.Done()
			}()
		}
	}
}

func (r *Router) Shutdown() {
	close(r.shutdown)
	r.wg.Wait()
}

func (r *Router) mapToHandler(ctx context.Context, u tg.Update) error {
	switch {
	case u.Message != nil:
		return r.mapToMessageHandler(ctx, u.Message)
	default:
		return ErrInvalidUpdate
	}
}

func (r *Router) mapToMessageHandler(ctx context.Context, m *tg.Message) error {
	logger.Get().Debug("message info",
		zap.String("text", m.Text),
		zap.Time("date", m.Time()))
	switch m.Text {
	case "/menu":
		return r.h.Menu(ctx, m)
	default:
		return ErrNoHandler
	}
}
