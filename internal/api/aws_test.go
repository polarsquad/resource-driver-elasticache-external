package api

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"humanitec.io/resources/driver-aws-external/internal/aws"
	"humanitec.io/resources/driver-aws-external/internal/aws/mock_aws"
	"humanitec.io/resources/driver-aws-external/internal/messages"
	"humanitec.io/resources/driver-aws-external/internal/model"
	"humanitec.io/resources/driver-aws-external/internal/model/mock_model"

	"github.com/golang/mock/gomock"
	"github.com/matryer/is"
)

// Custom matcher that ignores all dates in the in model.ResourceMetadata
type ignoreDateResourceMetadata struct{ model.ResourceMetadata }

func IgnoreDateResourceMetadata(m model.ResourceMetadata) gomock.Matcher {
	return &ignoreDateResourceMetadata{m}
}

func (m *ignoreDateResourceMetadata) Matches(x interface{}) bool {
	return AreEqualExceptDates(m.ResourceMetadata, x)
}

func (m *ignoreDateResourceMetadata) String() string {
	return fmt.Sprintf("%v", m.ResourceMetadata)
}

func TestCreateAWSResource_Existing(t *testing.T) {
	is := is.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mock_model.NewMockModeler(ctrl)
	s := Server{
		Model: m,
	}

	resourceID := "test-db-id"
	resType := "s3"
	params := map[string]interface{}{
		"region": "eu-west-1",
	}
	data := map[string]interface{}{
		"region": "eu-west-1",
		"bucket": "my-s3-bucket",
	}
	accessKeyId := "AWS_ACCESS_KEY_ID-value"
	secretAccessKey := "AWS_SECRET_ACCESS_KEY-value"

	drd := messages.DriverResourceDefinition{
		ID:             resourceID,
		Type:           "s3",
		ResourceParams: params,
		DriverParams: map[string]interface{}{
			"region": "eu-west-1",
		},
		DriverSecrets: map[string]interface{}{
			"account": map[string]interface{}{
				"aws_access_key_id":     accessKeyId,
				"aws_secret_access_key": secretAccessKey,
			},
		},
	}

	expectedResponseData := messages.ResourceData{
		Type: drd.Type,
		Data: messages.ValuesSecrets{
			Values: data,
			Secrets: map[string]interface{}{
				"aws_access_key_id":     accessKeyId,
				"aws_secret_access_key": secretAccessKey,
			},
		},
		DriverType: "aws",
		DriverData: messages.ValuesSecrets{},
	}

	metadata := model.ResourceMetadata{
		ID:        resourceID,
		Type:      resType,
		CreatedAt: time.Date(2020, 07, 16, 18, 12, 20, 0, time.UTC),
		UpdatedAt: time.Date(2020, 07, 16, 18, 12, 20, 0, time.UTC),
		DeletedAt: sql.NullTime{Valid: false},
		Params:    params,
		Data:      data,
	}

	m.
		EXPECT().
		SelectResourceMetadata(resourceID).
		Return(metadata, true, nil).
		Times(1)

	res := ExecuteRequest(s, http.MethodPost, "/", drd, t)

	var returnedResourceData messages.ResourceData
	json.Unmarshal(res.Body.Bytes(), &returnedResourceData)
	fmt.Printf("returnedResourceData: %v\n", returnedResourceData)
	is.Equal(expectedResponseData, returnedResourceData)
}

