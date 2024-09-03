package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/jgndev/jgn.dev/internal/application"
	"github.com/jgndev/jgn.dev/internal/models"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"os"
)

func main() {
	// Read environment variables on startup
	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		log.Fatal("AWS_REGION environment variable not set!")
	}
	log.Printf("starting with AWS_REGION: %v", awsRegion)

	s3BucketName := os.Getenv("S3_BUCKET_NAME")
	if s3BucketName == "" {
		log.Fatal("S3_BUCKET_NAME environment variable not set!")
	}
	log.Printf("starting with S3_BUCKET_NAME: %v", s3BucketName)

	dynamoTableName := os.Getenv("DYNAMODB_TABLE_NAME")
	if dynamoTableName == "" {
		log.Fatal("DYNAMODB_TABLE_NAME environment variable not set!")
	}
	log.Printf("starting with DYNAMODB_TABLE_NAME: %v", dynamoTableName)

	sqsQueueURL := os.Getenv("SQS_QUEUE_URL")
	if sqsQueueURL == "" {
		log.Fatal("SQS_QUEUE_URL environment variable not set!")
	}
	log.Printf("starting with SQS_QUEUE_URL: %v", sqsQueueURL)

	// AWS SDK configuration
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(awsRegion))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// S3 client
	s3Client := s3.NewFromConfig(cfg)

	// DynamoDB client
	dynamoClient := dynamodb.NewFromConfig(cfg)

	// SQS client
	sqsClient := sqs.NewFromConfig(cfg)

	var posts models.Posts
	if err = posts.LoadFromDynamoDB(ctx, dynamoClient, dynamoTableName); err != nil {
		log.Fatalf("Failed to load posts: %v\n", err)
	}

	// Instantiate a new echo app
	e := echo.New()

	// Instantiate new instance of Application
	app := application.NewApplication(posts, s3Client, dynamoClient, sqsClient, s3BucketName, dynamoTableName, sqsQueueURL)

	// Start SQS polling in a goroutine
	go app.PollSQS(ctx)

	// Enable gzip compression
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))

	// Static assets with caching
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Cache-Control", "public, max-age=86400")
			return next(c)
		}
	})

	// Static assets
	e.Static("/public", "public")
	e.File("/favicon.ico", "public/img/favicon.ico")
	e.File("/robots.txt", "public/txt/robots.txt")

	e.HTTPErrorHandler = app.CustomErrorHandler

	// Routes
	e.GET("/", app.Home)
	e.GET("/posts", app.Posts)
	e.GET("/posts/:slug", app.Post)
	e.GET("/about", app.About)
	e.GET("/contact", app.Contact)
	e.GET("/search", app.SearchPosts)
	e.GET("/health", app.Health)
	e.GET("/get-time", app.GetTime)
	e.GET("/sitemap.xml", app.SiteMap)

	// Start app
	e.Logger.Fatal(e.Start(":8080"))
}
