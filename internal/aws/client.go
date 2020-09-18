//go:generate mockgen -destination mock_aws/client_mock.go humanitec.io/resources/driver-aws-external/internal/aws Client

package aws

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Client interface {
	CreateBucket(bucketName string) (string, error)
	DeleteBucket(bucketName string) error
}

type awsClient struct {
	sess   *session.Session
	region string
}

func New(accessKeyId, secretAccessKey, region string) (Client, error) {
	creds := credentials.NewStaticCredentials(accessKeyId, secretAccessKey, "")
	sess, err := session.NewSession(&aws.Config{
		Region:      &region,
		Credentials: creds,
	})
	if err != nil {
		log.Printf(`Error creating AWS Session: %v`, err)
		return nil, fmt.Errorf(`creating aws session: %w`, err)
	}
	return awsClient{
		sess:   sess,
		region: region,
	}, nil
}

func (c awsClient) CreateBucket(bucketName string) (string, error) {
	input := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(c.region),
		},
	}
	svc := s3.New(c.sess)
	bucketResult, err := svc.CreateBucket(input)
	if err != nil {
		var aerr awserr.Error
		if errors.As(err, &aerr) {
			switch aerr.Code() {
			case s3.ErrCodeBucketAlreadyExists:
				log.Printf(`Attempted to create s3 bucket that already exists: "%s"`, bucketName)
				return "", fmt.Errorf(`s3 bucket name already exists "%s": %w`, bucketName, aerr)
			case s3.ErrCodeBucketAlreadyOwnedByYou:
				log.Printf(`Attempted to create s3 bucket that already exists: "%s"`, bucketName)
				return "", fmt.Errorf(`s3 bucket name already exists "%s": %w`, bucketName, aerr)
			}
		}
		log.Printf(`Error creating s3 bucket "%s": %v`, bucketName, err)
		return "", fmt.Errorf(`creating s3 bucket "%s": %w`, bucketName, err)
	}
	return *bucketResult.Location, nil
}

func (c awsClient) DeleteBucket(bucketName string) error {
	// NOTE: This is not a full implementation. Buckets need to be empty before they can be deleted.
	// See https://docs.aws.amazon.com/AmazonS3/latest/dev/delete-or-empty-bucket.html#delete-bucket-awssdks
	input := &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	}
	svc := s3.New(c.sess)
	_, err := svc.DeleteBucket(input)
	if err != nil {
		log.Printf(`Error creating s3 bucket "%s": %v`, bucketName, err)
		return fmt.Errorf(`creating s3 bucket "%s": %w`, bucketName, err)
	}
	return nil
}
