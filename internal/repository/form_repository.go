package repository

import (
	"context"
	"fmt"
	"time"

	"go-form-hub/internal/database"
	"go-form-hub/internal/model"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

type Form struct {
	Title     string    `db:"title"`
	ID        int64     `db:"id"`
	AuthorID  int64     `db:"author_id"`
	CreatedAt time.Time `db:"created_at"`
}

var (
	selectFields = []string{
		"f.id",
		"f.title",
		"f.created_at",
		"f.author_id",
		"u.id",
		"u.username",
		"u.first_name",
		"u.last_name",
		"u.email",
		"q.id",
		"q.title",
		"q.text",
		"q.type",
		"q.required",
		"a.id",
		"a.answer_text",
	}
)

var (
	selectFieldsResults = []string{
		"f.id",
		"f.title",
		"f.created_at",
		"f.description",
		"f.anonymous",
		"u.id",
		"u.username",
		"u.first_name",
		"u.last_name",
		"u.email",
		"u.id",
		"u.username",
		"u.first_name",
		"u.last_name",
		"u.email",
		"q.id",
		"q.title",
		"q.text",
		"q.type",
		"q.required",
		"a.id",
		"a.answer_text",
		"COUNT(DISTINCT q.id) AS NumberOfPassagesForm",
		"COUNT(DISTINCT a.id) AS NumberOfPassagesQuestion",
		"COUNT(DISTINCT a.answer_text) AS SelectedTimesAnswer",
	}
)

type formDatabaseRepository struct {
	db      database.ConnPool
	builder squirrel.StatementBuilderType
}

func NewFormDatabaseRepository(db database.ConnPool, builder squirrel.StatementBuilderType) FormRepository {
	return &formDatabaseRepository{
		db:      db,
		builder: builder,
	}
}

func (r *formDatabaseRepository) FindAll(ctx context.Context) (forms []*model.Form, err error) {
	query, _, err := r.builder.
		Select(selectFields...).
		From(fmt.Sprintf("%s.form as f", r.db.GetSchema())).
		Join(fmt.Sprintf("%s.user as u ON f.author_id = u.id", r.db.GetSchema())).
		LeftJoin(fmt.Sprintf("%s.question as q ON q.form_id = f.id", r.db.GetSchema())).
		LeftJoin(fmt.Sprintf("%s.answer as a ON a.question_id = q.id", r.db.GetSchema())).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("form_repository find_all failed to build query: %e", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("form_repository find_all failed to begin transaction: %e", err)
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit(ctx)
		default:
			_ = tx.Rollback(ctx)
		}
	}()

	rows, err := tx.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("form_repository find_all failed to execute query: %e", err)
	}

	return r.fromRows(rows)
}

func (r *formDatabaseRepository) FormsSearch(ctx context.Context, title string, userID uint) (forms []*model.FormTitle, err error) {
	const limit = 5
	query := fmt.Sprintf(`select id, title, created_at
	FROM (select title, id, created_at, similarity(title, $1::text) as sim
	FROM %s.form
	WHERE author_id = $2::integer
	order by sim desc, created_at) as res
	LIMIT $3::integer`, r.db.GetSchema())

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("form_repository form_search failed to begin transaction: %e", err)
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit(ctx)
		default:
			_ = tx.Rollback(ctx)
		}
	}()

	rows, err := tx.Query(ctx, query, title, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("form_repository form_search failed to execute query: %e", err)
	}

	return r.searchTitleFromRows(rows)
}

