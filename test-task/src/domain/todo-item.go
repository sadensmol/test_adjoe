package domain

import "time"

type ToDoItem struct {
	ID          int       `json:"id"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
}

func (t ToDoItem) WithID(id int64) ToDoItem {
	return ToDoItem{
		ID:          int(id),
		Description: t.Description,
		DueDate:     t.DueDate,
	}
}
