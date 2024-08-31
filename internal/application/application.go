package application

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"github.com/jgndev/jgn.dev/internal/models"
	"gopkg.in/yaml.v2"
)

// Application is a struct that holds AWS clients and application data
type Application struct {
	S3Client        *s3.Client
	DynamoClient    *dynamodb.Client
	SQSClient       *sqs.Client
	Blog            models.Posts
	S3BucketName    string
	DynamoTableName string
	SQSQueueURL     string
}

// NewApplication returns an instantiated instance of Application
func NewApplication(posts models.Posts, s3Client *s3.Client, dynamoClient *dynamodb.Client, sqsClient *sqs.Client, s3BucketName, dynamoTableName, sqsQueueURL string) *Application {
	return &Application{
		Blog:            posts,
		S3Client:        s3Client,
		DynamoClient:    dynamoClient,
		SQSClient:       sqsClient,
		S3BucketName:    s3BucketName,
		DynamoTableName: dynamoTableName,
		SQSQueueURL:     sqsQueueURL,
	}
}

// ReloadPosts will initiate a read from the DynamoDB table and reload
// all posts in memory.
func (a *Application) ReloadPosts(ctx context.Context) error {
	err := a.Blog.LoadFromDynamoDB(ctx, a.DynamoClient, a.DynamoTableName)
	if err != nil {
		log.Println("Failed to reload posts from DynamoDB.")
		return err
	}

	return nil
}

// ProcessS3Upload handles a new file upload to S3
func (a *Application) ProcessS3Upload(ctx context.Context, key string) error {
	// Get the file from S3
	result, err := a.S3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(a.S3BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	defer result.Body.Close()

	// Read the file content
	content, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return err
	}

	// Split frontmatter and content
	parts := strings.SplitN(string(content), "---", 3)
	if len(parts) != 3 {
		return fmt.Errorf("invalid markdown format")
	}

	// Parse frontmatter
	var post models.Post
	err = yaml.Unmarshal([]byte(parts[1]), &post)
	if err != nil {
		return err
	}

	// Parse markdown to HTML
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	parser := parser.NewWithExtensions(extensions)
	md := []byte(parts[2])
	html := markdown.ToHTML(md, parser, nil)

	post.HTMLContent = string(html)
	post.ID = key // Use S3 key as post ID

	// Save the post to DynamoDB
	err = post.SaveToDynamoDB(ctx, a.DynamoClient, a.DynamoTableName)
	if err != nil {
		return err
	}

	// Reload posts
	return a.ReloadPosts(ctx)
}

// PollSQS continuously polls the SQS queue for messages
func (a *Application) PollSQS(ctx context.Context) {
	for {
		msgResult, err := a.SQSClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(a.SQSQueueURL),
			MaxNumberOfMessages: 1,
			WaitTimeSeconds:     20,
		})

		if err != nil {
			log.Printf("Failed to receive message: %v", err)
			continue
		}

		for _, message := range msgResult.Messages {
			// Parse the message to get the S3 object key
			var s3Event struct {
				Records []struct {
					S3 struct {
						Object struct {
							Key string `json:"key"`
						} `json:"object"`
					} `json:"s3"`
				} `json:"Records"`
			}
			err := json.Unmarshal([]byte(*message.Body), &s3Event)
			if err != nil {
				log.Printf("Failed to parse S3 event: %v", err)
				continue
			}

			// Process the S3 upload
			if len(s3Event.Records) > 0 {
				err = a.ProcessS3Upload(ctx, s3Event.Records[0].S3.Object.Key)
				if err != nil {
					log.Printf("Failed to process S3 upload: %v", err)
					continue
				}
			}

			// Delete the message from the queue
			_, err = a.SQSClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
				QueueUrl:      aws.String(a.SQSQueueURL),
				ReceiptHandle: message.ReceiptHandle,
			})

			if err != nil {
				log.Printf("Failed to delete message: %v", err)
			}
		}
	}
}
