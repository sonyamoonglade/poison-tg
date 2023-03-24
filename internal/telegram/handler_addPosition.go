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
		//todo: translate
		sizeText = "NO size"
	}
	if err := h.cleanSend(tg.NewMessage(chatID, fmt.Sprintf("Твой размер: %s", sizeText))); err != nil {
		return err
	}
	text := "Выбери цвет кнопки (влияет на условия доставки и цену в дальнейшем)"
	return h.sendWithKeyboard(chatID, text, selectColorButtons)
}

func (h *handler) HandleButtonSelect(ctx context.Context, c *tg.CallbackQuery, button domain.Button) error {
	var (
		chatID     = c.From.ID
		telegramID = chatID
	)
	// validate state
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
	if err := h.cleanSend(tg.NewMessage(chatID, fmt.Sprintf("Цвет выбранной кнопки: %s", string(button)))); err != nil {
		return err
	}
	text := "Отправь стоимость товара в юанях (указана на выбранной кнопке)."
	return h.cleanSend(tg.NewMessage(chatID, text))
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

	if err := h.cleanSend(tg.NewMessage(chatID, fmt.Sprintf("Стоимость товара: %d ₽", priceRub))); err != nil {
		return err
	}
	text := "Отправь ссылку на выбранный товар (строго по инструкции)"
	return h.cleanSend(tg.NewMessage(chatID, text))
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
		if err := h.cleanSend(tg.NewMessage(chatID, "Неправильная ссылка! Смотрите шаг 4 в инструкции")); err != nil {
			return err
		}
		return h.cleanSend(tg.NewMessage(chatID, "Введите повторно корректную ссылку 😀"))
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

	if err := h.cleanSend(tg.NewMessage(chatID, fmt.Sprintf("Товар по ссылке: %s", link))); err != nil {
		return err
	}

	positionAddedMsg := tg.NewMessage(chatID, "Позиция успешно добавлена!")
	positionAddedMsg.ReplyMarkup = bottomMenuButtons
	return h.cleanSend(positionAddedMsg)
}

func (h *handler) AddPosition(ctx context.Context, m *tg.Message) error {
	return h.addPosition(ctx, m.Chat.ID)
}

func (h *handler) addPosition(ctx context.Context, chatID int64) error {
	text := "Шаг 1. Выбери размер. Если товар безразмерный, то отправь #"
	if err := h.sendWithKeyboard(chatID, text, bottomMenuWithoutAddPositionButtons); err != nil {
		return err
	}

	return h.customerRepo.UpdateState(ctx, chatID, domain.StateWaitingForSize)
}
