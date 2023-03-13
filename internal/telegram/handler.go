package telegram

import (
	"context"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ButtonProvider interface {
	Menu()
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

func (h *handler) Menu(ctx context.Context, m *tg.Message) error {
	var (
		msg     = h.templateManager.Menu()
		buttons = h.buttonProvider.Menu()
		chatID  = m.Chat.ID
	)

	h.b.Send()
}

func (h *handler) HandleError(ctx context.Context, err error, m tg.Update) {
	h.b.Send(tg.NewMessage(m.FromChat().ID, "Whoops!"))
}
