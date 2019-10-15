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
	// ErrMissingTemplate missing email template file
	ErrMissingTemplate = errors.New("Missing email template file")
)

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
func New(c *Config) *Agent {
	a := &Agent{}
	a.conf = c

	// Set up authentication information
	a.auth = smtp.PlainAuth("", c.Username, c.Password, c.Host)
	a.addr = fmt.Sprintf("%s:%s", c.Host, c.Port)
	return a
}

// Init initializes mailing agent
func (a *Agent) Init() error {
	tmpl, err := template.ParseFiles("email.tmpl")
	if err != nil {
		return err
	}
	a.tmpl = tmpl
	return nil
}

// Send sends e-mail
// From
// To
// Header
// Content
// Footer
func (a *Agent) Send(To []string, From, Header, Content, Footer string) error {
	if a.tmpl == nil {
		return ErrMissingTemplate
	}
	email := new(bytes.Buffer)
	tmpl := struct {
		to      []string
		from    string
		header  string
		content string
		footer  string
	}{
		to:      To,
		from:    From,
		header:  Header,
		content: Content,
		footer:  Footer,
	}

	err := a.tmpl.Execute(email, tmpl)
	if err != nil {
		return err
	}
	return smtp.SendMail(a.addr, a.auth, a.conf.FromAddress, To, email.Bytes())
}
