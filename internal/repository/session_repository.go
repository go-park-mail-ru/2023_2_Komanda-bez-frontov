package repository

type Session struct {
	SessionID string
	UserID    string
	Username  string
	CreatedAt int64
}

// TODO: Implement repository using database
