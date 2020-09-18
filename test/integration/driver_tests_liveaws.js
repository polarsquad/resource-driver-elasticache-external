const chai = require('chai');
const should = chai.should();
const chaiHttp = require('chai-http');
const fs = require('fs');
chai.use(chaiHttp);

// NOTE: These tests require the following to be run locally:
// Change docker-compose.xml by removing variable USE_FAKE_AWS_CLIENT
// Add a account.json file to the directory where the tests are run with valid static credentials for an AWS account.
describe('S3', () => {

  const driverService = "http://localhost:8080";

  // Create a random organization for our integration test
  const resourceId = "test-resource-" + Math.floor(Math.random()*65536).toString(16);


  const drd = {
    "id": resourceId,
    "type": "s3",
    "resource_params" : {},
    "driver_params": {
      "region": "eu-west-1"
    },
    "driver_secrets": {
      "account": {},
    }
  };

  
  function readFile(path) {
    return new Promise((accept, reject) => {
      fs.readFile(path, (err, data) => {
        if (err) {
          reject(err);
        } else {
          accept(data);
        }
      });
    });
  }

  before('load credentials from filesystem', () => {
    return readFile('account.json').then((accountJson) => drd.driver_secrets.account = JSON.parse(accountJson));
  });

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
    const base64AccountJson = Buffer.from(JSON.stringify(drd.driver_secrets.account)).toString('base64');
    chai.request(driverService)
      .delete(`/${resourceId}`)
      .set('Humanitec-Driver-Account', base64AccountJson)
      .end((err, res) => {
        should.not.exist(err);
        res.status.should.equal(204);
        done();
      });
  });
});
