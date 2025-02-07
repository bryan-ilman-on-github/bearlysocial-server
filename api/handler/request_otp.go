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

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"bearlysocial-backend/api/model"
	"bearlysocial-backend/util"
)

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

// Handles OTP request.
func RequestOTP(w http.ResponseWriter, r *http.Request) {
	senderEmail = os.Getenv("SENDER_EMAIL")
	emailPasskey = os.Getenv("EMAIL_PASSKEY")
	smtpHost = os.Getenv("SMTP_HOST")
	smtpPort = os.Getenv("SMTP_PORT")

	if r.Method != http.MethodGet {
		util.ReturnMessage(w, http.StatusBadRequest, "Method not allowed.")
		return
	}

	// Parse request body.
	var req model.RequestOTP
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.ReturnMessage(w, http.StatusBadRequest, "Invalid request format.")
		return
	}

	userEmail := strings.ToLower(strings.TrimSpace(req.EmailAddress))
	if !util.ValidEmail(userEmail) {
		util.ReturnMessage(w, http.StatusBadRequest, "Invalid email format.")
		return
	}

	otp := generateOTP()

	// Create a context with a timeout to prevent long-running database operations.
	ctx, cancel := context.WithTimeout(context.Background(), 8 * time.Second)
	defer cancel()

	now := time.Now()
	currentTimeMillis := now.UnixMilli()

	// Define a variable to store the retrieved user account.
	var user_acc model.UserAccount

	// Define a filter to find the account by email.
	filter := bson.M{"_id": userEmail}

	// Attempt to find the user account in the database.
	err := util.MongoCollection.FindOne(ctx, filter).Decode(&user_acc)

	if err == mongo.ErrNoDocuments {
		// If the account does not exist, create a new one.
		user_acc := model.UserAccount{
			ID: userEmail,
			OTP: &otp,
			OTP_AttemptCount: 0,
			OTP_ExpiryTime: util.RefInt64(now.Add(8 * time.Minute).UnixMilli()),
			CreatedAt: now,
			Schedule: bson.M{},
		}

		// Insert the new account into the database.
		_, err = util.MongoCollection.InsertOne(ctx, user_acc)
		if err != nil {
			util.ReturnMessage(w, http.StatusInternalServerError, "Failed to create account.")
			return
		}
	} else if err != nil {
		// Handle any other database errors.
		log.Printf("DATABASE ERROR: %v\n", err)
		util.ReturnMessage(w, http.StatusInternalServerError, "Database error.")
		return
	} else {
		// If the account exists and the user has attempted OTP requests less than 4 times.
		if user_acc.OTP_AttemptCount < 4 {
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
			if *user_acc.CooldownTime <= currentTimeMillis {
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
				cooldownTime := time.UnixMilli(*user_acc.CooldownTime)
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
