# UTE Consumption InfluxDB Exporter

## Configuration

### Arguments and env vars

| Description                    | Flag                     | Env var               |
|--------------------------------|--------------------------|-----------------------|
| Enable debug log level         | -debug string            | DEBUG               |
| From date                      | -from string             | FROM                |
| InfluxDB API URL               | -influxdb-api-url string | INFLUXDB_API_URL    |
| InfluxDB Bucket                | -influxdb-bucket string  | INFLUXDB_BUCKET     |
| InfluxDB Organization          | -influxdb-org string     | INFLUXDB_ORG        |
| InfluxDB Token                 | -influxdb-token string   | INFLUXDB_TOKEN      |
| To date                        | -to string               | TO                  |
| UTE Password                   | -ute-password string     | UTE_PASSWORD        |
| UTE Service Number             | -ute-service string      | UTE_SERVICE         |
| UTE Username                   | -ute-user string         | UTE_USER            |


### Env file example

```
UTE_USER=<user>
UTE_PASSWORD=<pasword>
UTE_SERVICE=<service>

INFLUXDB_API_URL=http://localhost:8086
INFLUXDB_ORG=ute
INFLUXDB_BUCKET=ute
INFLUXDB_TOKEN=my-super-secret-token
```

## Usage

Build the project with the following command:

```shell
make build
```

```shell
bin/export -from 01-08-2024 -to 03-08-2024
```

