// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package http

import "github.com/mainflux/mainflux/users"

type apiReq interface {
	validate() error
}

type userReq struct {
	user users.User
}

func (req userReq) validate() error {
	return req.user.Validate()
}

type viewUserInfoReq struct {
	token string
}

func (req viewUserInfoReq) validate() error {
	if req.token == "" {
		return users.ErrUnauthorizedAccess
	}
	return nil
}

type passwResetReq struct {
	Email string `json:"email,omitempty"`
	Host  string `json:"host,omitempty"`
}

func (req passwResetReq) validate() error {
	if req.Email == "" {
		return users.ErrMissingEmail
	}
	if req.Host == "" {
		return users.ErrMalformedEntity
	}
	return nil
}

type resetTokenReq struct {
	Token    string `json:"token,omitempty"`
	Password string `json:"password,omitempty"`
	ConfPass string `json:"confirmPassword,omitempty"`
}

func (req resetTokenReq) validate() error {
	if req.Token == "" {
		return users.ErrMissingResetToken
	}
	if req.Password == "" {
		return users.ErrMalformedEntity
	}
	if req.ConfPass == "" {
		return users.ErrMalformedEntity
	}
	if req.Password != req.ConfPass {
		return users.ErrMalformedEntity
	}
	return nil
}

type passResReq struct {
	token string
}

func (req viewUserInfoReq) validate() error {
	if req.token == "" {
		return users.ErrUnauthorizedAccess
	}
	return nil
}

type resetTokenReq struct {
	token    string
	email    string
	password string
}

func (req resetTokenReq) validate() error {
	if req.token == "" {
		return users.ErrMisingResetToken
	}
	if req.email == "" {
		return users.ErrMissingEmail
	}
	if req.password == "" {
		return users.ErrMalformedEntity
	}
	return nil
}
