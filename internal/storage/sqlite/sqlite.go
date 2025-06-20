package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"short-url/internal/http-server/model/domain"
	"short-url/internal/storage"

	sqlite3 "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

// event loaded from events table
type event struct {
	ID        int    `db:"id"`
	EventType string `db:"event_type"`
	Payload   string `db:"payload"`
}

const (
	eventType = "url_saved"
)

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS url(
		id INTEGER PRIMARY KEY,
		alias TEXT NOT NULL UNIQUE,
		url TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);

	CREATE TABLE IF NOT EXISTS events(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		event_type TEXT NOT NULL,
		payload TEXT NOT NULL,
		status TEXT NOT NULL DEAFULT 'new' CHECK (status IN ('new', 'done')),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);`)
	//TODO: add reserved_to TIMESTAMP DEFAULT NULL
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db}, nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) (id int64, err error) {
	const op = "storage.sqlite.SaveURL"
	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
		commitErr := tx.Commit()
		if commitErr != nil {
			err = fmt.Errorf("%s: %w", op, commitErr)
		}
	}()

	stmt, err := tx.Prepare("INSERT INTO url(url,alias) VALUES(?,?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	res, err := stmt.Exec(urlToSave, alias)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}
	}
	id, err = res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	//save event to events table
	payload := fmt.Sprintf(`{"id": "%d", "url": "%s", "alias": "%s"}`, id, urlToSave, alias)
	if err := s.saveEvent(tx, eventType, payload); err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

func (s *Storage) saveEvent(tx *sql.Tx, eventType, payload string) error {
	const op = "storage.sqlite.saveEvent"
	stmt, err := tx.Prepare("INSERT INTO events(event_type, payload) VALUES(?, ?)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	_, err = stmt.Exec(eventType, payload)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.sqlite.GetURL"
	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias=?")
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	var resURL string
	err = stmt.QueryRow(alias).Scan(&resURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("%s: %w", op, storage.ErrURLNotFound)
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return resURL, nil
}

func (s *Storage) DeleteURL(alias string) error {
	const op = "storage.sqlite.DeleteURL"
	stmt, err := s.db.Prepare("DELETE FROM url WHERE alias=?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	res, err := stmt.Exec(alias)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrURLNotFound)
	}
	return nil
}

func (s *Storage) GetNewEvent() (domain.Event, error) {
	const op = "storage.sqlite.GetNewEvent"
	//TODO: reserved_to logic
	//TODO: batch processing of events
	row := s.db.QueryRow("SELECT id, event_type, payload FROM events WHERE status='new' LIMIT 1")
	var e event
	err := row.Scan(&e.ID, &e.EventType, &e.Payload)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Event{}, fmt.Errorf("%s: %w", op, storage.ErrEventNotFound)
		}
		return domain.Event{}, fmt.Errorf("%s: %w", op, err)
	}
	return domain.Event{
		ID:        e.ID,
		EventType: e.EventType,
		Payload:   e.Payload,
	}, nil
}

func (s *Storage) MarkEventAsDone(eventID int) error {
	const op = "storage.sqlite.MarkEventAsDone"

	stmt, err := s.db.Prepare("UPDATE events SET status='done' WHERE id=?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	_, err = stmt.Exec(eventID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
