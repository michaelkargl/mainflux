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
	Email string
	Host  string
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
	Token    string
	Email    string
	Password string
}

func (req resetTokenReq) validate() error {
	if req.Token == "" {
		return users.ErrMisingResetToken
	}
	if req.Email == "" {
		return users.ErrMissingEmail
	}
	if req.Password == "" {
		return users.ErrMalformedEntity
	}
	return nil
}
