package telegram

import (
	"context"
	"errors"
	"fmt"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sonyamoonglade/poison-tg/internal/domain"
	"github.com/sonyamoonglade/poison-tg/internal/services"
	"github.com/sonyamoonglade/poison-tg/pkg/functools"
)

var (
	ErrInvalidState = errors.New("invalid state")
)

type handler struct {
	b               *Bot
	customerService services.Customer
}

func NewHandler(bot *Bot, customerService services.Customer) RouteHandler {
	return &handler{
		b:               bot,
		customerService: customerService,
	}
}

func (h *handler) Start(ctx context.Context, chatID int64) error {
	return h.cleanSend(tg.NewMessage(chatID, GetTemplate().Start))
}

func (h *handler) MakeOrder(ctx context.Context, m *tg.Message) error {
	panic("implement me")
}

func (h *handler) Menu(ctx context.Context, chatID int64) error {
	return h.sendWithKeyboard(chatID, GetTemplate().Menu, menuButtons)
}

func (h *handler) Catalog(ctx context.Context, chatID int64) error {
	return h.cleanSend(tg.NewMessage(chatID, "catalog"))
}

func (h *handler) GetBucket(ctx context.Context, chatID int64) error {
	return h.cleanSend(tg.NewMessage(chatID, "get bucket"))
}

func (h *handler) StartMakeOrderGuide(ctx context.Context, m *tg.Message) error {
	var (
		chatID     = m.Chat.ID
		telegramID = chatID
		firstName  = m.Chat.FirstName
		lastName   = m.Chat.LastName
		username   = domain.MakeUsername(firstName, lastName, m.Chat.UserName)
	)
	url := "https://picsum.photos/300/300"
	image := tg.NewInputMediaPhoto(tg.FileURL(url))
	image.Caption = "У меня есть желание привезти лишь то,что нужно, поэтому, (имя юзера), предупреждаю, китайцы уже позаботились о нас и предоставили к каждому размерному товару - размерную сетку разных стран, тебе лишь нужно выбрать подходящий размер. Не ошибись с выбором, Стрелок  🤠 Поехали?"
	group := tg.NewMediaGroup(chatID, []interface{}{
		image,
		tg.NewInputMediaPhoto(tg.FileURL(url)),
	})
	sentMsgs, err := h.b.client.SendMediaGroup(group)
	if err != nil {
		return err
	}

	msgIDs := functools.Map(func(m tg.Message) int64 {
		return int64(m.MessageID)
	}, sentMsgs)

	buttons := prepareOrderGuideButtons(orderGuideStep1Callback, msgIDs...)
	if err := h.sendWithKeyboard(chatID, "Кнопки для пролистывания инструкции", buttons); err != nil {
		return err
	}

	return h.sendWaitForSizeCommandAndUpdateCustomerState(ctx, telegramID, username)
}

func (h *handler) MakeOrderGuideStep1(ctx context.Context, chatID int64, controlButtonsMessageID int, instructionMsgIDs ...int64) error {
	// update content
	for i, imID := range instructionMsgIDs {
		url := "https://picsum.photos/300/301"
		image := tg.NewInputMediaPhoto(tg.FileURL(url))
		// update Caption only on first element on order to show text (see telegram docs)
		if i == 0 {
			image.Caption = "У меня есть желание привезти лишь то,что нужно, поэтому, (имя юзера), предупреждаю, китайцы уже позаботились о нас и предоставили к каждому размерному товару - размерную сетку разных стран, тебе лишь нужно выбрать подходящий размер. Не ошибись с выбором, Стрелок  🤠 Поехали?"
		}
		editOneMedia := &tg.EditMessageMediaConfig{
			BaseEdit: tg.BaseEdit{
				ChatID:    chatID,
				MessageID: int(imID),
			},
			Media: image,
		}
		if err := h.cleanSend(editOneMedia); err != nil {
			return err
		}
	}
	// update control buttons
	buttons := tg.NewEditMessageReplyMarkup(chatID, controlButtonsMessageID, prepareOrderGuideButtons(orderGuideStep1Callback, instructionMsgIDs...))
	return h.cleanSend(buttons)
}

func (h *handler) MakeOrderGuideStep2(ctx context.Context, chatID int64, controlButtonsMessageID int, instructionMsgIDs ...int64) error {
	// update content
	for i, imID := range instructionMsgIDs {
		url := "https://picsum.photos/300/302"
		image := tg.NewInputMediaPhoto(tg.FileURL(url))
		// update Caption only on first element on order to show text (see telegram docs)
		if i == 0 {
			image.Caption = "This is step 2 of instruction"
		}
		editOneMedia := &tg.EditMessageMediaConfig{
			BaseEdit: tg.BaseEdit{
				ChatID:    chatID,
				MessageID: int(imID),
			},
			Media: image,
		}
		if err := h.cleanSend(editOneMedia); err != nil {
			return err
		}
	}
	// update control buttons
	buttons := tg.NewEditMessageReplyMarkup(chatID, controlButtonsMessageID, prepareOrderGuideButtons(orderGuideStep2Callback, instructionMsgIDs...))
	return h.cleanSend(buttons)
}

