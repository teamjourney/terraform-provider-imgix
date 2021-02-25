terraform {
  required_providers {
    imgix = {
      source  = "teamjourney/imgix"
      version = "0.0.1-pre1"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.0"
    }
  }
}

provider "imgix" {}
provider "aws" {
  region = "eu-central-1"
}

resource "aws_s3_bucket" "this" {
  bucket_prefix = "test1-"

  acl = "private"
}

resource "aws_iam_user" "this" {
  name = aws_s3_bucket.this.bucket
}

resource "aws_iam_user_policy" "this" {
  user = aws_iam_user.this.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:ListBucket",
          "s3:GetBucketLocation",
          "s3:GetObject",
        ]
        Resource = [
          aws_s3_bucket.this.arn,
          "${aws_s3_bucket.this.arn}/*",
        ]
      }
    ]
  })
}

resource "aws_iam_access_key" "this" {
  user = aws_iam_user.this.id
}

resource "imgix_source" "this" {
  name = "test11"

  deployment {
    annotation       = "test1 annotation"
    type             = "s3"
    imgix_subdomains = ["test11"]

    s3_access_key = aws_iam_access_key.this.id
    s3_secret_key = "aa"
    s3_bucket     = aws_s3_bucket.this.bucket
    s3_prefix     = "test1"
  }
}