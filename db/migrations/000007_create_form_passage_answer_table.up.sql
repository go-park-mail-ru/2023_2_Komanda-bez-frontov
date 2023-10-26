CREATE TABLE nofronts.form_passage_answer (
    id BIGSERIAL PRIMARY KEY,
    form_passage_id BIGINT NOT NULL REFERENCES nofronts.form_passage(id),
    question_id BIGINT NOT NULL REFERENCES nofronts.question(id),
    answer_text TEXT NOT NULL
);
