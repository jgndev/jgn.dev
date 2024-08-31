package models

import (
	"context"
	"log"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Post is a struct of properties that make up a post in the application
type Post struct {
	ID          string   `dynamodbav:"id" yaml:"id"`
	Title       string   `dynamodbav:"title" yaml:"title"`
	Date        string   `dynamodbav:"date" yaml:"date"`
	Author      string   `dynamodbav:"author" yaml:"author"`
	Summary     string   `dynamodbav:"summary" yaml:"summary"`
	Slug        string   `dynamodbav:"slug" yaml:"slug"`
	HTMLContent string   `dynamodbav:"htmlContent" yaml:"-"`
	Tags        []string `dynamodbav:"tags" yaml:"tags"`
	Published   bool     `dynamodbav:"published" yaml:"published"`
}

// Posts is a collection of Post
type Posts []*Post

// LoadFromDynamoDB retrieves all posts from the DynamoDB table
func (p *Posts) LoadFromDynamoDB(ctx context.Context, client *dynamodb.Client, tableName string) error {
	params := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}

	paginator := dynamodb.NewScanPaginator(client, params)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return err
		}

		var posts []*Post
		err = attributevalue.UnmarshalListOfMaps(page.Items, &posts)
		if err != nil {
			return err
		}

		*p = append(*p, posts...)
	}

	// Sort the posts newest to oldest
	sort.Slice(*p, func(i, j int) bool {
		return (*p)[i].Date > (*p)[j].Date
	})

	return nil
}

// SaveToDynamoDB saves a post to DynamoDB
func (post *Post) SaveToDynamoDB(ctx context.Context, client *dynamodb.Client, tableName string) error {
	item, err := attributevalue.MarshalMap(post)
	if err != nil {
		return err
	}

	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})

	return err
}

// DeleteFromDynamoDB deletes a post from DynamoDB
func (post *Post) DeleteFromDynamoDB(ctx context.Context, client *dynamodb.Client, tableName string) error {
	_, err := client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: post.ID},
		},
	})

	return err
}

// GetAll returns all posts
func (p *Posts) GetAll() Posts {
	return *p
}

// GetByID returns a specific post from the collection and a boolean value that
// represents a successful retrieval
func (p *Posts) GetByID(id string) (*Post, bool) {
	for _, post := range *p {
		if post.ID == id {
			return post, true
		}
	}

	return nil, false
}

// GetBySlug returns a specific post by the value of the slug
func (p *Posts) GetBySlug(slug string) (*Post, bool) {
	for _, post := range *p {
		if post.Slug == slug {
			return post, true
		}
	}

	return nil, false
}

// GetByTag returns a collection of posts that match the provided tag
func (p *Posts) GetByTag(tag string) Posts {
	var taggedPosts Posts
	for _, post := range *p {
		for _, t := range post.Tags {
			if t == tag {
				taggedPosts = append(taggedPosts, post)
				break
			}
		}
	}

	return taggedPosts
}

// GetRecent returns a collection of the n most recent posts
func (p *Posts) GetRecent(n int) Posts {
	if len(*p) < n {
		return *p
	}

	return (*p)[:n]
}

// GetOldest returns a collection of posts that are the n oldest
func (p *Posts) GetOldest(n int) Posts {
	length := len(*p)
	if length < n {
		return *p
	}

	return (*p)[length-n:]
}

// FilterBySearchTerms returns a collection of posts that match a search term
func (p *Posts) FilterBySearchTerms(query string) Posts {
	log.Println("query:", query)

	var filteredPosts Posts
	uniqueMatches := make(map[string]bool)

	// Split the query into individual search terms
	searchTerms := strings.Fields(strings.ToLower(query))
	log.Println("search terms:", searchTerms)

	for _, post := range *p {
		matches := true

		for _, term := range searchTerms {
			lowerTitle := strings.ToLower(post.Title)
			lowerSummary := strings.ToLower(post.Summary)
			if !strings.Contains(lowerTitle, term) &&
				!strings.Contains(lowerSummary, term) &&
				!containsIgnoreCase(term, post.Tags) {
				matches = false
				break
			}
		}

		if matches && !uniqueMatches[post.ID] {
			uniqueMatches[post.ID] = true
			filteredPosts = append(filteredPosts, post)
		}
	}

	return filteredPosts
}

func containsIgnoreCase(term string, tags []string) bool {
	for _, tag := range tags {
		if strings.EqualFold(term, tag) {
			return true
		}
	}
	return false
}