// FormResults извлекает результаты формы по ID формы.
// Эта функция строит SQL-запрос для получения данных, связанных с формой,
// из базы данных, включая информацию о форме, вопросах, ответах и участниках.
// Результат представляет собой структурированный model.FormResult.
func (r *formDatabaseRepository) FormResults(ctx context.Context, id int64) (formResult *model.FormResult, err error) {
	formQuery, args, err := r.builder.
		Select(selectFieldsResults...).
		From(fmt.Sprintf("%s.form as f", r.db.GetSchema())).
		Join(fmt.Sprintf("%s.user as u ON f.author_id = u.id", r.db.GetSchema())).
		LeftJoin(fmt.Sprintf("%s.question as q ON q.form_id = f.id", r.db.GetSchema())).
		LeftJoin(fmt.Sprintf("%s.answer as a ON a.question_id = q.id", r.db.GetSchema())).
		LeftJoin(fmt.Sprintf("%s.passage_answer as pa ON q.id = pa.question_id", r.db.GetSchema())).
		Where(squirrel.Eq{"f.id": id}).
		GroupBy("f.id, f.title, f.created_at, f.description, f.author_id, u.id, u.username, u.first_name, u.last_name, u.email, q.id, q.title, q.text, q.type, q.required, a.id, a.answer_text, pa.user_id"). // Добавлено условие для группировки по user_id из passage_answer
		ToSql()

	fmt.Println("SQL Query:", formQuery)
	if err != nil {
		return nil, fmt.Errorf("form_repository form_results failed to build query: %e", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("form_repository form_results failed to begin transaction: %e", err)
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit(ctx)
		default:
			_ = tx.Rollback(ctx)
		}
	}()

	rows, err := tx.Query(ctx, formQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("form_repository form_results failed to execute query: %e", err)
	}

	formResults, err := r.formResultsFromRows(ctx, rows)
	if err != nil {
		return nil, err
	}

	if len(formResults) == 0 {
		return nil, nil
	}

	return formResults[0], nil
}

// formResultsFromRows обрабатывает строки, полученные из результата запроса к базе данных,
// и создает список структурированных результатов формы.
// Она заполняет карты для организации вопросов и ответов по ID формы
// и вычисляет количество прохождений для каждой формы.
// Кроме того, она извлекает информацию о участниках для каждой формы,
// вызывая функцию getParticipantsForForm.
func (r *formDatabaseRepository) formResultsFromRows(ctx context.Context, rows pgx.Rows) ([]*model.FormResult, error) {
	defer func() {
		rows.Close()
	}()

	formResultMap := map[int64]*model.FormResult{}
	questionsByFormID := map[int64][]*model.QuestionResult{}
	answersByQuestionID := map[int64][]*model.AnswerResult{}

	for rows.Next() {
		info, err := r.formResultsFromRow(rows)
		if err != nil {
			return nil, err
		}

		if info.formResult == nil {
			continue
		}

		if _, ok := formResultMap[info.formResult.ID]; !ok {
			formResultMap[info.formResult.ID] = &model.FormResult{
				ID:                   info.formResult.ID,
				Title:                info.formResult.Title,
				Description:          info.formResult.Description,
				CreatedAt:            info.formResult.CreatedAt,
				Author:               info.formResult.Author,
				NumberOfPassagesForm: 0,
				Questions:            []*model.QuestionResult{},
				Anonymous:            info.formResult.Anonymous,
			}
		}

		var questionExists bool
		var existingQuestion *model.QuestionResult

		for _, existingQuestion = range formResultMap[info.formResult.ID].Questions {
			if existingQuestion.ID == info.questionResult.ID {
				questionExists = true
				break
			}
		}

		if questionExists {
			existingQuestion.NumberOfPassagesQuestion++
			for _, existingAnswer := range existingQuestion.Answers {
				if existingAnswer.ID == info.answerResult.ID {
					existingAnswer.SelectedTimesAnswer++
					break
				}
			}
		} else {
			formResultMap[info.formResult.ID].Questions = append(formResultMap[info.formResult.ID].Questions, info.questionResult)
			info.questionResult.Answers = append(info.questionResult.Answers, info.answerResult)

			if _, ok := questionsByFormID[info.formResult.ID]; !ok {
				questionsByFormID[info.formResult.ID] = make([]*model.QuestionResult, 0)
			}

			questionsByFormID[info.formResult.ID] = append(questionsByFormID[info.formResult.ID], info.questionResult)

			if _, ok := answersByQuestionID[info.questionResult.ID]; !ok {
				answersByQuestionID[info.questionResult.ID] = make([]*model.AnswerResult, 0)
			}

			answersByQuestionID[info.questionResult.ID] = append(answersByQuestionID[info.questionResult.ID], info.answerResult)
		}
		formResultMap[info.formResult.ID].NumberOfPassagesForm++
	}

	formResults := make([]*model.FormResult, 0, len(formResultMap))

	for _, formResult := range formResultMap {
		formResult.Questions = questionsByFormID[formResult.ID]
		for _, questionResult := range formResult.Questions {
			questionResult.Answers = answersByQuestionID[questionResult.ID]
		}

		participants, err := r.getParticipantsForForm(ctx, formResult.ID)
		if err != nil {
			return nil, err
		}
		formResult.Participants = participants

		formResults = append(formResults, formResult)
	}

	return formResults, nil
}

