package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"shortener/internal/storage"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	PgErrUniqueViolation = "23505"
)

type Storage struct {
	db *sql.DB
}

func New(db *sql.DB) *Storage {

	return &Storage{db: db}
}

func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const op = "storage.postgres.SaveURL"

	const q = `
		INSERT INTO url(alias, url)
		VALUES ($1, $2)
		RETURNING id
	`

	var id int64
	err := s.db.QueryRow(q, alias, urlToSave).Scan(&id)
	if err == nil {
		return id, nil
	}

	// alias уже существует (unique_violation 23505)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == PgErrUniqueViolation {
		return 0, storage.ErrURLExists // если такого нет — замени на свою ошибку
	}

	return 0, fmt.Errorf("%s: %w", op, err)
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.postgres.GetURL"

	const q = `SELECT url FROM url WHERE alias = $1`

	var resURL string
	err := s.db.QueryRow(q, alias).Scan(&resURL)
	if errors.Is(err, sql.ErrNoRows) {
		return "", storage.ErrURLNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return resURL, nil
}

func (s *Storage) DeleteURL(alias string) (int64, error) {
	const op = "storage.postgres.DeleteURL"

	const q = `DELETE FROM url WHERE alias = $1`

	res, err := s.db.Exec(q, alias)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	if n == 0 {
		return 0, storage.ErrURLNotFound
	}

	return n, nil
}