//package models
//
//import (
//	"context"
//	"github.com/aws/aws-sdk-go-v2/aws"
//	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
//	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
//	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
//	"log"
//	"sort"
//	"strings"
//)
//
//// Post is a struct of properties that make up a post in the application
//type Post struct {
//	ID          string   `dynamodbav:"id" yaml:"id"`
//	Title       string   `dynamodbav:"title" yaml:"title"`
//	Date        string   `dynamodbav:"date" yaml:"date"`
//	Author      string   `dynamodbav:"author" yaml:"author"`
//	Summary     string   `dynamodbav:"summary" yaml:"summary"`
//	Slug        string   `dynamodbav:"slug" yaml:"slug"`
//	HTMLContent string   `dynamodbav:"htmlContent" yaml:"-"`
//	Tags        []string `dynamodbav:"tags" yaml:"tags"`
//	Published   bool     `dynamodbav:"published" yaml:"published"`
//}
//
//// Posts is a collection of Post
//type Posts []Post
//
//// LoadFromDynamoDB retrieves all posts from the DynamoDB table
//func (p *Posts) LoadFromDynamoDB(ctx context.Context, client *dynamodb.Client, tableName string) error {
//	params := &dynamodb.ScanInput{
//		TableName: aws.String(tableName),
//	}
//
//	paginator := dynamodb.NewScanPaginator(client, params)
//
//	for paginator.HasMorePages() {
//		page, err := paginator.NextPage(ctx)
//		if err != nil {
//			return err
//		}
//
//		var posts []Post
//		err = attributevalue.UnmarshalListOfMaps(page.Items, &posts)
//		if err != nil {
//			return err
//		}
//
//		*p = append(*p, posts...)
//	}
//
//	// Sort the posts newest to oldest
//	sort.Slice(*p, func(i, j int) bool {
//		return (*p)[i].Date > (*p)[j].Date
//	})
//
//	return nil
//}
//
//// SaveToDynamoDB saves a post to DynamoDB
//func (post *Post) SaveToDynamoDB(ctx context.Context, client *dynamodb.Client, tableName string) error {
//	item, err := attributevalue.MarshalMap(post)
//	if err != nil {
//		return err
//	}
//
//	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
//		TableName: aws.String(tableName),
//		Item:      item,
//	})
//
//	return err
//}
//
//// DeleteFromDynamoDB deletes a post from DynamoDB
//func (post *Post) DeleteFromDynamoDB(ctx context.Context, client *dynamodb.Client, tableName string) error {
//	_, err := client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
//		TableName: aws.String(tableName),
//		Key: map[string]types.AttributeValue{
//			"id": &types.AttributeValueMemberS{Value: post.ID},
//		},
//	})
//
//	return err
//}
//
//// GetAll returns all posts
//func (p *Posts) GetAll() Posts {
//	return p
//}
//
//// GetByID returns a specific post from the collection and a boolean value that
//// represents a successful retrieval
//func (p Posts) GetByID(id string) (*Post, bool) {
//	for i := range p {
//		if p[i].ID == id {
//			return &p[i], true
//		}
//	}
//
//	return nil, false
//}
//
//// GetBySlug returns a specific post by the value of the slug
//func (p Posts) GetBySlug(slug string) (*Post, bool) {
//	for i := range p {
//		if p[i].Slug == slug {
//			return &p[i], true
//		}
//	}
//
//	return nil, false
//}
//
//// GetByTag returns a collection of posts that match the provided slug
//func (p Posts) GetByTag(tag string) Posts {
//	var taggedPosts Posts
//	for _, post := range p {
//		for _, t := range post.Tags {
//			if t == tag {
//				taggedPosts = append(taggedPosts, post)
//				break
//			}
//		}
//	}
//
//	return taggedPosts
//}
//
//// GetRecent returns a collection of the n most recent posts
//func (p Posts) GetRecent(n int) Posts {
//	if len(p) < n {
//		return p
//	}
//
//	return p[:n]
//}
//
//// GetOldest returns a collection of posts that are the n oldest
//func (p Posts) GetOldest(n int) Posts {
//	length := len(p)
//	if len(p) < n {
//		return p
//	}
//
//	return p[length-n:]
//}
//
//// FilterBySearchTerms returns a collection of posts that match a search term
//func (p Posts) FilterBySearchTerms(query string) Posts {
//	log.Println("query:", query)
//
//	var filteredPosts Posts
//	uniqueMatches := make(map[string]bool)
//
//	// Split the query into individual search terms
//	searchTerms := strings.Fields(strings.ToLower(query))
//	log.Println("search terms:", searchTerms)
//
//	for _, post := range p {
//		matches := true
//
//		for _, term := range searchTerms {
//			lowerTitle := strings.ToLower(post.Title)
//			lowerSummary := strings.ToLower(post.Summary)
//			if !strings.Contains(lowerTitle, term) &&
//				!strings.Contains(lowerSummary, term) &&
//				!containsIgnoreCase(term, post.Tags) {
//				matches = false
//				break
//			}
//		}
//
//		if matches && !uniqueMatches[post.ID] {
//			uniqueMatches[post.ID] = true
//			filteredPosts = append(filteredPosts, post)
//		}
//	}
//
//	return filteredPosts
//}
//
//func containsIgnoreCase(term string, tags []string) bool {
//	for _, tag := range tags {
//		if strings.EqualFold(term, tag) {
//			return true
//		}
//	}
//	return false
//}
