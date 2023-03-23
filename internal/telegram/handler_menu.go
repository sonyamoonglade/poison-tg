package telegram

import (
	"context"
	"errors"
	"fmt"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sonyamoonglade/poison-tg/internal/domain"
	"github.com/sonyamoonglade/poison-tg/internal/repositories/dto"
	"github.com/sonyamoonglade/poison-tg/pkg/functools"
)

func (h *handler) Start(ctx context.Context, m *tg.Message) error {
	var (
		chatID       = m.Chat.ID
		telegramID   = chatID
		firstName    = m.From.FirstName
		lastName     = m.From.LastName
		chatUsername = m.From.UserName
		username     = domain.MakeUsername(firstName, lastName, chatUsername)
	)
	// register customer
	_, err := h.customerRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		if !errors.Is(err, domain.ErrCustomerNotFound) {
			return err
		}
		// save to db
		if err := h.customerRepo.Save(ctx, domain.NewCustomer(telegramID, username)); err != nil {
			return err
		}
	}
	return h.sendWithKeyboard(chatID, getTemplate().Start, initialMenuKeyboard)
}

func (h *handler) Menu(ctx context.Context, chatID int64) error {
	if err := h.customerRepo.UpdateState(ctx, chatID, domain.StateDefault); err != nil {
		return err
	}
	return h.sendWithKeyboard(chatID, getTemplate().Menu, menuButtons)
}

func (h *handler) Catalog(ctx context.Context, chatID int64) error {
	var (
		telegramID = chatID
		first      bool
	)

	customer, err := h.customerRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return err
	}

	if err := h.cleanSend(tg.NewMessage(chatID, getCatalog(*customer.Username))); err != nil {
		return err
	}

	// Load appropriate item
	item := h.catalogProvider.LoadAt(customer.CatalogOffset)

	thumbnails := functools.Map(func(url string) interface{} {
		thumbnail := tg.NewInputMediaPhoto(tg.FileURL(url))
		if !first {
			// add caption to first element
			thumbnail.Caption = item.GetCaption()
			first = true
		}
		return thumbnail
	}, item.ImageURLs)

	group := tg.NewMediaGroup(chatID, thumbnails)

	// Sends thumnails with caption
	sentMsgs, err := h.b.client.SendMediaGroup(group)
	if err != nil {
		return err
	}

	msgIDs := functools.Map(func(m tg.Message) int64 {
		return int64(m.MessageID)
	}, sentMsgs)

	// Prepare buttons for controlling prev, next
	var (
		currentOffset = customer.CatalogOffset
		hasNext       = h.catalogProvider.HasNext(currentOffset)
		hasPrev       = h.catalogProvider.HasPrev(currentOffset)
	)

	btnArgs := catalogButtonsArgs{
		hasNext: hasNext,
		hasPrev: hasPrev,
		msgIDs:  msgIDs,
	}

	if hasNext {
		next := h.catalogProvider.LoadNext(currentOffset)
		btnArgs.nextTitle = next.Title
	} else if hasPrev {
		prev := h.catalogProvider.LoadPrev(currentOffset)
		btnArgs.prevTitle = prev.Title
	}

	buttons := prepareCatalogButtons(btnArgs)
	return h.sendWithKeyboard(chatID, "Кнопки для пролистывания каталога", buttons)
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
