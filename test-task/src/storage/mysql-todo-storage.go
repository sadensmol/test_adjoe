package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-test-task/test-task/src/domain"
)

type MySqlTodo struct {
	DB *sql.DB
}

func (m *MySqlTodo) Save(ctx context.Context, item domain.ToDoItem) (int64, error) {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
	defer cancel()

	r, err := m.DB.ExecContext(ctx, "INSERT INTO ToDoItem (description, due_date) VALUES (?, ?)", item.Description, item.DueDate)
	if err != nil {
		return 0, err
	}
	id, err := r.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (m *MySqlTodo) Latest(ctx context.Context) (domain.ToDoItem, error) {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
	defer cancel()

	// Query the database for the latest item
	rows, err := m.DB.QueryContext(ctx, "SELECT id, description, due_date FROM ToDoItem ORDER BY id DESC LIMIT 1")
	if err != nil {
		return domain.ToDoItem{}, err
	}
	defer rows.Close()

	// Iterate over the rows
	for rows.Next() {
		var item domain.ToDoItem
		var dueDateBytes []byte
		err := rows.Scan(&item.ID, &item.Description, &dueDateBytes)
		if err != nil {
			return domain.ToDoItem{}, err
		}

		// Convert dueDateBytes to string and then parse it to time.Time
		dueDateStr := string(dueDateBytes)
		item.DueDate, err = time.Parse("2006-01-02 15:04:05", dueDateStr)
		if err != nil {
			return domain.ToDoItem{}, err
		}

		return item, nil
	}

	return domain.ToDoItem{}, errors.New("no items found")
}
