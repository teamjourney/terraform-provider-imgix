terraform {
  required_providers {
    imgix = {
      source = "journey.travel/terraform-providers/imgix"
      version = "1.0.0"
    }
    aws = {
      source = "hashicorp/aws"
      version = "~> 3.0"
    }
  }
}

provider "imgix" {}
provider "aws" {
  region = "eu-west-1"
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

resource "aws_iam_access_key" "this1" {
  user = aws_iam_user.this.id
}

//resource "imgix_source" "this1" {
//  name = "test11"
//
//  deployment {
//    annotation = "test11 annotation"
//    type = "s3"
//    imgix_subdomains = ["pj-test111", "pj-test112"]
//
//    s3_access_key = aws_iam_access_key.this1.id
//    s3_secret_key = aws_iam_access_key.this1.secret
//    s3_bucket = aws_s3_bucket.this.bucket
//  }
//}

resource "imgix_source" "this" {
  name = "test2"

  deployment {
    annotation = "test22 annotation."
    type = "s3"
    imgix_subdomains = ["pj-test2"]

    s3_access_key = aws_iam_access_key.this1.id
    s3_secret_key = aws_iam_access_key.this1.secret
    s3_bucket = aws_s3_bucket.this.bucket
  }
}
