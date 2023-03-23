package telegram

import (
	"context"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sonyamoonglade/poison-tg/internal/domain"
	"github.com/sonyamoonglade/poison-tg/internal/repositories/dto"
	"github.com/sonyamoonglade/poison-tg/pkg/functools"
)

// No need to call h.catalogProvider.HasNext. See h.Catalog impl
func (h *handler) HandleCatalogNext(ctx context.Context, chatID int64, controlButtonsMsgID int64, thumnailMsgIDs []int64) error {
	var telegramID = chatID
	customer, err := h.customerRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return err
	}

	next := h.catalogProvider.LoadNext(customer.CatalogOffset)
	// Increment the offset
	customer.CatalogOffset++

	return h.updateCatalog(ctx, chatID, thumnailMsgIDs, controlButtonsMsgID, customer, next)
}

func (h *handler) HandleCatalogPrev(ctx context.Context, chatID int64, controlButtonsMsgID int64, thumnailMsgIDs []int64) error {
	var telegramID = chatID
	customer, err := h.customerRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return err
	}

	prev := h.catalogProvider.LoadPrev(customer.CatalogOffset)
	// Decrement the offset
	customer.CatalogOffset--

	return h.updateCatalog(ctx, chatID, thumnailMsgIDs, controlButtonsMsgID, customer, prev)
}

func (h *handler) updateCatalog(ctx context.Context, chatID int64, thumnailMsgIDs []int64, controlButtonsMsgID int64, customer domain.Customer, item domain.CatalogItem) error {
	// Get next item images
	var first bool
	thumbnails := functools.Map(func(url string) interface{} {
		thumbnail := tg.NewInputMediaPhoto(tg.FileURL(url))
		if !first {
			// add caption to first element
			thumbnail.Caption = item.GetCaption()
			first = true
		}
		return thumbnail
	}, item.ImageURLs)

	var sentMsgIDs []int64
	// Draw it by updating
	for i, thumbMsgID := range thumnailMsgIDs {
		editOneMedia := &tg.EditMessageMediaConfig{
			BaseEdit: tg.BaseEdit{
				ChatID:    chatID,
				MessageID: int(thumbMsgID),
			},
			Media: thumbnails[i],
		}
		sentMessage, err := h.b.Send(editOneMedia)
		if err != nil {
			return err
		}
		sentMsgIDs = append(sentMsgIDs, int64(sentMessage.MessageID))
	}

	updateDTO := dto.UpdateCustomerDTO{
		CatalogOffset: &customer.CatalogOffset,
	}
	if err := h.customerRepo.Update(ctx, customer.CustomerID, updateDTO); err != nil {
		return err
	}

	// Load accordingly to next offset
	var (
		hasNext = h.catalogProvider.HasNext(customer.CatalogOffset)
		hasPrev = h.catalogProvider.HasPrev(customer.CatalogOffset)
	)

	// update buttons
	btnArgs := catalogButtonsArgs{
		hasNext: hasNext,
		hasPrev: hasPrev,
		msgIDs:  sentMsgIDs,
	}
	if hasNext {
		next := h.catalogProvider.LoadNext(customer.CatalogOffset)
		btnArgs.nextTitle = next.Title
	}

	if hasPrev {
		prev := h.catalogProvider.LoadPrev(customer.CatalogOffset)
		btnArgs.prevTitle = prev.Title
	}

	buttons := prepareCatalogButtons(btnArgs)
	editButtons := tg.NewEditMessageReplyMarkup(chatID, int(controlButtonsMsgID), buttons)

	return h.cleanSend(editButtons)
}
