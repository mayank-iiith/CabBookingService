-- This table tracks which drivers were offered a specific booking.
-- It implements a Many-to-Many relationship between Bookings and Drivers.

CREATE TABLE IF NOT EXISTS booking_notified_drivers (
    booking_id UUID NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    driver_id UUID NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (booking_id, driver_id)
);

-- Index for fast lookups during AcceptBooking (Where booking_id = ? AND driver_id = ?)
CREATE INDEX IF NOT EXISTS idx_booking_notified_lookup ON booking_notified_drivers(booking_id, driver_id);