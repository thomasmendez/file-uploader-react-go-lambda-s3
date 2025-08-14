package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var S3_BUCKET string
var REGION string

func main() {
	S3_BUCKET = os.Getenv("BUCKET_NAME")
	if S3_BUCKET == "" {
		fmt.Println("'BUCKET_NAME' environment variable is not set")
		os.Exit(1)
	}
	REGION = os.Getenv("REGION")
	if REGION == "" {
		fmt.Println("'REGION' environment variable is not set")
		os.Exit(1)
	}
	lambda.Start(handler)
}

type FileData struct {
	Filename    string
	Content     []byte
	ContentType string
}

type ResponseBody struct {
	Message     string `json:"message"`
	Method      string `json:"method"`
	Path        string `json:"path"`
	ImageSize   int    `json:"imageSize"`
	ContentType string `json:"contentType"`
	S3Key       string `json:"s3Key,omitempty"`
	S3URL       string `json:"s3Url,omitempty"`
}

type JSONPayload struct {
	ImageData   string `json:"imageData"`
	ContentType string `json:"contentType"`
	Filename    string `json:"filename"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Printf("Request Method: %s\n", request.HTTPMethod)
	fmt.Printf("Request Path: %s\n", request.Path)
	fmt.Printf("IsBase64Encoded: %t\n", request.IsBase64Encoded)
	fmt.Printf("Content-Type: %s\n", request.Headers["Content-Type"])
	fmt.Printf("Body length: %d\n", len(request.Body))

	response := ResponseBody{
		Method:      request.HTTPMethod,
		Path:        request.Path,
		ContentType: request.Headers["Content-Type"],
	}

	imageData, err := getRequestImageData(request)
	if err != nil {
		response.Message = fmt.Sprintf("Failed to get image data: %s", err.Error())
		return createResponse(400, response), nil
	}

	// Validate it's actually an image by checking magic bytes
	contentType := detectContentType(imageData)

	response.ContentType = contentType

	// Upload to S3
	s3Key, s3URL, err := uploadToS3(ctx, imageData, contentType)
	if err != nil {
		fmt.Printf("Failed to upload to S3: %s", err.Error())
		response.Message = fmt.Sprintf("Failed to upload to S3: %s", err.Error())
		return createResponse(500, response), nil
	}

	response.S3Key = s3Key
	response.S3URL = s3URL
	response.Message = "Image successfully uploaded to S3 bucket!"

	return createResponse(200, response), nil
}

func getRequestImageData(request events.APIGatewayProxyRequest) ([]byte, error) {
	if strings.Contains(request.Headers["Content-Type"], "application/json") {
		var jsonPayload JSONPayload
		if err := json.Unmarshal([]byte(request.Body), &jsonPayload); err == nil {
			fmt.Println("Processing JSON-wrapped base64 data...")

			if jsonPayload.ImageData == "" {
				return nil, fmt.Errorf("no image data found in JSON payload")
			}

			imageData, err := base64.StdEncoding.DecodeString(jsonPayload.ImageData)

			if err != nil {
				return nil, fmt.Errorf("failed to decode base64 image from JSON: %w", err)
			}

			return imageData, nil
		} else {
			return nil, fmt.Errorf("invalid JSON payload")
		}
	} else if strings.Contains(request.Headers["Content-Type"], "multipart/form-data") {
		fmt.Println("Processing multipart/form-data...")

		_, files, err := parseMultipartFormData(
			request.Body,
			request.Headers["Content-Type"],
			request.IsBase64Encoded,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to parse multipart data: %w", err)
		}

		// Check if image file exists
		imageFile, exists := files["image"]
		if !exists {
			return nil, fmt.Errorf("no image file found")
		}

		if imageFile.Content == nil {
			return nil, fmt.Errorf("no image data found in multipart/form-data")
		}

		return imageFile.Content, nil
	} else if request.IsBase64Encoded { // based on BinaryMediaTypes in template.yaml
		// API Gateway automatically base64 encodes request body as binary data
		fmt.Println("Decoding base64 data...")
		imageData, err := base64.StdEncoding.DecodeString(request.Body)

		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 image: %w", err)
		}

		if imageData == nil {
			return nil, fmt.Errorf("no image data found in base64 data")
		}

		return imageData, nil
	} else {
		fmt.Println("Processing raw binary data...")
		return []byte(request.Body), nil
	}
}

func parseMultipartFormData(body string, contentType string, isBase64Encoded bool) (map[string]string, map[string]FileData, error) {
	// Parse the content type to get the boundary
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse content type: %w", err)
	}

	boundary := params["boundary"]
	if boundary == "" {
		return nil, nil, fmt.Errorf("no boundary found in content type")
	}

	// Decode the body if it's base64 encoded
	var bodyBytes []byte
	if isBase64Encoded {
		bodyBytes, err = base64.StdEncoding.DecodeString(body)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decode base64 body: %w", err)
		}
	} else {
		bodyBytes = []byte(body)
	}

	// Create a multipart reader
	reader := multipart.NewReader(strings.NewReader(string(bodyBytes)), boundary)

	formData := make(map[string]string)
	files := make(map[string]FileData)

	// Parse each part
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read multipart part: %w", err)
		}

		// Read the content of this part
		content, err := io.ReadAll(part)
		if err != nil {
			part.Close()
			return nil, nil, fmt.Errorf("failed to read part content: %w", err)
		}

		fieldName := part.FormName()
		filename := part.FileName()

		if filename != "" {
			// This is a file
			files[fieldName] = FileData{
				Filename:    filename,
				Content:     content,
				ContentType: part.Header.Get("Content-Type"),
			}
		} else {
			// This is a regular form field
			formData[fieldName] = string(content)
		}

		part.Close()
	}

	return formData, files, nil
}

func uploadToS3(ctx context.Context, imageData []byte, contentType string) (string, string, error) {
	// Load AWS config
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(REGION))
	if err != nil {
		return "", "", fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	s3Client := s3.NewFromConfig(cfg)

	// Generate unique key for the image
	timestamp := time.Now().Unix()
	extension := getExtensionFromContentType(contentType)
	s3Key := fmt.Sprintf("uploads/%d%s", timestamp, extension)

	// Upload to S3
	fmt.Println("Uploading file to S3 bucket...")
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(S3_BUCKET),
		Key:         aws.String(s3Key),
		Body:        strings.NewReader(string(imageData)),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to upload to S3: %w", err)
	}
	fmt.Println("Uploaded file to S3 bucket!")
	// Generate S3 URL
	s3URL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", S3_BUCKET, s3Key)

	return s3Key, s3URL, nil
}

func detectContentType(data []byte) string {
	if len(data) < 8 {
		return "application/octet-stream"
	}

	// Check magic bytes for common image formats
	switch {
	case data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF:
		return "image/jpeg"
	case data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47:
		return "image/png"
	case data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46:
		return "image/gif"
	case data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 &&
		data[8] == 0x57 && data[9] == 0x45 && data[10] == 0x42 && data[11] == 0x50:
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}

func getExtensionFromContentType(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ".bin"
	}
}

func createResponse(statusCode int, body ResponseBody) events.APIGatewayProxyResponse {
	jsonResponse, _ := json.Marshal(body)

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       string(jsonResponse),
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Headers": "Content-Type",
			"Access-Control-Allow-Methods": "POST, GET, OPTIONS",
		},
	}
}
