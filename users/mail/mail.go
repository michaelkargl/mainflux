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
	Driver      string
	Host        string
	Port        string
	Username    string
	Password    string
	FromAddress string
	FromName    string
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
			Driver:      mainflux.Env(envMailDriver, defMailDriver),
			FromAddress: mainflux.Env(envMailFromAddress, defMailFromAddress),
			FromName:    mainflux.Env(envMailFromName, defMailFromName),
			Host:        mainflux.Env(envMailHost, defMailHost),
			Port:        mainflux.Env(envMailPort, defMailPort),
			Username:    mainflux.Env(envMailUsername, defMailUsername),
			Password:    mainflux.Env(envMailPassword, defMailPassword),
		}

		// Set up authentication information.
		instance.auth = smtp.PlainAuth("", instance.conf.Username, instance.conf.Password, instance.conf.Host)
		instance.addr = fmt.Sprintf("%s:%s", instance.conf.Host, instance.conf.Port)

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
		err := smtp.SendMail(a.addr, a.auth, a.conf.FromAddress, to, msg)
		if err != nil {
			a.log.Error(fmt.Sprintf("Failed to send mail:%s", err.Error()))
		}
	}()
}
