package application

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
	"github.com/spf13/viper"
	"io"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/jgndev/jgn.dev/internal/models"
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
		return fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer result.Body.Close()

	// Read the file
	//content, err := ioutil.ReadAll(result.Body)
	content, err := io.ReadAll(result.Body)
	if err != nil {
		return fmt.Errorf("failed to read file content: %w", err)
	}

	// Parse front matter and content
	frontMatter, bodyContent, err := FrontMatter(content)
	if err != nil {
		return fmt.Errorf("failed to parse front matter: %w", err)
	}

	// Convert markdown to HTML
	htmlContent, err := MarkdownToHTML([]byte(bodyContent))
	if err != nil {
		return fmt.Errorf("failed to convert markdown to HTML: %w", err)
	}

	// Create post
	post := models.Post{
		ID:          key, // Use S3 key as post ID
		Title:       frontMatter.Title,
		Date:        frontMatter.Date,
		Author:      frontMatter.Author,
		Summary:     frontMatter.Summary,
		Slug:        frontMatter.Slug,
		HTMLContent: htmlContent,
		Tags:        frontMatter.Tags,
		Published:   frontMatter.Published,
	}

	// Save the post to DynamoDB
	err = post.SaveToDynamoDB(ctx, a.DynamoClient, a.DynamoTableName)
	if err != nil {
		return fmt.Errorf("failed to save post to DynamoDB: %w", err)
	}

	// Reload posts
	return a.ReloadPosts(ctx)
}

// FrontMatter parses the front matter and body content from a given markdown byte slice.
// The front matter is expected to be enclosed in "---" delimiters at the beginning of the markdown.
// This function returns a FrontMatter struct containing the parsed front matter, a string containing
// the body content of the markdown (excluding the front matter), and an error if the front matter
// cannot be parsed into the struct. If no front matter is present, the function returns the entire
// markdown content as the body with an empty FrontMatter struct and no error.
//
// Parameters:
// - markdown: A byte slice containing the markdown content to be parsed.
//
// Returns:
// - A models.FrontMatter struct containing the parsed front matter.
// - A string containing the body content of the markdown.
// - An error if the parsing fails, nil otherwise.
func FrontMatter(markdown []byte) (models.FrontMatter, string, error) {
	parts := strings.SplitN(string(markdown), "---", 3)
	if len(parts) < 3 {
		return models.FrontMatter{}, string(markdown), nil
	}

	matter := parts[1]
	bodyContent := parts[2]

	v := viper.New()
	v.SetConfigType("yaml")
	err := v.ReadConfig(bytes.NewBufferString(matter))
	if err != nil {
		return models.FrontMatter{}, "", err
	}

	var frontMatter models.FrontMatter
	err = v.Unmarshal(&frontMatter)
	if err != nil {
		return models.FrontMatter{}, "", err
	}

	return frontMatter, bodyContent, nil
}

// MarkdownToHTML converts a markdown byte slice into sanitized HTML. It uses the Blackfriday
// Markdown processor and the Bluemonday sanitizer to safely convert markdown content to HTML.
// The function preserves classes on fenced code blocks for syntax highlighting and applies
// additional security and formatting policies using Bluemonday. It returns the HTML as a string
// and an error if the conversion process fails.
//
// Parameters:
// - bytes: A byte slice containing the markdown content to be converted.
//
// Returns:
// - A string containing the sanitized HTML representation of the markdown.
// - An error if the conversion fails, nil otherwise.
//
// Notes:
//   - The function configures Bluemonday to allow class attributes matching "language-[a-zA-Z0-9]+"
//     on "code" elements, requires parseable URLs, and adds target="_blank" to fully qualified links.
//   - For more information on Bluemonday's policy configuration, visit:
//     https://pkg.go.dev/github.com/microcosm-cc/bluemonday
func MarkdownToHTML(bytes []byte) (string, error) {
	// README:
	// Reference https://github.com/russross/blackfriday
	// Preserve the classes on fenced code blocks for syntax highlighting
	output := blackfriday.Run(bytes)
	p := bluemonday.UGCPolicy()
	// Additional policy configuration:
	// See https://pkg.go.dev/github.com/microcosm-cc/bluemonday for more information.
	p.AllowAttrs("class").Matching(regexp.MustCompile("^language-[a-zA-Z0-9]+$")).OnElements("code")
	p.RequireParseableURLs(true)
	p.AddTargetBlankToFullyQualifiedLinks(true)
	html := p.SanitizeBytes(output)

	return string(html), nil
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