// getParticipantsForForm извлекает информацию о участниках (UserGet) для данной формы.
// Она строит SQL-запрос для выбора уникальных участников, которые ответили на вопросы в форме.
// Результат представляет собой слайс структурированных model.UserGet.
func (r *formDatabaseRepository) getParticipantsForForm(ctx context.Context, formID int64) ([]*model.UserGet, error) {
	query, args, err := r.builder.
		Select("u.id", "u.username", "u.first_name", "u.last_name", "u.email").
		From(fmt.Sprintf("%s.passage_answer as pa", r.db.GetSchema())).
		Join(fmt.Sprintf("%s.user as u ON pa.user_id = u.id", r.db.GetSchema())).
		Join(fmt.Sprintf("%s.question as q ON pa.question_id = q.id", r.db.GetSchema())).
		Where(squirrel.Eq{"q.form_id": formID}).
		GroupBy("u.id").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("form_repository getParticipantsForForm failed to build query: %v", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("form_repository form_results failed to begin transaction: %e", err)
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
		return nil, fmt.Errorf("form_repository form_results failed to execute query: %e", err)
	}

	defer rows.Close()

	participants := make([]*model.UserGet, 0)

	for rows.Next() {
		var participant model.UserGet
		if err := rows.Scan(&participant.ID, &participant.Username, &participant.FirstName, &participant.LastName, &participant.Email); err != nil {
			return nil, fmt.Errorf("form_repository getParticipantsForForm failed to scan row: %v", err)
		}
		participants = append(participants, &participant)
	}

	return participants, nil
}

type formResultsFromRowReturn struct {
	formResult      *model.FormResult
	questionResult  *model.QuestionResult
	answerResult    *model.AnswerResult
	participantInfo *model.UserGet
}

// formResultsFromRowReturn представляет структурированные данные,
// возвращаемые при обработке одной строки в результате запроса к базе данных.
// Она включает информацию о форме, вопросе, ответе и участнике.
func (r *formDatabaseRepository) formResultsFromRow(row pgx.Row) (*formResultsFromRowReturn, error) {
	formResult := &model.FormResult{}
	questionResult := &model.QuestionResult{}
	answerResult := &model.AnswerResult{}
	formResult.Author = &model.UserGet{}
	participantInfo := &model.UserGet{}

	err := row.Scan(
		&formResult.ID,
		&formResult.Title,
		&formResult.CreatedAt,
		&formResult.Description,
		&formResult.Anonymous,
		&formResult.Author.ID,
		&formResult.Author.Username,
		&formResult.Author.FirstName,
		&formResult.Author.LastName,
		&formResult.Author.Email,
		&participantInfo.ID,
		&participantInfo.Username,
		&participantInfo.FirstName,
		&participantInfo.LastName,
		&participantInfo.Email,
		&questionResult.ID,
		&questionResult.Title,
		&questionResult.Description,
		&questionResult.Type,
		&questionResult.Required,
		&answerResult.ID,
		&answerResult.Text,
		&formResult.NumberOfPassagesForm,
		&questionResult.NumberOfPassagesQuestion,
		&answerResult.SelectedTimesAnswer,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("form_repository failed to scan row: %e", err)
	}

	return &formResultsFromRowReturn{formResult, questionResult, answerResult, participantInfo}, nil
}

