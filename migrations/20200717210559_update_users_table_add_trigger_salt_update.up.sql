CREATE TRIGGER update_salt BEFORE UPDATE ON users FOR EACH ROW EXECUTE PROCEDURE trigger_salt_update();