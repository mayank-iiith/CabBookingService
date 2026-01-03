-- 1. Create Reviews Table
CREATE TABLE IF NOT EXISTS reviews (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ,

    booking_id UUID NOT NULL,
    rating INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    note TEXT,
    reviewer_role VARCHAR(50)
);

-- 2. Add Foreign Keys to Bookings
ALTER TABLE bookings ADD COLUMN review_by_passenger_id UUID REFERENCES reviews(id);
ALTER TABLE bookings ADD COLUMN review_by_driver_id UUID REFERENCES reviews(id);

-- 3. Add Rating Columns to Profiles
ALTER TABLE passengers ADD COLUMN average_rating DOUBLE PRECISION DEFAULT 0;
ALTER TABLE passengers ADD COLUMN rating_count INT DEFAULT 0;

ALTER TABLE drivers ADD COLUMN average_rating DOUBLE PRECISION DEFAULT 0;
ALTER TABLE drivers ADD COLUMN rating_count INT DEFAULT 0;