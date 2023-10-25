package repository

type Session struct {
	SessionID int64
	UserID    int64
	Username  string
	CreatedAt int64
}

// TODO: Implement repository using database
