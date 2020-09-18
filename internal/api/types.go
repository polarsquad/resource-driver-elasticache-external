package api

import (
	"net/http"

	"humanitec.io/resources/driver-aws-external/internal/aws"
	"humanitec.io/resources/driver-aws-external/internal/doer"
	"humanitec.io/resources/driver-aws-external/internal/model"
)

// Server holds all dependancies that are necessary for the api to be able operate.
type Server struct {
	Model        model.Modeler
	Router       http.Handler
	ServingPort  string
	HttpClient   doer.Doer
	NewAwsClient func(string, string, string, int) (aws.Client, error)
	TimeoutLimit int
}

type AWSCredentials struct {
	AccessKeyID     string `json:"aws_access_key_id"`
	SecretAccessKey string `json:"aws_secret_access_key"`
}
