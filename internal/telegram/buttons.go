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
	mockCallback = iota
	menuCatalogCallback
	menuTrackOrderCallback
	menuCalculatorCallback
	menuMakeOrderCallback
	myOrdersCallback
	orderGuideStep0Callback
	orderGuideStep1Callback
	orderGuideStep2Callback
	orderGuideStep3Callback
	orderGuideStep4Callback
	orderGuideStep5Callback
	makeOrderCallback
	buttonTorqoiseSelectCallback
	buttonGreySelectCallback
	button95SelectCallback
	addPositionCallback
	editCartCallback
	izhLocationCallback
	izhLocationCalculatorCallback
	spbLocationCallback
	spbLocationCalculatorCallback
	othLocationCallback
	othLocationCalculatorCallback
	orderTypeNormalCallback
	orderTypeNormalCalculatorCallback
	orderTypeExpressCallback
	orderTypeExpressCalculatorCallback
	paymentCallback
)

const (
	editCartRemovePositionOffset = 1000
	catalogOffset                = 1200
)

const (
	catalogPrevCallback = iota + 1
	catalogNextCallback
)

var (
	initialMenuKeyboard                 = initialBottomMenu()
	menuButtons                         = menu()
	selectColorButtons                  = selectButtonColor()
	bottomMenuButtons                   = bottomMenu()
	bottomMenuWithoutAddPositionButtons = bottomMenuWithoutAddPosition()
	cartPreviewButtons                  = cartPreview()
	addPositionButtons                  = addPos()
	locationButtons                     = location()
	orderTypeButtons                    = orderType()
	locationCalculatorButtons           = locationCalculator()
	orderTypeCalculatorButtons          = orderTypeCalculator()
)

func injectMessageIDs(callback int, msgIDs ...int) string {
	var msgIDstr string
	for i, m := range msgIDs {
		if i < len(msgIDs)-1 {
			msgIDstr += strconv.Itoa(m) + ","
		} else {
			msgIDstr += strconv.Itoa(m)
		}
	}
	return "m" + msgIDstr + ":" + strconv.Itoa(callback)
}

func injectStringData(callback int, str string) string {
	return "s" + str + ":" + strconv.Itoa(callback)
}

func parseStringCallbackData(data string) (payload string, callback int, err error) {
	data = data[1:]
	var colonIdx int
	for i, ch := range data {
		if ch == ':' {
			colonIdx = i
			break
		}
	}
	callback, err = strconv.Atoi(data[colonIdx+1:])
	if err != nil {
		return "", 0, err
	}

	return data[0:colonIdx], callback, nil
}

func parseCallbackData(data string) (injectedData any, callback int, err error) {
	// raw callback
	if !strings.ContainsRune(data, ':') {
		callback, err := strconv.Atoi(data)
		if err != nil {
			return nil, 0, fmt.Errorf("strconv.Atoi: %w", err)
		}
		return nil, callback, nil
	}

	prefix := data[0]
	// means message id's are injected
	if prefix == 'm' {
		var (
			msgIDstrs []string
			msgIDints []int
		)
		spl := strings.Split(data[1:], ":")
		msgIDstrs = strings.Split(spl[0], ",")
		cbStr := spl[1]

		for _, m := range msgIDstrs {
			mInt, err := strconv.Atoi(m)
			if err != nil {
				return nil, 0, fmt.Errorf("strconv.Atoi msgID: %w", err)
			}
			msgIDints = append(msgIDints, mInt)
		}

		cbInt, err := strconv.Atoi(cbStr)
		if err != nil {
			return nil, 0, fmt.Errorf("strconv.Atoi cb: %w", err)
		}

		return msgIDints, cbInt, nil
	}

	// string data encoded
	if prefix == 's' {
		return parseStringCallbackData(data)
	}

	return
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
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Мои заказы", strconv.Itoa(myOrdersCallback)),
		),
	)
}

func prepareOrderGuideButtons(step int, msgIDs ...int) tg.InlineKeyboardMarkup {
	if step == orderGuideStep5Callback {
		return tg.NewInlineKeyboardMarkup(
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData(arrLeft, injectMessageIDs(step-1, msgIDs...)),
			),
		)
	} else if step == orderGuideStep0Callback {
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
			button := tg.NewInlineKeyboardButtonData(strconv.Itoa(current+1), injectMessageIDs(editCartRemovePositionOffset+current+1, previewCartMsgID))
			keyboard[row] = append(keyboard[row], button)
			current++
		}
	}

	return tg.NewInlineKeyboardMarkup(keyboard...)
}

func location() tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Ижевск", strconv.Itoa(izhLocationCallback)),
			tg.NewInlineKeyboardButtonData("Питер", strconv.Itoa(spbLocationCallback)),
			tg.NewInlineKeyboardButtonData("Другой город", strconv.Itoa(othLocationCallback)),
		))
}

func orderType() tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Экспресс", strconv.Itoa(orderTypeExpressCallback)),
			tg.NewInlineKeyboardButtonData("Обычный", strconv.Itoa(orderTypeNormalCallback)),
		))
}

func preparePaymentButton(orderShortID string) tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Оплачено", injectStringData(paymentCallback, orderShortID)),
		))
}

type catalogButtonsArgs struct {
	hasNext, hasPrev     bool
	nextTitle, prevTitle string
	msgIDs               []int
}

func prepareCatalogButtons(args catalogButtonsArgs) tg.InlineKeyboardMarkup {
	if args.hasNext && args.hasPrev {
		return tg.NewInlineKeyboardMarkup(
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData("< "+args.prevTitle, injectMessageIDs(catalogOffset+catalogPrevCallback, args.msgIDs...)),
				tg.NewInlineKeyboardButtonData(args.nextTitle+" >", injectMessageIDs(catalogOffset+catalogNextCallback, args.msgIDs...)),
			))
	} else if args.hasNext {
		return tg.NewInlineKeyboardMarkup(
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData(args.nextTitle+" >", injectMessageIDs(catalogOffset+catalogNextCallback, args.msgIDs...)),
			))
	}

	// only prev
	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("< "+args.prevTitle, injectMessageIDs(catalogOffset+catalogPrevCallback, args.msgIDs...)),
		))
}

func prepareAfterPaidButtons(shortOrderId string) tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData(fmt.Sprintf("Заказ %s оплачен ✅", shortOrderId), strconv.Itoa(mockCallback)),
		))
}

func locationCalculator() tg.InlineKeyboardMarkup {

	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Ижевск", strconv.Itoa(izhLocationCalculatorCallback)),
			tg.NewInlineKeyboardButtonData("Питер", strconv.Itoa(spbLocationCalculatorCallback)),
			tg.NewInlineKeyboardButtonData("Другой город", strconv.Itoa(othLocationCalculatorCallback)),
		))
}
func orderTypeCalculator() tg.InlineKeyboardMarkup {

	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Экспресс", strconv.Itoa(orderTypeExpressCalculatorCallback)),
			tg.NewInlineKeyboardButtonData("Обычный", strconv.Itoa(orderTypeNormalCalculatorCallback)),
		))
}
