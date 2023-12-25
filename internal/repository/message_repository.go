package repository

import (
	"context"
	"fmt"

	"go-form-hub/internal/database"
	"go-form-hub/internal/model"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

var (
	selectMessageFields = []string{
		"u.id",
		"u.username",
		"u.first_name",
		"u.last_name",
		"u.email",
		"u.avatar",
		"m.id",
		"m.sender_id",
		"m.text",
		"m.send_at",
		"m.is_read",
	}
)

type messageDatabaseRepository struct {
	db      database.ConnPool
	builder squirrel.StatementBuilderType
}

func NewMessageDatabaseRepository(db database.ConnPool, builder squirrel.StatementBuilderType) MessageRepository {
	return &messageDatabaseRepository{
		db:      db,
		builder: builder,
	}
}

func (r *messageDatabaseRepository) Insert(ctx context.Context, message *model.MessageSave) error {
	query, args, err := r.builder.
		Insert(fmt.Sprintf("%s.message", r.db.GetSchema())).
		Columns("sender_id", "receiver_id", "text", "send_at", "is_read").
		Values(message.SenderID, message.ReceiverID, message.Text, message.SendAt, false).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return fmt.Errorf("message_repository insert failed to build query: %e", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("message_repository insert failed to begin transaction: %e", err)
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit(ctx)
		default:
			_ = tx.Rollback(ctx)
		}
	}()

	_, err = tx.Exec(ctx, query, args...)

	return err
}

func (r *messageDatabaseRepository) CheckUnreadForUser(ctx context.Context, userID int64) (*model.CheckUnreadMessages, error) {
	query, args, err := r.builder.
		Select("count(*) as unread_messages").
		From(fmt.Sprintf("%s.message", r.db.GetSchema())).
		Where(squirrel.Eq{"receiver_id": userID, "is_read": false}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("message_repository check unread messages failed to build query: %v", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("message_repository insert failed to begin transaction: %e", err)
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit(ctx)
		default:
			_ = tx.Rollback(ctx)
		}
	}()

	var unreadMessages int
	err = tx.QueryRow(ctx, query, args...).Scan(&unreadMessages)
	if err != nil {
		return nil, fmt.Errorf("message_repository check unread messages failed to execute query: %v", err)
	}

	return &model.CheckUnreadMessages{Count: unreadMessages}, nil
}

func (r *messageDatabaseRepository) GetChatByIDs(ctx context.Context, id1, id2 int64) (*model.Chat, error) {
	query, args, err := r.builder.
		Select(selectMessageFields...).
		From(fmt.Sprintf("%s.message as m", r.db.GetSchema())).
		Join(fmt.Sprintf("%s.user as u ON u.id = %d", r.db.GetSchema(), id2)).
		Where(squirrel.Or{squirrel.And{squirrel.Eq{"sender_id": id1}, squirrel.Eq{"receiver_id": id2}}, squirrel.And{squirrel.Eq{"sender_id": id2}, squirrel.Eq{"receiver_id": id1}}}).
		OrderBy("m.id").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("message_repository get chat by IDs failed to build query: %v", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("message_repository get chat failed to begin transaction: %e", err)
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit(ctx)
		default:
			_ = tx.Rollback(ctx)
		}
	}()

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("message_repository get chat failed to execute query: %e", err)
	}

	chat, err := r.fromRows(rows)
	if len(chat) == 0 {
		return nil, nil
	}

	return chat[0], err
}

func (r *messageDatabaseRepository) GetChatListByUserID(ctx context.Context, userID int64) ([]*model.Chat, error) {
	query, args, err := r.builder.
		Select(selectMessageFields...).
		From(fmt.Sprintf("%s.message as m", r.db.GetSchema())).
		Join(fmt.Sprintf("%s.user as u ON u.id != %d AND (u.id = m.sender_id OR u.id = m.receiver_id)", r.db.GetSchema(), userID)).
		Where(fmt.Sprintf("m.sender_id = %d OR m.receiver_id = %d", userID, userID)).
		OrderBy("m.id").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("message_repository get_all_chats by IDs failed to build query: %v", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("message_repository get_all_chats failed to begin transaction: %e", err)
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit(ctx)
		default:
			_ = tx.Rollback(ctx)
		}
	}()

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("crepository get_all_chats failed to execute query: %e", err)
	}

	chats, err := r.fromRows(rows)

	return chats, err
}

func (r *messageDatabaseRepository) ReadAllInChat(ctx context.Context, id1, id2 int64) error {
	query, args, err := r.builder.
		Update(fmt.Sprintf("%s.message as m", r.db.GetSchema())).
		Set("is_read", true).
		Where(fmt.Sprintf("(m.sender_id = %d OR m.receiver_id = %d) AND m.receiver_id = %d", id2, id2, id1)).
		ToSql()
	if err != nil {
		return fmt.Errorf("message_repository read_all failed to build query: %e", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("message_repository read_all failed to begin transaction: %e", err)
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit(ctx)
		default:
			_ = tx.Rollback(ctx)
		}
	}()

	_, err = tx.Exec(ctx, query, args...)

	return err
}

type fromRowMessageReturn struct {
	user    *model.UserGet
	message *model.Message
}

func (r *messageDatabaseRepository) fromRow(row pgx.Row) (*fromRowMessageReturn, error) {
	user := &model.UserGet{}
	message := &model.Message{}

	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Avatar,
		&message.ID,
		&message.SenderID,
		&message.Text,
		&message.SendAt,
		&message.IsRead,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("message_repository from_row failed to scan row: %e", err)
	}

	return &fromRowMessageReturn{user, message}, nil
}

func (r *messageDatabaseRepository) fromRows(rows pgx.Rows) ([]*model.Chat, error) {
	defer func() {
		rows.Close()
	}()

	chatMap := map[int64]*model.Chat{}

	for rows.Next() {
		info, err := r.fromRow(rows)
		if err != nil {
			return nil, err
		}

		if info.message == nil {
			continue
		}

		if _, ok := chatMap[info.user.ID]; !ok {
			chatMap[info.user.ID] = &model.Chat{
				User: &model.UserGet{
					ID:        info.user.ID,
					Username:  info.user.Username,
					FirstName: info.user.FirstName,
					LastName:  info.user.LastName,
					Email:     info.user.Email,
					Avatar:    info.user.Avatar,
				},
			}
		}

		chatMap[info.user.ID].Messages = append(chatMap[info.user.ID].Messages, &model.Message{
			ID:       info.message.ID,
			SenderID: info.message.SenderID,
			SendAt:   info.message.SendAt,
			IsRead:   info.message.IsRead,
			Text:     info.message.Text,
		})
	}

	chats := make([]*model.Chat, 0, len(chatMap))

	for _, chat := range chatMap {
		chats = append(chats, chat)
	}

	return chats, nil
}
