package api

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"humanitec.io/resources/driver-aws-external/internal/messages"
)

func (s *Server) createS3Bucket(drd messages.DriverResourceDefinition) (messages.ValuesSecrets, error) {

	if _, exists := drd.DriverSecrets["account"]; !exists {
		log.Println(`"account" property in driver_secrets is missing`)
		return messages.ValuesSecrets{}, fmt.Errorf(`"account" property in driver_secrets is missing: create s3 bucket`)
	}

	account, ok := drd.DriverSecrets["account"].(map[string]interface{})
	if !ok {
		log.Printf(`"account" property in driver_secrets is not an object. Got: %T`, drd.DriverSecrets["account"])
		log.Printf(`value of driver_secrets.account: %#v`, drd.DriverSecrets["account"])
		return messages.ValuesSecrets{}, fmt.Errorf(`"account" property in driver_secrets is not an obeject: create s3 bucket`)
	}

	bucketNameUUID, err := uuid.NewRandom()
	if err != nil {
		log.Println("Unable to generate random UUID.")
		return messages.ValuesSecrets{}, fmt.Errorf("create s3 bucket, generating name: %w", err)
	}
	bucketName := bucketNameUUID.String()
	accessKeyId := account["aws_access_key_id"].(string)
	secretAccessKey := account["aws_secret_access_key"].(string)
	region := drd.DriverParams["region"].(string)

	client, err := s.NewAwsClient(accessKeyId, secretAccessKey, region)
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

	client, err := s.NewAwsClient(awsCreds.AccessKeyID, awsCreds.SecretAccessKey, region)
	if err != nil {
		return err
	}

	err = client.DeleteBucket(bucketName)

	if err != nil {
		return err
	}

	return nil
}
