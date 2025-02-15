package handler

import (
	"context"
	"math/rand"
	"net/http"
	"time"

	"bearlysocial-backend/util"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Benchmark(w http.ResponseWriter, r *http.Request) {
	_, err := util.GenerateToken("john_doe@example.com")
	if err != nil {
		util.ReturnMessage(w, http.StatusInternalServerError, "Failed to generate token.")
		return
	}

	data := make([]byte, 128*1024) // 128 KB.
	rand.Seed(time.Now().UnixNano())
	rand.Read(data) // Fill with random data.

	doc := bson.M{"_id": "benchmark_data", "data": data}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.Update().SetUpsert(true)
	_, err = util.MongoCollection.UpdateOne(ctx, bson.M{"_id": "benchmark_data"}, bson.M{"$set": doc}, opts)
	if err != nil {
		util.ReturnMessage(w, http.StatusInternalServerError, "Failed to insert/update data.")
		return
	}

	var result bson.M
	err = util.MongoCollection.FindOne(ctx, bson.M{"_id": "benchmark_data"}).Decode(&result)
	if err != nil {
		util.ReturnMessage(w, http.StatusInternalServerError, "Failed to retrieve data.")
		return
	}

	w.WriteHeader(http.StatusOK)
}
