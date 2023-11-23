create or replace function valid_passage_answer(passage_question_id bigint, passage_answer text)
returns boolean
language plpgsql
as
$$
declare
   question_type integer;
   total integer;
  answer_row record;
begin
   select type
   into question_type
   from question
   where id = passage_question_id;

   if question_type = 1 then
	return true;
  elsif question_type = 2 then
  	for answer_row in select answer_text from answer where answer.question_id = passage_question_id
  	loop
	  	if answer_row.answer_text = passage_answer then
	  		return true;
	  	end if;
  	end loop;
	return false;
  elsif question_type = 3 then
  	for answer_row in select answer_text from answer where answer.question_id = passage_question_id
  	loop
	  	if answer_row.answer_text = passage_answer then
	  		return true;
	  	end if;
  	end loop;
	return false;
  end if;
 return true;
end;
$$;
