package telegram

import (
	"strconv"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	catalogCallback = iota + 1
	makeOrderCallback
	getBucketCallback
	trackOrderCallback
	startMakeOrderCallback
)

type buttons struct {
	Menu           tg.InlineKeyboardMarkup
	StartMakeOrder tg.InlineKeyboardMarkup
}

type ButtonManager struct {
	b buttons
}

func NewButtonManager() ButtonProvider {
	return &ButtonManager{
		b: buttons{
			Menu:           menu(),
			StartMakeOrder: startMakeOrder(),
		},
	}
}

func (bm *ButtonManager) Menu() tg.InlineKeyboardMarkup {
	return bm.b.Menu
}

func (bm *ButtonManager) StartMakeOrder() tg.InlineKeyboardMarkup {
	return bm.b.StartMakeOrder
}

func menu() tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Каталог", strconv.Itoa(catalogCallback)),
		),

		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Сделать заказ", strconv.Itoa(makeOrderCallback)),
		),
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Посмотреть корзину", strconv.Itoa(getBucketCallback)),
		),
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Отследить посылку", strconv.Itoa(trackOrderCallback)),
		),
	)
}

func startMakeOrder() tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Приступить к созданию заказа!", strconv.Itoa(startMakeOrderCallback))),
	)
}
