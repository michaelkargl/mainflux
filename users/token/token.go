//Package token provides password recovery token generation with TTL
//Token is sent by email to user as part of recovery URL
//Token contains 32 bytes where first 4 bytes are exparation time and 28 bytes random string
// exparation-email signed by secret signature
package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/mainflux/mainflux/users/mail"
)

var (
	emailLength      = 254
	ttlLength        = 4
	secretLength     = 20
	tokenLength      = ttlLength + emailLength // Max token length is 4 bytes + max email length
	hashCost         = 10
	tokenDuration    = 5                              // Recovery token TTL in minutes, reperesents token time to live
	hmacSampleSecret = []byte("fcERNb7KpM3WyAmguJMZ") // Random string for secret key, required for signing

	// ErrMalformedToken malformed token
	ErrMalformedToken = errors.New("Malformed token")
	// ErrExpiredToken  password reset token has expired
	ErrExpiredToken = errors.New("Token expired")
	// ErrWrongSignature wrong signature
	ErrWrongSignature = errors.New("Wrong token signature")
)

// Generate generate new random token with defined TTL.
// offset can be used to manipulate token validity in time
// useful for testing.
func Generate(email string, offset int) (string, error) {

	exp := tokenDuration + offset
	if exp < 0 {
		exp = 0
	}
	expires := time.Now().Add(time.Minute * time.Duration(exp))

	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"nbf":   expires.Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(hmacSampleSecret)

	return tokenString, err
}

// Verify verifies token validity
func Verify(email, tok string, hashed string) error {

	token, err := jwt.Parse(tok, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrWrongSignature
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return hmacSampleSecret, nil
	})
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		fmt.Println(claims["email"], claims["nbf"])
		if claims.VerifyNotBefore(time.Now().Unix(), false) == false {
			return ErrExpiredToken
		}
		if email != claims["email"] {
			return ErrMalformedToken
		}
	} else {
		fmt.Println(err)
	}
	return nil
}

// SendToken sends password recovery link to user
func SendToken(host, email, token string) {
	body := buildBody(host, email, token)
	mail.Send([]string{email}, body)
}

// Builds recovery email body
func buildBody(host, email, token string) []byte {
	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: Reset Password!\r\n"+
		"\r\n"+
		"You have initiated password reset.\r\n"+
		"Follow the link below to reset password.\r\n"+
		"%s/passwd/reset?token=%s", email, host, token))

	return msg
}
