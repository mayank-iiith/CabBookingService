ALTER TABLE bookings
    DROP COLUMN IF EXISTS scheduled_time;

DROP INDEX IF EXISTS idx_bookings_scheduled_time;