package aws

type fakeClient struct {
	region string
}

func (c fakeClient) CreateBucket(bucketName string) (string, error) {
	return c.region, nil
}

func (c fakeClient) DeleteBucket(bucketName string) error {
	return nil
}

func (c fakeClient) CreateElastiCacheRedis(clusterId string, cacheNodeType string, cacheAz string) error {
	return nil
}
func (c fakeClient) DeleteElastiCacheRedis(clusterId string) error {
	return nil
}

func FakeNew(accessKeyId, secretAccessKey, region string) (Client, error) {
	return fakeClient{
		region: region,
	}, nil
}
