* [POST /api/v2/flows/jobs](https://developer.mypurecloud.com/api/rest/v2/architect/#post-api-v2-flows-jobs)
* [GET /api/v2/flows](https://developer.genesys.cloud/api/rest/v2/architect/#get-api-v2-flows)
* [GET /api/v2/flows/{flowId}](https://developer.genesys.cloud/api/rest/v2/architect/#get-api-v2-flows--flowId-)
* [GET /api/v2/flows/jobs/{jobId}](https://developer.mypurecloud.com/api/rest/v2/architect/#get-api-v2-flows-jobs--jobId-)
* [DELETE /api/v2/flows/{flowId}](https://developer.genesys.cloud/api/rest/v2/architect/#delete-api-v2-flows--flowId-)

**NOTE: Version 1.7.0 and lower had a defect that could cause improper variable substitution and an inadvertent deployment of a flow during a terraform plan. Please use version 1.8.0 or higher of the CX as Code provider.  With the newer versions of CX as Code you must set the file_content_hash attribute. See the example below on how to do this.**
* [POST /api/v2/flows/{flowId}/instances/settings/loglevels](https://developer.genesys.cloud/devapps/api-explorer#post--api-v2-flows--flowId--instances-settings-loglevels)
* [GET /api/v2/flows/{flowId}/instances/settings/loglevels](https://developer.genesys.cloud/devapps/api-explorer#get--api-v2-flows--flowId--instances-settings-loglevels)
* [GET /api/v2/flows/instances/settings/loglevels](https://developer.genesys.cloud/devapps/api-explorer#get--api-v2-flows-instances-settings-loglevels)
* [PUT /api/v2/flows/{flowId}/instances/settings/loglevels](https://developer.genesys.cloud/devapps/api-explorer#put--api-v2-flows--flowId--instances-settings-loglevels)
* [DELETE /api/v2/flows/{flowId}/instances/settings/loglevels](https://developer.genesys.cloud/devapps/api-explorer#delete--api-v2-flows--flowId--instances-settings-loglevels)
