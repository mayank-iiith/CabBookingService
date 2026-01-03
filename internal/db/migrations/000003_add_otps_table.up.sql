-- 1. Create OTPs Table
CREATE TABLE otps (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ,

    code VARCHAR(10) NOT NULL,
    sent_to_number VARCHAR(20) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);

-- 2. Add OTP Foreign Key to Bookings Table
ALTER TABLE bookings
    ADD COLUMN ride_start_otp_id UUID REFERENCES otps(id) ON DELETE SET NULL;