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
	defMailLogLevel    = "debug"
	defMailDriver      = "smtp"
	defMailHost        = "localhost"
	defMailPort        = "25"
	defMailUsername    = "root"
	defMailPassword    = ""
	defMailFromAddress = ""
	defMailFromName    = ""

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

var instance *agent
var once sync.Once

// Agent - Thread safe creation of mail agent
func initAgent() *agent {
	once.Do(func() {
		instance = &agent{}

		instance.conf = mail{
			driver:      mainflux.Env(envMailDriver, defMailDriver),
			fromAddress: mainflux.Env(envMailFromAddress, defMailFromAddress),
			fromName:    mainflux.Env(envMailFromName, defMailFromName),
			host:        mainflux.Env(envMailHost, defMailHost),
			port:        mainflux.Env(envMailPort, defMailPort),
			username:    mainflux.Env(envMailUsername, defMailUsername),
			password:    mainflux.Env(envMailPassword, defMailPassword),
		}

		// Set up authentication information.
		instance.auth = smtp.PlainAuth("", instance.conf.username, instance.conf.password, instance.conf.host)
		instance.addr = fmt.Sprintf("%s:%s", instance.conf.host, instance.conf.port)

		logLevel := mainflux.Env(envMailLogLevel, defMailLogLevel)
		logger, err := logger.New(os.Stdout, logLevel)
		if err != nil {
			log.Fatalf(err.Error())
		}

		instance.log = logger
	})
	return instance
}

// Send sends mail.
func Send(to []string, msg []byte) {
	go func() {
		a := initAgent()
		err := smtp.SendMail(a.addr, a.auth, a.conf.fromAddress, to, msg)
		if err != nil {
			a.log.Error(fmt.Sprintf("Failed to send mail:%s", err.Error()))
		}
	}()
}
