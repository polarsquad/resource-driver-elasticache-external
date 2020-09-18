package api

import (
	"testing"

	"humanitec.io/resources/driver-aws-external/internal/aws"
	"humanitec.io/resources/driver-aws-external/internal/aws/mock_aws"
	"humanitec.io/resources/driver-aws-external/internal/messages"
	"humanitec.io/resources/driver-aws-external/internal/model/mock_model"

	"github.com/golang/mock/gomock"
	"github.com/matryer/is"
)

func TestCreateRedis(t *testing.T) {
	is := is.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	accessKeyId := "AWS_ACCESS_KEY_ID-value"
	secretAccessKey := "AWS_SECRET_ACCESS_KEY-value"
	region := "eu-west-1"

	m := mock_model.NewMockModeler(ctrl)
	a := mock_aws.NewMockClient(ctrl)
	s := Server{
		Model: m,
		NewAwsClient: func(key, secret, reg string, timeoutLimit int) (aws.Client, error) {
			is.Equal(key, accessKeyId)
			is.Equal(secret, secretAccessKey)
			is.Equal(reg, region)
			return a, nil
		},
	}

	drd := messages.DriverResourceDefinition{
		ID:             "resource-id",
		Type:           "redis",
		ResourceParams: map[string]interface{}{},
		DriverParams: map[string]interface{}{
			"region":                  region,
			"cache_node_type":         "cache-node-type",
			"cache_availability_zone": "my-zone",
		},
		DriverSecrets: map[string]interface{}{
			"account": map[string]interface{}{
				"aws_access_key_id":     accessKeyId,
				"aws_secret_access_key": secretAccessKey,
			},
		},
	}
	redisHost := "redis-host"
	expectedData := messages.ValuesSecrets{
		Values: map[string]interface{}{
			"host": redisHost,
			"port": 6379,
		},
		Secrets: map[string]interface{}{},
	}
	awsCreds, _ := AccountMapToAWSCredentials(drd.DriverSecrets["account"])

	a.
		EXPECT().
		CreateElastiCacheRedis(gomock.AssignableToTypeOf(""), drd.DriverParams["cache_node_type"], drd.DriverParams["cache_availability_zone"]).
		Return(redisHost, nil).
		Times(1)

	responseData, err := s.createRedis(drd, awsCreds)

	is.NoErr(err)
	is.Equal(expectedData, responseData)
}

func TestDeleteRedis(t *testing.T) {
	is := is.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	accessKeyId := "AWS_ACCESS_KEY_ID-value"
	secretAccessKey := "AWS_SECRET_ACCESS_KEY-value"
	region := "eu-west-1"

	m := mock_model.NewMockModeler(ctrl)
	a := mock_aws.NewMockClient(ctrl)
	s := Server{
		Model: m,
		NewAwsClient: func(key, secret, reg string, timeoutLimit int) (aws.Client, error) {
			is.Equal(key, accessKeyId)
			is.Equal(secret, secretAccessKey)
			is.Equal(reg, region)
			return a, nil
		},
	}
	elastiCacheID := "elastic-cache-id"
	driverParams := map[string]interface{}{
		"region":                  region,
		"cache_node_type":         "cache-node-type",
		"cache_availability_zone": "my-zone",
	}
	driverSecrets := map[string]interface{}{
		"account": map[string]interface{}{
			"aws_access_key_id":     accessKeyId,
			"aws_secret_access_key": secretAccessKey,
		},
	}

	awsCreds, _ := AccountMapToAWSCredentials(driverSecrets["account"])
	a.
		EXPECT().
		DeleteElastiCacheRedis(elastiCacheID).
		Return(nil).
		Times(1)

	err := s.deleteRedis(elastiCacheID, driverParams, driverSecrets, awsCreds)

	is.NoErr(err)
}
