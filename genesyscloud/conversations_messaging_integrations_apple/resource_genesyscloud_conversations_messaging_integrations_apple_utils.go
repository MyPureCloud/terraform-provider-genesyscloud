package conversations_messaging_integrations_apple

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v171/platformclientv2"
)

/*
The resource_genesyscloud_apple_integration_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getAppleIntegrationFromResourceData maps data from schema ResourceData object to a platformclientv2.Appleintegrationrequest for create
func getConversationsMessagingIntegrationsAppleFromResourceData(d *schema.ResourceData) platformclientv2.Appleintegrationrequest {
	supportedContentId := d.Get("supported_content_id").(string)
	messagingSettingId := d.Get("messaging_setting_id").(string)

	request := platformclientv2.Appleintegrationrequest{
		Name:                  platformclientv2.String(d.Get("name").(string)),
		MessagesForBusinessId: platformclientv2.String(d.Get("messages_for_business_id").(string)),
		BusinessName:          platformclientv2.String(d.Get("business_name").(string)),
		LogoUrl:               platformclientv2.String(d.Get("logo_url").(string)),
		AppleIMessageApp:      buildAppleIMessageApp(d.Get("apple_i_message_app").([]interface{})),
		AppleAuthentication:   buildAppleAuthentication(d.Get("apple_authentication").([]interface{})),
		ApplePay:              buildApplePay(d.Get("apple_pay").([]interface{})),
	}

	if supportedContentId != "" {
		request.SupportedContent = &platformclientv2.Supportedcontentreference{Id: &supportedContentId}
	}
	if messagingSettingId != "" {
		request.MessagingSetting = &platformclientv2.Messagingsettingrequestreference{Id: &messagingSettingId}
	}

	return request
}

// getAppleIntegrationFromResourceDataForUpdate maps data from schema ResourceData object to a platformclientv2.Appleintegrationupdaterequest for update
func getConversationsMessagingIntegrationsAppleFromResourceDataForUpdate(d *schema.ResourceData) platformclientv2.Appleintegrationupdaterequest {
	supportedContentId := d.Get("supported_content_id").(string)
	messagingSettingId := d.Get("messaging_setting_id").(string)

	request := platformclientv2.Appleintegrationupdaterequest{
		Name:                platformclientv2.String(d.Get("name").(string)),
		BusinessName:        platformclientv2.String(d.Get("business_name").(string)),
		LogoUrl:             platformclientv2.String(d.Get("logo_url").(string)),
		AppleIMessageApp:    buildAppleIMessageApp(d.Get("apple_i_message_app").([]interface{})),
		AppleAuthentication: buildAppleAuthentication(d.Get("apple_authentication").([]interface{})),
		ApplePay:            buildApplePay(d.Get("apple_pay").([]interface{})),
	}

	if supportedContentId != "" {
		request.SupportedContent = &platformclientv2.Supportedcontentreference{Id: &supportedContentId}
	}
	if messagingSettingId != "" {
		request.MessagingSetting = &platformclientv2.Messagingsettingrequestreference{Id: &messagingSettingId}
	}

	return request
}

// buildAppleIMessageApp maps an []interface{} into a Genesys Cloud *platformclientv2.Appleimessageapp
func buildAppleIMessageApp(appleIMessageApp []interface{}) *platformclientv2.Appleimessageapp {
	if len(appleIMessageApp) == 0 {
		return nil
	}

	appleIMessageAppMap, ok := appleIMessageApp[0].(map[string]interface{})
	if !ok {
		return nil
	}

	var sdkAppleIMessageApp platformclientv2.Appleimessageapp
	resourcedata.BuildSDKStringValueIfNotNil(&sdkAppleIMessageApp.ApplicationName, appleIMessageAppMap, "application_name")
	resourcedata.BuildSDKStringValueIfNotNil(&sdkAppleIMessageApp.ApplicationId, appleIMessageAppMap, "application_id")
	resourcedata.BuildSDKStringValueIfNotNil(&sdkAppleIMessageApp.BundleId, appleIMessageAppMap, "bundle_id")

	return &sdkAppleIMessageApp
}

// buildAppleAuthentication maps an []interface{} into a Genesys Cloud *platformclientv2.Appleauthentication
func buildAppleAuthentication(appleAuthentication []interface{}) *platformclientv2.Appleauthentication {
	if len(appleAuthentication) == 0 {
		return nil
	}

	appleAuthenticationMap, ok := appleAuthentication[0].(map[string]interface{})
	if !ok {
		return nil
	}

	var sdkAppleAuthentication platformclientv2.Appleauthentication
	resourcedata.BuildSDKStringValueIfNotNil(&sdkAppleAuthentication.OauthClientId, appleAuthenticationMap, "oauth_client_id")
	resourcedata.BuildSDKStringValueIfNotNil(&sdkAppleAuthentication.OauthClientSecret, appleAuthenticationMap, "oauth_client_secret")
	resourcedata.BuildSDKStringValueIfNotNil(&sdkAppleAuthentication.OauthTokenUrl, appleAuthenticationMap, "oauth_token_url")
	resourcedata.BuildSDKStringValueIfNotNil(&sdkAppleAuthentication.OauthScope, appleAuthenticationMap, "oauth_scope")

	return &sdkAppleAuthentication
}

// buildApplePay maps an []interface{} into a Genesys Cloud *platformclientv2.Applepay
func buildApplePay(applePay []interface{}) *platformclientv2.Applepay {
	if len(applePay) == 0 {
		return nil
	}

	applePayMap, ok := applePay[0].(map[string]interface{})
	if !ok {
		return nil
	}

	var sdkApplePay platformclientv2.Applepay
	resourcedata.BuildSDKStringValueIfNotNil(&sdkApplePay.StoreName, applePayMap, "store_name")
	resourcedata.BuildSDKStringValueIfNotNil(&sdkApplePay.MerchantId, applePayMap, "merchant_id")
	resourcedata.BuildSDKStringValueIfNotNil(&sdkApplePay.DomainName, applePayMap, "domain_name")
	resourcedata.BuildSDKStringArrayValueIfNotNil(&sdkApplePay.PaymentCapabilities, applePayMap, "payment_capabilities")
	resourcedata.BuildSDKStringArrayValueIfNotNil(&sdkApplePay.SupportedPaymentNetworks, applePayMap, "supported_payment_networks")
	resourcedata.BuildSDKStringValueIfNotNil(&sdkApplePay.PaymentCertificateCredentialId, applePayMap, "payment_certificate_credential_id")
	resourcedata.BuildSDKStringValueIfNotNil(&sdkApplePay.PaymentGatewayUrl, applePayMap, "payment_gateway_url")
	resourcedata.BuildSDKStringValueIfNotNil(&sdkApplePay.FallbackUrl, applePayMap, "fallback_url")
	resourcedata.BuildSDKStringValueIfNotNil(&sdkApplePay.ShippingMethodUpdateUrl, applePayMap, "shipping_method_update_url")
	resourcedata.BuildSDKStringValueIfNotNil(&sdkApplePay.ShippingContactUpdateUrl, applePayMap, "shipping_contact_update_url")
	resourcedata.BuildSDKStringValueIfNotNil(&sdkApplePay.PaymentMethodUpdateUrl, applePayMap, "payment_method_update_url")
	resourcedata.BuildSDKStringValueIfNotNil(&sdkApplePay.OrderTrackingUrl, applePayMap, "order_tracking_url")

	return &sdkApplePay
}

// flattenAppleIMessageApp maps a Genesys Cloud *platformclientv2.Appleimessageapp into a []interface{}
func flattenAppleIMessageApp(appleIMessageApp *platformclientv2.Appleimessageapp) []interface{} {
	if appleIMessageApp == nil {
		return nil
	}

	appleIMessageAppMap := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(appleIMessageAppMap, "application_name", appleIMessageApp.ApplicationName)
	resourcedata.SetMapValueIfNotNil(appleIMessageAppMap, "application_id", appleIMessageApp.ApplicationId)
	resourcedata.SetMapValueIfNotNil(appleIMessageAppMap, "bundle_id", appleIMessageApp.BundleId)

	return []interface{}{appleIMessageAppMap}
}

// flattenAppleAuthentication maps a Genesys Cloud *platformclientv2.Appleauthentication into a []interface{}
func flattenAppleAuthentication(appleAuthentication *platformclientv2.Appleauthentication) []interface{} {
	if appleAuthentication == nil {
		return nil
	}

	appleAuthenticationMap := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(appleAuthenticationMap, "oauth_client_id", appleAuthentication.OauthClientId)
	resourcedata.SetMapValueIfNotNil(appleAuthenticationMap, "oauth_client_secret", appleAuthentication.OauthClientSecret)
	resourcedata.SetMapValueIfNotNil(appleAuthenticationMap, "oauth_token_url", appleAuthentication.OauthTokenUrl)
	resourcedata.SetMapValueIfNotNil(appleAuthenticationMap, "oauth_scope", appleAuthentication.OauthScope)

	return []interface{}{appleAuthenticationMap}
}

// flattenApplePay maps a Genesys Cloud *platformclientv2.Applepay into a []interface{}
func flattenApplePay(applePay *platformclientv2.Applepay) []interface{} {
	if applePay == nil {
		return nil
	}

	applePayMap := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(applePayMap, "store_name", applePay.StoreName)
	resourcedata.SetMapValueIfNotNil(applePayMap, "merchant_id", applePay.MerchantId)
	resourcedata.SetMapValueIfNotNil(applePayMap, "domain_name", applePay.DomainName)
	resourcedata.SetMapStringArrayValueIfNotNil(applePayMap, "payment_capabilities", applePay.PaymentCapabilities)
	resourcedata.SetMapStringArrayValueIfNotNil(applePayMap, "supported_payment_networks", applePay.SupportedPaymentNetworks)
	resourcedata.SetMapValueIfNotNil(applePayMap, "payment_certificate_credential_id", applePay.PaymentCertificateCredentialId)
	resourcedata.SetMapValueIfNotNil(applePayMap, "payment_gateway_url", applePay.PaymentGatewayUrl)
	resourcedata.SetMapValueIfNotNil(applePayMap, "fallback_url", applePay.FallbackUrl)
	resourcedata.SetMapValueIfNotNil(applePayMap, "shipping_method_update_url", applePay.ShippingMethodUpdateUrl)
	resourcedata.SetMapValueIfNotNil(applePayMap, "shipping_contact_update_url", applePay.ShippingContactUpdateUrl)
	resourcedata.SetMapValueIfNotNil(applePayMap, "payment_method_update_url", applePay.PaymentMethodUpdateUrl)
	resourcedata.SetMapValueIfNotNil(applePayMap, "order_tracking_url", applePay.OrderTrackingUrl)

	return []interface{}{applePayMap}
}
