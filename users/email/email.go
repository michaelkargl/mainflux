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
	defMailLogLevel    = "debug"
	defMailDriver      = "smtp"
	defMailHost        = "localhost"
	defMailPort        = "25"
	defMailUsername    = "root"
	defMailPassword    = ""
	defMailFromAddress = ""
	defMailFromName    = ""

	envMailDriver      = "MF_EMAIL_DRIVER"
	envMailHost        = "MF_EMAIL_HOST"
	envMailPort        = "MF_EMAIL_PORT"
	envMailUsername    = "MF_EMAIL_USERNAME"
	envMailPassword    = "MF_EMAIL_PASSWORD"
	envMailFromAddress = "MF_EMAIL_FROM_ADDRESS"
	envMailFromName    = "MF_EMAIL_FROM_NAME"
	envMailLogLevel    = "MF_EMAIL_LOG_LEVEL"
)

type mail struct {
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
	conf mail
	auth smtp.Auth
	addr string
	log  logger.Logger
}

var a *Agent
var once sync.Once

// Agent - Thread safe creation of mail agent
func Instance() *Agent {
	once.Do(func() {
		a = &Agent{}

		a.conf = mail{
			driver:      mainflux.Env(envMailDriver, defMailDriver),
			fromAddress: mainflux.Env(envMailFromAddress, defMailFromAddress),
			fromName:    mainflux.Env(envMailFromName, defMailFromName),
			host:        mainflux.Env(envMailHost, defMailHost),
			port:        mainflux.Env(envMailPort, defMailPort),
			username:    mainflux.Env(envMailUsername, defMailUsername),
			password:    mainflux.Env(envMailPassword, defMailPassword),
		}

		// Set up authentication information
		a.auth = smtp.PlainAuth("", a.conf.username, a.conf.password, a.conf.host)
		a.addr = fmt.Sprintf("%s:%s", a.conf.host, a.conf.port)

		logLevel := mainflux.Env(envMailLogLevel, defMailLogLevel)
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
