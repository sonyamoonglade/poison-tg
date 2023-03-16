package telegram

import (
	"encoding/binary"
	"fmt"
	"strconv"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	menuCatalogCallback = iota + 1
	menuMakeOrderCallback
	menuGetBucketCallback
	menuTrackOrderCallback
	orderGuideStep1Callback
	orderGuideStep2Callback
	orderGuideStep3Callback
	orderGuideStep4Callback
	makeOrderCallback
)

var (
	menuButtons                  = menu()
	orderGuideStep1Buttons       = orderGuideStep(orderGuideStep1Callback)
	orderGuideStep2Buttons       = orderGuideStep(orderGuideStep2Callback)
	orderGuideStep3Buttons       = orderGuideStep(orderGuideStep3Callback)
	orderGuideStep4Buttons       = orderGuideStep(orderGuideStep4Callback)
	orderGuideToMakeOrderButtons = orderGuideToMakeOrder()
)

// first 10 bits is msgID
// next 10 bits is callback
func injectMessageID(msgID int64, callback int) string {
	payload := make([]byte, 64)
	msgIdBuf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(msgIdBuf, msgID)
	msgIdBuf = msgIdBuf[:n]
	cbBuf := make([]byte, binary.MaxVarintLen64)
	n = binary.PutVarint(msgIdBuf, msgID)
	cbBuf = cbBuf[:n]

	return string(payload)
}

func ExtractMsgID(data string) (msgID int64, callback int, err error) {
	payload := []byte(data)
	msgIDstr := payload[56:]
	callbackStr := payload[:9]
	fmt.Printf("before: %v %v\n", msgIDstr, callbackStr)
	// find end of msgIDstr
	for i, b := range msgIDstr {
		if b == 0 {
			msgIDstr = msgIDstr[:i]
			break
		}
	}
	// find end of callback
	for i, b := range callbackStr {
		fmt.Printf("ite: %d %v %v\n", i, b, b == '0')
		if b == 0 {
			callbackStr = callbackStr[:i]
			break
		}
	}
	fmt.Printf("recv: msg: %v cb: %v\n", msgIDstr, callbackStr)
	fmt.Printf("recv: msg: %v cb: %v\n", string(msgIDstr), string(callbackStr))
	msgIDint, err := strconv.ParseInt(string(msgIDstr), 2, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("can't parse payload: %w", err)
	}

	callbackInt, err := strconv.ParseInt(string(callbackStr), 2, 32)
	if err != nil {
		return 0, 0, fmt.Errorf("can't parse payload: %w", err)
	}
	fmt.Printf("out: %d %d\n", msgIDint, callbackInt)
	return msgIDint, int(callbackInt), nil
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

func orderGuideStep(step int) tg.InlineKeyboardMarkup {
	if step == orderGuideStep4Callback {
		return tg.NewInlineKeyboardMarkup(
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData("<-", strconv.Itoa(step-1)),
			),
		)
	} else if step == orderGuideStep1Callback {
		return tg.NewInlineKeyboardMarkup(
			tg.NewInlineKeyboardRow(
				tg.NewInlineKeyboardButtonData("->", strconv.Itoa(step+1)),
			),
		)
	}

	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("<-", strconv.Itoa(step-1)),
			tg.NewInlineKeyboardButtonData("->", strconv.Itoa(step+1)),
		),
	)
}

func orderGuideToMakeOrder() tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Pereity k oformleniy zakaza", strconv.Itoa(makeOrderCallback))),
	)
}
