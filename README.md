# zdts3

zdts3 is a tool for archiving recent files in a directory by purging old files from the source directory, compressing remaining files into a zip archive and uploading the zip archive to an S3 or S3-compatible bucket. This tool is intended to be used in backing up periodic database dumps for recovery purposes.

build with `go build .` for a binary or `docker build .` for a docker image.

## Usage

### Configuration

zdts3 can be configured using environment variables or command-line flags.

#### Environment Variables

- `endpoint`: S3 or S3-compatible endpoint.
- `accesskeyid`: S3 access key ID.
- `secretaccesskey`: S3 secret access key.
- `bucket`: S3 bucket name.
- `sourcedir`: Source directory to archive.
- `loglevel`: Log level (debug, info, warn, error, fatal).

#### Command-Line Flags

- `-endpoint`: S3 or S3-compatible endpoint.
- `-accesskeyid`: S3 access key ID.
- `-secretaccesskey`: S3 secret access key.
- `-bucket`: S3 bucket name.
- `-dir`: Source directory to archive.
- `-loglevel`: Log level (debug, info, warn, error, fatal).

### Docker Compose

To run the zdts3 using Docker Compose, create a `.env` file with the following parameters:

```env
endpoint=<your-s3-endpoint>
accesskeyid=<your-access-key-id>
secretaccesskey=<your-secret-access-key>
bucket=<your-bucket-name>
sourcedir=<your-source-directory>
zipfile=<your-zip-filename>
loglevel=info
```

Create a `docker-compose.yml` file with the following content:

```yaml
version: '3.8'

services:
  zdts3:
    image: <zdts3-docker-image>
    restart: always
    command:
      [
        "-loglevel=${loglevel}",
        "-endpoint=${endpoint}",
        "-accesskeyid=${accesskeyid}",
        "-secretaccesskey=${secretaccesskey}",
        "-bucket=${bucket}",
        "-dir=${dir}",
      ]
```

To start the zdts3 service, run `docker-compose up`
