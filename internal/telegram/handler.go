package telegram

import (
	"context"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ButtonProvider interface {
	Menu() tg.InlineKeyboardMarkup
	StartMakeOrder() tg.InlineKeyboardMarkup
}

type handler struct {
	b               *Bot
	templateManager *TemplateManager
	buttonProvider  ButtonProvider
}

func NewHandler(bot *Bot, templateManager *TemplateManager, buttonProvider ButtonProvider) RouteHandler {
	return &handler{
		b:               bot,
		templateManager: templateManager,
		buttonProvider:  buttonProvider,
	}
}

func (h *handler) Menu(ctx context.Context, chatID int64) error {
	return h.sendWithKeyboard(chatID, h.templateManager.Menu(), h.buttonProvider.Menu())
}

func (h *handler) Catalog(ctx context.Context, chatID int64) error {
	return h.cleanSend(tg.NewMessage(chatID, "catalog"))
}

func (h *handler) GetBucket(ctx context.Context, chatID int64) error {
	return h.cleanSend(tg.NewMessage(chatID, "get bucket"))
}

func (h *handler) MakeOrder(ctx context.Context, chatID int64) error {
	return h.sendWithKeyboard(chatID, "missing template manager", h.buttonProvider.StartMakeOrder())
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
