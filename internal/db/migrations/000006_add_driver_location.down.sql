ALTER TABLE taskflow.subject
    DROP COLUMN IF EXISTS last_known_latitude,
    DROP COLUMN IF EXISTS last_known_longitude,