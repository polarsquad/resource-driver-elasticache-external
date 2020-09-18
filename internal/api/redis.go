package api

import (
	"fmt"
	"log"

	"humanitec.io/resources/driver-aws-external/internal/messages"
)

func (s *Server) createRedis(drd messages.DriverResourceDefinition, awsCreds AWSCredentials) (messages.ValuesSecrets, error) {

	var region string
	var ok bool
	if region, ok = drd.DriverParams["region"].(string); !ok {
		log.Printf(`"region" property in driver_params: Expected string, Got: %T`, drd.DriverParams["region"])
		return messages.ValuesSecrets{}, fmt.Errorf(`"region" property in driver_params: expected string, got %T`, drd.DriverParams["region"])
	}

	client, err := s.NewAwsClient(awsCreds.AccessKeyID, awsCreds.SecretAccessKey, region)
	if err != nil {
		return messages.ValuesSecrets{}, err
	}
	_ = client

	// Here we call the driver to create the elasticache with appropriate params
	/*
		err = client.CreateElastiCache(...)

		if err != nil {
			return err
		}
	*/
	return messages.ValuesSecrets{
		Values: map[string]interface{}{
			// This will hold access credentials
		},
		Secrets: map[string]interface{}{},
	}, nil
}

func (s *Server) deleteRedis(id string, driverParams, driverSecrets map[string]interface{}) error {

	if _, exists := driverSecrets["account"]; !exists {
		log.Println(`"account" property in driver_secrets is missing`)
		return fmt.Errorf(`"account" property in driver_secrets is missing: delete redis`)
	}
	var region string
	var ok bool
	if region, ok = driverParams["region"].(string); !ok {
		log.Printf(`"region" property in driver_params: Expected string, Got: %T`, driverParams["region"])
		return fmt.Errorf(`"region" property in driver_params: expected string, got %T`, driverParams["region"])
	}

	awsCreds, err := AccountMapToAWSCredentials(driverSecrets["account"])
	client, err := s.NewAwsClient(awsCreds.AccessKeyID, awsCreds.SecretAccessKey, region)
	if err != nil {
		return err
	}
	_ = client
	/*
		err = client.DeleteElastiCache(...)

		if err != nil {
			return err
		}
	*/
	return nil
}
