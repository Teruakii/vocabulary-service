package mail

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"mime"
	"net"
	"net/smtp"
	"strings"
	"time"
)

type Mailer interface {
	Send(ctx context.Context, to, subject, body string) error
}

// SMTP sends through a mail server, upgrading to STARTTLS when the server
// offers it, so port 587 works without extra setup.
type SMTP struct {
	addr     string
	host     string
	username string
	password string
	from     string
}

func NewSMTP(host, port, username, password, from string) *SMTP {
	return &SMTP{addr: net.JoinHostPort(host, port), host: host, username: username, password: password, from: from}
}

// sendTimeout bounds one whole send. smtp.SendMail has none, so a mail server
// that stops answering would hold the HTTP request open forever.
const sendTimeout = 15 * time.Second

func (m *SMTP) Send(ctx context.Context, to, subject, body string) error {
	if err := m.send(ctx, to, message(m.from, to, subject, body)); err != nil {
		return fmt.Errorf("send mail: %w", err)
	}
	return nil
}

func (m *SMTP) send(ctx context.Context, to string, msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, sendTimeout)
	defer cancel()
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", m.addr)
	if err != nil {
		return err
	}
	deadline, _ := ctx.Deadline()
	if err := conn.SetDeadline(deadline); err != nil {
		conn.Close()
		return err
	}
	c, err := smtp.NewClient(conn, m.host)
	if err != nil {
		conn.Close()
		return err
	}
	defer c.Close()

	if ok, _ := c.Extension("STARTTLS"); ok {
		if err := c.StartTLS(&tls.Config{ServerName: m.host}); err != nil {
			return err
		}
	}
	if m.username != "" {
		if err := c.Auth(smtp.PlainAuth("", m.username, m.password, m.host)); err != nil {
			return err
		}
	}
	if err := c.Mail(m.from); err != nil {
		return err
	}
	if err := c.Rcpt(to); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return c.Quit()
}

func message(from, to, subject, body string) []byte {
	var b strings.Builder
	header := func(k, v string) {
		// A CR or LF in a value would let it start a header of its own.
		v = strings.NewReplacer("\r", "", "\n", "").Replace(v)
		b.WriteString(k + ": " + v + "\r\n")
	}
	header("From", from)
	header("To", to)
	header("Subject", mime.QEncoding.Encode("utf-8", subject))
	header("Date", time.Now().Format(time.RFC1123Z))
	header("MIME-Version", "1.0")
	header("Content-Type", "text/plain; charset=UTF-8")
	b.WriteString("\r\n")
	b.WriteString(strings.ReplaceAll(strings.ReplaceAll(body, "\r\n", "\n"), "\n", "\r\n"))
	return []byte(b.String())
}

// Log prints messages instead of sending them. Development only: the output
// contains live password reset links.
type Log struct{}

func (Log) Send(_ context.Context, to, subject, body string) error {
	log.Printf("mail to %s: %s\n%s", to, subject, body)
	return nil
}
