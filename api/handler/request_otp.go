package handler

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"

	"bearlysocial-backend/util"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Represents the user account structure in MongoDB with snake_case fields.
type UserAccount struct {
	ID string `bson:"_id"`
	OTP string `bson:"otp"`
	OTP_AttemptCount int `bson:"otp_attempt_count"`
	OTP_ExpiryTime int64 `bson:"otp_expiry_time"`
	CooldownTime *int64 `bson:"cooldown_time"`
	CreatedAt time.Time `bson:"created_at"`
	Token int `bson:"token"`
	FirstName *string `bson:"first_name"`
	LastName *string `bson:"last_name"`
	Interests []string `bson:"interests"`
	Langs []string `bson:"langs"`
	InstaHandler *string `bson:"insta_handler"`
	FB_Handler *string `bson:"fb_handler"`
	LinkedinHandler *string `bson:"linkedin_handler"`
	Mood *string `bson:"mood"`
	Schedule bson.M `bson:"schedule"`
}

// Represents the incoming request structure.
type UserRequest struct {
	EmailAddress string `json:"email_address"`
}

var (
	senderEmail string
	emailPasskey string
	smtpHost string
	smtpPort string
)

// Generates a 6-character alphanumeric OTP (A-Z, 0-9).
func generateOTP() string {
	const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 6)

	// Fill b with random bytes from the system's secure random generator.
	if _, err := rand.Read(b); err != nil {
		// Use time-based indexing in case of error.
		for i := range b {
			b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		}
	} else {
		// Map random bytes to the allowed characters.
		for i, v := range b {
			b[i] = charset[int(v)%len(charset)]
		}
	}
	return string(b) // Convert byte slice to a string and return as OTP.
}

// Sends the OTP using standard net/smtp package.
func sendOTP(to, otp string) error {
	// Create an SMTP authentication object using the sender's email, password, and SMTP host.
	// This is necessary to authenticate with the SMTP server before sending an email. 
	// It ensures that the server knows the sender is authorized to send emails from this account.
	auth := smtp.PlainAuth("", senderEmail, emailPasskey, smtpHost)

	headers := map[string]string{
		"From":         fmt.Sprintf("BearlySocial <%s>", senderEmail),
		"To":           to,
		"Subject":      "Your One-Time Password (OTP)",
		"MIME-Version": "1.0",
		"Content-Type": "text/html; charset=UTF-8",
	}

	body := fmt.Sprintf(`<p style="font-size: 18px;">Your One-time Password (OTP) is:</p>
		<p style="font-size: 24px; font-weight: bold;">%s</p>
		<p style="font-size: 18px">The OTP is valid for only <span style="font-weight: bold;">8 minutes</span>.</p>`, otp)

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n" + body)

	return smtp.SendMail(
		smtpHost+":"+smtpPort,
		auth,
		senderEmail,
		[]string{to},
		[]byte(msg.String()),
	)
}

// Handles OTP requests.
func RequestOTP(w http.ResponseWriter, r *http.Request) {
	senderEmail = os.Getenv("SENDER_EMAIL")
	emailPasskey = os.Getenv("EMAIL_PASSKEY")
	smtpHost = os.Getenv("SMTP_HOST")
	smtpPort = os.Getenv("SMTP_PORT")

	if r.Method != http.MethodGet {
		util.ReturnMessage(w, http.StatusBadRequest, "Method not allowed.")
		return
	}

	var req UserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.ReturnMessage(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	userEmail := strings.TrimSpace(req.EmailAddress)
	if userEmail == "" {
		util.ReturnMessage(w, http.StatusBadRequest, "Email address is required.")
		return
	}

	otp := generateOTP()

	// Create a context with a timeout to prevent long-running database operations.
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	now := time.Now()
	currentTimeMillis := now.UnixMilli()

	// Define a variable to store the retrieved user account.
	var acc UserAccount

	// Define a filter to find the account by email.
	filter := bson.M{"_id": userEmail}

	// Attempt to find the user account in the database.
	err := util.MongoCollection.FindOne(ctx, filter).Decode(&acc)

	if err == mongo.ErrNoDocuments {
		// If the account does not exist, create a new one.
		acc := UserAccount{
			ID: userEmail,
			OTP: otp,
			OTP_AttemptCount: 0,
			OTP_ExpiryTime: now.Add(8 * time.Minute).UnixMilli(),
			CreatedAt: now,
			Schedule: bson.M{},
		}

		// Insert the new account into the database.
		_, err = util.MongoCollection.InsertOne(ctx, acc)
		if err != nil {
			util.ReturnMessage(w, http.StatusInternalServerError, "Failed to create account.")
			return
		}
	} else if err != nil {
		// Handle any other database errors.
		util.ReturnMessage(w, http.StatusInternalServerError, "Database error.")
		return
	} else {
		// If the account exists and the user has attempted OTP requests less than 4 times.
		if acc.OTP_AttemptCount < 4 {
			update := bson.M{
				"$set": bson.M{
					"otp": otp,
					"otp_expiry_time": currentTimeMillis + 8*60*1000, // Extend expiry time.
				},
			}
			_, err = util.MongoCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				util.ReturnMessage(w, http.StatusInternalServerError, "Failed to update OTP.")
				return
			}
		} else {
			// If the user exceeded OTP attempts, check if the cooldown period has expired.
			if *acc.CooldownTime <= currentTimeMillis {
				update := bson.M{
					"$set": bson.M{
						"otp": otp,
						"otp_attempt_count": 0, // Reset attempt count.
						"otp_expiry_time": currentTimeMillis + 8*60*1000, // Reset expiry time.
						"cooldown_time": nil, // Remove cooldown restriction.
					},
				}
				_, err = util.MongoCollection.UpdateOne(ctx, filter, update)
				if err != nil {
					util.ReturnMessage(w, http.StatusInternalServerError, "Failed to reset OTP attempts.")
					return
				}
			} else {
				// If still in cooldown, calculate the remaining time before retry is allowed.
				cooldownTime := time.UnixMilli(*acc.CooldownTime)
				remainingTime := time.Until(cooldownTime)
				message := fmt.Sprintf("Please wait %s before trying again.", util.HumanReadableDuration(remainingTime))

				util.ReturnMessage(w, http.StatusBadRequest, message)
				return				
			}
		}
	}

	// Send the OTP to the user's email.
	if err := sendOTP(userEmail, otp); err != nil {
		log.Printf("ERROR SENDING EMAIL: %v\n", err)
		util.ReturnMessage(w, http.StatusInternalServerError, "Failed to send OTP email.")
		return
	}

	w.WriteHeader(http.StatusOK)
}
