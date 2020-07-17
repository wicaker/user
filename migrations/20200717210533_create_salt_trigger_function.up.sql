CREATE OR REPLACE FUNCTION trigger_salt_update()   
RETURNS TRIGGER AS $$
BEGIN
    NEW.salt = digest(NEW.uuid || random()::text || clock_timestamp()::text, 'sha1');
    RETURN NEW;   
END;
$$ language 'plpgsql';