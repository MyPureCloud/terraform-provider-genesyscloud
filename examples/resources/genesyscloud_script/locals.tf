locals {
  working_dir = {
    script = "."
  }
  dependencies = {
    resource = [
      "../../common/random_uuid.tf"
    ]
  }
}
