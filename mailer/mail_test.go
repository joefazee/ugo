package mailer

import (
	"errors"
	"testing"
)

func getDemoMessage() Message {
	return Message{
		From:        "aj@test.com",
		FromName:    "aj",
		To:          "aj@demo.com",
		Template:    "test",
		Subject:     "test",
		Attachments: []string{"./testdata/mail/test.html.tmpl"},
	}
}

func TestMail_SendSMTPMessage(t *testing.T) {

	msg := getDemoMessage()

	err := mailer.SendSMTPMessage(msg)
	if err != nil {
		t.Error(err)
	}
}

func TestMail_SendSMTPMessageUsingChan(t *testing.T) {

	msg := getDemoMessage()
	mailer.Jobs <- msg

	res := <-mailer.Result
	if res.Error != nil {
		t.Error(errors.New("failed to send over channel"))
	}

	msg.To = "invalid_email"
	mailer.Jobs <- msg
	res = <-mailer.Result
	if res.Error == nil {
		t.Error(errors.New("we expect to get an error for invalid TO email"))
	}
}

func TestMail_SendUsingAPI(t *testing.T) {

	msg := getDemoMessage()
	mailer.API = "unknown"
	mailer.APIKey = "abc123"
	mailer.APIURL = "https://www.fakeapi.com"

	err := mailer.SendUsingAPI(msg, "unknown")
	if err == nil {
		t.Error("expect an error for unknown api")
	}

	mailer.API = ""
	mailer.APIKey = ""
	mailer.APIURL = ""

}

func TestMail_BuildHTMLMessage(t *testing.T) {

	msg := getDemoMessage()

	_, err := mailer.buildHTMLMessage(msg)
	if err != nil {
		t.Error(err)
	}
}

func TestMail_BuildPlainMessage(t *testing.T) {

	msg := getDemoMessage()

	_, err := mailer.buildPlainTextMessage(msg)
	if err != nil {
		t.Error(err)
	}
}

func TestMail_Send(t *testing.T) {

	msg := getDemoMessage()

	err := mailer.Send(msg)
	if err != nil {
		t.Error(err)
	}

	mailer.API = "unknown"
	mailer.APIKey = "abc123"
	mailer.APIURL = "https://www.fakeapi.com"

	err = mailer.Send(msg)
	if err == nil {
		t.Error("we expect an error for invalid credentials")
	}

	mailer.API = ""
	mailer.APIKey = ""
	mailer.APIURL = ""
}

func TestMail_ChooseAPI(t *testing.T) {

	msg := getDemoMessage()
	mailer.API = "unknown"

	err := mailer.ChooseAPI(msg)
	if err == nil {
		t.Error("we should get an error for invalid api")
	}
}
