package user

import "time"

type UserProfile struct {
	bio      *string
	location *string
	website  *string
	birthday *time.Time

	createdAt time.Time
	updatedAt time.Time
}

func NewProfile() *UserProfile {
	now := time.Now()

	return &UserProfile{
		createdAt: now,
		updatedAt: now,
	}
}

func (p *UserProfile) Update(
	bio, location, website *string,
	birthday *time.Time,
) {
	if bio != nil {
		p.bio = bio
	}

	if location != nil {
		p.location = location
	}

	if website != nil {
		p.website = website
	}

	if birthday != nil {
		p.birthday = birthday
	}

	p.updatedAt = time.Now()
}

func (p *UserProfile) Bio() *string {
	return p.bio
}

func (p *UserProfile) Location() *string {
	return p.location
}

func (p *UserProfile) Website() *string {
	return p.website
}

func (p *UserProfile) Birthday() *time.Time {
	return p.birthday
}

func (p *UserProfile) CreatedAt() time.Time {
	return p.createdAt
}

func (p *UserProfile) UpdatedAt() time.Time {
	return p.updatedAt
}

func RestoreProfile(
	bio *string,
	location *string,
	website *string,
	birthday *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) *UserProfile {
	return &UserProfile{
		bio:       bio,
		location:  location,
		website:   website,
		birthday:  birthday,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}
