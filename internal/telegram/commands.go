package telegram

import "strconv"

const (
	create = "`/create <categoryName e.g. escaperoom>`"
	add    = "`/add <categoryName e.g. escaperoom> - <itemName e.g. zed>`"
	titles = "/titles"
	list   = "/list"

	requestRateKeyboardCallback = "request_rate_keyboard:"
	rateCallback                = "rate:"
)

func rateCallbackFmt(score string, itemID int) string {
	return rateCallback + score + ":" + strconv.Itoa(itemID)
}
