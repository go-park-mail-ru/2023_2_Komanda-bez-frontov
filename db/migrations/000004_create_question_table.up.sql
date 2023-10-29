CREATE TABLE nofronts.question (
    id BIGSERIAL PRIMARY KEY,
    form_id BIGINT NOT NULL REFERENCES nofronts.form(id),
    type INTEGER NOT NULL CHECK (type IN (1, 2, 3)),
    title VARCHAR NOT NULL,
    text TEXT NOT NULL,
    shuffle BOOLEAN NOT NULL DEFAULT FALSE
);
