create or replace function make_anon()
returns trigger
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
 where id = new.question_id;

 select anonymous
 into anon
 from form
 where id = passage_form_id;

 if anon = true then
    new.user_id := null;
 end if;
 return new;
end;
$$;
