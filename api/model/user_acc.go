package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// Represents the user account structure in MongoDB with snake_case fields.
type UserAccount struct {
	ID string `bson:"_id" json:"uid"`
	OTP *string `bson:"otp" json:"otp"`
	OTP_AttemptCount int `bson:"otp_attempt_count" json:"otp_attempt_count"`
	OTP_ExpiryTime *int64 `bson:"otp_expiry_time" json:"otp_expiry_time"`
	CooldownTime *int64 `bson:"cooldown_time" json:"cooldown_time"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	Token *string `bson:"token" json:"token"`
	FirstName string `bson:"first_name" json:"first_name"`
	LastName string `bson:"last_name" json:"last_name"`
	Interests []string `bson:"interests" json:"interests"`
	Langs []string `bson:"langs" json:"langs"`
	InstaHandler string `bson:"insta_handler" json:"insta_handler"`
	FB_Handler string `bson:"fb_handler" json:"fb_handler"`
	LinkedinHandler string `bson:"linkedin_handler" json:"linkedin_handler"`
	Mood string `bson:"mood" json:"mood"`
	Schedule bson.M `bson:"schedule" json:"schedule"`
}
