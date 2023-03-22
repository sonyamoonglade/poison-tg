package telegram

import (
	"context"
	"fmt"
	"strings"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sonyamoonglade/poison-tg/internal/domain"
	"github.com/sonyamoonglade/poison-tg/internal/repositories/dto"
)

func (h *handler) AskForFIO(ctx context.Context, chatID int64) error {
	var telegramID = chatID
	customer, err := h.customerRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return err
	}
	var (
		isExpressOrder = *customer.Meta.NextOrderType == domain.OrderTypeExpress
		cart           = customer.Cart
	)
	if isExpressOrder && len(cart) > 1 {
		return h.cleanSend(tg.NewMessage(chatID, "Невозможно создать заказ с типом [Экспресс]\nКорзина должна состоять только из 1 элемента"))
	}
	if err := h.customerRepo.UpdateState(ctx, telegramID, domain.StateWaitingForFIO); err != nil {
		return err
	}
	return h.cleanSend(tg.NewMessage(chatID, "Отправь мне ФИО получателя"))
}

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

	shortID, err := h.orderRepo.GetFreeShortID(ctx)
	if err != nil {
		return err
	}

	isExpress := *customer.Meta.NextOrderType == domain.OrderTypeExpress
	order := domain.NewOrder(customer, address, isExpress, shortID)

	if err := h.orderRepo.Save(ctx, order); err != nil {
		return err
	}

	return h.prepareOrderPreview(ctx, customer, order, chatID)
}

func (h *handler) makeOrder(ctx context.Context, m *tg.Message) error {
	return nil
}

func (h *handler) prepareOrderPreview(ctx context.Context, customer domain.Customer, order domain.Order, chatID int64) error {
	out := getOrderStart(orderStartArgs{
		fullName:        *customer.FullName,
		shortOrderID:    order.ShortID,
		phoneNumber:     *customer.PhoneNumber,
		isExpress:       order.IsExpress,
		deliveryAddress: order.DeliveryAddress,
		nCartItems:      len(order.Cart),
	})

	for i, cartItem := range order.Cart {
		out += getPositionTemplate(cartPositionPreviewArgs{
			n:         i + 1,
			link:      cartItem.ShopLink,
			size:      cartItem.Size,
			priceRub:  cartItem.PriceRUB,
			priceYuan: cartItem.PriceYUAN,
		})
	}

	out += getOrderEnd(order.AmountRUB)

	if err := h.cleanSend(tg.NewMessage(chatID, out)); err != nil {
		return err
	}

	updateDTO := dto.UpdateCustomerDTO{
		Meta:  &domain.Meta{},
		Cart:  new(domain.Cart),
		State: &domain.StateDefault,
	}

	if err := h.customerRepo.Update(ctx, customer.CustomerID, updateDTO); err != nil {
		return err
	}

	requisites, err := h.businessRepo.GetRequisites(ctx)
	if err != nil {
		return err
	}

	requisitesMsg := tg.NewMessage(chatID, getRequisites(requisites, order.ShortID))
	sentRequisitesMsg, err := h.b.Send(requisitesMsg)
	if err != nil {
		return err
	}

	editButton := tg.NewEditMessageReplyMarkup(chatID, sentRequisitesMsg.MessageID, preparePaymentButton(sentRequisitesMsg.MessageID))
	return h.cleanSend(editButton)
}

func (h *handler) HandlePayment(ctx context.Context, shortOrderID string, c *tg.CallbackQuery) error {
	var (
	//chatID     = c.From.ID
	//telegramID = chatID
	)

	//customer, err := h.customerRepo.GetByTelegramID(ctx, telegramID)
	//if err != nil {
	//	return fmt.Errorf("customerRepo.GetByTelegramID: %w", err)
	//}

	//shortOrderID, err := extractShortOrderIDFromRequisites()
	//panic(err)
	return nil
}
