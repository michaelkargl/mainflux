// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package users

import (
	"context"
	"errors"

	"github.com/mainflux/mainflux/users/token"
)

var (
	// ErrConflict indicates usage of the existing email during account
	// registration.
	ErrConflict = errors.New("email already taken")

	// ErrMalformedEntity indicates malformed entity specification (e.g.
	// invalid username or password).
	ErrMalformedEntity = errors.New("malformed entity specification")

	// ErrUnauthorizedAccess indicates missing or invalid credentials provided
	// when accessing a protected resource.
	ErrUnauthorizedAccess = errors.New("missing or invalid credentials provided")

	// ErrNotFound indicates a non-existent entity request.
	ErrNotFound = errors.New("non-existent entity")

	// ErrScanMetadata indicates problem with metadata in db
	ErrScanMetadata = errors.New("Failed to scan metadata")

	// ErrMissingEmail indicates missing email for password reset request.
	ErrMissingEmail = errors.New("missing email for password reset")

	// ErrSavingRecoveryToken indicates error saving recovery token
	ErrSavingRecoveryToken = errors.New("error saving recovery token")

	// ErrDeletingRecoveryToken indicates error deleting recovery token
	ErrDeletingRecoveryToken = errors.New("error deleting recovery token")

	// ErrRetrievingRecoveryToken indicates error deleting recovery token
	ErrRetrievingRecoveryToken = errors.New("error deleting recovery token")

	// ErrMisingResetToken indicates malformed or missing reset token
	// for reseting password.
	ErrMisingResetToken = errors.New("error mising reset token")

	// ErrGeneratingResetToken indicates error in generating password recovery
	// token.
	ErrGeneratingResetToken = errors.New("error mising reset token")
)

// Service specifies an API that must be fullfiled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	// Register creates new user account. In case of the failed registration, a
	// non-nil error value is returned.
	Register(context.Context, User) error

	// Login authenticates the user given its credentials. Successful
	// authentication generates new access token. Failed invocations are
	// identified by the non-nil error values in the response.
	Login(context.Context, User) (string, error)

	// Identify validates user's token. If token is valid, user's id
	// is returned. If token is invalid, or invocation failed for some
	// other reason, non-nil error values are returned in response.
	Identify(string) (string, error)

	// Get authenticated user info for the given token.
	UserInfo(ctx context.Context, token string) (User, error)

	// SaveToken
	SaveToken(_ context.Context, email, token string) error

	// RetrieveToken
	RetrieveToken(_ context.Context, email string) (string, error)

	// DeleteToken
	DeleteToken(_ context.Context, email string) error

	// GenerateResetToken
	GenerateResetToken(_ context.Context, email string) (string, error)

	// ChangePassword
	ChangePassword(_ context.Context, email, token, password string) error
}

var _ Service = (*usersService)(nil)

type usersService struct {
	users  UserRepository
	hasher Hasher
	idp    IdentityProvider
}

// New instantiates the users service implementation.
func New(users UserRepository, hasher Hasher, idp IdentityProvider) Service {
	return &usersService{users: users, hasher: hasher, idp: idp}
}

func (svc usersService) Register(ctx context.Context, user User) error {
	hash, err := svc.hasher.Hash(user.Password)
	if err != nil {
		return ErrMalformedEntity
	}

	user.Password = hash
	return svc.users.Save(ctx, user)
}

func (svc usersService) Login(ctx context.Context, user User) (string, error) {
	dbUser, err := svc.users.RetrieveByID(ctx, user.Email)
	if err != nil {
		return "", ErrUnauthorizedAccess
	}

	if err := svc.hasher.Compare(user.Password, dbUser.Password); err != nil {
		return "", ErrUnauthorizedAccess
	}

	return svc.idp.TemporaryKey(user.Email)
}

func (svc usersService) Identify(token string) (string, error) {
	id, err := svc.idp.Identity(token)
	if err != nil {
		return "", ErrUnauthorizedAccess
	}
	return id, nil
}

func (svc usersService) UserInfo(ctx context.Context, token string) (User, error) {
	id, err := svc.idp.Identity(token)
	if err != nil {
		return User{}, ErrUnauthorizedAccess
	}

	dbUser, err := svc.users.RetrieveByID(ctx, id)
	if err != nil {
		return User{}, ErrUnauthorizedAccess
	}

	return User{
		Email:    id,
		Password: "",
		Metadata: dbUser.Metadata,
	}, nil

}

func (svc usersService) SaveToken(ctx context.Context, email, tok string) error {

	err := svc.users.SaveToken(ctx, email, tok)
	if err != nil {
		return ErrSavingRecoveryToken
	}

	return nil

}

func (svc usersService) RetrieveToken(ctx context.Context, email string) (string, error) {

	token, err := svc.users.RetrieveToken(ctx, email)
	if err != nil {
		return "", ErrSavingRecoveryToken
	}

	return token, nil

}

func (svc usersService) DeleteToken(ctx context.Context, email string) error {

	err := svc.users.DeleteToken(ctx, email)
	if err != nil {
		return ErrDeletingRecoveryToken
	}

	return nil

}

func (svc usersService) GenerateResetToken(_ context.Context, email string) (string, error) {

	tok, err := token.Generate(email)
	if err != nil {
		return "", ErrGeneratingResetToken
	}
	return tok, nil
}

func (svc usersService) ChangePassword(ctx context.Context, email, tok, password string) error {
	u, err := svc.users.RetrieveByID(ctx, email)
	if err != nil || u.Email == "" {
		return ErrNotFound
	}

	retToken, err := svc.users.RetrieveToken(ctx, email)
	if err != nil {
		return err
	}

	token.Verify(tok, retToken)
	if err != nil {
		return err
	}

	password, err = svc.hasher.Hash(password)
	if err != nil {
		return err
	}
	err = svc.users.ChangePassword(ctx, email, tok, password)
	return err

}
