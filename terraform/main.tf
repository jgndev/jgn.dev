terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# Provider configuration
provider "aws" {
  region = var.aws_region
}

# VPC
resource "aws_vpc" "main" {
  cidr_block = var.aws_vpc_cidr_block
  enable_dns_hostnames = true
  enable_dns_support = true

  tags = {
    Name = "${var.project_name}-vpc"
  }
}

# Primary Subnet
resource "aws_subnet" "primary" {
  vpc_id = aws_vpc.main.id
  cidr_block = var.aws_primary_subnet_cidr_block
  availability_zone = var.aws_primary_availability_zone
  map_public_ip_on_launch = true

  tags = {
    Name = "${var.project_name}-primary-subnet"
  }
}

# Secondary Subnet
resource "aws_subnet" "secondary" {
  vpc_id = aws_vpc.main.id
  cidr_block = var.aws_secondary_subnet_cidr_block
  availability_zone = var.aws_secondary_availability_zone
  map_public_ip_on_launch = true

  tags = {
    Name = "${var.project_name}-secondary-subnet"
  }
}

# Internet Gateway
resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "${var.project_name}-igw"
  }
}

# Route Tables
resource "aws_route_table" "main" {
  vpc_id = aws_vpc.main.id
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }
}

# Route Table Association
resource "aws_route_table_association" "primary" {
  subnet_id = aws_subnet.primary.id
  route_table_id = aws_route_table.main.id
}

resource "aws_route_table_association" "secondary" {
  subnet_id = aws_subnet.secondary.id
  route_table_id = aws_route_table.main.id
}

# Load Balancer
resource "aws_lb" "main" {
  name               = "${var.project_name}-lb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.lb.id]
  subnets            = [aws_subnet.primary.id, aws_subnet.secondary.id]
}

# Load Balancer Target Group
resource "aws_lb_target_group" "app" {
  name        = "${var.project_name}-tg"
  port        = 8080
  protocol    = "HTTP"
  vpc_id      = aws_vpc.main.id
  target_type = "ip"

  health_check {
    path = "/health"
    healthy_threshold = 2
    unhealthy_threshold = 10
    timeout = 30
    interval = 60
  }
}

