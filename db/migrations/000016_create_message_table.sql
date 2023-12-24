CREATE TABLE message (
    id SERIAL PRIMARY KEY,
    -- form_id BIGINT NOT NULL REFERENCES nofronts.form(id) ON DELETE CASCADE,
    sender_id BIGINT NOT NULL REFERENCES nofronts.user(id),
    receiver_id BIGINT NOT NULL REFERENCES nofronts.user(id),
    text TEXT,
    is_read BOOLEAN DEFAULT FALSE,
    send_at TIMESTAMPTZ DEFAULT (NOW() AT TIME ZONE 'utc')
);