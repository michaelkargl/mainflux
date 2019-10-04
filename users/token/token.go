// Package token provides password recovery token generation with jwt
// Token is sent by email to user as part of recovery URL
// Token is signed by secret signature
package token

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/users/mail"
)

var (
<<<<<<< HEAD

	// ErrMalformedToken malformed token
	ErrMalformedToken = errors.New("Malformed token")
	// ErrExpiredToken  password reset token has expired
	ErrExpiredToken = errors.New("Token is expired")
	// ErrWrongSignature wrong signature
	ErrWrongSignature = errors.New("Wrong token signature")
=======
	emailLength   = 254
	ttlLength     = 4
	secretLength  = 20
	tokenLength   = ttlLength + emailLength // Max token length is 4 bytes + max email length
	hashCost      = 10
	tokenDuration = 5                      // Recovery token TTL in minutes, reperesents token time to live
	secret        = "fcERNb7KpM3WyAmguJMZ" // Random string for secret key, required for signing

	// Errors
	errMalformedToken  = errors.New("Malformed token")
	errExpiredToken    = errors.New("Token expired")
	errWrongSignature  = errors.New("Wrong token signature")
	errTokenGeneration = errors.New("Token generation failed")
>>>>>>> addint test and update swagger for pass reset
)

const (
	defTokenSecret        = "mainflux-secret"
	defTokenDuration      = "5"
	defTokenLogLevel      = "debug"
	defTokenResetEndpoint = "/password/reset"

	envTokenSecret        = "MF_TOKEN_SECRET"
	envTokenDuration      = "MF_TOKEN_DURATION"
	envTokenLogLevel      = "MF_TOKEN_DEBUG_LEVEL"
	envTokenResetEndpoint = "MF_TOKEN_RESET_ENDPOINT"
)

var once sync.Once

type tokenizer struct {
	hmacSampleSecret []byte // secret for signing token
	tokenDuration    int    //token in duration in min
	logger           logger.Logger
	url              string
}

var t *tokenizer

// Agent - Thread safe creation of mail agent
func instance() *tokenizer {
	once.Do(func() {
		t = &tokenizer{}
		t.hmacSampleSecret = []byte(mainflux.Env(envTokenSecret, defTokenSecret))
		t.tokenDuration, _ = strconv.Atoi(mainflux.Env(envTokenDuration, defTokenDuration))
		t.url = mainflux.Env(envTokenResetEndpoint, defTokenResetEndpoint)
		logLevel := mainflux.Env(envTokenLogLevel, defTokenLogLevel)
		l, err := logger.New(os.Stdout, logLevel)
		if err != nil {
			log.Fatalf(err.Error())
		}

		t.logger = l
	})
	return t
}

// Generate generate new random token with defined TTL.
// offset can be used to manipulate token validity in time
// useful for testing.
func Generate(email string, offset int) (string, error) {
	return instance().generate(email, offset)
}

// Verify verifies token validity
func Verify(tok string) (string, error) {
	return instance().verify(tok)
}

// SendToken sends password recovery link to user
func SendToken(host, email, token string) {
	instance().sendToken(host, email, token)
}

func (t *tokenizer) generate(email string, offset int) (string, error) {

	exp := t.tokenDuration + offset
	if exp < 0 {
		exp = 0
	}
	expires := time.Now().Add(time.Minute * time.Duration(exp))
	nbf := time.Now()

	// Create a new token object, specifying signing method and the claims
	// you would like it to contain
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"exp":   expires.Unix(),
		"nbf":   nbf.Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(t.hmacSampleSecret)
	fmt.Println(tokenString)
	return tokenString, err
}

// Verify verifies token validity
func (t *tokenizer) verify(tok string) (string, error) {
	email := ""
	token, err := jwt.Parse(tok, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			t.logger.Error(ErrWrongSignature.Error())
			return nil, ErrWrongSignature
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return t.hmacSampleSecret, nil
	})
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if claims.VerifyExpiresAt(time.Now().Unix(), false) == false {
			t.logger.Error(ErrExpiredToken.Error())
			return "", ErrExpiredToken
		}
		email = claims["email"].(string)

	} else {
		return email, err
	}
	return email, nil
}

// SendToken sends password recovery link to user
func (t *tokenizer) sendToken(host, email, token string) {
	body := t.buildBody(host, email, token)
	mail.Send([]string{email}, body)
}

// Builds recovery email body
func (t *tokenizer) buildBody(host, email, token string) []byte {
	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: Reset Password!\r\n"+
		"\r\n"+
		"You have initiated password reset.\r\n"+
		"Follow the link below to reset password.\r\n"+
		"%s%s?token=%s", email, host, t.url, token))

	return msg
}
