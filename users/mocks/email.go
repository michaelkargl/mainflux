// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/mainflux/mainflux/users"
	"github.com/mainflux/mainflux/users/email"
)

// NewEmailer provides emailer instance for  the test
func NewEmailer() users.Emailer {
	return users.Emailer{ResetURL: "password/reset", Agent: email.Instance()}
}
