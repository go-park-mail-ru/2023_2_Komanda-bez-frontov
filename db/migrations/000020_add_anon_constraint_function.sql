create or replace function valid_anon_passage(passage_question_id bigint, user_id bigint)
returns boolean
language plpgsql
as
$$
declare
   passage_form_id bigint;
   anon boolean;
begin
 select form_id
 into passage_form_id
 from question
 where id = passage_question_id;

 select anonymous
 into anon
 from form
 where id = passage_form_id;

 if anon = true and user_id is null then
    return true;
 elsif anon = false and user_id is not null then
    return true;
end if;
 return false;
end;
$$;
