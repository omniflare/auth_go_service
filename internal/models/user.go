package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	FirebaseUID string        `gorm:"uniqueIndex;not null" json:"firebase_uid"`
	Email       string        `gorm:"unique" json:"email"`
	Name        string        `json:"name"`
	PhotoURL    string        `json:"photoUrl"`
	Role        string        `gorm:"default:'user'" json:"role"`
	Metadata    *UserMetadata `gorm:"serializer:json" json:"metadata"`
}

type UserMetadata struct {
	CreationTimestamp    int64 `json:"creationTimestamp"`
	LastSignInTimestamp  int64 `json:"lastSignInTimestamp"`
	LastRefreshTimestamp int64 `json:"lastRefreshTimestamp"`
}
