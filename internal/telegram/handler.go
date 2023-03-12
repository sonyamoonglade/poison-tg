package telegram

import (
	"context"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type handler struct {
	b *Bot
}

func NewHandler(bot *Bot) RouteHandler {
	return &handler{
		b: bot,
	}
}

func (h *handler) Menu(ctx context.Context, m *tg.Message) error {
	_, err := h.b.Send(tg.NewMessage(m.Chat.ID, "hello mate!"))
	return err
}

func (h *handler) HandleError(ctx context.Context, err error, m tg.Update) {
	h.b.Send(tg.NewMessage(m.FromChat().ID, "Whoops!"))
}
