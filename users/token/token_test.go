package token_test

import (
	"fmt"
	"testing"

	"github.com/mainflux/mainflux/users/token"
)

var email = "johnsnow@gmai.com"

func TestGenerate(t *testing.T) {
	hash, err := token.Generate(email)
	if err != nil {
		t.Errorf("Token generation failed.")
	}
	fmt.Println("Here it is: ", hash)
}

func TestVerify(t *testing.T) {
	tok, err := token.Generate(email)
	if err != nil {
		t.Errorf("Token generation faild.")
	}
	hashed, err := token.Hash(tok)
	if err != nil {
		t.Errorf("Token Hashing faild.")
	}

	if err := token.Verify(tok, hashed); err != nil {
		fmt.Println("Here is a error:", err)
		t.Errorf("Token verification faild.")
	}

}

func TestHash(t *testing.T) {

}

func TestSend(t *testing.T) {
	hash, err := token.Generate(email)
	if err != nil {
		t.Errorf("Token generation failed.")
	}
	token.SendToken("http://localhost", email, hash)
}
