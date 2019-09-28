// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"

	"github.com/lib/pq"
	"github.com/mainflux/mainflux/users"
)

var _ users.UserRepository = (*userRepository)(nil)

const errDuplicate = "unique_violation"

type userRepository struct {
	db Database
}

// New instantiates a PostgreSQL implementation of user
// repository.
func New(db Database) users.UserRepository {
	return &userRepository{
		db: db,
	}
}

func (ur userRepository) Save(ctx context.Context, user users.User) error {
	q := `INSERT INTO users (email, password, metadata) VALUES (:email, :password, :metadata)`

	dbu := toDBUser(user)
	if _, err := ur.db.NamedExecContext(ctx, q, dbu); err != nil {
		if pqErr, ok := err.(*pq.Error); ok && errDuplicate == pqErr.Code.Name() {
			return users.ErrConflict
		}
		return err
	}

	return nil
}

func (ur userRepository) RetrieveByID(ctx context.Context, email string) (users.User, error) {
	q := `SELECT password, metadata FROM users WHERE email = $1`

	dbu := dbUser{
		Email: email,
	}
	if err := ur.db.QueryRowxContext(ctx, q, email).StructScan(&dbu); err != nil {
		if err == sql.ErrNoRows {
			return users.User{}, users.ErrNotFound
		}
		return users.User{}, err
	}

	user := toUser(dbu)

	return user, nil
}

<<<<<<< HEAD
// dbMetadata type for handling metadata properly in database/sql
type dbMetadata map[string]interface{}

// Scan - Implement the database/sql scanner interface
func (m *dbMetadata) Scan(value interface{}) error {
	if value == nil {
		m = nil
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		m = &dbMetadata{}
		return users.ErrScanMetadata
	}

	if err := json.Unmarshal(b, m); err != nil {
		m = &dbMetadata{}
=======
func (ur userRepository) SaveToken(_ context.Context, email, token string) error {
	t, err := ur.retrieveTokenByID(email)
	if err != nil {
		return err
	}
	q := `INSERT INTO tokens (user_id, token) VALUES (:email, :token )`
	if len(t) > 0 {
		q = `UPDATE tokens SET ( token) VALUES :token  WHERE user_id = :email`
	}

	db := struct {
		Email string
		Token string
	}{
		email,
		token,
	}

	if _, err := ur.db.NamedExec(q, db); err != nil {
		if pqErr, ok := err.(*pq.Error); ok && errDuplicate == pqErr.Code.Name() {
			return users.ErrConflict
		}
>>>>>>> add token
		return err
	}

	return nil
}

<<<<<<< HEAD
// Value Implements valuer
func (m dbMetadata) Value() (driver.Value, error) {
	if len(m) == 0 {
		return nil, nil
	}

	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return b, err
=======
func (ur userRepository) RetrieveToken(_ context.Context, email string) (string, error) {
	return ur.retrieveTokenByID(email)
}

func (ur userRepository) DeleteToken(email string) (string, error) {
	q := `SELECT token from tokens WHERE email = $1`

	t := ""
	if err := ur.db.QueryRowx(q, email).StructScan(&t); err != nil {
		if err == sql.ErrNoRows {
			return t, users.ErrNotFound
		}
		return t, err
	}

	return t, nil
}

func (ur userRepository) retrieveTokenByID(email string) (string, error) {
	q := `SELECT token from tokens WHERE email = $1`

	t := ""
	if err := ur.db.QueryRowx(q, email).StructScan(&t); err != nil {
		if err == sql.ErrNoRows {
			return t, users.ErrNotFound
		}
		return t, err
	}

	return t, nil
>>>>>>> add token
}

type dbUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Metadata []byte `json:"metadata,omitempty"`
}

func toDBUser(u users.User) (dbUser, error) {
	m := []byte("{}")
	if len(u.Metadata) > 0 {
		data, err := json.Marshal(u.Metadata)
		if err != nil {
			return dbUser{}, err
		}
		m = data
	}

	return dbUser{
		Email:    u.Email,
		Password: u.Password,
		Metadata: m,
	}, nil
}

func toUser(dbu dbUser) users.User {

	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(dbu.Metadata), &metadata); err != nil {
		return users.User{}
	}

	return users.User{
		Email:    dbu.Email,
		Password: dbu.Password,
		Metadata: metadata,
	}
}
