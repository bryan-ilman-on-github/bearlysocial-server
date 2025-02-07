package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"bearlysocial-backend/api/model"
	"bearlysocial-backend/util"
)

// Define a key type to avoid context key collisions.
type contextKey string
const USER_ACCOUNT contextKey = "user_acc"

// Verifies the token and injects user data into the request context.
func ValidateToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			util.ReturnMessage(w, http.StatusBadRequest, "Method not allowed.")
			return
		}
	
		// Extract token from Authorization header.
		reqToken := r.Header.Get("Authorization")
		if !util.ValidToken(reqToken) {
			util.ReturnMessage(w, http.StatusUnauthorized, "Missing authorization header.")
			return
		}

		// Create a context with a timeout to prevent long-running database operations.
		ctx, cancel := context.WithTimeout(context.Background(), 8 * time.Second)
		defer cancel()

		uid := strings.Split(strings.ToLower(reqToken), "::")[0]

		// Define a variable to store the retrieved user account.
		var user_acc model.UserAccount

		// Define a filter to find the account by uid (email).
		filter := bson.M{"_id": uid}

		// Attempt to find the user account in the database.
		err := util.MongoCollection.FindOne(ctx, filter).Decode(&user_acc)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				util.ReturnMessage(w, http.StatusUnauthorized, "User account not found.")
			} else {
				log.Printf("DATABASE ERROR: %v\n", err)
				util.ReturnMessage(w, http.StatusInternalServerError, "Database error.")
			}
			return
		}

		// Validate the provided token.
		if reqToken != *user_acc.Token {
			util.ReturnMessage(w, http.StatusUnauthorized, "Invalid token.")
			return
		}

		// Generate a new token for the user.
		updateToken, err := util.GenerateToken(uid)
		if err != nil {
			util.ReturnMessage(w, http.StatusInternalServerError, "Server error while generating token.")
			return
		}
		user_acc.Token = &updateToken
		
		// Inject user data into request context and call next handler.
		ctx = context.WithValue(r.Context(), USER_ACCOUNT, user_acc)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
