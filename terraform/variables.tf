variable "aws_region" {
  description = "The AWS region to deploy resources in"
  type        = string
  default     = "us-west-2"  # Change this to your preferred region
}

variable "environment" {
  description = "The deployment environment (e.g., dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "bucket_name" {
  description = "Name of the S3 bucket to store Markdown files"
  type        = string
}

variable "dynamodb_table_name" {
  description = "Name of the DynamoDB table to store parsed posts"
  type        = string
  default     = "Posts"
}

variable "sqs_queue_name" {
  description = "Name of the SQS queue for notifications"
  type        = string
  default     = "post-notification-queue"
}