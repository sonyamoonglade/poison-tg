package telegram

import (
	"context"
	"errors"
	"fmt"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sonyamoonglade/poison-tg/internal/domain"
	"github.com/sonyamoonglade/poison-tg/internal/services"
	"github.com/sonyamoonglade/poison-tg/pkg/logger"
)

const (
	groupSendErrMsg = "json: cannot unmarshal array into Go value of type tgbotapi.Message"
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

func (h *handler) MakeOrder(ctx context.Context, m *tg.Message) error {
	var (
		telegramID int64  = m.Chat.ID
		firstName  string = m.Chat.FirstName
		lastName   string = m.Chat.LastName
		username   string = domain.MakeUsername(firstName, lastName, m.Chat.UserName)
	)

	_, err := h.customerService.GetByTelegramID(ctx, telegramID)
	// if no such customer yet create it
	if err != nil && errors.Is(err, domain.ErrCustomerNotFound) {
		if err := h.customerService.Save(ctx, domain.NewCustomer(telegramID, username)); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	// if customer exists then update it's state to waiting for screenshot
	if err := h.customerService.UpdateState(ctx, telegramID, domain.StateWaitingForLink); err != nil {
		return err
	}
	return h.cleanSend(tg.NewMessage(m.Chat.ID, "Pojalusta, otpravte image of your products"))
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

func (h *handler) StartMakeOrderGuide(ctx context.Context, chatID int64) error {
	logger.Get().Sugar().Debugf("start make order guide")
	url := "https://picsum.photos/200/300"
	imageWithCaption := tg.NewPhoto(chatID, tg.FileURL(url))
	imageWithCaption.Caption = "Vadim's message blabla\nblabla\nbla\nListay dalee dlya instrukcii!"
	if err := h.cleanSend(imageWithCaption); err != nil {
		return err
	}
	return h.sendWithKeyboard(chatID, "Knopku dlya upravlienia", orderGuideStep1Buttons)
}

func (h *handler) MakeOrderGuideStep2(ctx context.Context, m *tg.Message) error {
	logger.Get().Sugar().Debugf("guide step 2")
	url := "https://picsum.photos/200/200"
	imageWithCaption := tg.NewInputMediaPhoto(tg.FileURL(url))
	imageWithCaption.Caption = "This is step 2 of instruction"
	editMsg := &tg.EditMessageMediaConfig{
		BaseEdit: tg.BaseEdit{
			ChatID:      m.Chat.ID,
			MessageID:   m.MessageID,
			ReplyMarkup: &orderGuideStep2Buttons,
		},
		Media: imageWithCaption,
	}
	if err := h.cleanSend(editMsg); err != nil {
		return fmt.Errorf("step2 err: %w", err)
	}
	return nil
}

func (h *handler) MakeOrderGuideStep3(ctx context.Context, m *tg.Message) error {
	logger.Get().Sugar().Debugf("guide step 3")
	url := "https://picsum.photos/200/300"
	return h.sendImageWithTextAndKeyboard(m.Chat.ID, url, "step 3", orderGuideStep3Buttons)
}
func (h *handler) MakeOrderGuideStep4(ctx context.Context, m *tg.Message) error {
	logger.Get().Sugar().Debugf("guide step 4")
	url := "https://picsum.photos/200/300"
	return h.sendImageWithTextAndKeyboard(m.Chat.ID, url, "step 4", orderGuideStep4Buttons)
}

func (h *handler) AnswerCallback(callbackID string) error {
	return h.cleanSend(tg.NewCallback(callbackID, ""))
}

func (h *handler) HandleError(ctx context.Context, err error, m tg.Update) {
	h.b.Send(tg.NewMessage(m.FromChat().ID, "Whoops!"))
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

func (h *handler) cleanEdit(chatID int64, prevMsgID int, newText string, newKeyboard tg.InlineKeyboardMarkup) error {
	_, err := h.b.Edit(chatID, prevMsgID, newText, &newKeyboard)
	return err
}

func (h *handler) sendImageWithText(chatID int64, imageURL string, text string) error {
	resp := tg.NewPhoto(chatID, tg.FileURL(imageURL))
	resp.Caption = text
	return h.cleanSend(resp)
}

func (h *handler) sendImageWithTextAndKeyboard(chatID int64, imageURL string, text string, keyboard tg.InlineKeyboardMarkup) error {
	resp := tg.NewPhoto(chatID, tg.FileURL(imageURL))
	resp.Caption = text
	resp.ReplyMarkup = keyboard
	return h.cleanSend(resp)
}
