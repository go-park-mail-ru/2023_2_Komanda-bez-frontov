package repository

type Question struct {
	ID          int64   `db:"id"`
	FormID      int64   `db:"form_id"`
	Title       *string `db:"title"`
	Description *string `db:"text"`
	Type        int64   `db:"type"`
	Shuffle     bool    `db:"shuffle"`
}
