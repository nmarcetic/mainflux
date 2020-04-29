# MongoDB writer

MongoDB writer provides message repository implementation for MongoDB.

## Configuration

The service is configured using the environment variables presented in the
following table. Note that any unset variables will be replaced with their
default values.

| Variable                        | Description                                | Default                |
|---------------------------------|--------------------------------------------|------------------------|
| MF_NATS_URL                     | NATS instance URL                          | nats://localhost:4222  |
| MF_MONGO_WRITER_LOG_LEVEL       | Log level for MongoDB writer               | error                  |
| MF_MONGO_WRITER_PORT            | Service HTTP port                          | 8180                   |
| MF_MONGO_WRITER_DB              | Default MongoDB database name              | messages               |
| MF_MONGO_WRITER_DB_HOST         | Default MongoDB database host              | localhost              |
| MF_MONGO_WRITER_DB_PORT         | Default MongoDB database port              | 27017                  |
| MF_MONGO_WRITER_SUBJECTS_CONFIG | Configuration file path with subjects list | /config/subjects.toml  |
| MF_MONGO_WRITER_CONTENT_TYPE    | Message payload Content Type               | application/senml+json |

## Deployment

```yaml
  version: "3.7"
  mongodb-writer:
    image: mainflux/mongodb-writer:[version]
    container_name: [instance name]
    depends_on:
      - mongodb
      - nats
    expose:
      - [Service HTTP port]
    restart: on-failure
    environment:
      MF_NATS_URL: [NATS instance URL]
      MF_MONGO_WRITER_LOG_LEVEL: [MongoDB writer log level]
      MF_MONGO_WRITER_PORT: [Service HTTP port]
      MF_MONGO_WRITER_DB: [MongoDB name]
      MF_MONGO_WRITER_DB_HOST: [MongoDB host]
      MF_MONGO_WRITER_DB_PORT: [MongoDB port]
      MF_MONGO_WRITER_SUBJETCS_CONFIG: [Configuration file path with subjects list]
      MF_MONGO_WRITER_CONTENT_TYPE: [Message payload Content Type]
    ports:
      - [host machine port]:[configured HTTP port]
    volume:
      - ./subjects.yaml:/config/subjects.yaml
```

To start the service, execute the following shell script:

```bash
# download the latest version of the service
git clone https://github.com/mainflux/mainflux

cd mainflux

# compile the mongodb writer
make mongodb-writer

# copy binary to bin
make install

# Set the environment variables and run the service
MF_NATS_URL=[NATS instance URL] MF_MONGO_WRITER_LOG_LEVEL=[MongoDB writer log level] MF_MONGO_WRITER_PORT=[Service HTTP port] MF_MONGO_WRITER_DB=[MongoDB database name] MF_MONGO_WRITER_DB_HOST=[MongoDB database host] MF_MONGO_WRITER_DB_PORT=[MongoDB database port] MF_MONGO_WRITER_SUBJETCS_CONFIG=[Configuration file path with subjetcs list] $GOBIN/mainflux-mongodb-writer
```

## Usage

Starting service will start consuming normalized messages in SenML format.
