package models

// Gender represents the gender of a person
type Gender string

const (
	GenderMale   Gender = "MALE"
	GenderFemale Gender = "FEMALE"
	GenderOther  Gender = "OTHER"
)

func (gender Gender) String() string {
	return string(gender)
}

// BookingStatus defines the status of a booking
type BookingStatus string

const (
	BookingStatusRequested BookingStatus = "REQUESTED"
	BookingStatusAccepted  BookingStatus = "ACCEPTED"
	BookingStatusStarted   BookingStatus = "STARTED"
	BookingStatusCompleted BookingStatus = "COMPLETED"
	BookingStatusCancelled BookingStatus = "CANCELLED"
	BookingStatusScheduled BookingStatus = "SCHEDULED"
)

func (b BookingStatus) String() string {
	return string(b)
}

func (b BookingStatus) IsCancellable() bool {
	// Only REQUESTED and ACCEPTED bookings can be cancelled
	return b == BookingStatusRequested || b == BookingStatusAccepted
}
