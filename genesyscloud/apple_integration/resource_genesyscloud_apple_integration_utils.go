package apple_integration

import (
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v171/platformclientv2"
)

/*
The resource_genesyscloud_apple_integration_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// BuildSdkDomainEntityRef builds a Genesys Cloud *platformclientv2.Domainentityref from a resource data field
func BuildSdkDomainEntityRef(d *schema.ResourceData, idAttr string) *platformclientv2.Domainentityref {
	idVal := d.Get(idAttr).(string)
	if idVal == "" {
		return nil
	}
	return &platformclientv2.Domainentityref{Id: &idVal}
}

// getAppleIntegrationFromResourceData maps data from schema ResourceData object to a platformclientv2.Appleintegration
func getAppleIntegrationFromResourceData(d *schema.ResourceData) platformclientv2.Appleintegration {
	return platformclientv2.Appleintegration{
		Name:                  platformclientv2.String(d.Get("name").(string)),
		SupportedContent:      buildSupportedContentReference(d.Get("supported_content").([]interface{})),
		MessagingSetting:      buildMessagingSettingReference(d.Get("messaging_setting").([]interface{})),
		MessagesForBusinessId: platformclientv2.String(d.Get("messages_for_business_id").(string)),
		BusinessName:          platformclientv2.String(d.Get("business_name").(string)),
		LogoUrl:               platformclientv2.String(d.Get("logo_url").(string)),
		Status:                platformclientv2.String(d.Get("status").(string)),
		CreateStatus:          platformclientv2.String(d.Get("create_status").(string)),
		CreateError:           buildErrorBody(d.Get("create_error").([]interface{})),
		AppleIMessageApp:      buildAppleIMessageApp(d.Get("apple_i_message_app").([]interface{})),
		AppleAuthentication:   buildAppleAuthentication(d.Get("apple_authentication").([]interface{})),
		ApplePay:              buildApplePay(d.Get("apple_pay").([]interface{})),
		IdentityResolution:    buildAppleIdentityResolutionConfig(d.Get("identity_resolution").([]interface{})),
	}
}

// buildSupportedContentReference maps an []interface{} into a Genesys Cloud *platformclientv2.Supportedcontentreference
func buildSupportedContentReference(supportedContent []interface{}) *platformclientv2.Supportedcontentreference {
	if len(supportedContent) == 0 {
		return nil
	}

	supportedContentMap, ok := supportedContent[0].(map[string]interface{})
	if !ok {
		return nil
	}

	var sdkSupportedContent platformclientv2.Supportedcontentreference
	resourcedata.BuildSDKStringValueIfNotNil(&sdkSupportedContent.Id, supportedContentMap, "id")

	return &sdkSupportedContent
}

// buildMessagingSettingReference maps an []interface{} into a Genesys Cloud *platformclientv2.Messagingsettingreference
func buildMessagingSettingReference(messagingSetting []interface{}) *platformclientv2.Messagingsettingreference {
	if len(messagingSetting) == 0 {
		return nil
	}

	messagingSettingMap, ok := messagingSetting[0].(map[string]interface{})
	if !ok {
		return nil
	}

	var sdkMessagingSetting platformclientv2.Messagingsettingreference
	resourcedata.BuildSDKStringValueIfNotNil(&sdkMessagingSetting.Id, messagingSettingMap, "id")

	return &sdkMessagingSetting
}

// buildErrorBody maps an []interface{} into a Genesys Cloud *platformclientv2.Errorbody
func buildErrorBody(errorBody []interface{}) *platformclientv2.Errorbody {
	if len(errorBody) == 0 {
		return nil
	}
	// ErrorBody schema is empty - this is read-only data from API
	return &platformclientv2.Errorbody{}
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

// buildAppleIdentityResolutionConfig maps an []interface{} into a Genesys Cloud *platformclientv2.Appleidentityresolutionconfig
func buildAppleIdentityResolutionConfig(identityResolution []interface{}) *platformclientv2.Appleidentityresolutionconfig {
	if len(identityResolution) == 0 {
		return nil
	}

	identityResolutionMap, ok := identityResolution[0].(map[string]interface{})
	if !ok {
		return nil
	}

	var sdkIdentityResolution platformclientv2.Appleidentityresolutionconfig
	sdkIdentityResolution.ResolveIdentities = platformclientv2.Bool(identityResolutionMap["resolve_identities"].(bool))

	return &sdkIdentityResolution
}

// flattenSupportedContentReference maps a Genesys Cloud *platformclientv2.Supportedcontentreference into a []interface{}
func flattenSupportedContentReference(supportedContent *platformclientv2.Supportedcontentreference) []interface{} {
	if supportedContent == nil {
		return nil
	}

	supportedContentMap := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(supportedContentMap, "id", supportedContent.Id)
	resourcedata.SetMapValueIfNotNil(supportedContentMap, "self_uri", supportedContent.SelfUri)

	return []interface{}{supportedContentMap}
}

// flattenMessagingSettingReference maps a Genesys Cloud *platformclientv2.Messagingsettingreference into a []interface{}
func flattenMessagingSettingReference(messagingSetting *platformclientv2.Messagingsettingreference) []interface{} {
	if messagingSetting == nil {
		return nil
	}

	messagingSettingMap := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(messagingSettingMap, "id", messagingSetting.Id)
	resourcedata.SetMapValueIfNotNil(messagingSettingMap, "self_uri", messagingSetting.SelfUri)

	return []interface{}{messagingSettingMap}
}

// flattenErrorBody maps a Genesys Cloud *platformclientv2.Errorbody into a []interface{}
func flattenErrorBody(errorBody *platformclientv2.Errorbody) []interface{} {
	if errorBody == nil {
		return nil
	}

	errorBodyMap := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(errorBodyMap, "message", errorBody.Message)
	resourcedata.SetMapValueIfNotNil(errorBodyMap, "code", errorBody.Code)
	resourcedata.SetMapValueIfNotNil(errorBodyMap, "status", errorBody.Status)

	return []interface{}{errorBodyMap}
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

// flattenAppleIdentityResolutionConfig maps a Genesys Cloud *platformclientv2.Appleidentityresolutionconfig into a []interface{}
func flattenAppleIdentityResolutionConfig(identityResolution *platformclientv2.Appleidentityresolutionconfig) []interface{} {
	if identityResolution == nil {
		return nil
	}

	identityResolutionMap := make(map[string]interface{})
	resourcedata.SetMapValueIfNotNil(identityResolutionMap, "resolve_identities", identityResolution.ResolveIdentities)

	return []interface{}{identityResolutionMap}
}