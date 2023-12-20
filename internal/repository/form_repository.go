package repository

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"go-form-hub/internal/database"
	"go-form-hub/internal/model"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

type Form struct {
	Title       string    `db:"title"`
	ID          int64     `db:"id"`
	Description *string   `db:"description"`
	Anonymous   bool      `db:"anonymous"`
	PassageMax  int64     `db:"passage_max"`
	AuthorID    int64     `db:"author_id"`
	CreatedAt   time.Time `db:"created_at"`
}

var (
	selectFields = []string{
		"f.id",
		"f.title",
		"f.description",
		"f.created_at",
		"f.author_id",
		"f.anonymous",
		"f.passage_max",
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
	selectFieldsFormInfo = []string{
		"f.id",
		"f.title",
		"f.created_at",
		"COALESCE(f.description, '')",
		"f.anonymous",
		"f.passage_max",
		"u.id",
		"u.username",
		"u.first_name",
		"u.last_name",
		"u.email",
		"q.id",
		"COALESCE(q.title, '')",
		"q.text",
		"q.type",
		"COALESCE(a.answer_text, '')",
	}
	selectFieldsFormPassageInfo = []string{
		"fp.id",
		"ua.id",
		"COALESCE(ua.username, '')",
		"COALESCE(ua.first_name, '')",
		"COALESCE(ua.last_name, '')",
		"COALESCE(ua.email, '')",
		"q.id",
		"pa.answer_text",
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

func (r *formDatabaseRepository) FindAll(ctx context.Context) (forms []*model.FormTitle, err error) {
	query, args, err := r.builder.
		Select("id, title, created_at, count(fp.id) as number_of_passages").
		From(fmt.Sprintf("%s.form as f", r.db.GetSchema())).
		LeftJoin(fmt.Sprintf("%s.form_passage as fp ON fp.form_id = f.id", r.db.GetSchema())).
		GroupBy("f.id").
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

	return r.searchTitleFromRows(rows)
}

func (r *formDatabaseRepository) FormsSearch(ctx context.Context, title string, userID uint) (forms []*model.FormTitle, err error) {
	const limit = 5
	query := fmt.Sprintf(`SELECT id, title, created_at, number_of_passages
		FROM (
		  SELECT f.title as title, f.id as id, f.created_at as created_at, COUNT(fp.id) as number_of_passages, similarity(f.title, $1::text) as sim
		  FROM %s.form as f
		  LEFT JOIN %s.form_passage  as fp ON fp.form_id = f.id
		  WHERE f.author_id = $2::integer
		  GROUP BY f.id
		  ORDER BY sim DESC, f.created_at
		) AS res
		WHERE sim > 0
		LIMIT $3::integer`, r.db.GetSchema(), r.db.GetSchema())

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

func (r *formDatabaseRepository) FormResultsCsv(ctx context.Context, id int64) ([]byte, error) {
	form, err := r.FormResults(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("form_repository form_results_exel failed to run FormResults: %e", err)
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	formRow := []string{
		form.Title,
	}

	for _, question := range form.Questions {
		questionRow := []string{
			question.Title,
			fmt.Sprint(question.NumberOfPassagesQuestion),
		}

		for _, answer := range question.Answers {
			answerRow := []string{
				answer.Text,
				fmt.Sprint(answer.SelectedTimesAnswer),
			}

			questionRow = append(questionRow, answerRow...)
		}

		formRow = append(formRow, questionRow...)
	}

	err = writer.Write(formRow)
	if err != nil {
		return nil, fmt.Errorf("error writing to CSV: %e", err)
	}
	writer.Flush()

	return buf.Bytes(), nil
}

func (r *formDatabaseRepository) FormResultsExel(ctx context.Context, id int64) ([]byte, error) {
	form, err := r.FormResults(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("form_repository form_results_exel failed to run FormResults: %e", err)
	}

	excelFile, err := generateExcelFile(form)
	if err != nil {
		return nil, err
	}

	return excelFile, nil
}

func generateExcelFile(form *model.FormResult) ([]byte, error) {
	file := excelize.NewFile()

	fillExcelFile(file, form)

	buf, err := file.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func fillExcelFile(file *excelize.File, form *model.FormResult) {
	file.SetCellValue("Sheet1", "A1", "Form Name")
	file.SetCellValue("Sheet1", "B1", form.Title)

	file.SetCellValue("Sheet1", "A2", "Description")
	file.SetCellValue("Sheet1", "B2", form.Description)

	row := 4
	qcounter := 1

	for _, question := range form.Questions {
		file.SetCellValue("Sheet1", fmt.Sprintf("A%d", row), fmt.Sprintf("Question%d", qcounter))
		file.SetCellValue("Sheet1", fmt.Sprintf("B%d", row), question.Title)
		file.SetCellValue("Sheet1", fmt.Sprintf("C%d", row), fmt.Sprintf("NumberOfPassagesQuestion %d", question.NumberOfPassagesQuestion))
		row++

		acounter := 1
		for _, answer := range question.Answers {
			file.SetCellValue("Sheet1", fmt.Sprintf("B%d", row), fmt.Sprintf("Answer%d", acounter))
			file.SetCellValue("Sheet1", fmt.Sprintf("C%d", row), answer.Text)
			file.SetCellValue("Sheet1", fmt.Sprintf("D%d", row), fmt.Sprintf("SelectedTimesAnswer %d", answer.SelectedTimesAnswer))

			row++
			acounter++
		}
		qcounter++
	}
}

func (r *formDatabaseRepository) FormResults(ctx context.Context, id int64) (formResult *model.FormResult, err error) {
	formInfoQuery, formInfoArgs, err := r.builder.
		Select(selectFieldsFormInfo...).
		From(fmt.Sprintf("%s.form as f", r.db.GetSchema())).
		Join(fmt.Sprintf("%s.user as u ON f.author_id = u.id", r.db.GetSchema())).
		LeftJoin(fmt.Sprintf("%s.question as q ON q.form_id = f.id", r.db.GetSchema())).
		LeftJoin(fmt.Sprintf("%s.answer as a ON a.question_id = q.id", r.db.GetSchema())).
		Where(squirrel.Eq{"f.id": id}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("form_repository form_results failed to build form info query: %e", err)
	}

	formPassageInfoQuery, formPassageInfoArgs, err := r.builder.
		Select(selectFieldsFormPassageInfo...).
		From(fmt.Sprintf("%s.form_passage as fp", r.db.GetSchema())).
		LeftJoin(fmt.Sprintf("%s.user as ua ON fp.user_id = ua.id", r.db.GetSchema())).
		Join(fmt.Sprintf("%s.form_passage_answer as pa ON fp.id = pa.form_passage_id", r.db.GetSchema())).
		Join(fmt.Sprintf("%s.question as q ON pa.question_id = q.id", r.db.GetSchema())).
		Where(squirrel.Eq{"fp.form_id": id}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("form_repository form_passage_results failed to build form passage info query: %e", err)
	}

	formPassageCount, formPassageArgs, err := r.builder.
		Select("fp.form_id", "COUNT(DISTINCT fp.id) AS unique_response_count").
		From(fmt.Sprintf("%s.form_passage as fp", r.db.GetSchema())).
		Join(fmt.Sprintf("%s.form_passage_answer as pa ON fp.id = pa.form_passage_id", r.db.GetSchema())).
		Join(fmt.Sprintf("%s.question as q ON pa.question_id = q.id", r.db.GetSchema())).
		Where(squirrel.Eq{"fp.form_id": id}).
		GroupBy("fp.form_id").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("form_repository question_count failed to build form passage info query: %e", err)
	}

	questionPassageCount, questionPassageArgs, err := r.builder.
		Select("q.id", "COUNT(DISTINCT fp.id) AS unique_response_count").
		From(fmt.Sprintf("%s.form_passage as fp", r.db.GetSchema())).
		Join(fmt.Sprintf("%s.form_passage_answer as pa ON fp.id = pa.form_passage_id", r.db.GetSchema())).
		Join(fmt.Sprintf("%s.question as q ON pa.question_id = q.id", r.db.GetSchema())).
		Where(squirrel.Eq{"fp.form_id": id}).
		GroupBy("q.id").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("form_repository question_count failed to build form passage info query: %e", err)
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

	countQuestion, err := tx.Query(ctx, questionPassageCount, questionPassageArgs...)
	if err != nil {
		return nil, fmt.Errorf("form_repository question_count failed to execute form info query: %e", err)
	}

	countQuestionResults, err := r.countQuestionFromRows(countQuestion)
	if err != nil {
		return nil, err
	}

	countForm, err := tx.Query(ctx, formPassageCount, formPassageArgs...)
	if err != nil {
		return nil, fmt.Errorf("form_repository form_count failed to execute form info query: %e", err)
	}

	countFormResults, err := r.countFormFromRows(countForm)
	if err != nil {
		return nil, err
	}

	rowsFormInfo, err := tx.Query(ctx, formInfoQuery, formInfoArgs...)
	if err != nil {
		return nil, fmt.Errorf("form_repository form_results failed to execute form info query: %e", err)
	}

	formResults, err := r.formResultsFromRows(rowsFormInfo)
	if err != nil {
		return nil, err
	}

	if len(formResults) == 0 {
		return nil, nil
	}

	rowsFormPassageInfo, err := tx.Query(ctx, formPassageInfoQuery, formPassageInfoArgs...)
	if err != nil {
		return nil, fmt.Errorf("form_repository form_results failed to execute form passage info query: %e", err)
	}

	formPassageResults, err := r.formPassageResultsFromRows(rowsFormPassageInfo)
	if err != nil {
		return nil, err
	}

	for _, formPassageResult := range formPassageResults {
		formResult := formResults[0]
		for _, formCount := range countFormResults {
			if formCount.ID == formResult.ID {
				formResult.NumberOfPassagesForm = formCount.NumberOfPassagesForm
			}
		}
		for _, questionResult := range formResult.Questions {
			if questionResult.ID == formPassageResult.QuestionID {
				for _, questionCount := range countQuestionResults {
					if questionCount.ID == questionResult.ID {
						questionResult.NumberOfPassagesQuestion = questionCount.NumberOfPassagesQuestion
					}
				}
				answerExist := false
				for _, answerResult := range questionResult.Answers {
					if answerResult.Text == formPassageResult.AnswerText {
						answerResult.SelectedTimesAnswer++
						answerExist = true
						break
					}
				}
				if !answerExist {
					questionResult.Answers = append(questionResult.Answers, &model.AnswerResult{
						Text:                formPassageResult.AnswerText,
						SelectedTimesAnswer: 1,
					})
				}
			}
		}
		if !formResult.Anonymous {
			userExist := false
			for _, partisipantsResult := range formResult.Participants {
				if partisipantsResult.ID == formPassageResult.UserID.Int64 {
					userExist = true
					break
				}
			}
			if !userExist {
				formResult.Participants = append(formResult.Participants, &model.UserGet{
					ID:        formPassageResult.UserID.Int64,
					FirstName: formPassageResult.FirstName,
					Username:  formPassageResult.Username,
				})
			}
		}
	}
	return formResults[0], nil
}

func (r *formDatabaseRepository) formResultsFromRows(rows pgx.Rows) ([]*model.FormResult, error) {
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
			for _, existingAnswer := range existingQuestion.Answers {
				if existingAnswer == existingQuestion.Answers[len(existingQuestion.Answers)-1] {
					if existingAnswer.Text != "" {
						existingQuestion.Answers = append(existingQuestion.Answers, info.answerResult)
						if _, ok := answersByQuestionID[info.questionResult.ID]; !ok {
							answersByQuestionID[info.questionResult.ID] = make([]*model.AnswerResult, 0)
						}
						answersByQuestionID[info.questionResult.ID] = append(answersByQuestionID[info.questionResult.ID], info.answerResult)
					}
				}
			}
		} else {
			formResultMap[info.formResult.ID].Questions = append(formResultMap[info.formResult.ID].Questions, info.questionResult)

			info.questionResult.Answers = append(info.questionResult.Answers, info.answerResult)

			if _, ok := questionsByFormID[info.formResult.ID]; !ok {
				questionsByFormID[info.formResult.ID] = make([]*model.QuestionResult, 0)
			}

			questionsByFormID[info.formResult.ID] = append(questionsByFormID[info.formResult.ID], info.questionResult)
			if info.answerResult.Text != "" {
				if _, ok := answersByQuestionID[info.questionResult.ID]; !ok {
					answersByQuestionID[info.questionResult.ID] = make([]*model.AnswerResult, 0)
				}
				answersByQuestionID[info.questionResult.ID] = append(answersByQuestionID[info.questionResult.ID], info.answerResult)
			}
		}
	}

	formResults := make([]*model.FormResult, 0, len(formResultMap))

	for _, formResult := range formResultMap {
		formResult.Questions = questionsByFormID[formResult.ID]
		for _, questionResult := range formResult.Questions {
			questionResult.Answers = answersByQuestionID[questionResult.ID]
		}

		formResults = append(formResults, formResult)
	}

	return formResults, nil
}

func (r *formDatabaseRepository) formPassageResultsFromRows(rows pgx.Rows) ([]*model.FormPassageResult, error) {
	defer func() {
		rows.Close()
	}()

	formPassageResults := make([]*model.FormPassageResult, 0)

	for rows.Next() {
		result := &model.FormPassageResult{}
		err := rows.Scan(
			&result.FormID,
			&result.UserID,
			&result.Username,
			&result.FirstName,
			&result.LastName,
			&result.Email,
			&result.QuestionID,
			&result.AnswerText,
		)
		if err != nil {
			return nil, fmt.Errorf("form_repository formPassageResultsFromRows failed to scan row: %v", err)
		}
		formPassageResults = append(formPassageResults, result)
	}

	return formPassageResults, nil
}

func (r *formDatabaseRepository) countQuestionFromRows(rows pgx.Rows) ([]*model.QuestionResult, error) {
	defer func() {
		rows.Close()
	}()

	countQuestionPassageResults := make([]*model.QuestionResult, 0)

	for rows.Next() {
		result := &model.QuestionResult{}
		err := rows.Scan(
			&result.ID,
			&result.NumberOfPassagesQuestion,
		)
		if err != nil {
			return nil, fmt.Errorf("form_repository countQuestionFromRows failed to scan row: %v", err)
		}
		countQuestionPassageResults = append(countQuestionPassageResults, result)
	}

	return countQuestionPassageResults, nil
}

func (r *formDatabaseRepository) countFormFromRows(rows pgx.Rows) ([]*model.FormResult, error) {
	defer func() {
		rows.Close()
	}()

	countFormPassageResults := make([]*model.FormResult, 0)

	for rows.Next() {
		result := &model.FormResult{}
		err := rows.Scan(
			&result.ID,
			&result.NumberOfPassagesForm,
		)
		if err != nil {
			return nil, fmt.Errorf("form_repository countFormFromRows failed to scan row: %v", err)
		}
		countFormPassageResults = append(countFormPassageResults, result)
	}

	return countFormPassageResults, nil
}

type formResultsFromRowReturn struct {
	formResult     *model.FormResult
	questionResult *model.QuestionResult
	answerResult   *model.AnswerResult
}

// formResultsFromRowReturn представляет структурированные данные,
// возвращаемые при обработке одной строки в результате запроса к базе данных.
// Она включает информацию о форме, вопросе, ответе и участнике.
func (r *formDatabaseRepository) formResultsFromRow(row pgx.Row) (*formResultsFromRowReturn, error) {
	formResult := &model.FormResult{}
	questionResult := &model.QuestionResult{}
	answerResult := &model.AnswerResult{}
	formResult.Author = &model.UserGet{}

	err := row.Scan(
		&formResult.ID,
		&formResult.Title,
		&formResult.CreatedAt,
		&formResult.Description,
		&formResult.Anonymous,
		&formResult.PassageMax,
		&formResult.Author.ID,
		&formResult.Author.Username,
		&formResult.Author.FirstName,
		&formResult.Author.LastName,
		&formResult.Author.Email,
		&questionResult.ID,
		&questionResult.Title,
		&questionResult.Description,
		&questionResult.Type,
		&answerResult.Text,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("form_repository form_results failed to scan row: %e", err)
	}

	return &formResultsFromRowReturn{formResult, questionResult, answerResult}, nil
}

func (r *formDatabaseRepository) FindAllByUser(ctx context.Context, username string) (forms []*model.FormTitle, err error) {
	query, args, err := r.builder.
		Select("f.id, f.title, f.created_at, count(fp.id) as number_of_passages").
		From(fmt.Sprintf("%s.form as f", r.db.GetSchema())).
		Join(fmt.Sprintf("%s.user as u ON f.author_id = u.id", r.db.GetSchema())).
		LeftJoin(fmt.Sprintf("%s.form_passage as fp ON fp.form_id = f.id", r.db.GetSchema())).
		Where(squirrel.Eq{"u.username": username}).
		GroupBy("f.id").
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

	return r.searchTitleFromRows(rows)
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
		Columns("title", "author_id", "created_at", "description", "anonymous", "passage_max").
		Values(form.Title, form.Author.ID, form.CreatedAt, form.Description, form.Anonymous, form.PassageMax).
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

func (r *formDatabaseRepository) FormPassageSave(ctx context.Context, formPassage *model.FormPassage, userID uint64) error {
	var err error

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("form_facade insert failed to begin transaction: %e", err)
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit(ctx)
		default:
			_ = tx.Rollback(ctx)
		}
	}()

	formPassageQuery := fmt.Sprintf(`INSERT INTO %s.form_passage
	(user_id, form_id)
	VALUES($1::integer, $2::integer)
	RETURNING id`, r.db.GetSchema())

	var formPassageID int64
	err = tx.QueryRow(ctx, formPassageQuery, userID, formPassage.FormID).Scan(&formPassageID)
	if err != nil {
		return err
	}

	passageAnswerBatch := &pgx.Batch{}
	passageAnswerQuery := fmt.Sprintf(`INSERT INTO %s.form_passage_answer
	(answer_text, question_id, form_passage_id)
	VALUES($1::text, $2::integer, $3::integer)`, r.db.GetSchema())

	for _, passageAnswer := range formPassage.PassageAnswers {
		passageAnswerBatch.Queue(passageAnswerQuery, passageAnswer.Text,
			passageAnswer.QuestionID, formPassageID)
	}
	answerBatch := tx.SendBatch(ctx, passageAnswerBatch)
	answerBatch.Close()

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *formDatabaseRepository) FormPassageCount(ctx context.Context, formID int64) (int64, error) {
	var err error

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("form_facade insert failed to begin transaction: %e", err)
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit(ctx)
		default:
			_ = tx.Rollback(ctx)
		}
	}()

	formPassageQuery := fmt.Sprintf(`select count(*)
	from %s.form_passage
	where form_id = $1`, r.db.GetSchema())

	var total int64
	err = tx.QueryRow(ctx, formPassageQuery, formID).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (r *formDatabaseRepository) UserFormPassageCount(ctx context.Context, formID, userID int64) (int64, error) {
	var err error

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("form_facade insert failed to begin transaction: %e", err)
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit(ctx)
		default:
			_ = tx.Rollback(ctx)
		}
	}()

	formPassageQuery := fmt.Sprintf(`select count(*)
	from %s.form_passage
	where form_id = $1 and user_id = $2`, r.db.GetSchema())

	var total int64
	err = tx.QueryRow(ctx, formPassageQuery, formID, userID).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (r *formDatabaseRepository) Update(ctx context.Context, id int64, form *model.FormUpdate) (result *model.FormUpdate, err error) {
	query, args, err := r.builder.Update(fmt.Sprintf("%s.form", r.db.GetSchema())).
		Set("title", form.Title).
		Set("description", form.Description).
		Set("anonymous", form.Anonymous).
		Set("passage_max", form.PassageMax).
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
				ID:          &info.form.ID,
				Title:       info.form.Title,
				Description: info.form.Description,
				Anonymous:   info.form.Anonymous,
				PassageMax:  int(info.form.PassageMax),
				CreatedAt:   info.form.CreatedAt,
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
			ID:                   form.ID,
			Title:                form.Title,
			CreatedAt:            form.CreatedAt,
			NumberOfPassagesForm: form.NumberOfPassagesForm,
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

func (r *formDatabaseRepository) formTitleFromRow(row pgx.Row) (*model.FormTitle, error) {
	form := &model.FormTitle{}

	err := row.Scan(
		&form.ID,
		&form.Title,
		&form.CreatedAt,
		&form.NumberOfPassagesForm,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("form_repository form Title failed to scan row: %e", err)
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
		&form.Description,
		&form.CreatedAt,
		&form.AuthorID,
		&form.Anonymous,
		&form.PassageMax,
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

		return nil, fmt.Errorf("form_repository from_row failed to scan row: %e", err)
	}

	return &fromRowReturn{form, author, question, answer}, nil
}
