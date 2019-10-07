// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"
	"sync"

	"github.com/mainflux/mainflux/users"
	tok "github.com/mainflux/mainflux/users/token"
)

var _ users.UserRepository = (*userRepositoryMock)(nil)

type userRepositoryMock struct {
	mu     sync.Mutex
	users  map[string]users.User
	tokens map[string]string
}

// NewUserRepository creates in-memory user repository.
func NewUserRepository() users.UserRepository {
	return &userRepositoryMock{
		users:  make(map[string]users.User),
		tokens: make(map[string]string),
	}
}

func (urm *userRepositoryMock) Save(ctx context.Context, user users.User) error {
	urm.mu.Lock()
	defer urm.mu.Unlock()

	if _, ok := urm.users[user.Email]; ok {
		return users.ErrConflict
	}

	urm.users[user.Email] = user
	return nil
}

func (urm *userRepositoryMock) RetrieveByID(ctx context.Context, email string) (users.User, error) {
	urm.mu.Lock()
	defer urm.mu.Unlock()

	val, ok := urm.users[email]
	if !ok {
		return users.User{}, users.ErrNotFound
	}

	return val, nil
}
func (urm *userRepositoryMock) SaveToken(_ context.Context, email, token string) error {
	urm.mu.Lock()
	defer urm.mu.Unlock()

	if _, ok := urm.tokens[email]; ok {
		return users.ErrConflict
	}
	t, _ := tok.Hash(token)
	urm.tokens[email] = t
	return nil
}

// RetrieveToken
func (urm *userRepositoryMock) RetrieveToken(_ context.Context, email string) (string, error) {

	urm.mu.Lock()
	defer urm.mu.Unlock()

	val, ok := urm.tokens[email]
	if !ok {
		return "", users.ErrNotFound
	}

	return val, nil
}

// DeleteToken
func (urm *userRepositoryMock) DeleteToken(_ context.Context, email string) error {
	return nil
}

// ChangePassword
func (urm *userRepositoryMock) ChangePassword(_ context.Context, email, token, password string) error {
	urm.mu.Lock()
	defer urm.mu.Unlock()
	_, ok := urm.users[email]
	if !ok {
		return users.ErrNotFound
	}
	return nil
}
