# Configure the AWS Provider
provider "aws" {
  region = var.aws_region
}

# S3 bucket for storing Markdown files
resource "aws_s3_bucket" "post_bucket" {
  bucket = var.bucket_name

  tags = {
    Name        = "Post Bucket"
    Environment = var.environment
  }
}

# S3 bucket ACL
resource "aws_s3_bucket_ownership_controls" "post_bucket" {
  bucket = aws_s3_bucket.post_bucket.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_acl" "post_bucket" {
  depends_on = [aws_s3_bucket_ownership_controls.post_bucket]

  bucket = aws_s3_bucket.post_bucket.id
  acl    = "private"
}

# DynamoDB table for storing parsed posts
resource "aws_dynamodb_table" "post_table" {
  name           = var.dynamodb_table_name
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "id"

  attribute {
    name = "id"
    type = "S"
  }

  attribute {
    name = "slug"
    type = "S"
  }

  global_secondary_index {
    name               = "SlugIndex"
    hash_key           = "slug"
    projection_type    = "ALL"
  }

  tags = {
    Name        = "Post Table"
    Environment = var.environment
  }
}

# SQS queue for notifications
resource "aws_sqs_queue" "post_queue" {
  name                      = var.sqs_queue_name
  delay_seconds             = 90
  max_message_size          = 2048
  message_retention_seconds = 86400
  receive_wait_time_seconds = 10

  tags = {
    Name        = "Post Notification Queue"
    Environment = var.environment
  }
}

# S3 event notification to SQS
resource "aws_s3_bucket_notification" "bucket_notification" {
  bucket = aws_s3_bucket.post_bucket.id

  queue {
    queue_arn     = aws_sqs_queue.post_queue.arn
    events        = ["s3:ObjectCreated:*", "s3:ObjectRemoved:*"]
    filter_suffix = ".md"
  }
}

# IAM role for EC2 instance (or ECS task) to access S3, DynamoDB, and SQS
resource "aws_iam_role" "app_role" {
  name = "app_role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"  # Change this to ecs-tasks.amazonaws.com if using ECS
        }
      }
    ]
  })
}

# IAM policy for app role
resource "aws_iam_role_policy" "app_policy" {
  name = "app_policy"
  role = aws_iam_role.app_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "s3:GetObject",
          "s3:ListBucket",
        ]
        Effect   = "Allow"
        Resource = [
          aws_s3_bucket.post_bucket.arn,
          "${aws_s3_bucket.post_bucket.arn}/*",
        ]
      },
      {
        Action = [
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:UpdateItem",
          "dynamodb:DeleteItem",
          "dynamodb:Scan",
          "dynamodb:Query",
        ]
        Effect   = "Allow"
        Resource = aws_dynamodb_table.post_table.arn
      },
      {
        Action = [
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes",
        ]
        Effect   = "Allow"
        Resource = aws_sqs_queue.post_queue.arn
      },
    ]
  })
}

# Output values
output "bucket_name" {
  description = "Name of the S3 bucket"
  value       = aws_s3_bucket.post_bucket.id
}

output "dynamodb_table_name" {
  description = "Name of the DynamoDB table"
  value       = aws_dynamodb_table.post_table.name
}

output "sqs_queue_url" {
  description = "URL of the SQS queue"
  value       = aws_sqs_queue.post_queue.id
}

output "iam_role_arn" {
  description = "ARN of the IAM role for the application"
  value       = aws_iam_role.app_role.arn
}