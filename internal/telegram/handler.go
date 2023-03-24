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
	"github.com/sonyamoonglade/poison-tg/internal/repositories/dto"
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

func (h *handler) AskForCalculatorOrderType(ctx context.Context, chatID int64) error {
	if err := h.customerRepo.UpdateState(ctx, chatID, domain.StateWaitingForCalculatorOrderType); err != nil {
		return err
	}

	text := "Выбери тип доставки (время экспресс перевозки в среднем составляет 4 дня из Китая в СПб, обычная перевозка составляет 8-15 дней)"
	return h.sendWithKeyboard(chatID, text, orderTypeCalculatorButtons)
}

func (h *handler) HandleCalculatorOrderTypeInput(ctx context.Context, chatID int64, typ domain.OrderType) error {
	var (
		telegramID = chatID
		isExpress  = typ == domain.OrderTypeExpress
	)

	if err := h.checkRequiredState(ctx, domain.StateWaitingForCalculatorOrderType, chatID); err != nil {
		return err
	}

	customer, err := h.customerRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return err
	}

	customer.UpdateCalculatorMetaOrderType(typ)
	if isExpress {
		// If order type is express then it's no matter which location user would put,
		// so whatever
		customer.UpdateCalculatorMetaLocation(domain.LocationOther)
		// skip location state
		customer.TgState = domain.StateWaitingForCalculatorInput
	} else {
		customer.TgState = domain.StateWaitingForCalculatorLocation
	}

	updateDTO := dto.UpdateCustomerDTO{
		State: &customer.TgState,
		CalculatorMeta: &domain.Meta{
			NextOrderType: customer.CalculatorMeta.NextOrderType,
			Location:      customer.CalculatorMeta.Location,
		},
	}

	if err := h.customerRepo.Update(ctx, customer.CustomerID, updateDTO); err != nil {
		return err
	}

	var resp = "Тип заказа: "
	switch isExpress {
	case true:
		resp += "Экспресс"
		break
	case false:
		resp += "Обычный"
		break
	}
	if err := h.cleanSend(tg.NewMessage(chatID, resp)); err != nil {
		return err
	}

	if isExpress {
		// Skip location part because there's one formula for express orders
		return h.Calculator(ctx, chatID)
	}

	return h.askForCalculatorLocation(ctx, chatID)

}

func (h *handler) askForCalculatorLocation(ctx context.Context, chatID int64) error {
	text := "Из какого ты города?"
	return h.sendWithKeyboard(chatID, text, locationCalculatorButtons)
}

func (h *handler) HandleCalculatorLocationInput(ctx context.Context, chatID int64, loc domain.Location) error {
	var telegramID = chatID

	if err := h.checkRequiredState(ctx, domain.StateWaitingForCalculatorLocation, chatID); err != nil {
		return err
	}

	customer, err := h.customerRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return err
	}

	customer.UpdateCalculatorMetaLocation(loc)

	updateDTO := dto.UpdateCustomerDTO{
		CalculatorMeta: &customer.CalculatorMeta,
	}

	if err := h.customerRepo.Update(ctx, customer.CustomerID, updateDTO); err != nil {
		return err
	}
	var resp = "Выбран: "
	switch loc {
	case domain.LocationSPB:
		resp += "Питер"
		break
	case domain.LocationIZH:
		resp += "Ижевск"
		break
	case domain.LocationOther:
		resp += "Другой"
		break
	}

	if err := h.cleanSend(tg.NewMessage(chatID, resp)); err != nil {
		return err
	}

	return h.Calculator(ctx, chatID)
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

	// Get meta
	customer, err := h.customerRepo.GetByTelegramID(ctx, chatID)
	if err != nil {
		return err
	}

	isExpress := *customer.CalculatorMeta.NextOrderType == domain.OrderTypeExpress

	priceRub, err := h.yuanService.ApplyFormula(inputUint, services.UseFormulaArguments{
		Location:  *customer.CalculatorMeta.Location,
		IsExpress: isExpress,
	})
	if err != nil {
		return err
	}

	return h.cleanSend(tg.NewMessage(chatID, getCalculatorOutput(priceRub)))
}

func (h *handler) AnswerCallback(callbackID string) error {
	return h.cleanSend(tg.NewCallback(callbackID, ""))
}

func (h *handler) HandleError(ctx context.Context, err error, m tg.Update) {
	h.b.Send(tg.NewMessage(m.FromChat().ID, "Whoops!"))
}
