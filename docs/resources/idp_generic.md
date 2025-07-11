---
page_title: "genesyscloud_idp_generic Resource - terraform-provider-genesyscloud"
subcategory: ""
description: |-
  Genesys Cloud Single Sign-on Generic Identity Provider. See this page for detailed configuration instructions: https://help.mypurecloud.com/articles/add-a-generic-single-sign-on-provider/
---
# genesyscloud_idp_generic (Resource)

Genesys Cloud Single Sign-on Generic Identity Provider. See this page for detailed configuration instructions: https://help.mypurecloud.com/articles/add-a-generic-single-sign-on-provider/

## API Usage
The following Genesys Cloud APIs are used by this resource. Ensure your OAuth Client has been granted the necessary scopes and permissions to perform these operations:

* [GET /api/v2/identityproviders/generic](https://developer.mypurecloud.com/api/rest/v2/identityprovider/#get-api-v2-identityproviders-generic)
* [PUT /api/v2/identityproviders/generic](https://developer.mypurecloud.com/api/rest/v2/identityprovider/#put-api-v2-identityproviders-generic)
* [DELETE /api/v2/identityproviders/generic](https://developer.mypurecloud.com/api/rest/v2/identityprovider/#delete-api-v2-identityproviders-generic)

## Example Usage

```terraform
resource "genesyscloud_idp_generic" "generic" {
  name                     = "Generic Provider"
  certificates             = [local.generic_certificate]
  issuer_uri               = "https://example.com"
  target_uri               = "https://example.com/login"
  relying_party_identifier = "unique-id-from-provider"
  logo_image_data          = filebase64("${local.working_dir.idp_generic}/logo.svg")
  endpoint_compression     = false
  name_identifier_format   = "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `certificates` (List of String) PEM or DER encoded public X.509 certificates for SAML signature validation.
- `issuer_uri` (String) Issuer URI provided by the provider.
- `name` (String) Name of the provider.

### Optional

- `disabled` (Boolean) True if Generic provider is disabled. Defaults to `false`.
- `endpoint_compression` (Boolean) True if the Genesys Cloud authentication request should be compressed. Defaults to `false`.
- `logo_image_data` (String) Base64 encoded SVG image.
- `name_identifier_format` (String) SAML name identifier format. (urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified | urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress | urn:oasis:names:tc:SAML:1.1:nameid-format:X509SubjectName | urn:oasis:names:tc:SAML:1.1:nameid-format:WindowsDomainQualifiedName | urn:oasis:names:tc:SAML:2.0:nameid-format:kerberos | urn:oasis:names:tc:SAML:2.0:nameid-format:entity | urn:oasis:names:tc:SAML:2.0:nameid-format:persistent | urn:oasis:names:tc:SAML:2.0:nameid-format:transient) Defaults to `urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified`.
- `relying_party_identifier` (String) String used to identify Genesys Cloud to the identity provider.
- `slo_binding` (String) Valid values: HTTP Redirect, HTTP Post
- `slo_uri` (String) Provided on app creation.
- `target_uri` (String) Target URI provided by the provider.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `read` (String)
- `update` (String)

