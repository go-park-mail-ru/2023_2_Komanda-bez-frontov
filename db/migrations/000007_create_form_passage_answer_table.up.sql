CREATE TABLE nofronts.form_passage_answer (
    id BIGSERIAL PRIMARY KEY,
    form_passage_id BIGINT NOT NULL REFERENCES nofronts.form_passage(id) ON DELETE CASCADE,
    question_id BIGINT NOT NULL REFERENCES nofronts.question(id) ON DELETE CASCADE,
    answer_text TEXT NOT NULL
);
