package main

import (
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/handlers"
	"humanitec.io/resources/driver-aws-external/internal/api"
	"humanitec.io/resources/driver-aws-external/internal/aws"
	"humanitec.io/resources/driver-aws-external/internal/model"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	var s api.Server

	log.Println("Setting up Model")
	s.Model = model.Setup()

	log.Println("Setting up Routes")
	s.SetupRoutes()

	s.HttpClient = &http.Client{}

	s.NewAwsClient = aws.New
	if os.Getenv("USE_FAKE_AWS_CLIENT") != "" {
		s.NewAwsClient = aws.FakeNew
	}
	var err error
	s.TimeoutLimit = 300
	if os.Getenv("TIMEOUT_LIMIT") != "" {
		s.TimeoutLimit, err = strconv.Atoi(os.Getenv("TIMEOUT_LIMIT"))
		if err != nil || s.TimeoutLimit <= 0 {
			log.Fatalf(`Unable to set timeout limit to "%s"`, os.Getenv("TIMEOUT_LIMIT"))
		}
	}
	log.Printf("Timeout set to %d", s.TimeoutLimit)

	s.ServingPort = os.Getenv("PORT")
	if s.ServingPort == "" {
		s.ServingPort = "8080"
	}

	log.Printf("Listening on Port %s", s.ServingPort)
	log.Fatal(http.ListenAndServe(":"+s.ServingPort, handlers.LoggingHandler(os.Stdout, s.Router)))
}
