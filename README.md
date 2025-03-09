# Archiver

Archiver is a tool for archiving recent files in a directory by:
1. Purging old files from the source directory.
2. Compressing remaining files into a zip archive.
3. Uploading the zip archive to an S3 or S3-compatible bucket.

## Installation

To install Archiver, clone the repository and build the binary:

```sh
git clone <repository-url>
cd archiver
go build .
```

## Usage

### Configuration

Archiver can be configured using environment variables or command-line flags.

#### Environment Variables

- `endpoint`: S3 or S3-compatible endpoint
- `accesskeyid`: S3 access key ID
- `secretaccesskey`: S3 secret access key
- `bucket`: S3 bucket name
- `sourcedir`: Source directory to archive
- `zipfile`: Zip filename
- `loglevel`: Log level (debug, info, warn, error, fatal)

#### Command-Line Flags

- `-endpoint`: S3 or S3-compatible endpoint
- `-accesskeyid`: S3 access key ID
- `-secretaccesskey`: S3 secret access key
- `-bucket`: S3 bucket name
- `-dir`: Source directory to archive
- `-zipfile`: Zip filename
- `-loglevel`: Log level (debug, info, warn, error, fatal)

### Docker Compose

To run the archiver using Docker Compose, create a `.env` file with the following content:

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
  archiver:
    image: archiver:latest
    restart: always
    command:
      [
        "-loglevel=${loglevel}",
        "-endpoint=${endpoint}",
        "-accesskeyid=${accesskeyid}",
        "-secretaccesskey=${secretaccesskey}",
        "-bucket=${bucket}",
        "-filename=${filename}",
        "-dir=${dir}",
      ]
```

To start the Archiver service, run:

```sh
docker-compose up
```

This tool was inspired by 
