package telegram

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sonyamoonglade/poison-tg/internal/domain"
	"github.com/sonyamoonglade/poison-tg/pkg/logger"
	"go.uber.org/zap"
)

var (
	ErrNoHandler = errors.New("handler not found")
	ErrNoRoute   = errors.New("no route")
)

type StateProvider interface {
	GetState(ctx context.Context, telegramID int64) (domain.State, error)
}

type RouteHandler interface {
	// Main menu
	Start(ctx context.Context, m *tg.Message) error
	Menu(ctx context.Context, chatID int64) error
	Catalog(ctx context.Context, chatID int64) error
	Calculator(ctx context.Context, chatID int64) error

	HandleCalculatorInput(ctx context.Context, m *tg.Message) error

	GetCart(ctx context.Context, chatID int64) error
	EditCart(ctx context.Context, chatID int64, cartPreviewMsgID int) error
	RemoveCartPosition(ctx context.Context, chatID int64, callbackData int, originalMsgID, cartPreviewMsgID int64) error

	// Add position is like StartmakeOrderGuide but without instruction
	AddPosition(ctx context.Context, m *tg.Message) error

	HandleLocationInput(ctx context.Context, chatID int64, loc domain.Location) error
	HandleOrderTypeInput(ctx context.Context, chatID int64, typ domain.OrderType) error

	// StartMakeOrderGuide is initial guide handler
	StartMakeOrderGuide(ctx context.Context, m *tg.Message) error
	// MakeOrderGuideStep1
	// Can go to step 1 handler only from going backwards from step 2
	MakeOrderGuideStep1(ctx context.Context, chatID int64, controlButtonsMessageID int, instructionMsgIDs ...int64) error
	MakeOrderGuideStep2(ctx context.Context, chatID int64, controlButtonsMessageID int, instructionMsgIDs ...int64) error
	MakeOrderGuideStep3(ctx context.Context, chatID int64, controlButtonsMessageID int, instructionMsgIDs ...int64) error
	MakeOrderGuideStep4(ctx context.Context, chatID int64, controlButtonsMessageID int, instructionMsgIDs ...int64) error

	AskForFIO(ctx context.Context, chatID int64) error
	// use tg.Message because user types in and userID is user's
	HandleFIOInput(ctx context.Context, m *tg.Message) error
	HandlePhoneNumberInput(ctx context.Context, m *tg.Message) error
	HandleDeliveryAddressInput(ctx context.Context, m *tg.Message) error

	HandleSizeInput(ctx context.Context, m *tg.Message) error
	// use tg.CallbackQuery because callback is asosiated with c.User.ID, message is from bot
	HandleButtonSelect(ctx context.Context, c *tg.CallbackQuery, button domain.Button) error
	HandlePriceInput(ctx context.Context, m *tg.Message) error
	HandleLinkInput(ctx context.Context, m *tg.Message) error

	// Utils
	HandleError(ctx context.Context, err error, m tg.Update)
	AnswerCallback(callbackID string) error
}

type Router struct {
	h              RouteHandler
	updates        <-chan tg.Update
	shutdown       chan struct{}
	handlerTimeout time.Duration
	wg             *sync.WaitGroup
	stateProvider  StateProvider
}

