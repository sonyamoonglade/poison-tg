package telegram

import (
	"context"
	"fmt"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sonyamoonglade/poison-tg/internal/domain"
	"github.com/sonyamoonglade/poison-tg/internal/repositories/dto"
)

func (h *handler) Start(ctx context.Context, chatID int64) error {
	return h.sendWithKeyboard(chatID, getTemplate().Start, initialMenuKeyboard)
}

func (h *handler) Menu(ctx context.Context, chatID int64) error {
	return h.sendWithKeyboard(chatID, getTemplate().Menu, menuButtons)
}

func (h *handler) Catalog(ctx context.Context, chatID int64) error {
	return h.cleanSend(tg.NewMessage(chatID, "catalog"))
}

func (h *handler) Calculator(ctx context.Context, chatID int64) error {
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
