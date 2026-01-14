package domain

import "time"

type Patient struct {
	ID    int
	Name  string
	Email string
	Phone string
}

type Appointment struct {
	ID        int
	PatientID int
	StartTime time.Time
	Duration  time.Duration
}

type ScheduleConfig struct {
	DayOfWeek time.Weekday // time.Monday, time.Tuesday, etc...
	StartHour int
	EndHour   int
}

// Slot representa un hueco disponible para mostrar en el frontend
type Slot struct {
	Start     time.Time
	Available bool
}
