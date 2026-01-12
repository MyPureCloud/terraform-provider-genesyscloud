resource "genesyscloud_conversations_messaging_integrations_apple" "example" {
  name                     = "Apple Messages Integration"
  messages_for_business_id = "your-apple-messages-business-id"
  business_name            = "Your Business Name"
  logo_url                 = "https://example.com/logo.png"
  messaging_setting_id     = genesyscloud_conversations_messaging_settings.example_settings.id
  supported_content_id     = genesyscloud_conversations_messaging_supportedcontent.example_supported_content.id

  apple_i_message_app {
    application_name = "Your App Name"
    application_id   = "your-app-id"
    bundle_id        = "com.yourcompany.yourapp"
  }

  apple_authentication {
    oauth_client_id     = "your-oauth-client-id"
    oauth_client_secret = "your-oauth-client-secret"
    oauth_token_url     = "https://appleid.apple.com/auth/oauth2/token"
    oauth_scope         = "profile"
  }

  apple_pay {
    store_name                        = "Your Store"
    merchant_id                       = "merchant.com.yourcompany.yourstore"
    domain_name                       = "yourstore.com"
    payment_capabilities              = ["3DS"]
    supported_payment_networks        = ["visa", "masterCard", "amex"]
    payment_certificate_credential_id = "your-payment-certificate-credential-id"
    payment_gateway_url               = "https://yourstore.com/payment-gateway"
    fallback_url                      = "https://yourstore.com/fallback"
    shipping_method_update_url        = "https://yourstore.com/shipping-method-update"
    shipping_contact_update_url       = "https://yourstore.com/shipping-contact-update"
    payment_method_update_url         = "https://yourstore.com/payment-method-update"
    order_tracking_url                = "https://yourstore.com/order-tracking"
  }
}