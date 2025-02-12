package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"bearlysocial-backend/api/model"
	"bearlysocial-backend/util"
)

// Define a key type to avoid context key collisions.
type contextKey string
const USER_ACCOUNT contextKey = "user_acc"

// Verifies the token and injects user data into the request context.
func ValidateToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {	
		// Extract token from Authorization header.
		reqToken := r.Header.Get("Authorization")
		if !util.ValidToken(reqToken) {
			util.ReturnMessage(w, http.StatusUnauthorized, "Invalid token format.")
			return
		}

		// Create a context with a timeout to prevent long-running database operations.
		ctx, cancel := context.WithTimeout(context.Background(), 8 * time.Second)
		defer cancel()

		uid := strings.Split(strings.ToLower(reqToken), "::")[0] // Extract uid (email) from request token.
		updateToken, err := util.GenerateToken(uid) // Generate a new token for the user.
		if err != nil {
			util.ReturnMessage(w, http.StatusInternalServerError, "Token generation failed.")
			return
		  }

	    filter := bson.M{"_id": uid, "token": reqToken}
    	update := bson.M{"$set": bson.M{"token": updateToken}}

		// Define a variable to store the retrieved user account.
		var user_acc model.UserAccount

		// Atomically validate token and set new token.
		err = util.MongoCollection.FindOneAndUpdate(
			ctx,
			filter,
			update,
			options.FindOneAndUpdate().SetReturnDocument(options.After),
		  ).Decode(&user_acc)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				util.ReturnMessage(w, http.StatusUnauthorized, "Authorization failed.")
			} else {
				log.Printf("DATABASE ERROR: %v\n", err)
				util.ReturnMessage(w, http.StatusInternalServerError, "Database error.")
			}
			return
		}

		// Inject updated user data into context.
		ctx = context.WithValue(r.Context(), USER_ACCOUNT, user_acc)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
