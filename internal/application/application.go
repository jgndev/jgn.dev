package application

import (
	"context"
	"log"

	"github.com/jgndev/jgn.dev/internal/models"

	"cloud.google.com/go/firestore"
)

// Application is a struct that holds a Firestor client and collection of posts
type Application struct {
	FirestoreClient *firestore.Client
	Blog            models.Posts
}

// NewApplication returns an instantiated instance of Application
func NewApplication(posts models.Posts, client *firestore.Client) *Application {
	return &Application{
		Blog:            posts,
		FirestoreClient: client,
	}
}

// ReloadPosts will initiate a read from the collection in Firestore and reload
// all posts in memory.
func (a *Application) ReloadPosts(ctx context.Context) error {
	err := a.Blog.LoadFromFirestore(ctx, a.FirestoreClient)
	if err != nil {
		log.Println("Failed to reload posts from firestore.")
		return err
	}

	return nil
}
