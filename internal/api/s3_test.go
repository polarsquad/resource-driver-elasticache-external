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

func TestCreateS3Bucket(t *testing.T) {
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
		NewAwsClient: func(key, secret, reg string) (aws.Client, error) {
			is.Equal(key, accessKeyId)
			is.Equal(secret, secretAccessKey)
			is.Equal(reg, region)
			return a, nil
		},
	}

	drd := messages.DriverResourceDefinition{
		ID:             "resource-id",
		Type:           "s3",
		ResourceParams: map[string]interface{}{},
		DriverParams: map[string]interface{}{
			"region": region,
		},
		DriverSecrets: map[string]interface{}{
			"account": map[string]interface{}{
				"aws_access_key_id":     accessKeyId,
				"aws_secret_access_key": secretAccessKey,
			},
		},
	}
	expectedData := messages.ValuesSecrets{
		Values: map[string]interface{}{
			"region": region,
		},
		Secrets: map[string]interface{}{},
	}

	a.
		EXPECT().
		CreateBucket(gomock.AssignableToTypeOf("")).
		Do(func(bn interface{}) {
			expectedData.Values["bucket"] = bn.(string)
		}).
		Return(region, nil).
		Times(1)

	responseData, err := s.createS3Bucket(drd)

	is.NoErr(err)
	is.Equal(expectedData, responseData)
}
