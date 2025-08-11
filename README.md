# file-uploader-react-go-lambda-s3
Upload a image file to a s3 bucket through a react web app and go lambda backend

## Prerequisites

### Frontend Tools
- [Yarn](https://yarnpkg.com/getting-started/install)

### Backend Tools
- [Go 1.24.2](https://go.dev/doc/install)

### Containerization Tools
- [Docker](https://www.docker.com/products/docker-desktop)

### Cloud Tools
- [AWS Account](https://aws.amazon.com/account/)
- [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html)
- [SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html)

## Setup
- Clone the repository `git clone https://github.com/thomasmendez/file-uploader-react-go-lambda-s3.git`
- Install dependencies `cd frontend && yarn install`, `cd backend && go mod tidy`
- Configure AWS CLI
- Configure SAM CLI
- Deploy SAM template
### For Windows
- Install [build-lambda-zip](https://github.com/aws/aws-lambda-go?tab=readme-ov-file#for-developers-on-windows)

## Local Development
You will need to open a few terminals to run the application:

1. Run the React app `cd frontend && yarn dev`
2. Build the bootstrap file `cd backend && GOARCH=arm64 GOOS=linux go build -o bootstrap main.go`
3. Zip the bootstrap file `cd backend` and `C:\Users\<user>\go\bin\build-lambda-zip.exe -o lambda-handler.zip bootstrap` (on Windows Powershell)
4. Start Docker Desktop
5. For testing the lambda function locally, but still uploading the image to a S3 bucket, create a S3 bucket in the AWS Console and note the bucket name and region. Create a copy of the `env.sample.json` file with `cp env.sample.json env.json` and fill in the bucket name and region.
6. Run the SAM template `sam.cmd local start-api --template-file=template.yaml --env-vars env.json`
7. Upload a file through the React app [http://localhost:5173/](http://localhost:5173/)
8. The file will be uploaded to the designated S3 bucket

*Note: This is not a production ready application. This application assumes you have admin level access to AWS and SAM CLI configured.*

## License
MIT