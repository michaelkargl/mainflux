package token_test

import (
	"fmt"
	"testing"
)

var email = "johnsnow@gmai.com"

func TestGenerate(t *testing.T) {
	hash, err := recovery.Generate(email)
	if err != nil {
		t.Errorf("Token generation faild.")
	}

	fmt.Println("Here it is: ", hash)

}

func TestVerify(t *testing.T) {
	token, err := recovery.Generate(email)
	if err != nil {
		t.Errorf("Token generation faild.")
	}
	hashed, err := recovery.Hash(token)
	if err != nil {
		t.Errorf("Token Hashing faild.")
	}

	if err := recovery.Verify(token, hashed); err != nil {
		fmt.Println("Here is a error:", err)
		t.Errorf("Token verification faild.")
	}

}

func TestHash(t *testing.T) {

}

func TestSend(t *testing.T) {
	hash, err := recovery.Generate(email)
	if err != nil {
		t.Errorf("Token generation faild.")
	}
	if err := recovery.Send(email, hash); err != nil {
		t.Error("Faild to send email")
	}

}
