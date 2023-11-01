CREATE TABLE nofronts.answer (
    id BIGSERIAL PRIMARY KEY,
    question_id BIGINT NOT NULL REFERENCES nofronts.question(id),
    answer_text TEXT NOT NULL,
    UNIQUE (answer_text, question_id)
);
