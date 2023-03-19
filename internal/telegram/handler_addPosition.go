package telegram

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sonyamoonglade/poison-tg/internal/domain"
	"github.com/sonyamoonglade/poison-tg/internal/repositories/dto"
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

	if err := h.cleanSend(tg.NewMessage(chatID, "Thanks for size! Your size: "+sizeText)); err != nil {
		return err
	}

	return h.sendWithKeyboard(chatID, "Выберите цвет кнопки (шаг 2) 👍", selectColorButtons)
}

func (h *handler) HandleButtonSelect(ctx context.Context, m *tg.Message, button domain.Button) error {
	var (
		chatID     = m.Chat.ID
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
	if err := h.cleanSend(tg.NewMessage(chatID, fmt.Sprintf("Спасибо! Вы выбрали цвет: %s!", string(button)))); err != nil {
		return err
	}
	return h.cleanSend(tg.NewMessage(chatID, fmt.Sprintf("Отправьте прайс в юанях, соотвествующий кнопке [%s] (шаг 3) 👍", customer.LastEditPosition.Button)))
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
	priceRub, err := h.yuanService.ApplyFormula(priceYuan)
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

	if err := h.cleanSend(tg.NewMessage(chatID, fmt.Sprintf("your price in rub: %d ₽", priceRub))); err != nil {
		return err
	}

	return h.cleanSend(tg.NewMessage(chatID, "Отправьте ссылку на выбранный товар (шаг 4) 👍"))
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

	if err := h.cleanSend(tg.NewMessage(chatID, fmt.Sprintf("Ссылка [%s] принята!", link))); err != nil {
		return err
	}

	positionAddedMsg := tg.NewMessage(chatID, "Позиция успешно добавлена!")
	positionAddedMsg.ReplyMarkup = bottomMenuButtons
	return h.cleanSend(positionAddedMsg)
}

func (h *handler) AddPosition(ctx context.Context, m *tg.Message) error {
	var (
		chatID       = m.Chat.ID
		firstName    = m.From.UserName
		lastName     = m.From.LastName
		chatUsername = m.From.UserName
	)
	return h.addPosition(ctx, chatID, domain.MakeUsername(firstName, lastName, chatUsername))
}

func (h *handler) addPosition(ctx context.Context, telegramID int64, username string) error {
	chatID := telegramID
	_, err := h.customerRepo.GetByTelegramID(ctx, telegramID)
	// if no such customer yet then create it
	if err != nil {
		if !errors.Is(err, domain.ErrCustomerNotFound) {
			return err
		}
		// save to db
		if err := h.customerRepo.Save(ctx, domain.NewCustomer(telegramID, username)); err != nil {
			return err
		}
	}

	if err := h.sendWithKeyboard(chatID, "Отправьте размер выбранного вами товара (шаг 1) 👍", bottomMenuWithouAddPositionButtons); err != nil {
		return err
	}

	return h.customerRepo.UpdateState(ctx, chatID, domain.StateWaitingForSize)
}
