terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.0" # Use the latest 3.x version
    }
  }
}

# Provider configuration
provider "aws" {
  region = var.aws_region
}

# VPC
resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "${var.project_name}-vpc"
  }
}

# Internet Gateway
resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "${var.project_name}-igw"
  }
}

# Subnets
resource "aws_subnet" "main" {
  count                   = 2
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.${count.index}.0/24"
  availability_zone       = data.aws_availability_zones.available.names[count.index]
  map_public_ip_on_launch = true

  tags = {
    Name = "${var.project_name}-subnet-${count.index + 1}"
  }
}

# Route Table
resource "aws_route_table" "main" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = {
    Name = "${var.project_name}-rt"
  }
}

# Route Table Association
resource "aws_route_table_association" "main" {
  count          = 2
  subnet_id      = aws_subnet.main[count.index].id
  route_table_id = aws_route_table.main.id
}

# ECR Repository
resource "aws_ecr_repository" "app" {
  name                 = var.ecr_repository_name
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }
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
  bucket     = aws_s3_bucket.post_bucket.id
  acl        = "private"
}

# S3 bucket for storing publicly accessible resume
resource "aws_s3_bucket" "resume_bucket" {
  bucket = "${var.project_name}-public-resume"

  tags = {
    Name        = "Public Resume Bucket"
    Environment = var.environment
  }
}

# S3 bucket ownership controls
resource "aws_s3_bucket_ownership_controls" "resume_bucket_ownership" {
  bucket = aws_s3_bucket.resume_bucket.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

# S3 bucket public access block
resource "aws_s3_bucket_public_access_block" "resume_bucket_public_access" {
  bucket = aws_s3_bucket.resume_bucket.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

# S3 bucket ACL
resource "aws_s3_bucket_acl" "resume_bucket_acl" {
  depends_on = [
    aws_s3_bucket_ownership_controls.resume_bucket_ownership,
    aws_s3_bucket_public_access_block.resume_bucket_public_access,
  ]

  bucket = aws_s3_bucket.resume_bucket.id
  acl    = "public-read"
}

# S3 bucket policy
resource "aws_s3_bucket_policy" "resume_bucket_policy" {
  bucket = aws_s3_bucket.resume_bucket.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "PublicReadGetObject"
        Effect    = "Allow"
        Principal = "*"
        Action    = "s3:GetObject"
        Resource  = "${aws_s3_bucket.resume_bucket.arn}/*"
      },
    ]
  })
}

resource "aws_s3_bucket_public_access_block" "resume_bucket_public_access" {
  bucket = aws_s3_bucket.resume_bucket.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

# DynamoDB table for storing parsed posts
resource "aws_dynamodb_table" "post_table" {
  name         = var.dynamodb_table_name
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  attribute {
    name = "id"
    type = "S"
  }

  attribute {
    name = "slug"
    type = "S"
  }

  global_secondary_index {
    name            = "SlugIndex"
    hash_key        = "slug"
    projection_type = "ALL"
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

resource "aws_sqs_queue_policy" "post_queue_policy" {
  queue_url = aws_sqs_queue.post_queue.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "s3.amazonaws.com"
        }
        Action = [
          "sqs:SendMessage"
        ]
        Resource = aws_sqs_queue.post_queue.arn
        Condition = {
          ArnEquals = {
            "aws:SourceArn" = aws_s3_bucket.post_bucket.arn
          }
        }
      }
    ]
  })
}

# S3 event notification to SQS
resource "aws_s3_bucket_notification" "bucket_notification" {
  bucket = aws_s3_bucket.post_bucket.id

  queue {
    queue_arn     = aws_sqs_queue.post_queue.arn
    events        = ["s3:ObjectCreated:*", "s3:ObjectRemoved:*"]
    filter_suffix = ".md"
  }

  depends_on = [aws_sqs_queue_policy.post_queue_policy]
}

# ECS Cluster
resource "aws_ecs_cluster" "main" {
  name = "${var.project_name}-cluster"
}

# ECS Task Definition
resource "aws_ecs_task_definition" "app" {
  family                   = "${var.project_name}-task"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_execution_role.arn
  task_role_arn            = aws_iam_role.ecs_task_role.arn

  container_definitions = jsonencode([{
    name  = "${var.project_name}-container"
    image = "${aws_ecr_repository.app.repository_url}:latest"
    portMappings = [{
      containerPort = 8080
      hostPort      = 8080
    }]
    environment = [
      { name = "AWS_REGION", value = var.aws_region },
      { name = "S3_BUCKET_NAME", value = aws_s3_bucket.post_bucket.id },
      { name = "DYNAMODB_TABLE_NAME", value = aws_dynamodb_table.post_table.name },
      { name = "SQS_QUEUE_URL", value = aws_sqs_queue.post_queue.url }
    ]
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        awslogs-group         = aws_cloudwatch_log_group.ecs_logs.name
        awslogs-region        = var.aws_region
        awslogs-stream-prefix = "ecs"
      }
    }
  }])
}

