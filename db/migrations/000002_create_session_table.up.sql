CREATE TABLE nofronts.session(
	id VARCHAR(128) PRIMARY KEY,
	created_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc'),
	user_id BIGINT REFERENCES nofronts.user(id)
);
