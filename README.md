# resources/driver-aws-external

`resources/driver-aws-external` is a driver for AWS resources - currently only supports S3 buckets. It implements the Humanitec driver specification.


## Configuration
It takes the following environment variables:

### Service

| Variable | Description |
|---|---|
| `USE_FAKE_AWS_CLIENT` | [Optional] If set does not actually contact AWS. Useful for local testing. |
| `PORT` | [Optional] The port number the server should be exposed on. It defaults to `8080`. |

### Metadata Database

| Variable | Description |
|---|---|
| `DATABASE_NAME` | The name of the Postgress DB to connect to. |
| `DATABASE_USER` | The userame that the service should access the database under. |
| `DATABASE_PASSWORD` | The password associated with the useranme. |
| `DATABASE_HOST` | The DNS name or IP address that the database server resides on. |
| `DATABASE_PORT` | [Optional] The port on the server that the database is listening on. It defaults to `5432`. |

**NOTE:** You can find examples of all the above variables in the `docker-compose.yml` file in the root of the repo.

## Supported endpoints

| Method | Path Template | Description |
| --- | --- | ---|
| `POST` | `/` | Create or Update a resource. Payload should be a DriverResourceDefinition. |
| `DELETE` | `/{resourceId}` | Deletes a resource. |

### System Endpoints
| Method | Path Template | Description |
| --- | --- | ---|
| `GET` | `/alive` | Should be used for liveness probe |
| `GET` | `/health | Should be used for readiness probe |

## Running locally

The service can be built with:

    $ go build humanitec.io/resources/driver-aws-external/cmd/driver

Mocks can be generated with:

    $ go generate ./...

Tests can be run with:

    $ go test ./...


## Testing with a database

The Go unit tests do not cover any of the database code. Tests on this can be run as follows:

Build the image and run it with docker-compose (in the root of the repo):

    $ docker-compose up --build

Build the package and run the integration tests (in the `/test/integration` directory:

    $ npm install
    $ npm run test

## Testing against real AWS

This requires some additional setup:
* Change `docker-compose.yml` by removing variable `USE_FAKE_AWS_CLIENT`
* Add an `account.json` file to  `/test/integration` containing valid credentials for an AWS Static Token.

The format of the `account.json` file is as follows:
    {
      "aws_access_key_id": "<AWS_ACCESS_KEY_ID>",
      "aws_secret_access_key": "<AWS_SECFRET_ACCESS_KEY>"
    }

Where `<AWS_ACCESS_KEY_ID>` and `<AWS_SECFRET_ACCESS_KEY>` should be replaces by the relevant values.

After starting with the modified `docker-compose.yml`, Run the tests with the following commend in `/test/integration`

    $ ./node_modules/.bin/mocha driver_tests_liveaws.js
