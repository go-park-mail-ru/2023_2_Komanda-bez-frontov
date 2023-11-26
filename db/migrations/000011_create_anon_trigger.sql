CREATE or replace TRIGGER anon_form
  BEFORE INSERT
  ON form_passage
  FOR EACH ROW
  EXECUTE PROCEDURE make_anon();
