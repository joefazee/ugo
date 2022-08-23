package mailer

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/ainsleyclark/go-mail/drivers"
	apimail "github.com/ainsleyclark/go-mail/mail"
	"github.com/vanng822/go-premailer/premailer"
	"github.com/xhit/go-simple-mail/v2"
	"html/template"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"
)

type Mail struct {
	Domain      string
	Templates   string
	Host        string
	Port        int
	Username    string
	Password    string
	Encryption  string
	FromAddress string
	FromName    string
	Jobs        chan Message // channel of jobs to send
	Result      chan Result  // channel of results from sending jobs
	API         string
	APIKey      string
	APIURL      string
}

type Message struct {
	From        string
	FromName    string
	To          string
	Subject     string
	Template    string
	Attachments []string
	Data        interface{}
}

type Result struct {
	Success bool
	Error   error
}

func (m *Mail) ListenForMail() {
	log.Println("Mail service started running")
	for {
		msg := <-m.Jobs
		err := m.Send(msg)
		if err != nil {
			m.Result <- Result{Success: false, Error: err}
		} else {
			m.Result <- Result{Success: true, Error: nil}
		}
	}
}

func (m *Mail) Send(msg Message) error {

	if len(m.API) > 0 && len(m.APIKey) > 0 && len(m.APIURL) > 0 && m.API != "smtp" {
		// send via API
		return m.ChooseAPI(msg)
	}

	return m.SendSMTPMessage(msg)

}

func (m *Mail) SendSMTPMessage(msg Message) error {

	formattedMessage, err := m.buildHTMLMessage(msg)
	if err != nil {
		return err
	}

	plainMessage, err := m.buildPlainTextMessage(msg)
	if err != nil {
		return err
	}

	server := mail.NewSMTPClient()
	server.Host = m.Host
	server.Port = m.Port
	server.Username = m.Username
	server.Password = m.Password
	server.Encryption = m.getEncryption(m.Encryption)
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	smtpClient, err := server.Connect()
	if err != nil {
		return err
	}

	email := mail.NewMSG()
	email.SetFrom(msg.From).
		AddTo(msg.To).
		SetSubject(msg.Subject)

	email.SetBody(mail.TextHTML, formattedMessage)
	email.AddAlternative(mail.TextPlain, plainMessage)

	if len(msg.Attachments) > 0 {
		for _, attachment := range msg.Attachments {
			email.AddAttachment(attachment)
		}
	}

	return email.Send(smtpClient)
}

func (m *Mail) buildHTMLMessage(msg Message) (string, error) {
	templateToRender := fmt.Sprintf("%s/%s.html.tmpl", m.Templates, msg.Template)

	t, err := template.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	err = t.ExecuteTemplate(&tpl, "body", msg.Data)
	if err != nil {
		return "", err
	}

	formattedEmail := tpl.String()
	formattedEmail, err = m.inlineCSS(formattedEmail)

	if err != nil {
		return "", err
	}

	return formattedEmail, nil
}

func (m *Mail) buildPlainTextMessage(msg Message) (string, error) {

	templateToRender := fmt.Sprintf("%s/%s.plain.txt", m.Templates, msg.Template)

	t, err := template.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	err = t.ExecuteTemplate(&tpl, "body", msg.Data)
	if err != nil {
		return "", err
	}

	plainMessage := tpl.String()

	return plainMessage, nil
}

func (m *Mail) getEncryption(encryption string) mail.Encryption {

	switch encryption {
	case "tls":
		return mail.EncryptionSTARTTLS
	case "ssl":
		return mail.EncryptionSSL
	case "none":
		return mail.EncryptionNone
	default:
		return mail.EncryptionSTARTTLS
	}

}

func (m *Mail) inlineCSS(s string) (string, error) {

	options := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}

	prem, err := premailer.NewPremailerFromString(s, &options)
	if err != nil {
		return "", err
	}

	html, err := prem.Transform()
	if err != nil {
		return "", err
	}

	return html, nil

}

func (m *Mail) ChooseAPI(msg Message) error {
	switch m.API {
	case "mailgun", "sparkpost", "sendgrid":
		return m.SendUsingAPI(msg, m.API)
	default:
		return fmt.Errorf("API not supported %s", m.API)
	}
}

func (m *Mail) SendUsingAPI(msg Message, api string) error {
	if msg.From == "" {
		msg.From = m.FromAddress
	}

	if msg.FromName == "" {
		msg.FromName = m.FromName
	}

	cfg := apimail.Config{
		URL:         m.APIURL,
		APIKey:      m.APIKey,
		Domain:      m.Domain,
		FromName:    msg.FromName,
		FromAddress: msg.From,
	}

	var driver apimail.Mailer
	var err error

	switch api {
	case "mailgun":
		driver, err = drivers.NewMailgun(cfg)
		if err != nil {
			return err
		}

	case "sparkpost":
		driver, err = drivers.NewSparkPost(cfg)
		if err != nil {
			return err
		}
	case "sendgrid":
		driver, err = drivers.NewSendGrid(cfg)
		if err != nil {
			return err
		}
	default:
		return errors.New("invalid api")
	}

	formattedMessage, err := m.buildHTMLMessage(msg)
	if err != nil {
		return err
	}

	plainMessage, err := m.buildPlainTextMessage(msg)
	if err != nil {
		return err
	}

	tx := &apimail.Transmission{
		Recipients: []string{msg.To},
		Subject:    msg.Subject,
		HTML:       formattedMessage,
		PlainText:  plainMessage,
	}

	err = m.addAPIAttachments(msg, tx)
	if err != nil {
		return err
	}

	_, err = driver.Send(tx)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mail) addAPIAttachments(msg Message, tx *apimail.Transmission) error {
	if len(msg.Attachments) > 0 {

		var attachments []apimail.Attachment

		for _, at := range msg.Attachments {
			var attach apimail.Attachment
			content, err := ioutil.ReadFile(at)
			if err != nil {
				return err
			}

			attach.Bytes = content
			attach.Filename = filepath.Base(at)
			attachments = append(attachments, attach)
		}

		tx.Attachments = attachments
	}

	return nil
}