func NewRouter(updates <-chan tg.Update, h RouteHandler, stateProvider StateProvider, timeout time.Duration) *Router {
	return &Router{
		h:              h,
		updates:        updates,
		handlerTimeout: timeout,
		stateProvider:  stateProvider,
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
					var username string = "default"
					var id int64 = 0
					if update.FromChat() != nil {
						username = update.FromChat().UserName
						id = update.FromChat().ID
					}
					logger.Get().Error("error in handler occurred",
						zap.String("from", username),
						zap.Int64("userID", id),
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
		return r.mapToCommandHandler(ctx, u.Message)
	case u.CallbackQuery != nil:
		return r.mapToCallbackHandler(ctx, u.CallbackQuery)
	default:
		return ErrNoRoute
	}
}

func (r *Router) mapToCommandHandler(ctx context.Context, m *tg.Message) error {
	var (
		chatID = m.Chat.ID
		cmd    = r.command(m.Text)
	)
	// get state and route accordingly
	logger.Get().Debug("message info",
		zap.String("text", m.Text),
		zap.Time("date", m.Time()))
	switch true {
	case cmd(startCommand):
		return r.h.Start(ctx, m)
	case cmd(menuCommand):
		return r.h.Menu(ctx, chatID)
	case cmd(getCartCommand):
		return r.h.GetCart(ctx, chatID)
	case cmd(addPositionCommand):
		return r.h.AddPosition(ctx, m)
	default:
		customerState, err := r.stateProvider.GetState(ctx, chatID)
		if err != nil {
			return err
		}
		switch customerState {
		case domain.StateWaitingForSize:
			return r.h.HandleSizeInput(ctx, m)
		case domain.StateWaitingForPrice:
			return r.h.HandlePriceInput(ctx, m)
		case domain.StateWaitingForLink:
			return r.h.HandleLinkInput(ctx, m)
			// It's here for consistency. It's not needed.
			// Users don't send messages in this state, they click buttons.
		case domain.StateWaitingForCartPositionToEdit:
			return nil
		case domain.StateWaitingForCalculatorInput:
			return r.h.HandleCalculatorInput(ctx, m)
		case domain.StateWaitingForFIO:
			return r.h.HandleFIOInput(ctx, m)
		case domain.StateWaitingForPhoneNumber:
			return r.h.HandlePhoneNumberInput(ctx, m)
		case domain.StateWaitingForDeliveryAddress:
			return r.h.HandleDeliveryAddressInput(ctx, m)
		case domain.StateDefault:
			return ErrNoHandler
		default:
			return ErrNoHandler
		}
	}
}

func (r *Router) mapToCallbackHandler(ctx context.Context, c *tg.CallbackQuery) error {
	defer logger.Get().Debug("callback info",
		zap.String("data", c.Data),
		zap.Time("date", c.Message.Time()))
	defer r.h.AnswerCallback(c.ID)
	var (
		chatID             = c.From.ID
		msgID              = c.Message.MessageID
		intCallbackData    int
		callbackDataMsgIDs []int64
	)
	injectedMsgIDs, callback, err := parseCallbackData(c.Data)
	if err != nil {
		return fmt.Errorf("parseCallbackData: %w", err)
	}
	intCallbackData = callback
	callbackDataMsgIDs = injectedMsgIDs

	switch intCallbackData {
	case menuMakeOrderCallback:
		return r.h.StartMakeOrderGuide(ctx, c.Message)
	case menuCalculatorCallback:
		return r.h.Calculator(ctx, chatID)
	case orderGuideStep1Callback:
		return r.h.MakeOrderGuideStep1(ctx, chatID, msgID, callbackDataMsgIDs...)
	case orderGuideStep2Callback:
		return r.h.MakeOrderGuideStep2(ctx, chatID, msgID, callbackDataMsgIDs...)
	case orderGuideStep3Callback:
		return r.h.MakeOrderGuideStep3(ctx, chatID, msgID, callbackDataMsgIDs...)
	case orderGuideStep4Callback:
		return r.h.MakeOrderGuideStep4(ctx, chatID, msgID, callbackDataMsgIDs...)
	case makeOrderCallback:
		return r.h.AskForFIO(ctx, chatID)
	case menuTrackOrderCallback:
		// TODO
		return nil
	case menuCatalogCallback:
		return r.h.Catalog(ctx, chatID)
	case buttonTorqoiseSelectCallback:
		return r.h.HandleButtonSelect(ctx, c, domain.ButtonTorqoise)
	case buttonGreySelectCallback:
		return r.h.HandleButtonSelect(ctx, c, domain.ButtonGrey)
	case button95SelectCallback:
		return r.h.HandleButtonSelect(ctx, c, domain.Button95)
	case editCartCallback:
		return r.h.EditCart(ctx, chatID, msgID)
	case addPositionCallback:
		return r.h.AddPosition(ctx, c.Message)
	case izhLocationCallback:
		return r.h.HandleLocationInput(ctx, chatID, domain.LocationIZH)
	case spbLocationCallback:
		return r.h.HandleLocationInput(ctx, chatID, domain.LocationSPB)
	case othLocationCallback:
		return r.h.HandleLocationInput(ctx, chatID, domain.LocationOther)
	case orderTypeNormalCallback:
		return r.h.HandleOrderTypeInput(ctx, chatID, domain.OrderTypeNormal)
	case orderTypeExpressCallback:
		return r.h.HandleOrderTypeInput(ctx, chatID, domain.OrderTypeExpress)
	default:
		if intCallbackData < editCartRemovePositionOffset {
			return ErrNoHandler
		}
		// remove position callback
		// callbackDataMsgIDs[0] - id of preview cart message
		return r.h.RemoveCartPosition(ctx, chatID, intCallbackData, int64(msgID), callbackDataMsgIDs[0])
	}
}

func (r *Router) command(actual string) func(string) bool {
	return func(want string) bool {
		return actual == want
	}
}
