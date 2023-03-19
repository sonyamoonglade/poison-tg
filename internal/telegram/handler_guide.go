package telegram

import (
	"context"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sonyamoonglade/poison-tg/internal/domain"
	"github.com/sonyamoonglade/poison-tg/pkg/functools"
)

func (h *handler) StartMakeOrderGuide(ctx context.Context, m *tg.Message) error {
	var (
		chatID       = m.Chat.ID
		firstName    = m.From.FirstName
		lastName     = m.From.LastName
		chatUsername = m.From.UserName
		telegramID   = chatID
		username     = domain.MakeUsername(firstName, lastName, chatUsername)
	)
	url := "https://picsum.photos/300/300"
	image := tg.NewInputMediaPhoto(tg.FileURL(url))
	image.Caption = "У меня есть желание привезти лишь то,что нужно, поэтому, (имя юзера), предупреждаю, китайцы уже позаботились о нас и предоставили к каждому размерному товару - размерную сетку разных стран, тебе лишь нужно выбрать подходящий размер. Не ошибись с выбором, Стрелок  🤠 Поехали?"
	group := tg.NewMediaGroup(chatID, []interface{}{
		image,
		tg.NewInputMediaPhoto(tg.FileURL(url)),
	})
	sentMsgs, err := h.b.client.SendMediaGroup(group)
	if err != nil {
		return err
	}

	msgIDs := functools.Map(func(m tg.Message) int64 {
		return int64(m.MessageID)
	}, sentMsgs)

	buttons := prepareOrderGuideButtons(orderGuideStep1Callback, msgIDs...)
	if err := h.sendWithKeyboard(chatID, "Кнопки для пролистывания инструкции", buttons); err != nil {
		return err
	}

	return h.addPosition(ctx, telegramID, username)
}

func (h *handler) MakeOrderGuideStep1(ctx context.Context, chatID int64, controlButtonsMessageID int, instructionMsgIDs ...int64) error {
	// update content
	for i, imID := range instructionMsgIDs {
		url := "https://picsum.photos/300/301"
		image := tg.NewInputMediaPhoto(tg.FileURL(url))
		// update Caption only on first element on order to show text (see telegram docs)
		if i == 0 {
			image.Caption = "У меня есть желание привезти лишь то,что нужно, поэтому, (имя юзера), предупреждаю, китайцы уже позаботились о нас и предоставили к каждому размерному товару - размерную сетку разных стран, тебе лишь нужно выбрать подходящий размер. Не ошибись с выбором, Стрелок  🤠 Поехали?"
		}
		editOneMedia := &tg.EditMessageMediaConfig{
			BaseEdit: tg.BaseEdit{
				ChatID:    chatID,
				MessageID: int(imID),
			},
			Media: image,
		}
		if err := h.cleanSend(editOneMedia); err != nil {
			return err
		}
	}
	// update control buttons
	buttons := tg.NewEditMessageReplyMarkup(chatID, controlButtonsMessageID, prepareOrderGuideButtons(orderGuideStep1Callback, instructionMsgIDs...))
	return h.cleanSend(buttons)
}

func (h *handler) MakeOrderGuideStep2(ctx context.Context, chatID int64, controlButtonsMessageID int, instructionMsgIDs ...int64) error {
	// update content
	for i, imID := range instructionMsgIDs {
		url := "https://picsum.photos/300/302"
		image := tg.NewInputMediaPhoto(tg.FileURL(url))
		// update Caption only on first element on order to show text (see telegram docs)
		if i == 0 {
			image.Caption = "This is step 2 of instruction"
		}
		editOneMedia := &tg.EditMessageMediaConfig{
			BaseEdit: tg.BaseEdit{
				ChatID:    chatID,
				MessageID: int(imID),
			},
			Media: image,
		}
		if err := h.cleanSend(editOneMedia); err != nil {
			return err
		}
	}
	// update control buttons
	buttons := tg.NewEditMessageReplyMarkup(chatID, controlButtonsMessageID, prepareOrderGuideButtons(orderGuideStep2Callback, instructionMsgIDs...))
	return h.cleanSend(buttons)
}

func (h *handler) MakeOrderGuideStep3(ctx context.Context, chatID int64, controlButtonsMessageID int, instructionMsgIDs ...int64) error {
	// update content
	for i, imID := range instructionMsgIDs {
		url := "https://picsum.photos/300/303"
		image := tg.NewInputMediaPhoto(tg.FileURL(url))
		// update Caption only on first element on order to show text (see telegram docs)
		if i == 0 {
			image.Caption = "This is step 3 of instruction"
		}
		editOneMedia := &tg.EditMessageMediaConfig{
			BaseEdit: tg.BaseEdit{
				ChatID:    chatID,
				MessageID: int(imID),
			},
			Media: image,
		}
		if err := h.cleanSend(editOneMedia); err != nil {
			return err
		}
	}
	// update control buttons
	buttons := tg.NewEditMessageReplyMarkup(chatID, controlButtonsMessageID, prepareOrderGuideButtons(orderGuideStep3Callback, instructionMsgIDs...))
	return h.cleanSend(buttons)
}

func (h *handler) MakeOrderGuideStep4(ctx context.Context, chatID int64, controlButtonsMessageID int, instructionMsgIDs ...int64) error {
	// update content
	for i, imID := range instructionMsgIDs {
		url := "https://picsum.photos/300/304"
		image := tg.NewInputMediaPhoto(tg.FileURL(url))
		// update Caption only on first element on order to show text (see telegram docs)
		if i == 0 {
			image.Caption = "This is step 4 of instruction"
		}
		editOneMedia := &tg.EditMessageMediaConfig{
			BaseEdit: tg.BaseEdit{
				ChatID:    chatID,
				MessageID: int(imID),
			},
			Media: image,
		}
		if err := h.cleanSend(editOneMedia); err != nil {
			return err
		}
	}
	// update control buttons
	buttons := tg.NewEditMessageReplyMarkup(chatID, controlButtonsMessageID, prepareOrderGuideButtons(orderGuideStep4Callback, instructionMsgIDs...))
	return h.cleanSend(buttons)
}
