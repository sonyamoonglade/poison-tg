package telegram

import (
	"context"
	"errors"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sonyamoonglade/poison-tg/internal/repositories"
	"github.com/sonyamoonglade/poison-tg/internal/telegram/catalog"
)

var (
	ErrInvalidState      = errors.New("invalid state")
	ErrInvalidPriceInput = errors.New("invalid price input")
)

type RateProvider interface {
	GetYuanRate() float64
}

type Bot interface {
	Send(c tg.Chattable) (tg.Message, error)
	CleanRequest(c tg.Chattable) error
	SendMediaGroup(c tg.MediaGroupConfig) ([]tg.Message, error)
}

type handler struct {
	b               Bot
	customerRepo    repositories.Customer
	orderRepo       repositories.Order
	rateProvider    RateProvider
	catalogProvider *catalog.CatalogProvider
}

func NewHandler(bot Bot,
	repositories repositories.Repositories,
	rateProvider RateProvider,
	catalogProvider *catalog.CatalogProvider) *handler {
	return &handler{
		b:               bot,
		customerRepo:    repositories.Customer,
		orderRepo:       repositories.Order,
		catalogProvider: catalogProvider,
		rateProvider:    rateProvider,
	}
}

func (h *handler) AnswerCallback(callbackID string) error {
	return h.cleanSend(tg.NewCallback(callbackID, ""))
}

func (h *handler) HandleError(ctx context.Context, err error, m tg.Update) {
	if errors.Is(err, ErrInvalidPriceInput) {
		h.sendMessage(m.FromChat().ID, "Неправильный формат ввода")
		return
	}
	h.b.Send(tg.NewMessage(m.FromChat().ID, "Извини, я не понимаю тебя :("))
}