func (r *formDatabaseRepository) FindAllByUser(ctx context.Context, username string) (forms []*model.Form, err error) {
	query, args, err := r.builder.
		Select(selectFields...).
		From(fmt.Sprintf("%s.form as f", r.db.GetSchema())).
		Join(fmt.Sprintf("%s.user as u ON f.author_id = u.id", r.db.GetSchema())).
		LeftJoin(fmt.Sprintf("%s.question as q ON q.form_id = f.id", r.db.GetSchema())).
		LeftJoin(fmt.Sprintf("%s.answer as a ON a.question_id = q.id", r.db.GetSchema())).
		Where(squirrel.Eq{"u.username": username}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("form_repository find_all failed to build query: %e", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("form_repository find_all failed to begin transaction: %e", err)
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
		return nil, fmt.Errorf("form_repository find_all failed to execute query: %e", err)
	}

	return r.fromRows(rows)
}

func (r *formDatabaseRepository) FindByID(ctx context.Context, id int64) (form *model.Form, err error) {
	query, args, err := r.builder.
		Select(selectFields...).
		From(fmt.Sprintf("%s.form as f", r.db.GetSchema())).
		Join(fmt.Sprintf("%s.user as u ON f.author_id = u.id", r.db.GetSchema())).
		LeftJoin(fmt.Sprintf("%s.question as q ON q.form_id = f.id", r.db.GetSchema())).
		LeftJoin(fmt.Sprintf("%s.answer as a ON a.question_id = q.id", r.db.GetSchema())).
		Where(squirrel.Eq{"f.id": id}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("form_repository find_by_title failed to build query: %e", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("form_repository find_by_title failed to begin transaction: %e", err)
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
		return nil, fmt.Errorf("form_repository find_by_title failed to execute query: %e", err)
	}

	forms, err := r.fromRows(rows)
	if len(forms) == 0 {
		return nil, nil
	}

	return forms[0], err
}

func (r *formDatabaseRepository) Insert(ctx context.Context, form *model.Form, tx pgx.Tx) (*model.Form, error) {
	var err error

	if tx == nil {
		tx, err = r.db.Begin(ctx)
		if err != nil {
			return nil, fmt.Errorf("form_facade insert failed to begin transaction: %e", err)
		}
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit(ctx)
		default:
			_ = tx.Rollback(ctx)
		}
	}()

	formQuery, args, err := r.builder.
		Insert(fmt.Sprintf("%s.form", r.db.GetSchema())).
		Columns("title", "author_id", "created_at").
		Values(form.Title, form.Author.ID, form.CreatedAt).
		Suffix("RETURNING id").
		ToSql()
	err = tx.QueryRow(ctx, formQuery, args...).Scan(&form.ID)
	if err != nil {
		return nil, err
	}

	questionBatch := &pgx.Batch{}
	questionQuery := r.builder.
		Insert(fmt.Sprintf("%s.question", r.db.GetSchema())).
		Columns("title", "text", "type", "required", "form_id").
		Suffix("RETURNING id")

	for _, question := range form.Questions {
		q, args, err := questionQuery.Values(question.Title, question.Description, question.Type, question.Required, form.ID).ToSql()
		if err != nil {
			return nil, err
		}

		questionBatch.Queue(q, args...)
	}
	questionResults := tx.SendBatch(ctx, questionBatch)

	answerBatch := &pgx.Batch{}
	answerQuery := r.builder.
		Insert(fmt.Sprintf("%s.answer", r.db.GetSchema())).
		Columns("answer_text", "question_id").
		Suffix("RETURNING id")

	for _, question := range form.Questions {
		questionID := int64(0)
		err = questionResults.QueryRow().Scan(&questionID)
		if err != nil {
			return nil, err
		}

		question.ID = &questionID
		for _, answer := range question.Answers {
			q, args, err := answerQuery.Values(answer.Text, question.ID).ToSql()
			if err != nil {
				return nil, err
			}

			answerBatch.Queue(q, args...)
		}
	}
	questionResults.Close()

	answerResults := tx.SendBatch(ctx, answerBatch)
	for _, question := range form.Questions {
		for _, answer := range question.Answers {
			answerID := int64(0)
			err = answerResults.QueryRow().Scan(&answerID)
			if err != nil {
				return nil, err
			}

			answer.ID = &answerID
		}
	}
	answerResults.Close()

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return form, nil
}

func (r *formDatabaseRepository) Update(ctx context.Context, id int64, form *model.Form) (result *model.Form, err error) {
	query, args, err := r.builder.Update(fmt.Sprintf("%s.form", r.db.GetSchema())).
		Set("title", form.Title).
		Where(squirrel.Eq{"id": id}).
		Suffix("RETURNING id, title, created_at").ToSql()
	if err != nil {
		return nil, fmt.Errorf("form_repository update failed to build query: %e", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("form_repository update failed to begin transaction: %e", err)
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
	if err != nil {
		return nil, fmt.Errorf("form_repository update failed to execute query: %e", err)
	}

	return form, nil
}

func (r *formDatabaseRepository) Delete(ctx context.Context, id int64) (err error) {
	query, args, err := r.builder.Delete(fmt.Sprintf("%s.form", r.db.GetSchema())).
		Where(squirrel.Eq{"id": id}).ToSql()
	if err != nil {
		return fmt.Errorf("form_repository delete failed to build query: %e", err)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("form_repository delete failed to begin transaction: %e", err)
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
	if err != nil {
		return fmt.Errorf("form_repository delete failed to execute query: %e", err)
	}

	return nil
}

func (r *formDatabaseRepository) fromRows(rows pgx.Rows) ([]*model.Form, error) {
	defer func() {
		rows.Close()
	}()

	formMap := map[int64]*model.Form{}
	questionsByFormID := map[int64][]*model.Question{}
	answersByQuestionID := map[int64][]*model.Answer{}

	questionWasAppended := map[int64]bool{}

	for rows.Next() {
		info, err := r.fromRow(rows)
		if err != nil {
			return nil, err
		}

		if info.form == nil {
			continue
		}

		if _, ok := formMap[info.form.ID]; !ok {
			formMap[info.form.ID] = &model.Form{
				ID:        &info.form.ID,
				Title:     info.form.Title,
				CreatedAt: info.form.CreatedAt,
				Author: &model.UserGet{
					ID:        info.author.ID,
					Username:  info.author.Username,
					FirstName: info.author.FirstName,
					LastName:  info.author.LastName,
					Email:     info.author.Email,
					Avatar:    info.author.Avatar,
				},
			}
		}

		if _, ok := questionWasAppended[info.question.ID]; !ok {
			questionsByFormID[info.form.ID] = append(questionsByFormID[info.form.ID], &model.Question{
				ID:          &info.question.ID,
				Title:       info.question.Title,
				Description: info.question.Text,
				Type:        info.question.Type,
				Required:    info.question.Required,
			})
			questionWasAppended[info.question.ID] = true
		}

		if _, ok := answersByQuestionID[info.question.ID]; !ok {
			answersByQuestionID[info.question.ID] = make([]*model.Answer, 0, 1)
		}

		answersByQuestionID[info.question.ID] = append(answersByQuestionID[info.question.ID], &model.Answer{
			ID:   &info.answer.ID,
			Text: info.answer.AnswerText,
		})
	}

	forms := make([]*model.Form, 0, len(formMap))

	for _, form := range formMap {
		form.Questions = questionsByFormID[*form.ID]
		for _, question := range form.Questions {
			question.Answers = answersByQuestionID[*question.ID]
		}
		forms = append(forms, form)
	}

	return forms, nil
}

func (r *formDatabaseRepository) searchTitleFromRows(rows pgx.Rows) ([]*model.FormTitle, error) {
	defer func() {
		rows.Close()
	}()

	formTitleArray := make([]*model.FormTitle, 0)

	for rows.Next() {
		form, err := r.formTitleFromRow(rows)
		if err != nil {
			return nil, err
		}

		if form == nil {
			continue
		}

		formTitleArray = append(formTitleArray, &model.FormTitle{
			ID:        form.ID,
			Title:     form.Title,
			CreatedAt: form.CreatedAt,
		})
	}

	return formTitleArray, nil
}

type fromRowReturn struct {
	form     *Form
	author   *User
	question *Question
	answer   *Answer
}

func (r *formDatabaseRepository) formTitleFromRow(row pgx.Row) (*Form, error) {
	form := &Form{}

	err := row.Scan(
		&form.ID,
		&form.Title,
		&form.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("form_repository failed to scan row: %e", err)
	}

	return form, nil
}

func (r *formDatabaseRepository) fromRow(row pgx.Row) (*fromRowReturn, error) {
	form := &Form{}
	author := &User{}
	question := &Question{}
	answer := &Answer{}

	err := row.Scan(
		&form.ID,
		&form.Title,
		&form.CreatedAt,
		&form.AuthorID,
		&author.ID,
		&author.Username,
		&author.FirstName,
		&author.LastName,
		&author.Email,
		&question.ID,
		&question.Title,
		&question.Text,
		&question.Type,
		&question.Required,
		&answer.ID,
		&answer.AnswerText,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("form_repository failed to scan row: %e", err)
	}

	return &fromRowReturn{form, author, question, answer}, nil
}
