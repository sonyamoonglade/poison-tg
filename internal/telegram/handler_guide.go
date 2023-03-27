package telegram

import (
	"context"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sonyamoonglade/poison-tg/internal/domain"
	"github.com/sonyamoonglade/poison-tg/internal/repositories/dto"
	"github.com/sonyamoonglade/poison-tg/pkg/functools"
)

func (h *handler) askForOrderType(ctx context.Context, chatID int64) error {
	text := "Выбери тип доставки"
	return h.sendWithKeyboard(chatID, text, orderTypeButtons)
}

func (h *handler) HandleOrderTypeInput(ctx context.Context, chatID int64, typ domain.OrderType) error {
	var (
		telegramID = chatID
		isExpress  = typ == domain.OrderTypeExpress
	)

	if err := h.checkRequiredState(ctx, domain.StateWaitingForOrderType, chatID); err != nil {
		return err
	}

	customer, err := h.customerRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return err
	}

	customer.UpdateMetaOrderType(typ)
	if isExpress {
		// If order type is express then it's no matter which location user would put,
		// so whatever
		customer.UpdateMetaLocation(domain.LocationOther)
		// skip location state
		customer.TgState = domain.StateWaitingForSize
	} else {
		customer.TgState = domain.StateWaitingForLocation
	}

	updateDTO := dto.UpdateCustomerDTO{
		Meta: &domain.Meta{
			NextOrderType: customer.Meta.NextOrderType,
			Location:      customer.Meta.Location,
		},
		State: &customer.TgState,
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
	if err := h.sendMessage(chatID, resp); err != nil {
		return err
	}

	if isExpress {
		return h.addPosition(ctx, chatID)
	}

	return h.askForLocation(ctx, chatID)
}

func (h *handler) askForLocation(ctx context.Context, chatID int64) error {
	text := "Из какого ты города?"
	return h.sendWithKeyboard(chatID, text, locationButtons)
}

func (h *handler) HandleLocationInput(ctx context.Context, chatID int64, loc domain.Location) error {
	var telegramID = chatID

	if err := h.checkRequiredState(ctx, domain.StateWaitingForLocation, chatID); err != nil {
		return err
	}

	customer, err := h.customerRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return err
	}

	customer.UpdateMetaLocation(loc)

	updateDTO := dto.UpdateCustomerDTO{
		Meta: &customer.Meta,
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

	if err := h.sendMessage(chatID, resp); err != nil {
		return err
	}

	return h.addPosition(ctx, telegramID)
}

func (h *handler) StartMakeOrderGuide(ctx context.Context, m *tg.Message) error {
	var (
		chatID     = m.Chat.ID
		telegramID = chatID
	)

	thumbnails := makeThumbnails(getTemplate().GuideStep1, guideStep1Thumbnail1, guideStep1Thumbnail2)
	group := tg.NewMediaGroup(chatID, thumbnails)

	sentMsgs, err := h.b.client.SendMediaGroup(group)
	if err != nil {
		return err
	}

	msgIDs := functools.Map(func(m tg.Message) int {
		return m.MessageID
	}, sentMsgs)

	buttons := prepareOrderGuideButtons(orderGuideStep0Callback, msgIDs...)
	if err := h.sendWithKeyboard(chatID, "Кнопки для пролистывания инструкции", buttons); err != nil {
		return err
	}

	customer, err := h.customerRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return err
	}

	// If cart is not empty then skip location and order type ask
	if len(customer.Cart) > 0 {
		return h.addPosition(ctx, chatID)
	}

	if err := h.customerRepo.UpdateState(ctx, telegramID, domain.StateWaitingForOrderType); err != nil {
		return err
	}

	return h.askForOrderType(ctx, chatID)
}

func (h *handler) MakeOrderGuideStep1(ctx context.Context, chatID int64, controlButtonsMessageID int, guideMsgIDs []int) error {
	thumbnails := makeThumbnails(getTemplate().GuideStep1, guideStep1Thumbnail1, guideStep1Thumbnail2)
	return h.updateGuideStep(chatID, guideMsgIDs, controlButtonsMessageID, orderGuideStep0Callback, thumbnails)
}

func (h *handler) MakeOrderGuideStep2(ctx context.Context, chatID int64, controlButtonsMessageID int, guideMsgIDs []int) error {
	thumbnails := makeThumbnails(getTemplate().GuideStep2, guideStep2Thumbnail1, guideStep2Thumbnail2)
	return h.updateGuideStep(chatID, guideMsgIDs, controlButtonsMessageID, orderGuideStep1Callback, thumbnails)
}

func (h *handler) MakeOrderGuideStep3(ctx context.Context, chatID int64, controlButtonsMessageID int, guideMsgIDs []int) error {
	thumbnails := makeThumbnails(getTemplate().GuideStep3, guideStep3Thumbnail1, guideStep3Thumbnail2)
	return h.updateGuideStep(chatID, guideMsgIDs, controlButtonsMessageID, orderGuideStep2Callback, thumbnails)
}

func (h *handler) MakeOrderGuideStep4(ctx context.Context, chatID int64, controlButtonsMessageID int, guideMsgIDs []int) error {
	thumbnails := makeThumbnails(getTemplate().GuideStep4, guideStep4Thumbnail1, guideStep4Thumbnail2)
	return h.updateGuideStep(chatID, guideMsgIDs, controlButtonsMessageID, orderGuideStep3Callback, thumbnails)
}

func (h *handler) MakeOrderGuideStep5(ctx context.Context, chatID int64, controlButtonsMessageID int, guideMsgIDs []int) error {
	thumbnails := makeThumbnails(getTemplate().GuideStep5, guideStep5Thumbnail1, guideStep5Thumbnail2)
	return h.updateGuideStep(chatID, guideMsgIDs, controlButtonsMessageID, orderGuideStep4Callback, thumbnails)
}

func (h *handler) MakeOrderGuideStep6(ctx context.Context, chatID int64, controlButtonsMessageID int, guideMsgIDs []int) error {
	thumbnails := makeThumbnails(getTemplate().GuideStep6, guideStep6Thumbnail1, guideStep6Thumbnail2)
	return h.updateGuideStep(chatID, guideMsgIDs, controlButtonsMessageID, orderGuideStep5Callback, thumbnails)
}

func (h *handler) updateGuideStep(chatID int64, guideMsgIDs []int, controlButtonsMessageID int, nextCallback int, thumbnails []interface{}) error {
	for i, t := range thumbnails {
		editOneMedia := &tg.EditMessageMediaConfig{
			BaseEdit: tg.BaseEdit{
				ChatID:    chatID,
				MessageID: guideMsgIDs[i],
			},
			Media: t,
		}
		if err := h.cleanSend(editOneMedia); err != nil {
			return err
		}
	}

	// update control buttons
	buttons := tg.NewEditMessageReplyMarkup(chatID, controlButtonsMessageID, prepareOrderGuideButtons(nextCallback, guideMsgIDs...))
	return h.cleanSend(buttons)

}
