// Command server runs the route planner API locally, outside Google Cloud Functions.
//
// Usage:
//
//	go run ./cmd/server
//	curl 'localhost:8080/?lat=51.2665&lon=0.2010&distance=5'
package main

import (
	"log"
	"net/http"
	"os"

	routeplanner "github.com/yurachistic1/routeplanner-backend"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", routeplanner.RoutePlannerAPI)

	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
