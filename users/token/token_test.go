// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package token_test

import (
	"fmt"
	"testing"

	"github.com/mainflux/mainflux/users/token"
)

var email = "johnsnow@gmai.com"

func TestGenerate(t *testing.T) {
	hash, err := token.Instance().Generate(email, 0)
	if err != nil {
		t.Errorf("Token generation failed.")
	}
	fmt.Println("Here it is: ", hash)
}

func TestVerify(t *testing.T) {
	tok, err := token.Instance().Generate(email, 0)
	if err != nil {
		t.Errorf("Token generation failed.")
	}

	e, err := token.Instance().Verify(tok)
	if err != nil {
		fmt.Println("Here is a error:", err)
		t.Errorf("Token verification failed.")
	}
	if e != email {
		t.Errorf("Token verification failed.")
	}

}
