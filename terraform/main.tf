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

# S3 post bucket for storing Markdown files
resource "aws_s3_bucket" "post_bucket" {
  bucket = var.bucket_name

  tags = {
    Name        = "Post Bucket"
    Environment = var.environment
  }
}

# S3 post bucket ACL
resource "aws_s3_bucket_ownership_controls" "post_bucket" {
  bucket = aws_s3_bucket.post_bucket.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

# S3 post bucket ACL
resource "aws_s3_bucket_acl" "post_bucket" {
  depends_on = [aws_s3_bucket_ownership_controls.post_bucket]
  bucket     = aws_s3_bucket.post_bucket.id
  acl        = "private"
}

# S3 resume bucket for storing publicly accessible resume
resource "aws_s3_bucket" "resume_bucket" {
  bucket = "${var.project_name}-public-resume"

  tags = {
    Name        = "Public Resume Bucket"
    Environment = var.environment
  }
}

# S3 resume bucket ownership controls
resource "aws_s3_bucket_ownership_controls" "resume_bucket_ownership" {
  bucket = aws_s3_bucket.resume_bucket.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

# S3 resume bucket public access block
resource "aws_s3_bucket_public_access_block" "resume_bucket_public_access" {
  bucket = aws_s3_bucket.resume_bucket.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

# S3 resume bucket ACL
resource "aws_s3_bucket_acl" "resume_bucket_acl" {
  depends_on = [
    aws_s3_bucket_ownership_controls.resume_bucket_ownership,
    aws_s3_bucket_public_access_block.resume_bucket_public_access,
  ]

  bucket = aws_s3_bucket.resume_bucket.id
  acl    = "public-read"
}

# S3 pubic img bucket policy
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

# S3 public img bucket for storing publicly accessible resume
resource "aws_s3_bucket" "public_img_bucket" {
  bucket = "${var.project_name}-public-image"

  tags = {
    Name        = "Public Image Bucket"
    Environment = var.environment
  }
}

# S3 public img bucket ownership controls
resource "aws_s3_bucket_ownership_controls" "public_img_bucket_ownership" {
  bucket = aws_s3_bucket.public_img_bucket.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

# S3 public img bucket public access block
resource "aws_s3_bucket_public_access_block" "pubic_img_bucket_public_access" {
  bucket = aws_s3_bucket.public_img_bucket.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

# S3 public img bucket ACL
resource "aws_s3_bucket_acl" "public_img_bucket_acl" {
  depends_on = [
    aws_s3_bucket_ownership_controls.public_img_bucket_ownership,
    aws_s3_bucket_public_access_block.pubic_img_bucket_public_access,
  ]

  bucket = aws_s3_bucket.public_img_bucket.id
  acl    = "public-read"
}

# S3 public img bucket policy
resource "aws_s3_bucket_policy" "public_img_bucket_policy" {
  bucket = aws_s3_bucket.public_img_bucket.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "PublicReadGetObject"
        Effect    = "Allow"
        Principal = "*"
        Action    = "s3:GetObject"
        Resource  = "${aws_s3_bucket.public_img_bucket.arn}/*"
      },
    ]
  })
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

# Route 53 Hosted Zone
resource "aws_route53_zone" "main" {
  name = "jgn.dev"
}

# Update Route 53 records to point to Elastic Beanstalk environment
resource "aws_route53_record" "apex" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "jgn.dev"
  type    = "A"

  alias {
    name                   = aws_elastic_beanstalk_environment.app_env.cname
    zone_id                = aws_elastic_beanstalk_environment.app_env.cname_prefix == "" ? aws_elastic_beanstalk_environment.app_env.load_balancers[0] : aws_elastic_beanstalk_environment.app_env.cname_prefix
    evaluate_target_health = true
  }
}

# Route 53 Record for www subdomain
resource "aws_route53_record" "www" {
  zone_id = aws_route53_zone.main.zone_id
  name    = "www.jgn.dev"
  type    = "A"

  alias {
    name                   = aws_elastic_beanstalk_environment.app_env.cname
    zone_id                = aws_elastic_beanstalk_environment.app_env.cname_prefix == "" ? aws_elastic_beanstalk_environment.app_env.load_balancers[0] : aws_elastic_beanstalk_environment.app_env.cname_prefix
    evaluate_target_health = true
  }
}

# ACM Certificate
resource "aws_acm_certificate" "main" {
  domain_name       = "jgn.dev"
  validation_method = "DNS"

  subject_alternative_names = ["www.jgn.dev"]

  lifecycle {
    create_before_destroy = true
  }
}

# Route 53 Record for ACM DNS Validation
resource "aws_route53_record" "acm_validation" {
  for_each = {
    for dvo in aws_acm_certificate.main.domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  }

  allow_overwrite = true
  name            = each.value.name
  records         = [each.value.record]
  ttl             = 60
  type            = each.value.type
  zone_id         = aws_route53_zone.main.zone_id
}

# ACM Certificate Validation
resource "aws_acm_certificate_validation" "main" {
  certificate_arn         = aws_acm_certificate.main.arn
  validation_record_fqdns = [for record in aws_route53_record.acm_validation : record.fqdn]
}

# Elastic Beanstalk Application
resource "aws_elastic_beanstalk_application" "app" {
  name        = var.project_name
  description = "Application for ${var.project_name}"
}

# Elastic Beanstalk Environment
resource "aws_elastic_beanstalk_environment" "app_env" {
  name                = "${var.project_name}-env"
  application         = aws_elastic_beanstalk_application.app.name
  solution_stack_name = "64bit Amazon Linux 2 v3.5.4 running Docker"

  setting {
    namespace = "aws:ec2:vpc"
    name      = "VPCId"
    value     = aws_vpc.main.id
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "Subnets"
    value     = join(",", aws_subnet.main[*].id)
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = aws_iam_instance_profile.eb_instance_profile.name
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "InstanceType"
    value     = "t3.micro"
  }

  setting {
    namespace = "aws:autoscaling:asg"
    name      = "MinSize"
    value     = "1"
  }

  setting {
    namespace = "aws:autoscaling:asg"
    name      = "MaxSize"
    value     = "2"
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "LoadBalancerType"
    value     = "application"
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "ServiceRole"
    value     = aws_iam_role.eb_service_role.name
  }

  setting {
    namespace = "aws:elbv2:listener:443"
    name      = "Protocol"
    value     = "HTTPS"
  }

  setting {
    namespace = "aws:elbv2:listener:443"
    name      = "SSLCertificateArns"
    value     = aws_acm_certificate.main.arn
  }

  setting {
    namespace = "aws:elasticbeanstalk:application:environment"
    name      = "AWS_REGION"
    value     = var.aws_region
  }

  setting {
    namespace = "aws:elasticbeanstalk:application:environment"
    name      = "S3_BUCKET_NAME"
    value     = aws_s3_bucket.post_bucket.id
  }

  setting {
    namespace = "aws:elasticbeanstalk:application:environment"
    name      = "DYNAMODB_TABLE_NAME"
    value     = aws_dynamodb_table.post_table.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:application:environment"
    name      = "SQS_QUEUE_URL"
    value     = aws_sqs_queue.post_queue.url
  }
}

# IAM Instance Profile for Elastic Beanstalk
resource "aws_iam_instance_profile" "eb_instance_profile" {
  name = "${var.project_name}-eb-instance-profile"
  role = aws_iam_role.eb_instance_role.name
}

resource "aws_iam_role" "eb_instance_role" {
  name = "${var.project_name}-eb-instance-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "eb_web_tier" {
  policy_arn = "arn:aws:iam::aws:policy/AWSElasticBeanstalkWebTier"
  role       = aws_iam_role.eb_instance_role.name
}

resource "aws_iam_role_policy_attachment" "eb_worker_tier" {
  policy_arn = "arn:aws:iam::aws:policy/AWSElasticBeanstalkWorkerTier"
  role       = aws_iam_role.eb_instance_role.name
}

resource "aws_iam_role_policy_attachment" "eb_docker" {
  policy_arn = "arn:aws:iam::aws:policy/AWSElasticBeanstalkMulticontainerDocker"
  role       = aws_iam_role.eb_instance_role.name
}

# Attach policies for S3, DynamoDB, and SQS access
resource "aws_iam_role_policy" "eb_instance_policy" {
  name = "${var.project_name}-eb-instance-policy"
  role = aws_iam_role.eb_instance_role.id

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

# Elastic Beanstalk Service Role
resource "aws_iam_role" "eb_service_role" {
  name = "${var.project_name}-eb-service-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "elasticbeanstalk.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "eb_enhanced_health" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSElasticBeanstalkEnhancedHealth"
  role       = aws_iam_role.eb_service_role.name
}

resource "aws_iam_role_policy_attachment" "eb_service" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSElasticBeanstalkService"
  role       = aws_iam_role.eb_service_role.name
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

# Output the nameservers
output "nameservers" {
  value       = aws_route53_zone.main.name_servers
  description = "Nameservers for the Route 53 zone. Update these in Porkbun."
}

# Update outputs
output "elastic_beanstalk_env_name" {
  description = "Name of the Elastic Beanstalk environment"
  value       = aws_elastic_beanstalk_environment.app_env.name
}

output "elastic_beanstalk_env_cname" {
  description = "CNAME of the Elastic Beanstalk environment"
  value       = aws_elastic_beanstalk_environment.app_env.cname
}

output "elastic_beanstalk_env_endpoint" {
  description = "Endpoint URL of the Elastic Beanstalk environment"
  value       = aws_elastic_beanstalk_environment.app_env.endpoint_url
}

output "website_url" {
  description = "URL of the website"
  value       = "https://jgn.dev"
}