package telegram

import (
	"context"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sonyamoonglade/poison-tg/internal/domain"
	"github.com/sonyamoonglade/poison-tg/internal/repositories/dto"
	"github.com/sonyamoonglade/poison-tg/pkg/functools"
)

func (h *handler) askForOrderType(ctx context.Context, chatID int64) error {
	text := "Shag 0. Выбери тип доставки (время экспресс перевозки в среднем составляет 4 дня из Китая в СПб, обычная перевозка составляет 8-15 дней)"
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

	var updateDTO dto.UpdateCustomerDTO
	customer.UpdateMetaOrderType(typ)
	if isExpress {
		// If order type is express then it's no matter which location user would put,
		// so whatever
		customer.UpdateMetaLocation(domain.LocationOther)
	}

	if err := h.customerRepo.Update(ctx, customer.CustomerID, updateDTO); err != nil {
		return err
	}
	//todo: translate
	var resp = "Тип заказа: "
	switch isExpress {
	case true:
		resp += "Express"
		break
	case false:
		resp += "Normal"
		break
	}
	if err := h.cleanSend(tg.NewMessage(chatID, resp)); err != nil {
		return err
	}

	if err := h.customerRepo.UpdateState(ctx, telegramID, domain.StateWaitingForLocation); err != nil {
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
		resp += "SPB"
		break
	case domain.LocationIZH:
		resp += "IZh"
		break
	case domain.LocationOther:
		resp += "other"
		break
	}

	if err := h.cleanSend(tg.NewMessage(chatID, resp)); err != nil {
		return err
	}

	return h.addPosition(ctx, telegramID)
}

func (h *handler) StartMakeOrderGuide(ctx context.Context, m *tg.Message) error {
	var (
		chatID     = m.Chat.ID
		telegramID = chatID
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

	customer, err := h.customerRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return err
	}

	if len(customer.Cart) > 0 {
		return h.addPosition(ctx, chatID)
	}

	if err := h.customerRepo.UpdateState(ctx, telegramID, domain.StateWaitingForOrderType); err != nil {
		return err
	}

	return h.askForOrderType(ctx, chatID)
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
