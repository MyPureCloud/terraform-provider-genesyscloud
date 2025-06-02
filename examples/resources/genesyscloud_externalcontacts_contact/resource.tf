resource "genesyscloud_externalcontacts_contact" "contact" {
  first_name  = "jean"
  middle_name = "jacques"
  last_name   = "dupont"
  salutation  = "salutation"
  title       = "genesys staff"
  work_phone {
    display      = "+33 03 20 45 67 89"
    extension    = 4
    accepts_sms  = false
    e164         = "+33320456789"
    country_code = "FR"
  }
  cell_phone {
    display      = "+33 09 20 45 67 89"
    accepts_sms  = true
    e164         = "+33920456789"
    country_code = "FR"
  }
  home_phone {
    display      = "+33 02 12 32 30 30"
    accepts_sms  = false
    e164         = "+33212323030"
    country_code = "FR"
  }
  other_phone {
    display      = "+33 02 12 32 30 30"
    extension    = 4
    accepts_sms  = false
    e164         = "+33212323030"
    country_code = "FR"
  }
  work_email     = "workEmail@example.com"
  personal_email = "personalEmail@example.com"
  other_email    = "otherEmail@example.com"
  address {
    address1     = "1 rue de la paix"
    address2     = "2 rue de la paix"
    city         = "Paris"
    state        = "île-de-France"
    postal_code  = "75000"
    country_code = "FR"
  }
  twitter_id {
    id          = "1725137533"
    name        = "@KMbappe"
    screen_name = "KMbappe"
  }
  line_id {
    ids {
      user_id = "1234"
    }
    display_name = "lineName"
  }
  whatsapp_id {
    phone_number {
      display      = "+33 01 72 80 92 96"
      extension    = 4
      accepts_sms  = false
      e164         = "+33172809296"
      country_code = "FR"
    }
    display_name = "whatsappName"
  }
  facebook_id {
    ids {
      scoped_id = "myscopeIdUrl"
    }
    display_name = "facebookName"
  }
  survey_opt_out           = false
  external_system_url      = "https://systemUrl.com"
  external_organization_id = genesyscloud_externalcontacts_organization.example_org.id
}
