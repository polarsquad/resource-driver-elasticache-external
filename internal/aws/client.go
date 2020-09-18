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
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Client interface {
	CreateBucket(bucketName string) (string, error)
	DeleteBucket(bucketName string) error
	CreateElastiCacheRedis(clusterId string, cacheNodeType string, cacheAz string) error
	DeleteElastiCacheRedis(clusterId string) error
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

func (c awsClient) CreateElastiCacheRedis(clusterId string, cacheNodeType string, cacheAz string) error {
	input := &elasticache.CreateCacheClusterInput{
		AutoMinorVersionUpgrade:   aws.Bool(true),
		CacheClusterId:            aws.String(clusterId),
		CacheNodeType:             aws.String(cacheNodeType),
		CacheSubnetGroupName:      aws.String("default"),
		Engine:                    aws.String("redis"),
		EngineVersion:             aws.String("5.0.6"),
		NumCacheNodes:             aws.Int64(1),
		Port:                      aws.Int64(6379),
		PreferredAvailabilityZone: aws.String(cacheAz),
		SnapshotRetentionLimit:    aws.Int64(7),
	}

	svc := elasticache.New(c.sess)
	_, err := svc.CreateCacheCluster(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elasticache.ErrCodeReplicationGroupNotFoundFault:
				log.Printf(`Replication group not found`)
				return fmt.Errorf(`Replication group not found`)
			case elasticache.ErrCodeInvalidReplicationGroupStateFault:
				log.Printf(`Invalid replication group state`)
				return fmt.Errorf(`Invalid replication group state`)
			case elasticache.ErrCodeCacheClusterAlreadyExistsFault:
				log.Printf(`Cache cluster already exists`)
				return fmt.Errorf(`Cache cluster already exists`)
			case elasticache.ErrCodeInsufficientCacheClusterCapacityFault:
				log.Printf(`Insufficient cache cluster capacity`)
				return fmt.Errorf(`Insufficient cache cluster capacity`)
			case elasticache.ErrCodeCacheSecurityGroupNotFoundFault:
				log.Printf(`Cache security group not found`)
				return fmt.Errorf(`Cache security group not found`)
			case elasticache.ErrCodeCacheSubnetGroupNotFoundFault:
				log.Printf(`Subnet group not found`)
				return fmt.Errorf(`Subnet group not found`)
			case elasticache.ErrCodeClusterQuotaForCustomerExceededFault:
				log.Printf(`Cluster quota for customer exceeded`)
				return fmt.Errorf(`Cluster quota for customer exceeded`)
			case elasticache.ErrCodeNodeQuotaForClusterExceededFault:
				log.Printf(`Quota for cluster exceeded`)
				return fmt.Errorf(`Quota for cluster exceeded`)
			case elasticache.ErrCodeNodeQuotaForCustomerExceededFault:
				log.Printf(`Node quota for customer exceeded`)
				return fmt.Errorf(`Node quota for customer exceeded`)
			case elasticache.ErrCodeCacheParameterGroupNotFoundFault:
				log.Printf(`Cache parameter group not found`)
				return fmt.Errorf(`Cache parameter group not found`)
			case elasticache.ErrCodeInvalidVPCNetworkStateFault:
				log.Printf(`Invalid VPC network state`)
				return fmt.Errorf(`Invalid VPC network state`)
			case elasticache.ErrCodeTagQuotaPerResourceExceeded:
				log.Printf(`Tag quota per resource exceeded`)
				return fmt.Errorf(`Tag quota per resource exceeded`)
			case elasticache.ErrCodeInvalidParameterValueException:
				log.Printf(`Invalid parameter value exception`)
				return fmt.Errorf(`Invalid parameter value exception`)
			case elasticache.ErrCodeInvalidParameterCombinationException:
				log.Printf(`Invalid parameter combination`)
				return fmt.Errorf(`Invalid parameter combination`)
			default:
				fmt.Println(aerr.Error())
			}
		}
		log.Printf(`Error creating Elasticache cluster "%s": %v`, clusterId, err)
		return fmt.Errorf(`creating Elasticache cluster "%s": %w`, clusterId, err)
	}
	return nil
}

func (c awsClient) DeleteElastiCacheRedis(clusterId string) error {
	input := &elasticache.DeleteCacheClusterInput{
		CacheClusterId: aws.String(clusterId),
	}

	svc := elasticache.New(c.sess)

	_, err := svc.DeleteCacheCluster(input)
	if err != nil {
		log.Printf(`Error deleting elasticache redis cluster "%s": %v`, clusterId, err)
		return fmt.Errorf(`deleting elasticache redis cluster "%s": %v`, clusterId, err)
	}
	return nil
}
