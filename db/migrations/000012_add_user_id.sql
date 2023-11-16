ALTER TABLE nofronts.passage_answer
ADD user_id BIGINT REFERENCES nofronts.user(id) ON DELETE CASCADE;