func TestCreateAWSResource_New(t *testing.T) {
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
	resourceID := "test-db-id"
	resType := "s3"
	params := map[string]interface{}{
		"region": "eu-west-1",
	}
	data := map[string]interface{}{
		"region": "eu-west-1",
		"bucket": "",
	}
	drd := messages.DriverResourceDefinition{
		ID:             resourceID,
		Type:           resType,
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

	expectedResponseData := messages.ResourceData{
		Type: drd.Type,
		Data: messages.ValuesSecrets{
			Values: data,
			Secrets: map[string]interface{}{
				"aws_access_key_id":     accessKeyId,
				"aws_secret_access_key": secretAccessKey,
			},
		},
		DriverType: "aws",
		DriverData: messages.ValuesSecrets{},
	}

	metadata := model.ResourceMetadata{
		ID:        resourceID,
		Type:      resType,
		CreatedAt: time.Date(2020, 07, 16, 18, 12, 20, 0, time.UTC),
		UpdatedAt: time.Date(2020, 07, 16, 18, 12, 20, 0, time.UTC),
		DeletedAt: sql.NullTime{Valid: false},
		Params:    params,
		Data:      data,
	}

	m.
		EXPECT().
		SelectResourceMetadata(resourceID).
		Return(model.ResourceMetadata{}, false, nil).
		Times(1)
	a.
		EXPECT().
		CreateBucket(gomock.AssignableToTypeOf("")).
		Do(func(bn interface{}) {
			data["bucket"] = bn.(string)
		}).
		Return(region, nil).
		Times(1)

	m.
		EXPECT().
		InsertOrUpdateResourceMetadata(IgnoreDateResourceMetadata(metadata)).
		Return(nil).
		Times(1)

	res := ExecuteRequest(s, http.MethodPost, "/", drd, t)

	var returnedResourceData messages.ResourceData
	json.Unmarshal(res.Body.Bytes(), &returnedResourceData)
	fmt.Printf("returnedResourceData: %v\n", returnedResourceData)
	is.Equal(expectedResponseData, returnedResourceData)
}

func TestDeleteAWSResource_Exists(t *testing.T) {
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
	resourceID := "test-db-id"
	resType := "s3"
	bucketName := "s3-bucket-name"
	params := map[string]interface{}{
		"region": "eu-west-1",
	}
	data := map[string]interface{}{
		"region": "eu-west-1",
		"bucket": bucketName,
	}
	account := AWSCredentials{
		AccessKeyID:     accessKeyId,
		SecretAccessKey: secretAccessKey,
	}

	metadata := model.ResourceMetadata{
		ID:        resourceID,
		Type:      resType,
		CreatedAt: time.Date(2020, 07, 16, 18, 12, 20, 0, time.UTC),
		UpdatedAt: time.Date(2020, 07, 16, 18, 12, 20, 0, time.UTC),
		DeletedAt: sql.NullTime{Valid: false},
		Params:    params,
		Data:      data,
	}

	m.
		EXPECT().
		SelectResourceMetadata(resourceID).
		Return(metadata, true, nil).
		Times(1)
	a.
		EXPECT().
		DeleteBucket(bucketName).
		Return(nil).
		Times(1)

	m.
		EXPECT().
		DeleteResourceMetadata(resourceID, gomock.AssignableToTypeOf(time.Now())).
		Return(nil).
		Times(1)

	header := http.Header{}
	jsonSecrets, _ := json.Marshal(map[string]interface{}{"account": account})
	header.Add("Humanitec-Driver-Secrets", base64.StdEncoding.EncodeToString(jsonSecrets))

	res := ExecuteRequestHeader(s, http.MethodDelete, "/"+resourceID, nil, header, t)

	is.Equal(res.Code, http.StatusNoContent)
}

func TestDeleteAWSResource_DoesNotExist(t *testing.T) {
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
	resourceID := "test-db-id"

	account := AWSCredentials{
		AccessKeyID:     accessKeyId,
		SecretAccessKey: secretAccessKey,
	}

	m.
		EXPECT().
		SelectResourceMetadata(resourceID).
		Return(model.ResourceMetadata{}, false, nil).
		Times(1)

	header := http.Header{}
	jsonSecrets, _ := json.Marshal(map[string]interface{}{"account": account})
	header.Add("Humanitec-Driver-Secrets", base64.StdEncoding.EncodeToString(jsonSecrets))

	res := ExecuteRequestHeader(s, http.MethodDelete, "/"+resourceID, nil, header, t)

	is.Equal(res.Code, http.StatusNotFound)
}
