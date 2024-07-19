- [POST /api/v2/users](https://developer.mypurecloud.com/api/rest/v2/users/#post-api-v2-users)
- [GET /api/v2/users/{userId}](https://developer.mypurecloud.com/api/rest/v2/users/#get-api-v2-users--userId-)
- [PATCH /api/v2/users/{userId}](https://developer.mypurecloud.com/api/rest/v2/users/#patch-api-v2-users--userId-)
- [DELETE /api/v2/users/{userId}](https://developer.mypurecloud.com/api/rest/v2/users/#delete-api-v2-users--userId-)
- [PUT /api/v2/users/{userId}/routingskills/bulk](https://developer.mypurecloud.com/api/rest/v2/users/#put-api-v2-users--userId--routingskills-bulk)
- [DELETE /api/v2/users/{userId}/routinglanguages/{languageId}](https://developer.mypurecloud.com/api/rest/v2/users/#delete-api-v2-users--userId--routinglanguages--languageId-)
- [PATCH /api/v2/users/{userId}/routinglanguages/bulk](https://developer.mypurecloud.com/api/rest/v2/users/#patch-api-v2-users--userId--routinglanguages-bulk)
- [GET /api/v2/users/{userId}/routinglanguages](https://developer.mypurecloud.com/api/rest/v2/users/#get-api-v2-users--userId--routinglanguages)
- [PUT /api/v2/users/{userId}/profileskills](https://developer.mypurecloud.com/api/rest/v2/users/#put-api-v2-users--userId--profileskills)
- [GET /api/v2/routing/users/{userId}/utilization](https://developer.mypurecloud.com/api/rest/v2/users/#get-api-v2-routing-users--userId--utilization)
- [PUT /api/v2/routing/users/{userId}/utilization](https://developer.mypurecloud.com/api/rest/v2/users/#put-api-v2-routing-users--userId--utilization)
- [DELETE /api/v2/routing/users/{userId}/utilization](https://developer.mypurecloud.com/api/rest/v2/users/#delete-api-v2-routing-users--userId--utilization)

---

This resource has a feature toggle experimenting with new logic for returning phone addresses in the read action.
You can enable this by setting the `USE_NEW_USER_ADDRESS_LOGIC` environmental variable before running `terraform apply`.
