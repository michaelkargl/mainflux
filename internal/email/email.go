// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package email

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/smtp"

	"github.com/mainflux/mainflux/logger"
)

var (
	// ErrMissingEmailTmpl missing email template file
	ErrMissingEmailTmpl = errors.New("Missing email template file")
)

type emailTemplate struct {
	To      []string
	From    string
	Header  string
	Content string
	Footer  string
}

// Config email agent configuration.
type Config struct {
	Driver      string
	Host        string
	Port        string
	Username    string
	Password    string
	FromAddress string
	FromName    string
}

// Agent for mailing
type Agent struct {
	conf *Config
	auth smtp.Auth
	addr string
	log  logger.Logger
	tmpl *template.Template
}

// New creates new email agent
func New(c *Config, t *template.Template) *Agent {
	a := &Agent{}
	a.conf = c

	// Set up authentication information
	a.auth = smtp.PlainAuth("", c.Username, c.Password, c.Host)
	a.addr = fmt.Sprintf("%s:%s", c.Host, c.Port)
	if t != nil {
		a.tmpl = t
		return a
	}

	tmpl, _ := template.ParseFiles("email.tmpl")
	a.tmpl = tmpl
	return a
}

// Init initializes mailing agent
func (a *Agent) SetTemplate(t *template.Template) {
	a.tmpl = t
}

// Send sends e-mail
// From
// To
// Header
// Content
// Footer
func (a *Agent) Send(To []string, From, Header, Content, Footer string) error {
	if a.tmpl == nil {
		return ErrMissingEmailTmpl
	}
	email := new(bytes.Buffer)
	tmpl := emailTemplate{
		To:      To,
		From:    From,
		Header:  Header,
		Content: Content,
		Footer:  Footer,
	}

	err := a.tmpl.Execute(email, tmpl)
	if err != nil {
		return err
	}
	err = smtp.SendMail(a.addr, a.auth, a.conf.FromAddress, To, email.Bytes())
	return err
}