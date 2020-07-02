// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/mainflux/mainflux/certs"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/errors"
)

const duplicateErr = "unique_violation"

var (
	errSaveDB = errors.New("failed to save certificate to database")
	errRemove = errors.New("failed to remove certificate from database")
)

var _ certs.Repository = (*certsRepository)(nil)

type Cert struct {
	ThingID string
	Serial  string
	Expire  time.Time
}

type certsRepository struct {
	db  *sqlx.DB
	log logger.Logger
}

// NewRepository instantiates a PostgreSQL implementation of certs
// repository.
func NewRepository(db *sqlx.DB, log logger.Logger) certs.Repository {
	return &certsRepository{db: db, log: log}
}

func (cr certsRepository) RetrieveAll(ctx context.Context, ownerID string, offset, limit uint64) (certs.Page, error) {
	q := `SELECT FROM certs WHERE owner_id = $1 ORDER BY expire LIMIT $2 OFFSET $3;`
	rows, err := cr.db.Query(q, ownerID, limit, offset)
	if err != nil {
		cr.log.Error(fmt.Sprintf("Failed to retrieve configs due to %s", err))
		return certs.Page{}, err
	}
	defer rows.Close()

	certificates := []certs.Cert{}

	for rows.Next() {
		c := certs.Cert{}
		if err := rows.Scan(&c.ThingID, &c.Serial, &c.Expire); err != nil {
			cr.log.Error(fmt.Sprintf("Failed to read retrieved config due to %s", err))
			return certs.Page{}, err

		}
		certificates = append(certificates, c)
	}

	q = fmt.Sprintf(`SELECT COUNT(*) FROM certs WHERE owner_id = $1`)
	var total uint64
	if err := cr.db.QueryRow(q, ownerID).Scan(&total); err != nil {
		cr.log.Error(fmt.Sprintf("Failed to count certs due to %s", err))
		return certs.Page{}, err
	}

	return certs.Page{
		Total:  total,
		Limit:  limit,
		Offset: offset,
		Certs:  certificates,
	}, nil
}

func (cr certsRepository) Save(ctx context.Context, cert certs.Cert) (string, error) {
	q := `INSERT INTO certs (thing_id, serial, expire)
		  VALUES (:thing_id, :serial, :expire)`

	tx, err := cr.db.Beginx()
	if err != nil {
		return "", errors.Wrap(errSaveDB, err)
	}

	dbcrt := toDBCert(cert)

	if _, err := tx.NamedExec(q, dbcrt); err != nil {
		e := err
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == duplicateErr {
			e = errors.New("error conflict")
		}

		cr.rollback("Failed to insert a Cert", tx, err)

		return "", errors.Wrap(errSaveDB, e)
	}

	if err := tx.Commit(); err != nil {
		cr.rollback("Failed to commit Config save", tx, err)
	}

	return cert.Serial, nil
}

func (cr certsRepository) Remove(ctx context.Context, c certs.Cert) error {
	q := `DELETE FROM certs WHERE serial = :serial`
	dbcrt := toDBCert(c)
	if _, err := cr.db.NamedExecContext(ctx, q, dbcrt); err != nil {
		return errors.Wrap(errRemove, err)
	}

	return nil
}

func (cr certsRepository) rollback(content string, tx *sqlx.Tx, err error) {
	cr.log.Error(fmt.Sprintf("%s %s", content, err))

	if err := tx.Rollback(); err != nil {
		cr.log.Error(fmt.Sprintf("Failed to rollback due to %s", err))
	}
}

type dbCert struct {
	ThingID string    `db:"thing_id"`
	Serial  string    `db:"serial"`
	Expire  time.Time `db:"expire"`
}

func toDBCert(c certs.Cert) dbCert {
	return dbCert{
		ThingID: c.ThingID,
		Serial:  c.Serial,
		Expire:  c.Expire,
	}
}