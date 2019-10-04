// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/mainflux/mainflux/users"
	"github.com/mainflux/mainflux/users/token"
)

func registrationEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(userReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		err := svc.Re[gister(ctx, req.user)
		return tokenRes{}, err
	}
}

// Password reset endpoint serves post request with email of the user
// for whom password reset flow is to be initiated.
// If request is successful email with reset link will be sent to the
// email specified in the request.
// Link contains token that has TTL that needs to be verified.
// When user gets email with reset password link that hits this endpoint
// which will return response into the ui form which then can
// be used to enter email and new password and then request can be submitted
// to the post endpoint which actually changes password.
func passwordResetRequestEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(passwResetReq)
		res := resetPassRes{}
		email := req.Email
		err := svc.GenerateResetToken(ctx, email, req.Host)
		if err != nil {
			res.Msg = err.Error()
			return res, nil
		}
		res.Msg = MailSent
		return res, nil
	}
}

// This is post request endpoint that actually sets new password. It requires a token
// generated in the password reset request endpoint.
// Token is verified for the TTL and against generated token saved in DB.
func passwordResetPutEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(resetTokenReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		res := resetPassRes{}
		email, err := token.Verify(req.Token)
		if err != nil {
			res.Msg = err.Error()
			return res, nil
		}

		err = svc.UpdatePassword(ctx, email, req.Password)
		if err != nil {
			res.Msg = err.Error()
			return res, nil
		}
		res.Msg = ""
		return res, nil
	}
}

func userUpdateEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(userReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		// TO DO
		// u, err := svc.UserInfo(ctx, req.token)
		// if err != nil {
		// 	return nil, err
		// }

		// change this return value
		return identityRes{"", map[string]interface{}{}}, nil
	}
}

func passwordUpdateEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {

		// req : = request.(passwordChange)

		// TO DO

		return identityRes{"", map[string]interface{}{}}, nil
	}
}

func passwordResetRequestEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(passResReq)
		// TO DO
		// This endpoint will initiate the reset procedure
		// it will prepare and send a link for reset to the users email

		return nil, nil

	}
}

func passwordResetEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(resetTokenReq)
		err := svc.ChangePassword(ctx, req.email, req.token, req.password)
		if err != nil {
			res.Error = err.Error()
			return res, nil
		}
		res.Error = ""
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