func (h *handler) MakeOrderGuideStep3(ctx context.Context, chatID int64, controlButtonsMessageID int, instructionMsgIDs ...int64) error {
	// update content
	for i, imID := range instructionMsgIDs {
		url := "https://picsum.photos/300/303"
		image := tg.NewInputMediaPhoto(tg.FileURL(url))
		// update Caption only on first element on order to show text (see telegram docs)
		if i == 0 {
			image.Caption = "This is step 3 of instruction"
		}
		editOneMedia := &tg.EditMessageMediaConfig{
			BaseEdit: tg.BaseEdit{
				ChatID:    chatID,
				MessageID: int(imID),
			},
			Media: image,
		}
		if err := h.cleanSend(editOneMedia); err != nil {
			return err
		}
	}
	// update control buttons
	buttons := tg.NewEditMessageReplyMarkup(chatID, controlButtonsMessageID, prepareOrderGuideButtons(orderGuideStep3Callback, instructionMsgIDs...))
	return h.cleanSend(buttons)
}

func (h *handler) MakeOrderGuideStep4(ctx context.Context, chatID int64, controlButtonsMessageID int, instructionMsgIDs ...int64) error {
	// update content
	for i, imID := range instructionMsgIDs {
		url := "https://picsum.photos/300/304"
		image := tg.NewInputMediaPhoto(tg.FileURL(url))
		// update Caption only on first element on order to show text (see telegram docs)
		if i == 0 {
			image.Caption = "This is step 4 of instruction"
		}
		editOneMedia := &tg.EditMessageMediaConfig{
			BaseEdit: tg.BaseEdit{
				ChatID:    chatID,
				MessageID: int(imID),
			},
			Media: image,
		}
		if err := h.cleanSend(editOneMedia); err != nil {
			return err
		}
	}
	// update control buttons
	buttons := tg.NewEditMessageReplyMarkup(chatID, controlButtonsMessageID, prepareOrderGuideButtons(orderGuideStep4Callback, instructionMsgIDs...))
	return h.cleanSend(buttons)
}

func (h *handler) HandleSizeInput(ctx context.Context, m *tg.Message) error {
	var (
		chatID     = m.Chat.ID
		telegramID = chatID
	)
	// validate state
	if err := h.checkRequiredState(ctx, domain.StateWaitingForSize, chatID); err != nil {
		return err
	}
	customer, err := h.customerService.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return fmt.Errorf("customerService.GetByTelegramID: %w", err)
	}
	position := domain.NewEmptyPosition()
	sizeText := m.Text
	if err := customer.SetLastEditPosition(position); err != nil {
		// handle if last pos already exists
		return err
	}
	customer.UpdateLastEditPositionSize(sizeText)
	fmt.Println("customer ent: ", customer.LastEditPosition)
	defer func() {
		customer, _ := h.customerService.GetByTelegramID(ctx, telegramID)
		fmt.Println(customer.LastEditPosition)
	}()

	if err := h.customerService.Save(ctx, customer); err != nil {
		return fmt.Errorf("customerService.Save: %w", err)
	}

	if err := h.cleanSend(tg.NewMessage(chatID, "Thanks for size! Your size: "+sizeText)); err != nil {
		return err
	}

	return h.cleanSend(tg.NewMessage(chatID, "select the button"))
}

func (h *handler) HandlePriceInput(ctx context.Context, m *tg.Message) error {
	//TODO implement me
	panic("implement me")
}

func (h *handler) AnswerCallback(callbackID string) error {
	return h.cleanSend(tg.NewCallback(callbackID, ""))
}

func (h *handler) HandleError(ctx context.Context, err error, m tg.Update) {
	h.b.Send(tg.NewMessage(m.FromChat().ID, "Whoops!"))
}

func (h *handler) checkRequiredState(ctx context.Context, want domain.State, telegramID int64) error {
	customer, err := h.customerService.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return err
	}
	if customer.TgState != want {
		return ErrInvalidState
	}
	return nil
}

func (h *handler) sendWaitForSizeCommandAndUpdateCustomerState(ctx context.Context, telegramID int64, username string) error {
	chatID := telegramID
	_, err := h.customerService.GetByTelegramID(ctx, telegramID)
	// if no such customer yet then create it
	if err != nil && errors.Is(err, domain.ErrCustomerNotFound) {
		// save to db
		defer h.customerService.PrintDb()
		if err := h.customerService.Save(ctx, domain.NewCustomer(telegramID, username)); err != nil {
			return err
		}

	} else if err != nil {
		// report internal error
		return err
	}

	initialMakeOrderCommand := tg.NewMessage(chatID, "Отправьте размер выбранного вами товара (шаг 1) 👍")
	if err := h.cleanSend(initialMakeOrderCommand); err != nil {
		return err
	}
	return h.customerService.UpdateState(ctx, chatID, domain.StateWaitingForSize)
}

func (h *handler) sendWithKeyboard(chatID int64, text string, keyboard interface{}) error {
	m := tg.NewMessage(chatID, text)
	m.ReplyMarkup = keyboard
	return h.cleanSend(m)
}

func (h *handler) cleanSend(c tg.Chattable) error {
	_, err := h.b.Send(c)
	return err
}
