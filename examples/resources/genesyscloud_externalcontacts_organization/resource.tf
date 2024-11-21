resource "genesyscloud_externalcontacts_organization" "organization" {
  first_name  = "jean"
  middle_name = "jacques"
  last_name   = "dupont"
  salutation  = "salutation"
  title       = "genesys staff"
  phone_number {
    display      = "+33 0 00 00 00 00"
    extension    = 4
    accepts_sms  = false
    e164         = "+330000000000"
    country_code = "FR"
  }
  address {
    address1     = "1 rue de la paix"
    address2     = "2 rue de la paix"
    city         = "Paris"
    state        = "île-de-France"
    postal_code  = "75000"
    country_code = "FR"
  }
  twitter {
    twitter_id  = "RealABCNews"
    name        = "ABCNews"
    screen_name = "ABCNewsCorp"
  }
  tickers {
    symbol   = "ABC"
    exchange = "NYSE"
  }
  external_system_url = "https://systemUrl.com"
}
