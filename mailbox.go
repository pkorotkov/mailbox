package mailbox

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/mail"
	"net/smtp"
	"path/filepath"
	"strings"
)

const boundary = "2a1ef074a9c58c3205dd2611e439701bc68318"

type Credentials struct {
	serverAddress string
	auth          smtp.Auth
}

func NewCredentials(serverAddress string) *Credentials {
	creds := new(Credentials)
	creds.serverAddress = serverAddress
	return creds
}

func (creds *Credentials) SetPLAINAuth(login, password string) {
	sa := strings.Split(creds.serverAddress, ":")
	creds.auth = smtp.PlainAuth("", login, password, sa[0])
}

type Message struct {
	from        mail.Address
	to          []mail.Address
	subject     string
	body        string
	attachments map[string][]byte
}

func (m *Message) From(name, address string) *Message {
	m.from = mail.Address{name, address}
	return m
}

func (m *Message) To(name, address string) *Message {
	m.to = append(m.to, mail.Address{name, address})
	return m
}

func (m *Message) Subject(subject string) *Message {
	m.subject = subject
	return m
}

func (m *Message) Body(body string) *Message {
	m.body = body
	return m
}

func (m *Message) Attach(filePath string) error {
	bs, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	_, fileName := filepath.Split(filePath)
	if m.attachments == nil {
		m.attachments = make(map[string][]byte, 3)
	}
	m.attachments[fileName] = bs
	return nil
}

func (m *Message) getToAddresses() []string {
	var addrs []string
	for _, toa := range m.to {
		addrs = append(addrs, toa.Address)
	}
	return addrs
}

func SendMessage(creds *Credentials, m *Message) error {
	switch {
	case creds == nil:
		return fmt.Errorf("Message not sent: Input credentials must not be nil.")
	case m == nil:
		return fmt.Errorf("Message not sent: Input message must not be nil.")
	}
	
	buf := new(bytes.Buffer)
	buf.WriteString("From: " + m.from.String() + "\n")
	buf.WriteString("To: " + m.to[0].String())
	for i := 1; i < len(m.to); i++ {
		buf.WriteString("," + m.to[i].String())
	}
	buf.WriteString("\nSubject: " + strings.Trim((&mail.Address{m.subject, ""}).String(), " <>") + "\n")
	buf.WriteString("MIME-Version: 1.0\n")

	if m.attachments != nil {
		buf.WriteString("Content-Type: multipart/mixed; boundary=" + boundary + "\n")
		buf.WriteString("--" + boundary + "\n")
	}

	buf.WriteString("Content-Type: text/plain; charset=\"utf-8\"\n")
	buf.WriteString("Content-Transfer-Encoding: base64\n")
	buf.WriteString("\n" + base64.StdEncoding.EncodeToString([]byte(m.body)))

	if m.attachments != nil {
		for fn, fbs := range m.attachments {
			buf.WriteString("\n\n--" + boundary + "\n")
			buf.WriteString("Content-Type: application/octet-stream\n")
			buf.WriteString("Content-Transfer-Encoding: base64\n")
			buf.WriteString("Content-Disposition: attachment; filename=\"" + fn + "\"\n\n")

			b := make([]byte, base64.StdEncoding.EncodedLen(len(fbs)))
			base64.StdEncoding.Encode(b, fbs)
			if _, err := buf.Write(b); err != nil {
				return fmt.Errorf("Could not attach %s: %s", fn, err.Error())
			}
			buf.WriteString("\n--" + boundary)
		}

		buf.WriteString("--")
	}

	return smtp.SendMail(creds.serverAddress, creds.auth, m.from.Address, m.getToAddresses(), buf.Bytes())
}
