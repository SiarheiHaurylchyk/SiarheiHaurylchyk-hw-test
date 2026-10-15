package sqlstorage

import (
	"context"
	"database/sql"
	"time"

	"github.com/SiarheiHaurylchyk/SiarheiHaurylchyk-hw-test/hw12_13_14_15_calendar/internal/storage"
	_ "github.com/lib/pq" // драйвер PostgreSQL
)

type Storage struct {
	dsn string
	db  *sql.DB
}

func New(dsn string) *Storage {
	return &Storage{dsn: dsn}
}

func (s *Storage) Connect(ctx context.Context) error {
	db, err := sql.Open("postgres", s.dsn)
	if err != nil {
		return err
	}
	if err = db.PingContext(ctx); err != nil {
		return err
	}

	s.db = db
	return nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) CreateEvent(ctx context.Context, event storage.Event) error {
	isBusy, err := s.isTimeBusy(ctx, event)
	if err != nil {
		return err
	}
	if isBusy {
		return storage.ErrDateBusy
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO events (id, title, start_time, end_time, description, user_id, notify_before)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		event.ID, event.Title, event.StartTime, event.EndTime,
		event.Description, event.UserID, int64(event.NotifyBefore),
	)
	return err
}

func (s *Storage) UpdateEvent(ctx context.Context, event storage.Event) error {
	isBusy, err := s.isTimeBusy(ctx, event)
	if err != nil {
		return err
	}
	if isBusy {
		return storage.ErrDateBusy
	}

	result, err := s.db.ExecContext(ctx,
		`UPDATE events
		 SET title = $2, start_time = $3, end_time = $4, description = $5, user_id = $6, notify_before = $7
		 WHERE id = $1`,
		event.ID, event.Title, event.StartTime, event.EndTime,
		event.Description, event.UserID, int64(event.NotifyBefore),
	)
	if err != nil {
		return err
	}
	return checkRowWasChanged(result)
}

func (s *Storage) DeleteEvent(ctx context.Context, eventID string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM events WHERE id = $1`, eventID)
	if err != nil {
		return err
	}
	return checkRowWasChanged(result)
}

func (s *Storage) ListEvents(ctx context.Context, from, to time.Time) ([]storage.Event, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, title, start_time, end_time, description, user_id, notify_before
		 FROM events
		 WHERE start_time >= $1 AND start_time < $2
		 ORDER BY start_time`,
		from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]storage.Event, 0)
	for rows.Next() {
		var event storage.Event
		var notifyBefore int64
		err = rows.Scan(
			&event.ID, &event.Title, &event.StartTime, &event.EndTime,
			&event.Description, &event.UserID, &notifyBefore,
		)
		if err != nil {
			return nil, err
		}
		event.NotifyBefore = time.Duration(notifyBefore)
		result = append(result, event)
	}
	return result, rows.Err()
}

func (s *Storage) isTimeBusy(ctx context.Context, event storage.Event) (bool, error) {
	var isBusy bool
	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS (
			SELECT 1 FROM events
			WHERE user_id = $1 AND id <> $2 AND start_time < $4 AND end_time > $3
		)`,
		event.UserID, event.ID, event.StartTime, event.EndTime,
	).Scan(&isBusy)
	return isBusy, err
}

func checkRowWasChanged(result sql.Result) error {
	changedRows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if changedRows == 0 {
		return storage.ErrEventNotFound
	}
	return nil
}
