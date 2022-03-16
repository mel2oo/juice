package notify

import (
	"github.com/mel2oo/juice/pkg/mail"
	"github.com/mel2oo/juice/transport/http"
	"go.uber.org/zap"
)

func OnPanicNotify(ctx http.Context, options *mail.Options, err interface{}, stackInfo string) {

	subject, body, htmlErr := NewPanicHTMLEmail(
		ctx.Method(),
		ctx.Host(),
		ctx.URI(),
		ctx.Trace().ID(),
		err,
		stackInfo,
	)
	if htmlErr != nil {
		ctx.Logger().Error("NewPanicHTMLEmail error", zap.Error(htmlErr))
		return
	}

	options.Subject = subject
	options.Body = body

	sendErr := mail.Send(options)
	if sendErr != nil {
		ctx.Logger().Error("Mail Send error", zap.Error(sendErr))
	}

	return
}
