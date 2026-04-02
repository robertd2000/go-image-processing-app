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
	p.bio = bio
	p.location = location
	p.website = website
	p.birthday = birthday
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
