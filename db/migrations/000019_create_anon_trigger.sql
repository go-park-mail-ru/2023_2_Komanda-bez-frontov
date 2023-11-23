CREATE or replace TRIGGER anon_form
  BEFORE INSERT
  ON passage_answer
  FOR EACH ROW
  EXECUTE PROCEDURE make_anon();
