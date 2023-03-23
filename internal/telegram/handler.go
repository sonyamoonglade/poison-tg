package telegram

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sonyamoonglade/poison-tg/internal/domain"
	"github.com/sonyamoonglade/poison-tg/internal/repositories"
	"github.com/sonyamoonglade/poison-tg/internal/services"
	"github.com/sonyamoonglade/poison-tg/internal/telegram/catalog"
)

var (
	ErrInvalidState      = errors.New("invalid state")
	ErrInvalidPriceInput = errors.New("invalid price input")
)

type handler struct {
	b               *Bot
	customerRepo    repositories.Customer
	businessRepo    repositories.Business
	orderRepo       repositories.Order
	yuanService     services.Yuan
	catalogProvider *catalog.CatalogProvider
}

func NewHandler(bot *Bot,
	customerRepo repositories.Customer,
	businessRepo repositories.Business,
	orderRepo repositories.Order,
	yuanService services.Yuan,
	catalogProvider *catalog.CatalogProvider) *handler {
	return &handler{
		b:               bot,
		customerRepo:    customerRepo,
		businessRepo:    businessRepo,
		orderRepo:       orderRepo,
		yuanService:     yuanService,
		catalogProvider: catalogProvider,
	}
}

func (h *handler) HandleCalculatorInput(ctx context.Context, m *tg.Message) error {
	var (
		chatID = m.Chat.ID
		input  = strings.TrimSpace(m.Text)
	)
	if err := h.checkRequiredState(ctx, domain.StateWaitingForCalculatorInput, chatID); err != nil {
		return err
	}

	inputUint, err := strconv.ParseUint(input, 10, 64)
	if err != nil {
		return fmt.Errorf("strconv.ParseUint: %w", err)
	}

	// todo: fixme
	priceRub, err := h.yuanService.ApplyFormula(inputUint, services.UseFormulaArguments{
		Location:  0,
		IsExpress: false,
	})
	if err != nil {
		return err
	}

	return h.cleanSend(tg.NewMessage(chatID, getCalculatorOutput(priceRub, priceRub+100)))
}

func (h *handler) AnswerCallback(callbackID string) error {
	return h.cleanSend(tg.NewCallback(callbackID, ""))
}

func (h *handler) HandleError(ctx context.Context, err error, m tg.Update) {
	h.b.Send(tg.NewMessage(m.FromChat().ID, "Whoops!"))
}
