package telegram

import (
	"fmt"
	"strconv"
	"strings"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	arrLeft  = "⬅"
	arrRight = "➡"
)

// DO NOT CHANGE ORDER
// LOGIC DEMANDS ON IOTA
// todo: change from iota
const (
	menuCatalogCallback = iota + 1
	menuGetBucketCallback
	menuTrackOrderCallback
	menuMakeOrderCallback
	orderGuideStep1Callback
	orderGuideStep2Callback
	orderGuideStep3Callback
	orderGuideStep4Callback
	makeOrderCallback
	buttonTorqoiseSelectCallback
	buttonGreySelectCallback
	button95SelectCallback
)

var (
	menuButtons        = menu()
	selectColorButtons = selectButtonColor()
)

func injectMessageIDs(callback int, msgIDs ...int64) string {
	var msgIDstr string
	for i, m := range msgIDs {
		if i < len(msgIDs)-1 {
			msgIDstr += strconv.Itoa(int(m)) + ","
		} else {
			msgIDstr += strconv.Itoa(int(m))
		}
	}
	return msgIDstr + ":" + strconv.Itoa(callback)
}

func parseCallbackData(data string) ([]int64, int, error) {
	if !strings.ContainsRune(data, ':') {
		callback, err := strconv.Atoi(data)
		if err != nil {
			return nil, 0, fmt.Errorf("strconv.Atoi: %w", err)
		}
		return nil, callback, nil
	}
	var (
		msgIDstrs []string
		msgIDints []int64
	)

	spl := strings.Split(data, ":")
	msgIDstrs = strings.Split(spl[0], ",")
	cbStr := spl[1]

	for _, m := range msgIDstrs {
		mInt, err := strconv.Atoi(m)
		if err != nil {
			return nil, 0, fmt.Errorf("strconv.Atoi msgID: %w", err)
		}
		msgIDints = append(msgIDints, int64(mInt))
	}

	cbInt, err := strconv.Atoi(cbStr)
	if err != nil {
		return nil, 0, fmt.Errorf("strconv.Atoi cb: %w", err)
	}

	return msgIDints, cbInt, nil
}

func menu() tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Каталог", strconv.Itoa(menuCatalogCallback)),
		),

		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Сделать заказ", strconv.Itoa(menuMakeOrderCallback)),
		),
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Посмотреть корзину", strconv.Itoa(menuGetBucketCallback)),
		),
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Отследить посылку", strconv.Itoa(menuTrackOrderCallback)),
		),
	)
}

func prepareOrderGuideButtons(step int, msgIDs ...int64) tg.InlineKeyboardMarkup {
	if step == orderGuideStep4Callback {
		return tg.NewInlineKeyboardMarkup(
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData(arrLeft, injectMessageIDs(step-1, msgIDs...)),
			),
		)
	} else if step == orderGuideStep1Callback {
		return tg.NewInlineKeyboardMarkup(
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData(arrRight, injectMessageIDs(step+1, msgIDs...)),
			),
		)
	}

	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData(arrLeft, injectMessageIDs(step-1, msgIDs...)),
			tg.NewInlineKeyboardButtonData(arrRight, injectMessageIDs(step+1, msgIDs...)),
		),
	)
}

func selectButtonColor() tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Бирюзовый", strconv.Itoa(buttonTorqoiseSelectCallback)),
			tg.NewInlineKeyboardButtonData("Серый", strconv.Itoa(buttonGreySelectCallback)),
			tg.NewInlineKeyboardButtonData("95% БУ", strconv.Itoa(button95SelectCallback)),
		),
	)
}
