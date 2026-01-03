-- Migration to add last known location columns to drivers table
ALTER TABLE drivers
    ADD COLUMN IF NOT EXISTS last_known_latitude DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS last_known_longitude DOUBLE PRECISION;