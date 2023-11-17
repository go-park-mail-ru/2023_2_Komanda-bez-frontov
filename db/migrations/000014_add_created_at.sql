ALTER TABLE nofronts.passage_answer
ADD created_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'utc');
