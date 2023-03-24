package telegram

import (
	"context"
	"fmt"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sonyamoonglade/poison-tg/internal/domain"
	"github.com/sonyamoonglade/poison-tg/pkg/functools"
)

func (h *handler) sendWithKeyboard(chatID int64, text string, keyboard interface{}) error {
	m := tg.NewMessage(chatID, text)
	m.ReplyMarkup = keyboard
	return h.cleanSend(m)
}

func (h *handler) cleanSend(c tg.Chattable) error {
	_, err := h.b.Send(c)
	return err
}

func (h *handler) checkRequiredState(ctx context.Context, want domain.State, telegramID int64) error {
	customer, err := h.customerRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return fmt.Errorf("checkRequiredState: %w", err)
	}
	if customer.TgState != want {
		return ErrInvalidState
	}
	return nil
}

func (h *handler) deleteUnusedMedia(offset int, chatID int64, msgIDs []int) error {
	// Delete the rest of medias
	for i := offset; i < len(msgIDs); i++ {
		del := tg.NewDeleteMessage(chatID, msgIDs[i])
		_, err := h.b.client.Request(del)
		if err != nil {
			return err
		}
	}
	return nil
}

func makeThumbnails(caption string, urls ...string) []interface{} {
	var first bool
	return functools.Map(func(url string) interface{} {
		thumbnail := tg.NewInputMediaPhoto(tg.FileURL(url))
		if !first {
			// add caption to first element
			thumbnail.Caption = caption
			first = true
		}
		return thumbnail
	}, urls)
}
