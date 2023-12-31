CREATE TABLE nofronts.form(
	id BIGSERIAL PRIMARY KEY,
	author_id BIGINT NOT NULL REFERENCES nofronts.user(id),
	title VARCHAR(255) NOT NULL,
	anonymous boolean NOT NULL DEFAULT FALSE,
	description text,
	created_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc')
);
