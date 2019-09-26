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
	// defMailLogLevel    = "debug"
	// defMailDriver      = "smtp"
	// defMailHost        = "localhost"
	// defMailPort        = "25"
	// defMailUsername    = "root"
	// defMailPassword    = ""
	// defMailFromAddress = ""
	// defMailFromName    = ""

	// 	MF_USERS_MAIL_DRIVER=smtp
	// MF_USERS_MAIL_HOST=smtp.mailtrap.io
	// MF_USERS_MAIL_PORT=2525
	// MF_USERS_MAIL_USERNAME=18bf7f70705139
	// MF_USERS_MAIL_PASSWORD=2b0d302e775b1e
	// MF_USERS_MAIL_FROM_ADDRESS=from@example.com
	// MF_USERS_MAIL_FROM_NAME=Example

	defMailLogLevel    = "debug"
	defMailDriver      = "smtp"
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
			Password:    mainflux.Env(envMailPassword, envMailPassword),
		}

		logLevel := mainflux.Env(envMailLogLevel, defMailLogLevel)
		logger, err := logger.New(os.Stdout, logLevel)
		if err != nil {
			log.Fatalf(err.Error())
		}

		instance.log = logger
	})
	return instance
}

// Send sends mail
func Send(to []string, msg []byte) {
	// Set up authentication information.
	a := initAgent()
	auth := smtp.PlainAuth("", a.conf.Username, a.conf.Password, a.conf.Host)
	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	addr := fmt.Sprintf("%s:%s", a.conf.Host, a.conf.Port)
	go func() {
		err := smtp.SendMail(addr, auth, a.conf.FromAddress, to, msg)
		if err != nil {
			log.Fatal(err)
		}
	}()

}
