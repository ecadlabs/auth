BEGIN;

CREATE FUNCTION ip_overlap_check() RETURNS trigger AS $$
    DECLARE
        ip inet;
    BEGIN
        SELECT service_account_ip.addr INTO ip FROM service_account_ip WHERE service_account_ip.addr && NEW.addr;
        IF FOUND THEN
            RAISE EXCEPTION '% conflicts with existing network or host address %', NEW.addr, ip USING ERRCODE = 'integrity_constraint_violation', CONSTRAINT = 'service_account_ip_addr_key';
        END IF;
        RETURN NEW;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER ip_overlap_trigger
    BEFORE INSERT OR UPDATE ON service_account_ip
    FOR EACH ROW
    EXECUTE PROCEDURE ip_overlap_check();

COMMIT;
