// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package email

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"sync"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/logger"
)

const (
	defEmailLogLevel    = "debug"
	defEmailDriver      = "smtp"
	defEmailHost        = "localhost"
	defEmailPort        = "25"
	defEmailUsername    = "root"
	defEmailPassword    = ""
	defEmailFromAddress = ""
	defEmailFromName    = ""

	envEmailDriver      = "MF_EMAIL_DRIVER"
	envEmailHost        = "MF_EMAIL_HOST"
	envEmailPort        = "MF_EMAIL_PORT"
	envEmailUsername    = "MF_EMAIL_USERNAME"
	envEmailPassword    = "MF_EMAIL_PASSWORD"
	envEmailFromAddress = "MF_EMAIL_FROM_ADDRESS"
	envEmailFromName    = "MF_EMAIL_FROM_NAME"
	envEmailLogLevel    = "MF_EMAIL_LOG_LEVEL"
)

type email struct {
	driver      string
	host        string
	port        string
	username    string
	password    string
	fromAddress string
	fromName    string
}

// Agent for mailing
type Agent struct {
	conf email
	auth smtp.Auth
	addr string
	log  logger.Logger
}

var a *Agent
var once sync.Once

// Instance - Thread safe creation of mail agent
func Instance() *Agent {
	once.Do(func() {
		a = &Agent{}

		a.conf = email{
			driver:      mainflux.Env(envEmailDriver, defEmailDriver),
			fromAddress: mainflux.Env(envEmailFromAddress, defEmailFromAddress),
			fromName:    mainflux.Env(envEmailFromName, defEmailFromName),
			host:        mainflux.Env(envEmailHost, defEmailHost),
			port:        mainflux.Env(envEmailPort, defEmailPort),
			username:    mainflux.Env(envEmailUsername, defEmailUsername),
			password:    mainflux.Env(envEmailPassword, defEmailPassword),
		}

		// Set up authentication information
		a.auth = smtp.PlainAuth("", a.conf.username, a.conf.password, a.conf.host)
		a.addr = fmt.Sprintf("%s:%s", a.conf.host, a.conf.port)

		logLevel := mainflux.Env(envEmailLogLevel, defEmailLogLevel)
		logger, err := logger.New(os.Stdout, logLevel)
		if err != nil {
			log.Fatalf(err.Error())
		}

		a.log = logger
	})
	return a
}

// Send sends e-mail
func (a *Agent) Send(to []string, msg []byte) {
	go func() {
		err := smtp.SendMail(a.addr, a.auth, a.conf.fromAddress, to, msg)
		if err != nil {
			a.log.Error(fmt.Sprintf("Failed to send email:%s", err.Error()))
		}
	}()
}
