package models

import (
	"time"
)

type Gender string

const (
	GenderMale   Gender = "MALE"
	GenderFemale Gender = "FEMALE"
)

type Profile struct {
	ID          int64
	Name        string
	Surname     string
	DateOfBirth time.Time
	Gender      Gender
	Interests   string
	City        string
	Bio         string
	OwnerID     int64
}

type ProfileDTO struct {
	Name        string    `json:"name"`
	Surname     string    `json:"surname"`
	DateOfBirth time.Time `json:"date_of_birth"`
	Gender      Gender    `json:"gender"`
	Interests   string    `json:"interests"`
	City        string    `json:"city"`
	Bio         string    `json:"bio"`
}

type GetProfileDTO struct {
	ProfileDTO
	ProfileID int64 `json:"profile_id"`
}

type QueryDTO struct {
	Query   string `json:"query"`
	Gender  Gender `json:"gender"`
	City    string `json:"city"`
	AgeFrom int    `json:"age_from"`
	AgeTo   int    `json:"age_to"`
	Count   int    `json:"count"`
}
