CREATE TABLE nofronts.question (
    id BIGSERIAL PRIMARY KEY,
    form_id BIGINT NOT NULL REFERENCES nofronts.form(id),
    question_type INTEGER,
    question_title VARCHAR NOT NULL,
    question_text TEXT NOT NULL DEFAULT '',
    shuffle BOOLEAN DEFAULT FALSE
);
