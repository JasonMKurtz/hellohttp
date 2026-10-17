resource "aws_ecr_repository" "hellohttp" {
  name = "hellohttp"

  # Let `terraform destroy` remove the repo even if it still holds images.
  force_delete = true
}

# Keep the registry from growing forever.
resource "aws_ecr_lifecycle_policy" "hellohttp" {
  repository = aws_ecr_repository.hellohttp.name

  policy = jsonencode({
    rules = [{
      rulePriority = 1
      description  = "Keep the 20 most recent images"
      selection = {
        tagStatus   = "any"
        countType   = "imageCountMoreThan"
        countNumber = 20
      }
      action = { type = "expire" }
    }]
  })
}
