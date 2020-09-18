package api

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"humanitec.io/resources/driver-aws-external/internal/messages"
)

func (s *Server) createS3Bucket(drd messages.DriverResourceDefinition, awsCreds AWSCredentials) (messages.ValuesSecrets, error) {

	var region string
	var ok bool
	if region, ok = drd.DriverParams["region"].(string); !ok {
		log.Printf(`"region" property in driver_params: Expected string, Got: %T`, drd.DriverParams["region"])
		return messages.ValuesSecrets{}, fmt.Errorf(`"region" property in driver_params: expected string, got %T`, drd.DriverParams["region"])
	}

	bucketNameUUID, err := uuid.NewRandom()
	if err != nil {
		log.Println("Unable to generate random UUID.")
		return messages.ValuesSecrets{}, fmt.Errorf("create s3 bucket, generating name: %w", err)
	}
	bucketName := bucketNameUUID.String()

	client, err := s.NewAwsClient(awsCreds.AccessKeyID, awsCreds.SecretAccessKey, region, s.TimeoutLimit)
	if err != nil {
		return messages.ValuesSecrets{}, err
	}

	generatedRegion, err := client.CreateBucket(bucketName)
	if err != nil {
		return messages.ValuesSecrets{}, err
	}

	return messages.ValuesSecrets{
		Values: map[string]interface{}{
			"region": generatedRegion,
			"bucket": bucketName,
		},
		Secrets: map[string]interface{}{},
	}, nil
}

func (s *Server) deleteS3Bucket(bucketName, region string, awsCreds AWSCredentials) error {

	client, err := s.NewAwsClient(awsCreds.AccessKeyID, awsCreds.SecretAccessKey, region, s.TimeoutLimit)
	if err != nil {
		return err
	}

	err = client.DeleteBucket(bucketName)

	if err != nil {
		return err
	}

	return nil
}
