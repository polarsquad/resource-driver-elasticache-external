package api

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"humanitec.io/resources/driver-aws-external/internal/messages"
)

func (s *Server) createRedis(drd messages.DriverResourceDefinition, awsCreds AWSCredentials) (messages.ValuesSecrets, error) {

	var region string
	var ok bool
	if region, ok = drd.DriverParams["region"].(string); !ok {
		log.Printf(`"region" property in driver_params: Expected string, Got: %T`, drd.DriverParams["region"])
		return messages.ValuesSecrets{}, fmt.Errorf(`"region" property in driver_params: expected string, got %T`, drd.DriverParams["region"])
	}

	client, err := s.NewAwsClient(awsCreds.AccessKeyID, awsCreds.SecretAccessKey, region, s.TimeoutLimit)
	if err != nil {
		return messages.ValuesSecrets{}, err
	}

	// Here we call the driver to create the elasticache with appropriate params
	clusterUUID, err := uuid.NewRandom()
	if err != nil {
		log.Println("Unable to generate random UUID.")
		return messages.ValuesSecrets{}, fmt.Errorf("create s3 bucket, generating name: %w", err)
	}
	clusterId := clusterUUID.String()

	var cacheNodeType string
	if cacheNodeType, ok = drd.DriverParams["cache_node_type"].(string); !ok {
		log.Printf(`"cache_node_type" property in driver_params: Expected string, Got: %T`, drd.DriverParams["cache_node_type"])
		return messages.ValuesSecrets{}, fmt.Errorf(`"cache_node_type" property in driver_params: expected string, got %T`, drd.DriverParams["cache_node_type"])
	}

	var cacheAz string
	if cacheAz, ok = drd.DriverParams["cache_availability_zone"].(string); !ok {
		log.Printf(`"cache_availability_zone" property in driver_params: Expected string, Got: %T`, drd.DriverParams["cache_availability_zone"])
		return messages.ValuesSecrets{}, fmt.Errorf(`"cache_availability_zone" property in driver_params: expected string, got %T`, drd.DriverParams["cache_availability_zone"])
	}

	endpoint, err := client.CreateElastiCacheRedis(clusterId, cacheNodeType, cacheAz)

	if err != nil {
		return messages.ValuesSecrets{}, err
	}
	return messages.ValuesSecrets{
		Values: map[string]interface{}{
			"host": endpoint,
			"port": 6379,
		},
		Secrets: map[string]interface{}{},
	}, nil
}

func (s *Server) deleteRedis(id string, driverParams, driverSecrets map[string]interface{}, awsCreds AWSCredentials) error {

	var region string
	var ok bool
	if region, ok = driverParams["region"].(string); !ok {
		log.Printf(`"region" property in driver_params: Expected string, Got: %T`, driverParams["region"])
		return fmt.Errorf(`"region" property in driver_params: expected string, got %T`, driverParams["region"])
	}

	client, err := s.NewAwsClient(awsCreds.AccessKeyID, awsCreds.SecretAccessKey, region, s.TimeoutLimit)
	if err != nil {
		return err
	}

	err = client.DeleteElastiCacheRedis(id)

	if err != nil {
		return err
	}

	return nil
}
