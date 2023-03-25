package telegram

import (
	"context"
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sonyamoonglade/poison-tg/internal/domain"
	"github.com/sonyamoonglade/poison-tg/internal/repositories/dto"
	"github.com/sonyamoonglade/poison-tg/internal/services"
	"strconv"
	"strings"
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

	var resp = "Тип доставки: "
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
		return h.askForCalculatorInput(ctx, chatID)
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

	return h.askForCalculatorInput(ctx, chatID)
}

func (h *handler) askForCalculatorInput(ctx context.Context, chatID int64) error {
	var telegramID = chatID

	customer, err := h.customerRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return fmt.Errorf("customerRepo.GetByTelegramID: %w", err)
	}

	updateDTO := dto.UpdateCustomerDTO{
		State: &domain.StateWaitingForCalculatorInput,
	}

	if err := h.customerRepo.Update(ctx, customer.CustomerID, updateDTO); err != nil {
		return fmt.Errorf("customerRepo.Update: %w", err)
	}

	return h.cleanSend(tg.NewMessage(chatID, "Отправь мне цену товара в юанях, а я скажу сколько это будет стоить в переводе на рубли для тебя"))
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
		if err := h.cleanSend(tg.NewMessage(chatID, "Неправильный формат ввода")); err != nil {
			return err
		}
		return fmt.Errorf("strconv.ParseUint: %w", err)
	}

	customer, err := h.customerRepo.GetByTelegramID(ctx, chatID)
	if err != nil {
		return err
	}

	if customer.CalculatorMeta.NextOrderType == nil || customer.CalculatorMeta.Location == nil {
		return fmt.Errorf("calculator meta values are nil")
	}
	isExpress := *customer.CalculatorMeta.NextOrderType == domain.OrderTypeExpress

	priceRub, err := h.yuanService.ApplyFormula(inputUint, services.UseFormulaArguments{
		Location:  *customer.CalculatorMeta.Location,
		IsExpress: isExpress,
	})
	if err != nil {
		return err
	}

	if err := h.customerRepo.UpdateState(ctx, chatID, domain.StateDefault); err != nil {
		return err
	}

	return h.sendWithKeyboard(chatID, getCalculatorOutput(priceRub), calculateMoreButtons)
}
