package telegram

import (
	"context"
	"fmt"
	"strings"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sonyamoonglade/poison-tg/internal/domain"
	"github.com/sonyamoonglade/poison-tg/internal/repositories/dto"
)

func (h *handler) HandleFIOInput(ctx context.Context, m *tg.Message) error {
	var (
		chatID     = m.From.ID
		telegramID = chatID
		fullName   = strings.TrimSpace(m.Text)
	)

	// validate state
	if err := h.checkRequiredState(ctx, domain.StateWaitingForFIO, chatID); err != nil {
		return err
	}

	customer, err := h.customerRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return fmt.Errorf("customerRepo.GetByTelegramID: %w", err)
	}

	if !domain.IsValidFullName(fullName) {
		return h.cleanSend(tg.NewMessage(chatID, "Неправильный формат полного имени.\n Отправь полное имя в формате - Иванов Иван Иванович"))
	}

	updateDTO := dto.UpdateCustomerDTO{
		State:    &domain.StateWaitingForPhoneNumber,
		FullName: &fullName,
	}

	if err := h.customerRepo.Update(ctx, customer.CustomerID, updateDTO); err != nil {
		return fmt.Errorf("customerRepo.Update: %w", err)
	}

	if err := h.cleanSend(tg.NewMessage(chatID, fmt.Sprintf("Спасибо, %s. ", fullName))); err != nil {
		return err
	}

	return h.cleanSend(tg.NewMessage(chatID, "Отправь мне свой контактный номер телефона в формате:\n\t>79999999999"))
}

func (h *handler) HandlePhoneNumberInput(ctx context.Context, m *tg.Message) error {
	var (
		chatID      = m.From.ID
		telegramID  = chatID
		phoneNumber = strings.TrimSpace(m.Text)
	)

	// validate state
	if err := h.checkRequiredState(ctx, domain.StateWaitingForPhoneNumber, chatID); err != nil {
		return err
	}

	if !domain.IsValidPhoneNumber(phoneNumber) {
		return h.cleanSend(tg.NewMessage(chatID, "Неправильный формат номера телефона.\n Отправь номер в формате - 79999999999"))
	}

	customer, err := h.customerRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return fmt.Errorf("customerRepo.GetByTelegramID: %w", err)
	}

	updateDTO := dto.UpdateCustomerDTO{
		State:       &domain.StateWaitingForDeliveryAddress,
		PhoneNumber: &phoneNumber,
	}

	if err := h.customerRepo.Update(ctx, customer.CustomerID, updateDTO); err != nil {
		return fmt.Errorf("customerRepo.Update: %w", err)
	}

	if err := h.cleanSend(tg.NewMessage(chatID, fmt.Sprintf("Спасибо, номер [%s] принят!", phoneNumber))); err != nil {
		return err
	}

	return h.cleanSend(tg.NewMessage(chatID, "Отправь адрес ближайшего постамата PickPoint или отделения Сбера (Сбербанк).\nЯ доставлю твой заказ туда!"))
}

func (h *handler) HandleDeliveryAddressInput(ctx context.Context, m *tg.Message) error {
	var (
		chatID     = m.From.ID
		telegramID = chatID
		address    = strings.TrimSpace(m.Text)
	)

	// validate state
	if err := h.checkRequiredState(ctx, domain.StateWaitingForDeliveryAddress, chatID); err != nil {
		return err
	}

	customer, err := h.customerRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return fmt.Errorf("customerRepo.GetByTelegramID: %w", err)
	}

	updateDTO := dto.UpdateCustomerDTO{
		LastPosition: &domain.Position{},
		Cart:         &domain.Cart{},
		State:        &domain.StateDefault,
	}

	if err := h.customerRepo.Update(ctx, customer.CustomerID, updateDTO); err != nil {
		return fmt.Errorf("customerRepo.Update: %w", err)
	}

	// todo: finish flow
	order := domain.NewOrder(customer, address)
	_ = order
	return nil
}

func (h *handler) makeOrder(ctx context.Context, m *tg.Message) error {
	return nil
}
