resource "genesyscloud_externalcontacts_organization" "example_org" {
  name           = "ExampleCorporation"
  company_type   = "Software"
  employee_count = 450
  phone_number {
    display      = "+1 321-700-1243"
    country_code = "US"
  }
  address {
    address1     = "51 Example Street"
    city         = "Irvine"
    state        = "California"
    postal_code  = "45678"
    country_code = "US"
  }
  twitter {
    twitter_id  = "ExampleTwitterId"
    name        = "ExampleTwitterName"
    screen_name = "ExampleScreenName"
  }
  tickers {
    symbol   = "EXPC"
    exchange = "NASDAQ"
  }
  external_system_url = "https://systemUrl.com"
}
