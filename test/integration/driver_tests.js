const chai = require('chai');
const should = chai.should();
const chaiHttp = require('chai-http');
chai.use(chaiHttp);

// NOTE: These tests use the configuration in docker-compose.xml
describe('S3', () => {

  const driverService = "http://localhost:8080";

  // Create a random organization for our integration test
  const resourceId = "test-resource-" + Math.floor(Math.random()*65536).toString(16);


const accountJson = {
  "aws_access_key_id": "key",
  "aws_secret_access_key": "secret"
};

const drd = {
  "id": resourceId,
  "type": "s3",
  "resource_params" : {},
  "driver_params": {
    "region": "eu-west-1"
  },
  "driver_secrets": {
    "account": accountJson,
  }
};

  let resourceData;
  it('should create a new bucket.', (done) => {
    chai.request(driverService)
      .post(`/`)
      .set('Content-Type', 'application/json')
      .send(drd)
      .end((err, res) => {
        should.not.exist(err);
        res.status.should.equal(200);
        resourceData = res.body;
        done();
      });
  });

  it('should return the same bucket.', (done) => {
    chai.request(driverService)
      .post(`/`)
      .set('Content-Type', 'application/json')
      .send(drd)
      .end((err, res) => {
        should.not.exist(err);
        res.status.should.equal(200);
        res.body.data.values.bucket.should.equal(resourceData.data.values.bucket);
        done();
      });
  });
  
  it('should delete the bucket.', (done) => {
    const base64DriverParamsJson = Buffer.from(JSON.stringify(drd.driver_params)).toString('base64');
    const base64DriverSecretsJson = Buffer.from(JSON.stringify(drd.driver_secrets)).toString('base64');
    chai.request(driverService)
      .delete(`/${resourceId}`)
      .set('Humanitec-Driver-Params', base64DriverParamsJson)
      .set('Humanitec-Driver-Secrets', base64DriverSecretsJson)
      .end((err, res) => {
        should.not.exist(err);
        res.status.should.equal(204);
        done();
      });
  });
});
