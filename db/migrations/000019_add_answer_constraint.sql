alter table passage_answer
add constraint chk_PassageAnswer
check (valid_passage_answer(question_id, answer_text) = true);
