package mail

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
	defMailLogLevel = "debug"
	defMailDriver   = "smtp"
	// defMailHost        = "localhost"
	// defMailPort        = "25"
	// defMailUsername    = "root"
	// defMailPassword    = ""
	// defMailFromAddress = ""
	// defMailFromName    = ""

	defMailHost        = "smtp.mailtrap.io"
	defMailPort        = "2525"
	defMailUsername    = "18bf7f70705139"
	defMailPassword    = "2b0d302e775b1e"
	defMailFromAddress = "from@example.com"
	defMailFromName    = "Example"

	envMailDriver      = "MF_MAIL_DRIVER"
	envMailHost        = "MF_MAIL_HOST"
	envMailPort        = "MF_MAIL_PORT"
	envMailUsername    = "MF_MAIL_USERNAME"
	envMailPassword    = "MF_MAIL_PASSWORD"
	envMailFromAddress = "MF_MAIL_FROM_ADDRESS"
	envMailFromName    = "MF_MAIL_FROM_NAME"
	envMailLogLevel    = "MF_MAIL_LOG_LEVEL"
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
type agent struct {
	conf mail
	auth smtp.Auth
	addr string
	log  logger.Logger
}

var a *agent
var once sync.Once

// Agent - Thread safe creation of mail agent
func instance() *agent {
	once.Do(func() {
		a = &agent{}

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
func Send(to []string, msg []byte) {
	go func() {
		a := instance()
		err := smtp.SendMail(a.addr, a.auth, a.conf.fromAddress, to, msg)
		if err != nil {
			a.log.Error(fmt.Sprintf("Failed to send mail:%s", err.Error()))
		}
	}()
}
