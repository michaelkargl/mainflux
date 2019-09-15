//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package postgres

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/lib/pq"
	"github.com/mainflux/mainflux/users"
)

var _ users.UserRepository = (*userRepository)(nil)

const errDuplicate = "unique_violation"

type userRepository struct {
	db *sqlx.DB
}

// New instantiates a PostgreSQL implementation of user
// repository.
func New(db *sqlx.DB) users.UserRepository {
	return &userRepository{db}
}

func (ur userRepository) Save(_ context.Context, user users.User) error {
	q := `INSERT INTO users (email, password, metadata) VALUES (:email, :password, :metadata)`

	dbu := toDBUser(user)
	if _, err := ur.db.NamedExec(q, dbu); err != nil {
		if pqErr, ok := err.(*pq.Error); ok && errDuplicate == pqErr.Code.Name() {
			return users.ErrConflict
		}
		return err
	}

	return nil
}

func (ur userRepository) RetrieveByID(_ context.Context, email string) (users.User, error) {
	q := `SELECT password, metadata FROM users WHERE email = $1`

	dbu := dbUser{
		Email: email,
	}
	if err := ur.db.QueryRowx(q, email).StructScan(&dbu); err != nil {
		if err == sql.ErrNoRows {
			return users.User{}, users.ErrNotFound
		}
		return users.User{}, err
	}

	user := toUser(dbu)

	return user, nil
}

type dbUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Metadata map[string]interface{}
}

func toDBUser(u users.User) dbUser {
	return dbUser{
		Email:    u.Email,
		Password: u.Password,b
		Metadata: u.Metadata,
	}
}

func toUser(dbu dbUser) users.User {
	return users.User{
		Email:    dbu.Email,
		Password: dbu.Password,
		Metadata: dbu.Metadata,
	}
}