# Security Groups
resource "aws_security_group" "lb" {
  name   = "${var.project_name}-lb"
  vpc_id = aws_vpc.main.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port = 443
    to_port = 443
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "ecs" {
  name   = var.ecs_security_group_name
  vpc_id = aws_vpc.main.id

  ingress {
    from_port       = 8080
    to_port         = 8080
    protocol        = "tcp"
    security_groups = [aws_security_group.lb.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# ECR Repository
resource "aws_ecr_repository" "app" {
  name = var.ecr_repo_name
}

# ECS Cluster
resource "aws_ecs_cluster" "main" {
  name = var.ecs_cluster_name
}

# ECS Task Definition
resource "aws_ecs_task_definition" "app" {
  family                   = var.ecs_task_definition_family
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = 256
  memory                   = 512
  execution_role_arn       = aws_iam_role.ecs_execution.arn

  container_definitions = jsonencode([{
    name         = "${var.project_name}-container"
    image        = "${aws_ecr_repository.app.repository_url}:latest"
    essential    = true
    portMappings = [{
      containerPort = 8080
      hostPort      = 8080
      protocol      = "tcp"
    }]
    # Each secret is injected as its own environment variable
    secrets = [
      {
        name      = "GITHUB_OWNER"
        valueFrom = aws_secretsmanager_secret.github_owner.arn
      },
      {
        name      = "GITHUB_REPO"
        valueFrom = aws_secretsmanager_secret.github_repo.arn
      },
      {
        name      = "GITHUB_TOKEN"
        valueFrom = aws_secretsmanager_secret.github_token.arn
      },
      {
        name      = "WEBHOOK_SECRET"
        valueFrom = aws_secretsmanager_secret.webhook_secret.arn
      }
    ]
  }])
}

# ECS Service
resource "aws_ecs_service" "app" {
  name            = var.ecs_service_name
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.app.arn
  launch_type     = "FARGATE"

  network_configuration {
    subnets         = [aws_subnet.primary.id, aws_subnet.secondary.id]
    security_groups = [aws_security_group.ecs.id]
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.app.arn
    container_name   = "${var.project_name}-container"
    container_port   = 8080
  }

  desired_count = 1

  depends_on = [
    aws_lb_listener.http,
    aws_lb_listener.https,
    aws_lb_target_group.app,
    aws_iam_role_policy_attachment.ecs_execution
  ]
}

# IAM
resource "aws_iam_role" "ecs_execution" {
  name = var.ecs_execution_name

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

resource "aws_iam_role_policy_attachment" "ecs_execution" {
  role       = aws_iam_role.ecs_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_iam_policy" "ecs_secrets_policy" {
  name        = "ecs-secrets-policy"
  description = "Policy for ECS tasks to read GitHub secrets"
  policy      = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Action = [
          "secretsmanager:GetSecretValue"
        ],
        Resource = [
          aws_secretsmanager_secret.github_owner.arn,
          aws_secretsmanager_secret.github_repo.arn,
          aws_secretsmanager_secret.github_token.arn,
          aws_secretsmanager_secret.webhook_secret.arn
        ]
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "ecs_execution_secrets" {
  role       = aws_iam_role.ecs_execution.name
  policy_arn = aws_iam_policy.ecs_secrets_policy.arn
}

# DNS (Route 53)
resource "aws_route53_zone" "main" {
  name = var.aws_route53_zone
}

resource "aws_route53_record" "apex" {
  zone_id = aws_route53_zone.main.id
  name    = var.aws_route53_domain
  type    = "A"

  alias {
    name                   = aws_lb.main.dns_name
    zone_id                = aws_lb.main.zone_id
    evaluate_target_health = true
  }
}

resource "aws_route53_record" "www" {
  zone_id = aws_route53_zone.main.id
  name    = "www.${var.aws_route53_domain}"
  type    = "A"

  alias {
    name                   = aws_lb.main.dns_name
    zone_id                = aws_lb.main.zone_id
    evaluate_target_health = true
  }
}

# ACM Certificate
resource "aws_acm_certificate" "main" {
  domain_name       = var.aws_route53_domain
  validation_method = "DNS"

  subject_alternative_names = ["www.${var.aws_route53_domain}"]

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_route53_record" "aws_route53_record_validation" {
  for_each = {
    for dvo in aws_acm_certificate.main.domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  }

  allow_overwrite = true
  zone_id = aws_route53_zone.main.id
  name    = each.value.name
  type    = each.value.type
  records = [each.value.record]
  ttl     = 60

  depends_on = [aws_route53_zone.main]
}

resource "aws_acm_certificate_validation" "acm_certificate_validation" {
  certificate_arn         = aws_acm_certificate.main.arn
  validation_record_fqdns = [for record in aws_route53_record.aws_route53_record_validation : record.fqdn]
  timeouts {
    create = "60m"
  }
}

# Load Balancer HTTP Listener
resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.main.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type = "redirect"
    redirect {
      port = "443"
      protocol = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}

# Update Load Balancer listener to HTTPS
resource "aws_lb_listener" "https" {
  load_balancer_arn = aws_lb.main.arn
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_acm_certificate.main.arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.app.arn
  }

  depends_on = [aws_acm_certificate_validation.acm_certificate_validation]
}

# Secrets and Environment Variables
resource "aws_secretsmanager_secret" "github_owner" {
  name = "stratocraft/github_owner"
  description = "Environment variables required by the app"

  tags = {
    Environment = terraform.workspace
    Application = "stratocraft-dev"
  }
}

resource "aws_secretsmanager_secret_version" "github_owner" {
  secret_id = aws_secretsmanager_secret.github_owner.id
  secret_string = var.github_owner
}

resource "aws_secretsmanager_secret" "github_repo" {
  name = "stratocraft/github_repo"
  description = "Environment variables required by the app"

  tags = {
    Environment = terraform.workspace
    Application = "stratocraft-dev"
  }
}

resource "aws_secretsmanager_secret_version" "github_repo" {
  secret_id = aws_secretsmanager_secret.github_repo.id
  secret_string = var.github_repo
}

resource "aws_secretsmanager_secret" "github_token" {
  name = "stratocraft/github_token"
  description = "Environment variables required by the app"

  tags = {
    Environment = terraform.workspace
    Application = "stratocraft-dev"
  }
}

resource "aws_secretsmanager_secret_version" "github_token" {
  secret_id = aws_secretsmanager_secret.github_token.id
  secret_string = var.github_token
}

resource "aws_secretsmanager_secret" "webhook_secret" {
  name = "stratocraft/webhook_secret"
  description = "Environment variables required by the app"

  tags = {
    Environment = terraform.workspace
    Application = "stratocraft-dev"
  }
}

resource "aws_secretsmanager_secret_version" "webhook_secret" {
  secret_id = aws_secretsmanager_secret.webhook_secret.id
  secret_string = var.webhook_secret
}

# outputs
output "nameservers" {
  description = "Nameservers for the Route 53 zone. Update these in your registrar's settings for the domain."
  value = aws_route53_zone.main.name_servers
}

output "ecr_repository_url" {
  description = "URL od the ECR repository"
  value = aws_ecr_repository.app.repository_url
}

output "app_url_http" {
  value = "http://${aws_lb.main.dns_name}"
  description = "Application URL using ALB DNS name"
}

output "app_url_https" {
  value = "https://${aws_lb.main.dns_name}"
  description = "Application URL using ALB DNS name"
}

output "custom_domain" {
  value = aws_route53_record.apex.name
  description = "Custom domain name"
}

output "load_balancer_dns" {
  value = aws_lb.main.dns_name
  description = "Load balancer DNS name"
}
