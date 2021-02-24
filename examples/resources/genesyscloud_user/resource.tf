resource "genesyscloud_user" "test_user" {
  email       = "test@example.com"
  name        = "Test User"
  password    = "initial-password"
  division_id = "505e1036-6f04-405c-a630-de94a8ad2eb8"
  state       = "active"
  department  = "Development"
  title       = "Senior Director"
  manager     = "165d19ca-7224-4a9b-ad08-2f4c9fe6356c"
  other_emails {
    address = "test@gmail.com"
    type    = "HOME"
  }
  phone_numbers {
    number     = "3174181234"
    media_type = "PHONE"
    type       = "MOBILE"
  }
}
