package main

import (
	"github.com/notegio/openrelay/ingest"
	"github.com/notegio/openrelay/channels"
	"github.com/notegio/openrelay/affiliates"
	"github.com/notegio/openrelay/accounts"
	"encoding/hex"
	"net/http"
	"gopkg.in/redis.v3"
	"os"
	"log"
	"github.com/rs/cors"
)

func main() {
	redisURL := os.Args[1]
	defaultFeeRecipientString := os.Args[2]
	var port string
	if len(os.Args) >= 4 {
		port = os.Args[3]
	} else {
		port = "8080"
	}
	if redisURL == "" {
		log.Fatalf("Please specify redis URL")
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisURL,
	})
	defaultFeeRecipientSlice, err := hex.DecodeString(defaultFeeRecipientString)
	if err != nil {
		log.Fatalf(err.Error())
	}
	defaultFeeRecipientBytes := [20]byte{}
	copy(defaultFeeRecipientBytes[:], defaultFeeRecipientSlice[:])
	affiliateService := affiliates.NewRedisAffiliateService(redisClient)
	accountService := accounts.NewRedisAccountService(redisClient)
	publisher := channels.NewRedisQueuePublisher("ingest", redisClient)
	handler := ingest.Handler(publisher, accountService, affiliateService)
	feeHandler := ingest.FeeHandler(publisher, accountService, affiliateService, defaultFeeRecipientBytes)

	mux := http.NewServeMux()
	mux.HandleFunc("/v0.0/order", handler)
	mux.HandleFunc("/v0.0/fees", feeHandler)
	corsHandler := cors.Default().Handler(mux)
	log.Printf("Order Ingest Serving on :%v", port)
	http.ListenAndServe(":"+port, corsHandler)
}
