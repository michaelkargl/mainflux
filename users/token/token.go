// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

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
	"github.com/mainflux/mainflux/users"
)

var (

	// ErrMalformedToken malformed token
	ErrMalformedToken = errors.New("Malformed token")
	// ErrExpiredToken  password reset token has expired
	ErrExpiredToken = errors.New("Token is expired")
	// ErrWrongSignature wrong signature
	ErrWrongSignature = errors.New("Wrong token signature")
)

const (
	defTokenSecret        = "mainflux-secret"
	defTokenDuration      = "5"
	defTokenLogLevel      = "debug"
	defTokenResetEndpoint = "/password/reset"

	envTokenSecret   = "MF_TOKEN_SECRET"
	envTokenDuration = "MF_TOKEN_DURATION"
	envTokenLogLevel = "MF_TOKEN_DEBUG_LEVEL"
)

var once sync.Once

type tokenizer struct {
	hmacSampleSecret []byte // secret for signing token
	tokenDuration    int    // token in duration in min
	logger           logger.Logger
}

var t *tokenizer

// Instance - Thread safe creation singleton instance of tokenizer.
// Used for creating password reset token.
func Instance() users.Tokenizer {
	once.Do(func() {
		t = &tokenizer{}
		t.hmacSampleSecret = []byte(mainflux.Env(envTokenSecret, defTokenSecret))
		t.tokenDuration, _ = strconv.Atoi(mainflux.Env(envTokenDuration, defTokenDuration))

		logLevel := mainflux.Env(envTokenLogLevel, defTokenLogLevel)
		l, err := logger.New(os.Stdout, logLevel)
		if err != nil {
			log.Fatalf(err.Error())
		}

		t.logger = l
	})
	return t
}

func (t *tokenizer) Generate(email string, offset int) (string, error) {
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
func (t *tokenizer) Verify(tok string) (string, error) {
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
