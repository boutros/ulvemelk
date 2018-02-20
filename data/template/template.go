package template

import (
	"context"

	"github.com/boutros/ulvemelk/data/locale"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func init() {
	for k, v := range locale.English {
		message.SetString(language.English, k, v)
	}
	for k, v := range locale.Norwegian {
		message.SetString(language.Norwegian, k, v)
	}

}

type contextKey int

const (
	MessagePrinterKey contextKey = iota + 1
)

func getPrinter(ctx context.Context) *message.Printer {
	return ctx.Value(MessagePrinterKey).(*message.Printer)
}
