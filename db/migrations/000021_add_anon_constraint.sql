alter table passage_answer
add constraint chk_PassageAnswerAnon
check (valid_anon_passage(question_id, user_id) = true);
