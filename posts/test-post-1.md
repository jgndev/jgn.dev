---
ID: "test-post-1.md"
title: "Let's build a blog"
date: 2024-01-01
author: "jgn"
tags: ["Go", "GCP"]
summary: "How I built this blog from scratch with Go and Google Cloud Platform"
slug: "test-post-1"
image: ""
published: false
---

![logo](https://storage.googleapis.com/jgn-dev-pub/go-logo.svg)

In this series of posts I talk about how this blog was built with [Go](https://go.dev)
and [Google Cloud Platform](https://cloud.google.com).

The post you are reading right now is part one.


## Blog design overview

I had a few goals in mind when creating a shiny new blog for 2024. 

- I want to create and edit posts in Markdown locally on my machine and store
them in a GitHub repo. Markdown with YAML frontmatter seems like the most flexible
and re-usable format going for documents and blog posts so I'll be using it here.

- It would be great to have some kind of ingestion service that just does everything for me
when I upload, change or delete a file from Cloud Storage. It should handle parsing
the files and do the creating, updating or deleting entries in Firestore automagically. 
That would be sweet. 

- The blog website is a good excuse to play around with [echo](https://linktoecho.com),
[a-h/templ](https://linktotempl.com) and [htmx](https://htmxorsomething.com). 
The blog should read posts from Firestore into memory on startup and *somehow* know
when there is a change to reload them. 


## Another heading

### And another heading


Some go code:


```go
func main() {
  fmt.Println("Hello, Interwebs")
}
```
