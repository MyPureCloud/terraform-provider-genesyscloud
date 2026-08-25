<!-- sources
genesyscloud/access_policy/genesyscloud_access_policy_proxy.go
-->

## APIs Used

The following Genesys Cloud APIs are used by this resource:

- POST /api/v2/authorization/policies/targets/{targetName} - Create an access policy
- GET /api/v2/authorization/policies/{policyId} - Get an access policy by ID
- PUT /api/v2/authorization/policies/{policyId} - Update an access policy
- DELETE /api/v2/authorization/policies/targets/{targetName}/subject/{subjectId} - Delete an access policy
- GET /api/v2/authorization/policies - List all access policies