# Application Load Balancer
resource "aws_lb" "main" {
  name               = "${var.project_name}-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = aws_subnet.main[*].id

  enable_deletion_protection = false
}

# ALB Listener
resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.main.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.app.arn
  }
}

# ALB Target Group
resource "aws_lb_target_group" "app" {
  name        = "${var.project_name}-tg"
  port        = 8080
  protocol    = "HTTP"
  vpc_id      = aws_vpc.main.id
  target_type = "ip"

  health_check {
    healthy_threshold   = "3"
    interval            = "30"
    protocol            = "HTTP"
    matcher             = "200"
    timeout             = "3"
    path                = "/"
    unhealthy_threshold = "2"
  }
}

# ECS Service
resource "aws_ecs_service" "main" {
  name            = "${var.project_name}-service"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.app.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = aws_subnet.main[*].id
    assign_public_ip = true
    security_groups  = [aws_security_group.ecs_tasks.id]
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.app.arn
    container_name   = "${var.project_name}-container"
    container_port   = 8080
  }
}

# Security Group for ECS Tasks
resource "aws_security_group" "ecs_tasks" {
  name        = "${var.project_name}-sg-ecs-tasks"
  description = "Allow inbound access from the ALB only"
  vpc_id      = aws_vpc.main.id

  ingress {
    protocol        = "tcp"
    from_port       = 8080
    to_port         = 8080
    security_groups = [aws_security_group.alb.id]
  }

  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# Security Group for ALB
resource "aws_security_group" "alb" {
  name        = "${var.project_name}-sg-alb"
  description = "Allow inbound traffic to ALB"
  vpc_id      = aws_vpc.main.id

  ingress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 80
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# IAM Roles
resource "aws_iam_role" "ecs_execution_role" {
  name = "${var.project_name}-ecs-execution-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ecs-tasks.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "ecs_execution_role_policy" {
  role       = aws_iam_role.ecs_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_iam_role" "ecs_task_role" {
  name = "${var.project_name}-ecs-task-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ecs-tasks.amazonaws.com"
      }
    }]
  })
}

# Attach policies to the ECS task role
resource "aws_iam_role_policy" "ecs_task_role_policy" {
  name = "${var.project_name}-ecs-task-role-policy"
  role = aws_iam_role.ecs_task_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:ListBucket",
        ]
        Resource = [
          aws_s3_bucket.post_bucket.arn,
          "${aws_s3_bucket.post_bucket.arn}/*",
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:UpdateItem",
          "dynamodb:DeleteItem",
          "dynamodb:Scan",
          "dynamodb:Query",
        ]
        Resource = aws_dynamodb_table.post_table.arn
      },
      {
        Effect = "Allow"
        Action = [
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes",
        ]
        Resource = aws_sqs_queue.post_queue.arn
      },
    ]
  })
}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "ecs_logs" {
  name              = "/ecs/${var.project_name}-logs"
  retention_in_days = 30
}

# Data source for availability zones
data "aws_availability_zones" "available" {}

# Outputs
output "ecr_repository_url" {
  description = "URL of the ECR repository"
  value       = aws_ecr_repository.app.repository_url
}

output "bucket_name" {
  description = "Name of the S3 bucket"
  value       = aws_s3_bucket.post_bucket.id
}

# Output the bucket name and URL
output "resume_bucket_name" {
  description = "Name of the public resume S3 bucket"
  value       = aws_s3_bucket.resume_bucket.id
}

output "resume_bucket_url" {
  description = "URL of the public resume S3 bucket"
  value       = "https://${aws_s3_bucket.resume_bucket.bucket_regional_domain_name}"
}

output "dynamodb_table_name" {
  description = "Name of the DynamoDB table"
  value       = aws_dynamodb_table.post_table.name
}

output "sqs_queue_url" {
  description = "URL of the SQS queue"
  value       = aws_sqs_queue.post_queue.url
}

output "ecs_cluster_name" {
  description = "Name of the ECS cluster"
  value       = aws_ecs_cluster.main.name
}

output "ecs_service_name" {
  description = "Name of the ECS service"
  value       = aws_ecs_service.main.name
}

output "alb_dns_name" {
  description = "DNS name of the Application Load Balancer"
  value       = aws_lb.main.dns_name
}