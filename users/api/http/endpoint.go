// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/mainflux/mainflux/users"
)

func registrationEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(userReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		err := svc.Register(ctx, req.user)
		return tokenRes{}, err
	}
}

// Password reset request endpoint serves post request with email of the user
// for whom password reset flow is to be initiated.
// If request is successful email with reset link will be sent to the
// email specified in the request where link is configured using MF_TOKEN_RESET_ENDPOINT.
// Link generate contains token that needs to be verified.
func passwordResetRequestEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(passwResetReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		res := passwChangeRes{}
		email := req.Email

		if err := svc.GenerateResetToken(ctx, email, req.Host); err != nil {
			res.Msg = err.Error()
			return res, nil
		}
		res.Msg = MailSent
		return res, nil
	}
}

// This is post request endpoint that actually sets new password.
// When user clicks on a link in the email he lands on UI page ( configured with MF_TOKEN_RESET_ENDPOINT )
// UI should have form that accepts new password and confirm password.
// When the form is submitted it will make PUT request to this endpoint.
func passwordResetEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(resetTokenReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		res := passwChangeRes{}

		if err := svc.UpdatePassword(ctx, req.Token, req.Password); err != nil {
			res.Msg = err.Error()
			return res, nil
		}
		res.Msg = ""
		return res, nil
	}
}

func userInfoEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(viewUserInfoReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		u, err := svc.UserInfo(ctx, req.token)
		if err != nil {
			return nil, err
		}

		return identityRes{u.Email, u.Metadata}, nil
	}
}

func passwordChangeEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(passwChangeReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		res := passwChangeRes{}

		u, err := svc.UserInfo(ctx, req.Token)
		if err != nil {
			res.Msg = err.Error()
			return res, nil
		}

		u.Password = req.OldPassword
		if _, err = svc.Login(ctx, u); err != nil {
			res.Msg = err.Error()
			return res, nil
		}

		if err = svc.UpdatePassword(ctx, req.Token, req.Password); err != nil {
			res.Msg = err.Error()
			return res, nil
		}

		return res, nil
	}
}

func loginEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(userReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		token, err := svc.Login(ctx, req.user)
		if err != nil {
			return nil, err
		}

		return tokenRes{token}, nil
	}
}
