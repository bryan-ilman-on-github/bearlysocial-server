package handler

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"bearlysocial-backend/api/model"
	"bearlysocial-backend/util"
)

// Creates a secure random token using crypto/rand.
func generateToken() (string, error) {
	b := make([]byte, 32)

	// Fill the byte slice with cryptographically random bytes.
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	// Create a new SHA-256 hash instance with random bytes written into the hash function.
	hasher := sha256.New()
	hasher.Write(b)

	// Compute the hash and return its hexadecimal representation.
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// Handles OTP validation.
func ValidateOTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		util.ReturnMessage(w, http.StatusBadRequest, "Method not allowed.")
		return
	}

	// Parse request body.
	var req model.ValidateOTP
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.ReturnMessage(w, http.StatusBadRequest, "Invalid request format.")
		return
	}

	userEmail := strings.TrimSpace(req.EmailAddress)
	userOTP := strings.TrimSpace(req.OTP)
	if !util.ValidEmail(userEmail) || !util.ValidOTP(userOTP) {
		util.ReturnMessage(w, http.StatusBadRequest, "Invalid email or OTP format.")
		return
	}

	// Create a context with a timeout to prevent long-running database operations.
	ctx, cancel := context.WithTimeout(context.Background(), 8 * time.Second)
	defer cancel()

	// Define a variable to store the retrieved user account.
	var acc UserAccount

	// Define a filter to find the account by email.
	filter := bson.M{"_id": userEmail}

	// Attempt to find the user account in the database.
	err := util.MongoCollection.FindOne(ctx, filter).Decode(&acc)

	if err == mongo.ErrNoDocuments || acc.OTP == nil {
		// If user not found or missing OTP, ask the user to request an OTP first.
		util.ReturnMessage(w, http.StatusBadRequest, "Please request an OTP first.")
		return
	}
	if err != nil {
		// Handle any other database errors.
		log.Printf("DATABASE ERROR: %v\n", err)
		util.ReturnMessage(w, http.StatusInternalServerError, "Database error.")
		return
	}

	currentTime := time.Now().UnixMilli()
	if acc.OTP_AttemptCount < 4 {
		if currentTime > *acc.OTP_ExpiryTime {
			util.ReturnMessage(w, http.StatusBadRequest, "Your OTP has expired.")
			return
		} else {
			if strings.EqualFold(*acc.OTP, req.OTP) {
				token, err := generateToken()
				if err != nil {
					util.ReturnMessage(w, http.StatusInternalServerError, "Failed to generate token.")
					return
				}

				// Update the database by resetting OTP fields and setting the new token.
				update := bson.M{
					"$set": bson.M{
						"token": token,
						"otp": nil,
						"otp_attempt_count": 0,
						"otp_expiry_time": nil,
						"cooldown_time": nil,
					},
				}

				_, err = util.MongoCollection.UpdateOne(ctx, filter, update)
				if err != nil {
					util.ReturnMessage(w, http.StatusInternalServerError, "Failed to update token.")
					return
				}

				// Setting fields to reflect the update in the response after successful verification.
				acc.OTP = nil
				acc.OTP_AttemptCount = 0
				acc.OTP_ExpiryTime = nil
				acc.CooldownTime = nil
				acc.Token = &token

				// Return a success response with the updated user data.
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(acc)
			} else {
				// If the OTP is incorrect, increment the attempt count.
				update := bson.M{
					"$inc": bson.M{"otp_attempt_count": 1},
				}
				var msg string

				if acc.OTP_AttemptCount + 1 >= 4 {
					cooldownTime := time.Now().Add(1 * time.Hour).UnixMilli()
					msg = "Too many failed attempts. Please request a new OTP in an hour."

					// Update the account with cooldown information and clear OTP fields.
					update["$set"] = bson.M{
						"cooldown_time":   cooldownTime,
						"otp":             nil,
						"otp_expiry_time": nil,
					}
				} else {
					msg = "The OTP you provided is incorrect."
				}

				_, err = util.MongoCollection.UpdateOne(ctx, filter, update)
				if err != nil {
					util.ReturnMessage(w, http.StatusInternalServerError, "Failed to update attempt count.")
					return
				}

				util.ReturnMessage(w, http.StatusBadRequest, msg)
				return
			}
		}
	} else {
		remaining := time.Until(time.UnixMilli(*acc.CooldownTime))

		// Construct message based on remaining time.
		msg := "Please request a new OTP."
		if remaining > 0 {
			msg = fmt.Sprintf("Please request a new OTP in %s.", util.HumanReadableDuration(remaining))
		}

		util.ReturnMessage(w, http.StatusBadRequest, msg)
		return
	}
}
