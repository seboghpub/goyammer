package internal

import (
	"github.com/mqu/go-notify"
)

func Notify(summary, message, icon string) {
	n := notify.NotificationNew(summary, message, icon)
	//n.SetUrgency(notify.NOTIFY_URGENCY_CRITICAL)
	n.Show()
}
