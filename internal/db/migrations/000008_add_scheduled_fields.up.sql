ALTER TABLE bookings
    ADD COLUMN IF NOT EXISTS scheduled_time TIMESTAMPTZ;

-- Partial index to optimize queries for scheduled bookings
CREATE INDEX idx_bookings_scheduled_time ON bookings(scheduled_time) WHERE status = 'SCHEDULED';