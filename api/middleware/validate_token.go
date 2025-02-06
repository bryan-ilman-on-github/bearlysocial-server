package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"bearlysocial-backend/api/handler"
	"bearlysocial-backend/util"
)

// Define a key type to avoid context key collisions.
type contextKey string
const userKey contextKey = "user"

// Verifies the token and injects user data into the request context.
func ValidateToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header.
		token := r.Header.Get("Authorization")
		if !util.ValidToken(token) {
			util.ReturnMessage(w, http.StatusUnauthorized, "Missing authorization header.")
			return
		}

		// Create a context with a timeout to prevent long-running database operations.
		ctx, cancel := context.WithTimeout(context.Background(), 8 * time.Second)
		defer cancel()

		uid := strings.Split(strings.ToLower(token), "::")[0]

		// Define a variable to store the retrieved user account.
		var acc handler.UserAccount

		// Define a filter to find the account by uid (email).
		filter := bson.M{"_id": uid}

		// Attempt to find the user account in the database.
		err := util.MongoCollection.FindOne(ctx, filter).Decode(&acc)

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
		if token != *acc.Token {
			util.ReturnMessage(w, http.StatusUnauthorized, "Invalid token.")
			return
		}

		// Generate a new token for the user.
		updateToken, err := util.GenerateToken(uid)
		if err != nil {
			util.ReturnMessage(w, http.StatusInternalServerError, "Server error while generating token.")
			return
		}
		acc.Token = &updateToken
		
		// Inject user data into request context and call next handler.
		ctx = context.WithValue(r.Context(), userKey, acc)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
