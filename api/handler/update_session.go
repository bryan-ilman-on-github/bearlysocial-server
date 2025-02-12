package handler

import (
	"context"
	"net/http"
	"time"

	"bearlysocial-backend/api/middleware"
	"bearlysocial-backend/api/model"
	"bearlysocial-backend/util"

	"go.mongodb.org/mongo-driver/bson"
)

// Handles session update.
func UpdateSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		util.ReturnMessage(w, http.StatusBadRequest, "Method not allowed.")
		return
	}

	// Retrieve user data from context.
	user_acc, ok := r.Context().Value(middleware.USER_ACCOUNT).(model.UserAccount)
	if !ok {
		util.ReturnMessage(w, http.StatusInternalServerError, "Failed to retrieve user session.")
		return
	}

	// Create a MongoDB context with a timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 8 * time.Second)
	defer cancel()

	// Define the filter using the user's ID field.
	filter := bson.M{"_id": user_acc.ID}

	// Replace the existing document with the latest user data.
	_, err := util.MongoCollection.ReplaceOne(ctx, filter, user_acc)
	if err != nil {
		util.ReturnMessage(w, http.StatusInternalServerError, "Failed to update session in database.")
		return
	}

	w.WriteHeader(http.StatusOK)
}
