package notify

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"github.com/mel2oo/juice/transport/http/middleware/notify/templates"
)

func NewPanicHTMLEmail(method, host, uri, id string, msg interface{}, stack string) (subject string, body string, err error) {
	mailData := &struct {
		URL   string
		ID    string
		Msg   string
		Stack string
		Year  int
	}{
		URL:   fmt.Sprintf("%s %s%s", method, host, uri),
		ID:    id,
		Msg:   fmt.Sprintf("%+v", msg),
		Stack: stack,
		Year:  time.Now().Year(),
	}

	mailTplContent, err := getEmailHTMLContent(templates.PanicMail, mailData)
	return fmt.Sprintf("[系统异常]-%s", uri), mailTplContent, err
}

func getEmailHTMLContent(mailTpl string, mailData interface{}) (string, error) {
	tpl, err := template.New("email tpl").Parse(mailTpl)
	if err != nil {
		return "", err
	}
	buffer := new(bytes.Buffer)
	err = tpl.Execute(buffer, mailData)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}
