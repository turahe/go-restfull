package valueobjects

import (
	"errors"
	"strings"
	"time"
)

// Gender represents user gender
type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
	GenderOther  Gender = "other"
)

// UserProfile represents a user profile value object
type UserProfile struct {
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	FullName    string     `json:"full_name"`
	Avatar      string     `json:"avatar,omitempty"`
	Bio         string     `json:"bio,omitempty"`
	Website     string     `json:"website,omitempty"`
	Location    string     `json:"location,omitempty"`
	Gender      Gender     `json:"gender,omitempty"`
	DateOfBirth *time.Time `json:"date_of_birth,omitempty"`
}

// NewUserProfile creates a new user profile value object
func NewUserProfile(firstName, lastName, avatar, bio, website, location string, gender Gender, dateOfBirth *time.Time) (UserProfile, error) {
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)
	bio = strings.TrimSpace(bio)
	website = strings.TrimSpace(website)
	location = strings.TrimSpace(location)

	if firstName == "" {
		return UserProfile{}, errors.New("first name cannot be empty")
	}

	if lastName == "" {
		return UserProfile{}, errors.New("last name cannot be empty")
	}

	if len(firstName) > 50 {
		return UserProfile{}, errors.New("first name cannot exceed 50 characters")
	}

	if len(lastName) > 50 {
		return UserProfile{}, errors.New("last name cannot exceed 50 characters")
	}

	if len(bio) > 500 {
		return UserProfile{}, errors.New("bio cannot exceed 500 characters")
	}

	if len(website) > 255 {
		return UserProfile{}, errors.New("website cannot exceed 255 characters")
	}

	if len(location) > 100 {
		return UserProfile{}, errors.New("location cannot exceed 100 characters")
	}

	if gender != "" && gender != GenderMale && gender != GenderFemale && gender != GenderOther {
		return UserProfile{}, errors.New("invalid gender value")
	}

	// Validate date of birth (must be in the past and reasonable)
	if dateOfBirth != nil {
		now := time.Now()
		if dateOfBirth.After(now) {
			return UserProfile{}, errors.New("date of birth cannot be in the future")
		}
		
		// Check if age is reasonable (not more than 150 years old)
		if now.Sub(*dateOfBirth).Hours() > 150*365*24 {
			return UserProfile{}, errors.New("date of birth is not reasonable")
		}
	}

	fullName := strings.TrimSpace(firstName + " " + lastName)

	return UserProfile{
		FirstName:   firstName,
		LastName:    lastName,
		FullName:    fullName,
		Avatar:      avatar,
		Bio:         bio,
		Website:     website,
		Location:    location,
		Gender:      gender,
		DateOfBirth: dateOfBirth,
	}, nil
}

// GetAge returns the user's age based on date of birth
func (p UserProfile) GetAge() *int {
	if p.DateOfBirth == nil {
		return nil
	}

	now := time.Now()
	age := int(now.Sub(*p.DateOfBirth).Hours() / (365 * 24))
	return &age
}

// Equals checks if two user profiles are equal
func (p UserProfile) Equals(other UserProfile) bool {
	return p.FirstName == other.FirstName &&
		p.LastName == other.LastName &&
		p.Avatar == other.Avatar &&
		p.Bio == other.Bio &&
		p.Website == other.Website &&
		p.Location == other.Location &&
		p.Gender == other.Gender &&
		((p.DateOfBirth == nil && other.DateOfBirth == nil) ||
			(p.DateOfBirth != nil && other.DateOfBirth != nil && p.DateOfBirth.Equal(*other.DateOfBirth)))
}