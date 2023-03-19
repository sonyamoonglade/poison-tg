package telegram

import (
	"fmt"
	"math"
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
	menuTrackOrderCallback
	menuCalculatorCallback
	menuMakeOrderCallback
	orderGuideStep1Callback
	orderGuideStep2Callback
	orderGuideStep3Callback
	orderGuideStep4Callback
	makeOrderCallback
	buttonTorqoiseSelectCallback
	buttonGreySelectCallback
	button95SelectCallback
	addPositionCallback
	editCartCallback
)

const editCartRemovePositionOffset = 1000

var (
	initialMenuKeyboard                = initialBottomMenu()
	menuButtons                        = menu()
	selectColorButtons                 = selectButtonColor()
	bottomMenuButtons                  = bottomMenu()
	bottomMenuWithouAddPositionButtons = bottomMenuWithoutAddPosition()
	cartPreviewButtons                 = cartPreview()
	addPositionButtons                 = addPos()
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
			tg.NewInlineKeyboardButtonData("Калькулятор стоимости", strconv.Itoa(menuCalculatorCallback)),
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

func bottomMenu() tg.ReplyKeyboardMarkup {
	return tg.NewReplyKeyboard(
		tg.NewKeyboardButtonRow(
			tg.NewKeyboardButton(menuCommand),
		),
		tg.NewKeyboardButtonRow(
			tg.NewKeyboardButton(getCartCommand),
		),
		tg.NewKeyboardButtonRow(
			tg.NewKeyboardButton(addPositionCommand),
		),
	)
}

func bottomMenuWithoutAddPosition() tg.ReplyKeyboardMarkup {
	return tg.NewReplyKeyboard(
		tg.NewKeyboardButtonRow(
			tg.NewKeyboardButton(menuCommand),
		),
		tg.NewKeyboardButtonRow(
			tg.NewKeyboardButton(getCartCommand),
		),
	)
}

func initialBottomMenu() tg.ReplyKeyboardMarkup {
	return tg.NewReplyKeyboard(
		tg.NewKeyboardButtonRow(
			tg.NewKeyboardButton(menuCommand),
		),
	)
}

func cartPreview() tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Оформить заказ", strconv.Itoa(makeOrderCallback)),
		),
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Редактировать корзину", strconv.Itoa(editCartCallback)),
			tg.NewInlineKeyboardButtonData("Добавить позицию", strconv.Itoa(addPositionCallback)),
		),
	)
}

func addPos() tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Добавить позицию", strconv.Itoa(addPositionCallback)),
		))
}

func prepareEditCartButtons(n int, previewCartMsgID int) tg.InlineKeyboardMarkup {
	keyboard := make([][]tg.InlineKeyboardButton, 0)

	var (
		numRows = int(math.Ceil(float64(n) / 3))
		current int
	)

	for row := 0; row < numRows; row++ {
		keyboard = append(keyboard, tg.NewInlineKeyboardRow())
		for col := 0; col < 3 && current < n; col++ {
			button := tg.NewInlineKeyboardButtonData(strconv.Itoa(current+1), injectMessageIDs(editCartRemovePositionOffset+current+1, int64(previewCartMsgID)))
			keyboard[row] = append(keyboard[row], button)
			current++
		}
	}

	return tg.NewInlineKeyboardMarkup(keyboard...)
}
