package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sonyamoonglade/poison-tg/internal/domain"
	"github.com/sonyamoonglade/poison-tg/internal/repositories/dto"
	"github.com/sonyamoonglade/poison-tg/internal/services"
	"github.com/sonyamoonglade/poison-tg/pkg/utils/url"
)

func (h *handler) HandleSizeInput(ctx context.Context, m *tg.Message) error {
	var (
		chatID     = m.Chat.ID
		telegramID = chatID
		sizeText   = strings.TrimSpace(m.Text)
	)
	// validate state
	if err := h.checkRequiredState(ctx, domain.StateWaitingForSize, chatID); err != nil {
		return err
	}
	customer, err := h.customerRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return fmt.Errorf("customerRepo.GetByTelegramID: %w", err)
	}

	position := domain.NewEmptyPosition()
	customer.SetLastEditPosition(position)
	customer.UpdateLastEditPositionSize(sizeText)

	updateDTO := dto.UpdateCustomerDTO{
		LastPosition: customer.LastEditPosition,
		State:        &domain.StateWaitingForButton,
	}

	if err := h.customerRepo.Update(ctx, customer.CustomerID, updateDTO); err != nil {
		return fmt.Errorf("customerRepo.Update: %w", err)
	}
	if sizeText == "#" {
		sizeText = "БЕЗ размера"
	}
	if err := h.sendMessage(chatID, fmt.Sprintf("Твой размер: %s", sizeText)); err != nil {
		return err
	}
	return h.sendWithKeyboard(chatID, askForButtonColorTemplate, selectColorButtons)
}

func (h *handler) HandleButtonSelect(ctx context.Context, c *tg.CallbackQuery, button domain.Button) error {
	var (
		chatID     = c.From.ID
		telegramID = chatID
	)

	if err := h.checkRequiredState(ctx, domain.StateWaitingForButton, chatID); err != nil {
		return err
	}

	customer, err := h.customerRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return fmt.Errorf("customerRepo.GetByTelegramID: %w", err)
	}

	customer.UpdateLastEditPositionButtonColor(button)
	updateDTO := dto.UpdateCustomerDTO{
		LastPosition: customer.LastEditPosition,
		State:        &domain.StateWaitingForPrice,
	}

	if err := h.customerRepo.Update(ctx, customer.CustomerID, updateDTO); err != nil {
		return fmt.Errorf("customerRepo.Update: %w", err)
	}
	if err := h.sendMessage(chatID, fmt.Sprintf("Цвет выбранной кнопки: %s", string(button))); err != nil {
		return err
	}

	return h.sendMessage(chatID, askForPriceTemplate)
}

func (h *handler) HandlePriceInput(ctx context.Context, m *tg.Message) error {
	var (
		chatID     = m.Chat.ID
		telegramID = chatID
		input      = strings.TrimSpace(m.Text)
	)
	// validate state
	if err := h.checkRequiredState(ctx, domain.StateWaitingForPrice, chatID); err != nil {
		return err
	}

	customer, err := h.customerRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return fmt.Errorf("customerRepo.GetByTelegramID: %w", err)
	}

	priceYuan, err := strconv.ParseUint(input, 10, 64)
	if err != nil {
		return ErrInvalidPriceInput
	}

	priceRub, err := h.yuanService.ApplyFormula(priceYuan, services.UseFormulaArguments{
		Location:  *customer.Meta.Location,
		IsExpress: *customer.Meta.NextOrderType == domain.OrderTypeExpress,
	})
	if err != nil {
		return err
	}

	customer.UpdateLastEditPositionPrice(priceRub, priceYuan)

	updateDTO := dto.UpdateCustomerDTO{
		LastPosition: customer.LastEditPosition,
		State:        &domain.StateWaitingForLink,
	}

	if err := h.customerRepo.Update(ctx, customer.CustomerID, updateDTO); err != nil {
		return fmt.Errorf("customerRepo.Update: %w", err)
	}

	if err := h.sendMessage(chatID, fmt.Sprintf("Стоимость товара: %d ₽", priceRub)); err != nil {
		return err
	}
	return h.sendMessage(chatID, askForLinkTemplate)
}

func (h *handler) HandleLinkInput(ctx context.Context, m *tg.Message) error {
	var (
		chatID     = m.From.ID
		telegramID = chatID
		link       = strings.TrimSpace(m.Text)
	)

	// validate state
	if err := h.checkRequiredState(ctx, domain.StateWaitingForLink, chatID); err != nil {
		return err
	}

	customer, err := h.customerRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return fmt.Errorf("customerRepo.GetByTelegramID: %w", err)
	}

	if ok := url.IsValidDW4URL(link); !ok {
		if err := h.sendMessage(chatID, "Неправильная ссылка! Смотри инструкцию"); err != nil {
			return err
		}
		return h.sendMessage(chatID, "Введи повторно корректную ссылку 😀")
	}

	customer.UpdateLastEditPositionLink(link)
	customer.Cart.Add(*customer.LastEditPosition)
	updateDTO := dto.UpdateCustomerDTO{
		LastPosition: customer.LastEditPosition,
		Cart:         &customer.Cart,
		State:        &domain.StateDefault,
	}

	if err := h.customerRepo.Update(ctx, customer.CustomerID, updateDTO); err != nil {
		return fmt.Errorf("customerRepo.Update: %w", err)
	}

	if err := h.sendMessage(chatID, fmt.Sprintf("Товар по ссылке: %s", link)); err != nil {
		return err
	}

	positionAddedMsg := tg.NewMessage(chatID, "Позиция успешно добавлена!")
	positionAddedMsg.ReplyMarkup = bottomMenuButtons
	return h.cleanSend(positionAddedMsg)
}

func (h *handler) AddPosition(ctx context.Context, m *tg.Message) error {
	var (
		chatID     = m.Chat.ID
		telegramID = chatID
	)
	if err := h.sendMessage(chatID, newPositionWarnTemplate); err != nil {
		return err
	}
	customer, err := h.customerRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return err
	}
	if len(customer.Cart) > 0 {
		// Means that we can use customer.Meta and it's valid
		return h.addPosition(ctx, m.Chat.ID)
	}
	if err := h.customerRepo.UpdateState(ctx, telegramID, domain.StateWaitingForOrderType); err != nil {
		return err
	}
	// Otherwise start from order type selection
	return h.askForOrderType(ctx, chatID)
}

func (h *handler) addPosition(ctx context.Context, chatID int64) error {
	if err := h.sendWithKeyboard(chatID, askForSizeTemplate, bottomMenuWithoutAddPositionButtons); err != nil {
		return err
	}

	return h.customerRepo.UpdateState(ctx, chatID, domain.StateWaitingForSize)
}
