ALTER TABLE drivers DROP COLUMN IF EXISTS average_rating;
ALTER TABLE drivers DROP COLUMN IF EXISTS rating_count;
ALTER TABLE passengers DROP COLUMN IF EXISTS average_rating;
ALTER TABLE passengers DROP COLUMN IF EXISTS rating_count;
ALTER TABLE bookings DROP COLUMN IF EXISTS review_by_passenger_id;
ALTER TABLE bookings DROP COLUMN IF EXISTS review_by_driver_id;
DROP TABLE IF EXISTS reviews;