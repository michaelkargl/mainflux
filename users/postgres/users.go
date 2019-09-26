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
	"database/sql/driver"
	"encoding/json"
	"errors"

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

	dbu, err := toDBUser(user)
	if err != nil {
		return users.ErrMalformedEntity
	}

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

// DbMetadata type for handling metada properly in database/sql
type DbMetadata map[string]interface{}

// Scan - Implement the database/sql scanner interface
func (m *DbMetadata) Scan(value interface{}) error {
	// if value is nil, return empty
	if value == nil {
		m = new(DbMetadata)
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		m = new(DbMetadata)
		return errors.New("Failed to scan metadata")
	}

	if err := json.Unmarshal(b, m); err != nil {
		m = new(DbMetadata)
		return err
	}

	return nil
}

// Value Implements valuer
func (m DbMetadata) Value() (driver.Value, error) {
	if len(m) == 0 {
		return nil, nil
	}

	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return b, err
}

type dbUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Metadata DbMetadata
}

func toDBUser(u users.User) (dbUser, error) {
	m := u.Metadata
	return dbUser{
		Email:    u.Email,
		Password: u.Password,
		Metadata: m,
	}, nil
}

func toUser(dbu dbUser) users.User {

	return users.User{
		Email:    dbu.Email,
		Password: dbu.Password,
		Metadata: dbu.Metadata,
	}
}
