package telegram

import (
	"context"
	"errors"
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sonyamoonglade/poison-tg/internal/domain"
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

	if err := h.sendWithKeyboard(chatID, getStartTemplate(username), initialMenuKeyboard); err != nil {
		return err
	}

	yuanRate, err := h.yuanService.GetRate()
	if err != nil {
		return err
	}

	return h.cleanSend(tg.NewMessage(chatID, fmt.Sprintf("Курс юаня на сегодня: %.2f ₽", yuanRate)))
}

func (h *handler) Menu(ctx context.Context, chatID int64) error {
	if err := h.customerRepo.UpdateState(ctx, chatID, domain.StateDefault); err != nil {
		return err
	}
	return h.sendWithKeyboard(chatID, getTemplate().Menu, menuButtons)
}

func (h *handler) MyOrders(ctx context.Context, chatID int64) error {
	var telegramID = chatID

	customer, err := h.customerRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return err
	}

	orders, err := h.orderRepo.GetAll(ctx, customer.CustomerID)
	if err != nil {
		if errors.Is(err, domain.ErrNoOrders) {
			return h.cleanSend(tg.NewMessage(chatID, "У вас пока нет заказов"))
		}

		return err
	}
	if len(orders) == 0 {
		return h.cleanSend(tg.NewMessage(chatID, "У вас еще нет заказов :("))
	}
	var name string
	if customer.FullName != nil {
		name = *customer.FullName
	} else {
		name = *customer.Username
	}
	out := getMyOrdersStart(name)
	for _, o := range orders {
		out += getSingleOrderPreview(singleOrderArgs{
			shortID:         o.ShortID,
			isExpress:       o.IsExpress,
			isPaid:          o.IsPaid,
			isApproved:      o.IsApproved,
			cartLen:         len(o.Cart),
			deliveryAddress: o.DeliveryAddress,
			totalYuan:       o.AmountYUAN,
			totalRub:        o.AmountRUB,
		})
		for nCartItem, cartItem := range o.Cart {
			out += getPositionTemplate(cartPositionPreviewArgs{
				n:         nCartItem + 1,
				link:      cartItem.ShopLink,
				size:      cartItem.Size,
				priceRub:  cartItem.PriceRUB,
				priceYuan: cartItem.PriceYUAN,
			})
		}

		out += getTemplate().MyOrdersEnd
	}

	return h.cleanSend(tg.NewMessage(chatID, out))
}
