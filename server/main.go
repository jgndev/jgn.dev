package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	"github.com/jgndev/jgn.dev/internal/application"
	"github.com/jgndev/jgn.dev/internal/models"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Read environment variables on startup
	project := os.Getenv("GCP_PROJECT_ID")
	if project == "" {
		log.Fatal("GCP_PROJECT_ID environment variable not set!")
	}
	log.Printf("starting with GCP_PROJECT_ID: %v", project)

	topic := os.Getenv("GCP_TOPIC_NAME")
	if topic == "" {
		log.Fatal("GCP_TOPIC_NAME environment variable not set!")
	}
	log.Printf("starting with GCP_TOPIC_NAME: %v", topic)

	// Firestore
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, project)
	if err != nil {
		log.Fatalf("failed to create Firestore client: %v\n", err)
	}
	defer client.Close()

	var posts models.Posts
	if err = posts.LoadFromFirestore(ctx, client); err != nil {
		log.Fatalf("Failed to load posts: %v\n", err)
	}

	// Instantiate a new echo app
	e := echo.New()

	// Instantiate new instance of Application
	app := application.NewApplication(posts, client)

	// PubSub
	pubsubClient, err := pubsub.NewClient(ctx, topic)
	if err != nil {
		log.Printf("Failed to create pubsub client: %v\n", err)
	}
	defer pubsubClient.Close()

	// sub := pubsubClient.Subscription(topic)
	// go func() {
	// 	err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
	// 		log.Println("Message received, reloading posts")
	// 		if err = app.ReloadPosts(ctx); err != nil {
	// 			log.Printf("Failed to reload posts: %v\n", err)
	// 			msg.Nack()
	// 		} else {
	// 			msg.Ack()
	// 		}
	// 	})
	// 	if err != nil {
	// 		log.Printf("Failed to receive messages: %v\n", err)
	// 	}
	// }()

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
	e.GET("/sitemap.xml", app.SiteMap)
	e.GET("/get-time", func(c echo.Context) error {
		loc, _ := time.LoadLocation("America/Chicago")
		currentTime := time.Now().In(loc).Format("3:04 PM")
		return c.String(http.StatusOK, currentTime)
	})

	// Start app
	e.Logger.Fatal(e.Start(":8080"))
}
