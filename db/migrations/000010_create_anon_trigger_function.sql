create or replace function make_anon()
returns trigger
language plpgsql
as
$$
declare
   anon boolean;
begin
 select anonymous
 into anon
 from form
 where id = new.form_id;

 if anon = true then
    new.user_id := null;
 end if;
 return new;
end;
$$;
