package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sonyamoonglade/poison-tg/internal/domain"
	"github.com/sonyamoonglade/poison-tg/internal/repositories/dto"
)

func (h *handler) AskForCalculatorOrderType(ctx context.Context, chatID int64) error {
	if err := h.customerRepo.UpdateState(ctx, chatID, domain.StateWaitingForCalculatorOrderType); err != nil {
		return err
	}

	text := "Выбери тип доставки"
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

	var updateDTO = dto.UpdateCustomerDTO{
		CalculatorMeta: &domain.CalculatorMeta{
			NextOrderType: customer.CalculatorMeta.NextOrderType,
		},
	}
	if isExpress {
		// If order type is express then it's no matter which location user would put,
		// so whatever
		customer.UpdateCalculatorMetaLocation(domain.LocationOther)
		updateDTO.CalculatorMeta.Location = customer.CalculatorMeta.Location
	}

	if err := h.customerRepo.Update(ctx, customer.CustomerID, updateDTO); err != nil {
		return err
	}

	var resp = "Тип доставки: "
	switch isExpress {
	case true:
		resp += "Экспресс"
		break
	case false:
		resp += "Обычный"
		break
	}
	if err := h.sendMessage(chatID, resp); err != nil {
		return err
	}

	if isExpress {
		// Skip location part because there's one formula for express orders
		return h.AskForCalculatorCategory(ctx, chatID)
	}

	return h.askForCalculatorLocation(ctx, chatID)
}

func (h *handler) askForCalculatorLocation(ctx context.Context, chatID int64) error {
	if err := h.customerRepo.UpdateState(ctx, chatID, domain.StateWaitingForCalculatorLocation); err != nil {
		return err
	}
	text := "Из какого ты города? 🌄"
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
		CalculatorMeta: &domain.CalculatorMeta{
			Location: customer.Meta.Location,
		},
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

	if err := h.sendMessage(chatID, resp); err != nil {
		return err
	}

	return h.AskForCalculatorCategory(ctx, chatID)
}

func (h *handler) AskForCalculatorCategory(ctx context.Context, chatID int64) error {
	if err := h.customerRepo.UpdateState(ctx, chatID, domain.StateWaitingForCalculatorCategory); err != nil {
		return err
	}
	return h.sendWithKeyboard(chatID, askForCategoryTemplate, categoryCalculatorButtons)
}

func (h *handler) HandleCalculatorCategoryInput(ctx context.Context, chatID int64, cat domain.Category) error {
	var telegramID = chatID

	if err := h.checkRequiredState(ctx, domain.StateWaitingForCalculatorCategory, chatID); err != nil {
		return err
	}

	customer, err := h.customerRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return err
	}

	customer.UpdateCalculatorMetaCategory(cat)
	updateDTO := dto.UpdateCustomerDTO{
		CalculatorMeta: &domain.CalculatorMeta{
			Category: &cat,
		},
	}

	if err := h.customerRepo.Update(ctx, customer.CustomerID, updateDTO); err != nil {
		return fmt.Errorf("customerRepo.Update: %w", err)
	}

	if err := h.sendMessage(chatID, fmt.Sprintf("Выбрана категория: %s", string(cat))); err != nil {
		return err
	}

	return h.askForCalculatorInput(ctx, chatID)
}

func (h *handler) askForCalculatorInput(ctx context.Context, chatID int64) error {
	var telegramID = chatID

	if err := h.customerRepo.UpdateState(ctx, telegramID, domain.StateWaitingForCalculatorInput); err != nil {
		return fmt.Errorf("customerRepo.Update: %w", err)
	}

	return h.sendMessage(chatID, askForCalculatorInputTemplate)
}

func (h *handler) HandleCalculatorInput(ctx context.Context, m *tg.Message) error {
	var (
		chatID = m.Chat.ID
		input  = strings.TrimSpace(m.Text)
	)
	if err := h.checkRequiredState(ctx, domain.StateWaitingForCalculatorInput, chatID); err != nil {
		return err
	}

	priceYuan, err := strconv.ParseUint(input, 10, 64)
	if err != nil {
		if err := h.sendMessage(chatID, "Неправильный формат ввода"); err != nil {
			return err
		}
		return fmt.Errorf("strconv.ParseUint: %w", err)
	}

	customer, err := h.customerRepo.GetByTelegramID(ctx, chatID)
	if err != nil {
		return err
	}

	var (
		ordTyp = customer.Meta.NextOrderType
		loc    = customer.Meta.Location
	)
	if ordTyp == nil || loc == nil {
		return fmt.Errorf("order type or location in meta is nil")
	}

	// We should apply customer.Meta and customer.CalculatorMeta.Category in order to calculate correctly
	args := domain.ConvertYuanArgs{
		X:         priceYuan,
		Rate:      h.rateProvider.GetYuanRate(),
		OrderType: *ordTyp,
		Location:  *loc,
		Category:  *customer.CalculatorMeta.Category,
	}

	priceRub := domain.ConvertYuan(args)

	if err != nil {
		return err
	}

	if err := h.customerRepo.UpdateState(ctx, chatID, domain.StateDefault); err != nil {
		return err
	}

	return h.sendWithKeyboard(chatID, getCalculatorOutput(priceRub), calculateMoreButtons)
}
