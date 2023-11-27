CREATE TABLE nofronts.form_passage (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES nofronts.user(id),
    form_id BIGINT NOT NULL REFERENCES nofronts.form(id) ON DELETE CASCADE,
    finished_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC')
);
