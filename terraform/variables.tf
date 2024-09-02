variable "aws_region" {
  description = "The AWS region to deploy resources in"
  type        = string
  default     = "us-east-2"
}

variable "environment" {
  description = "The deployment environment (e.g., dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "bucket_name" {
  description = "Name of the S3 bucket to store Markdown files"
  type        = string
  default     = "jgn-posts-bucket"
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

variable "ecr_repository_name" {
  description = "Name of the ECR repository"
  type        = string
  default     = "jgn-dev-image"
}

variable "project_name" {
  description = "Name of the project"
  type        = string
  default     = "jgn-dev"
}

variable "subnet_ids" {
  description = "Name of the Subnet IDs"
  type        = list(string)
}

variable "vpc_id" {
  description = "Name of the VPC"
  type        = string
  default     = "jgn-dev-vpc"
}