package telegram

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sonyamoonglade/poison-tg/pkg/logger"
	"github.com/sonyamoonglade/poison-tg/pkg/utils/numeric"
	"go.uber.org/zap"
)

var (
	ErrNoHandler     = errors.New("handler not found")
	ErrInvalidUpdate = errors.New("invalid update")
)

type RouteHandler interface {
	Start(ctx context.Context, chatID int64) error
	Menu(ctx context.Context, chatID int64) error
	Catalog(ctx context.Context, chatID int64) error
	GetBucket(ctx context.Context, chatID int64) error
	// StartMakeOrderGuide is initial guide handler
	StartMakeOrderGuide(ctx context.Context, m *tg.Message) error
	// MakeOrderGuideStep1
	// Can go to step 1 handler only from going backwards from step 2
	MakeOrderGuideStep1(ctx context.Context, chatID int64, controlButtonsMessageID int, instructionMsgIDs ...int64) error
	MakeOrderGuideStep2(ctx context.Context, chatID int64, controlButtonsMessageID int, instructionMsgIDs ...int64) error
	MakeOrderGuideStep3(ctx context.Context, chatID int64, controlButtonsMessageID int, instructionMsgIDs ...int64) error
	MakeOrderGuideStep4(ctx context.Context, chatID int64, controlButtonsMessageID int, instructionMsgID ...int64) error
	MakeOrder(ctx context.Context, m *tg.Message) error
	HandleSizeInput(ctx context.Context, m *tg.Message) error
	HandlePriceInput(ctx context.Context, m *tg.Message) error
	HandleError(ctx context.Context, err error, m tg.Update)
	AnswerCallback(callbackID string) error
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

			ctx, cancel := context.WithTimeout(context.Background(), r.handlerTimeout)
			r.wg.Add(1)
			go func() {
				if err := r.mapToHandler(ctx, update); err != nil {
					logger.Get().Error("error in handler occurred",
						zap.String("from", update.FromChat().UserName),
						zap.Int64("userID", update.FromChat().ID),
						zap.Error(err))
					r.h.HandleError(ctx, err, update)
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
	case u.CallbackQuery != nil:
		return r.mapToCallbackHandler(ctx, u.CallbackQuery)
	default:
		return ErrInvalidUpdate
	}
}

func (r *Router) mapToMessageHandler(ctx context.Context, m *tg.Message) error {
	var (
		chatID = m.Chat.ID
		cmd    = m.Text
	)

	logger.Get().Debug("message info",
		zap.String("text", m.Text),
		zap.Time("date", m.Time()))

	switch true {
	case r.command(cmd, "/start"):
		return r.h.Start(ctx, chatID)
	case r.command(cmd, "/menu"):
		return r.h.Menu(ctx, chatID)
	case r.isSizeReply(cmd):
		return r.h.HandleSizeInput(ctx, m)
	case r.isPriceReply(cmd):
		return r.h.HandlePriceInput(ctx, m)
	default:
		return ErrNoHandler
	}
}

func (r *Router) mapToCallbackHandler(ctx context.Context, c *tg.CallbackQuery) error {

	defer logger.Get().Debug("callback info",
		zap.String("data", c.Data),
		zap.Time("date", c.Message.Time()))
	defer r.h.AnswerCallback(c.ID)

	var (
		chatID             = c.Message.Chat.ID
		msgID              = c.Message.MessageID
		intCallbackData    int
		callbackDataMsgIDs []int64
	)
	injectedMsgID, callback, err := parseCallbackData(c.Data)
	if err != nil {
		return fmt.Errorf("parseCallbackData: %w", err)
	}
	intCallbackData = callback
	callbackDataMsgIDs = injectedMsgID

	switch intCallbackData {
	case menuMakeOrderCallback:
		return r.h.StartMakeOrderGuide(ctx, c.Message)
	case orderGuideStep1Callback:
		return r.h.MakeOrderGuideStep1(ctx, chatID, msgID, callbackDataMsgIDs...)
	case orderGuideStep2Callback:
		return r.h.MakeOrderGuideStep2(ctx, chatID, msgID, callbackDataMsgIDs...)
	case orderGuideStep3Callback:
		return r.h.MakeOrderGuideStep3(ctx, chatID, msgID, callbackDataMsgIDs...)
	case orderGuideStep4Callback:
		return r.h.MakeOrderGuideStep4(ctx, chatID, msgID, callbackDataMsgIDs...)
	case makeOrderCallback:
		return r.h.MakeOrder(ctx, c.Message)
	case menuGetBucketCallback:
		return r.h.GetBucket(ctx, c.From.ID)
	case menuTrackOrderCallback:
		return nil
	case menuCatalogCallback:
		return r.h.Catalog(ctx, c.From.ID)
	default:
		return ErrNoHandler
	}
}

func (r *Router) isSizeReply(text string) bool {
	if text == "#" {
		return true
	}
	if strings.HasPrefix(text, "http") || strings.HasPrefix(text, "https") {
		return false
	}
	for _, r := range "0123456789" {
		if strings.ContainsRune(text, r) {
			return true
		}
	}
	return false
}

func (r *Router) isPriceReply(text string) bool {
	if strings.HasPrefix(text, "http") || strings.HasPrefix(text, "https") {
		return false
	}
	return numeric.AllAreDigits(text)
}

func (r *Router) command(actual, want string) bool {
	return actual == want
}
