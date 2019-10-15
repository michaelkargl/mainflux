// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0
package emailer

import (
	"fmt"

	"github.com/mainflux/mainflux/email"
	"github.com/mainflux/mainflux/users"
)

const (
	message = `You have initiated password reset.
			   Follow the link below to reset password.`
)

var _ users.Emailer = (*bcryptHasher)(nil)

type emailer struct {
	resetURL string
	agent    *email.Agent
}

// New creates new emailer utility
func New() Emailer {
	return &emailer{ResetURL: url, email.New()}
}

func (e *emailer) SendPasswordReset(To []string, host string, string token) error {
	url = fmt.Sprintf("%s/%s?token=%s", host, e.resetURL, token)
	content = fmt.Sprintf("%s\r\n%s\r\n", message, url)
	e.agent.Send(To, "", "", content, "")
}
