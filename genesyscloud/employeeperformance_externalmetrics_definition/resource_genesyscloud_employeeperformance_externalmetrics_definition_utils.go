package employeeperformance_externalmetrics_definition

import (
	gcloud "terraform-provider-genesyscloud/genesyscloud"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v119/platformclientv2"
)

/*
The resource_genesyscloud_employeeperformance_externalmetrics_definition_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.
*/

// getEmployeeperformanceExternalmetricsDefinitionFromResourceData maps data from schema ResourceData object to a platformclientv2.Domainorganizationrole
func getEmployeeperformanceExternalmetricsDefinitionFromResourceData(d *schema.ResourceData) platformclientv2.Domainorganizationrole {
	return platformclientv2.Domainorganizationrole{
		Name:          platformclientv2.String(d.Get("name").(string)),
		Description:   platformclientv2.String(d.Get("description").(string)),
		DefaultRoleId: platformclientv2.String(d.Get("default_role_id").(string)),
		// TODO: Handle permissions property
		// TODO: Handle unused_permissions property
		PermissionPolicies: buildDomainPermissionPolicys(d.Get("permission_policies").([]interface{})),
		UserCount:          platformclientv2.Int(d.Get("user_count").(int)),
		RoleNeedsUpdate:    platformclientv2.Bool(d.Get("role_needs_update").(bool)),
		Base:               platformclientv2.Bool(d.Get("base").(bool)),
		Default:            platformclientv2.Bool(d.Get("default").(bool)),
	}
}

// buildChats maps an []interface{} into a Genesys Cloud *[]platformclientv2.Chat
func buildChats(chats []interface{}) *[]platformclientv2.Chat {
	chatsSlice := make([]platformclientv2.Chat, 0)
	for _, chat := range chats {
		var sdkChat platformclientv2.Chat
		chatsMap, ok := chat.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkChat.JabberId, chatsMap, "jabber_id")

		chatsSlice = append(chatsSlice, sdkChat)
	}

	return &chatsSlice
}

// buildContacts maps an []interface{} into a Genesys Cloud *[]platformclientv2.Contact
func buildContacts(contacts []interface{}) *[]platformclientv2.Contact {
	contactsSlice := make([]platformclientv2.Contact, 0)
	for _, contact := range contacts {
		var sdkContact platformclientv2.Contact
		contactsMap, ok := contact.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkContact.Address, contactsMap, "address")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkContact.Display, contactsMap, "display")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkContact.MediaType, contactsMap, "media_type")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkContact.Type, contactsMap, "type")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkContact.Extension, contactsMap, "extension")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkContact.CountryCode, contactsMap, "country_code")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkContact.Integration, contactsMap, "integration")

		contactsSlice = append(contactsSlice, sdkContact)
	}

	return &contactsSlice
}

// buildUsers maps an []interface{} into a Genesys Cloud *[]platformclientv2.User
func buildUsers(users []interface{}) *[]platformclientv2.User {
	usersSlice := make([]platformclientv2.User, 0)
	for _, user := range users {
		var sdkUser platformclientv2.User
		usersMap, ok := user.(map[string]interface{})
		if !ok {
			continue
		}

		usersSlice = append(usersSlice, sdkUser)
	}

	return &usersSlice
}

// buildUserImages maps an []interface{} into a Genesys Cloud *[]platformclientv2.Userimage
func buildUserImages(userImages []interface{}) *[]platformclientv2.Userimage {
	userImagesSlice := make([]platformclientv2.Userimage, 0)
	for _, userImage := range userImages {
		var sdkUserImage platformclientv2.Userimage
		userImagesMap, ok := userImage.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkUserImage.Resolution, userImagesMap, "resolution")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkUserImage.ImageUri, userImagesMap, "image_uri")

		userImagesSlice = append(userImagesSlice, sdkUserImage)
	}

	return &userImagesSlice
}

// buildEducations maps an []interface{} into a Genesys Cloud *[]platformclientv2.Education
func buildEducations(educations []interface{}) *[]platformclientv2.Education {
	educationsSlice := make([]platformclientv2.Education, 0)
	for _, education := range educations {
		var sdkEducation platformclientv2.Education
		educationsMap, ok := education.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkEducation.School, educationsMap, "school")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkEducation.FieldOfStudy, educationsMap, "field_of_study")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkEducation.Notes, educationsMap, "notes")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkEducation.DateStart, educationsMap, "date_start")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkEducation.DateEnd, educationsMap, "date_end")

		educationsSlice = append(educationsSlice, sdkEducation)
	}

	return &educationsSlice
}

// buildBiographys maps an []interface{} into a Genesys Cloud *[]platformclientv2.Biography
func buildBiographys(biographys []interface{}) *[]platformclientv2.Biography {
	biographysSlice := make([]platformclientv2.Biography, 0)
	for _, biography := range biographys {
		var sdkBiography platformclientv2.Biography
		biographysMap, ok := biography.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkBiography.Biography, biographysMap, "biography")
		resourcedata.BuildSDKStringArrayValueIfNotNil(&sdkBiography.Interests, biographysMap, "interests")
		resourcedata.BuildSDKStringArrayValueIfNotNil(&sdkBiography.Hobbies, biographysMap, "hobbies")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkBiography.Spouse, biographysMap, "spouse")
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkBiography.Education, biographysMap, "education", buildEducations)

		biographysSlice = append(biographysSlice, sdkBiography)
	}

	return &biographysSlice
}

// buildEmployerInfos maps an []interface{} into a Genesys Cloud *[]platformclientv2.Employerinfo
func buildEmployerInfos(employerInfos []interface{}) *[]platformclientv2.Employerinfo {
	employerInfosSlice := make([]platformclientv2.Employerinfo, 0)
	for _, employerInfo := range employerInfos {
		var sdkEmployerInfo platformclientv2.Employerinfo
		employerInfosMap, ok := employerInfo.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkEmployerInfo.OfficialName, employerInfosMap, "official_name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkEmployerInfo.EmployeeId, employerInfosMap, "employee_id")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkEmployerInfo.EmployeeType, employerInfosMap, "employee_type")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkEmployerInfo.DateHire, employerInfosMap, "date_hire")

		employerInfosSlice = append(employerInfosSlice, sdkEmployerInfo)
	}

	return &employerInfosSlice
}

// buildRoutingStatuss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Routingstatus
func buildRoutingStatuss(routingStatuss []interface{}) *[]platformclientv2.Routingstatus {
	routingStatussSlice := make([]platformclientv2.Routingstatus, 0)
	for _, routingStatus := range routingStatuss {
		var sdkRoutingStatus platformclientv2.Routingstatus
		routingStatussMap, ok := routingStatus.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkRoutingStatus.UserId, routingStatussMap, "user_id")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkRoutingStatus.Status, routingStatussMap, "status")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkRoutingStatus.StartTime, routingStatussMap, "start_time")

		routingStatussSlice = append(routingStatussSlice, sdkRoutingStatus)
	}

	return &routingStatussSlice
}

// buildPresenceDefinitions maps an []interface{} into a Genesys Cloud *[]platformclientv2.Presencedefinition
func buildPresenceDefinitions(presenceDefinitions []interface{}) *[]platformclientv2.Presencedefinition {
	presenceDefinitionsSlice := make([]platformclientv2.Presencedefinition, 0)
	for _, presenceDefinition := range presenceDefinitions {
		var sdkPresenceDefinition platformclientv2.Presencedefinition
		presenceDefinitionsMap, ok := presenceDefinition.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkPresenceDefinition.SystemPresence, presenceDefinitionsMap, "system_presence")

		presenceDefinitionsSlice = append(presenceDefinitionsSlice, sdkPresenceDefinition)
	}

	return &presenceDefinitionsSlice
}

// buildUserPresences maps an []interface{} into a Genesys Cloud *[]platformclientv2.Userpresence
func buildUserPresences(userPresences []interface{}) *[]platformclientv2.Userpresence {
	userPresencesSlice := make([]platformclientv2.Userpresence, 0)
	for _, userPresence := range userPresences {
		var sdkUserPresence platformclientv2.Userpresence
		userPresencesMap, ok := userPresence.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkUserPresence.Name, userPresencesMap, "name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkUserPresence.Source, userPresencesMap, "source")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkUserPresence.SourceId, userPresencesMap, "source_id")
		sdkUserPresence.Primary = platformclientv2.Bool(userPresencesMap["primary"].(bool))
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUserPresence.PresenceDefinition, userPresencesMap, "presence_definition", buildPresenceDefinition)
		resourcedata.BuildSDKStringValueIfNotNil(&sdkUserPresence.Message, userPresencesMap, "message")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkUserPresence.ModifiedDate, userPresencesMap, "modified_date")

		userPresencesSlice = append(userPresencesSlice, sdkUserPresence)
	}

	return &userPresencesSlice
}

// buildMediaSummaryDetails maps an []interface{} into a Genesys Cloud *[]platformclientv2.Mediasummarydetail
func buildMediaSummaryDetails(mediaSummaryDetails []interface{}) *[]platformclientv2.Mediasummarydetail {
	mediaSummaryDetailsSlice := make([]platformclientv2.Mediasummarydetail, 0)
	for _, mediaSummaryDetail := range mediaSummaryDetails {
		var sdkMediaSummaryDetail platformclientv2.Mediasummarydetail
		mediaSummaryDetailsMap, ok := mediaSummaryDetail.(map[string]interface{})
		if !ok {
			continue
		}

		sdkMediaSummaryDetail.Active = platformclientv2.Int(mediaSummaryDetailsMap["active"].(int))
		sdkMediaSummaryDetail.Acw = platformclientv2.Int(mediaSummaryDetailsMap["acw"].(int))

		mediaSummaryDetailsSlice = append(mediaSummaryDetailsSlice, sdkMediaSummaryDetail)
	}

	return &mediaSummaryDetailsSlice
}

// buildMediaSummarys maps an []interface{} into a Genesys Cloud *[]platformclientv2.Mediasummary
func buildMediaSummarys(mediaSummarys []interface{}) *[]platformclientv2.Mediasummary {
	mediaSummarysSlice := make([]platformclientv2.Mediasummary, 0)
	for _, mediaSummary := range mediaSummarys {
		var sdkMediaSummary platformclientv2.Mediasummary
		mediaSummarysMap, ok := mediaSummary.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkMediaSummary.ContactCenter, mediaSummarysMap, "contact_center", buildMediaSummaryDetail)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkMediaSummary.Enterprise, mediaSummarysMap, "enterprise", buildMediaSummaryDetail)

		mediaSummarysSlice = append(mediaSummarysSlice, sdkMediaSummary)
	}

	return &mediaSummarysSlice
}

// buildUserConversationSummarys maps an []interface{} into a Genesys Cloud *[]platformclientv2.Userconversationsummary
func buildUserConversationSummarys(userConversationSummarys []interface{}) *[]platformclientv2.Userconversationsummary {
	userConversationSummarysSlice := make([]platformclientv2.Userconversationsummary, 0)
	for _, userConversationSummary := range userConversationSummarys {
		var sdkUserConversationSummary platformclientv2.Userconversationsummary
		userConversationSummarysMap, ok := userConversationSummary.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkUserConversationSummary.UserId, userConversationSummarysMap, "user_id")
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUserConversationSummary.Call, userConversationSummarysMap, "call", buildMediaSummary)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUserConversationSummary.Callback, userConversationSummarysMap, "callback", buildMediaSummary)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUserConversationSummary.Email, userConversationSummarysMap, "email", buildMediaSummary)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUserConversationSummary.Message, userConversationSummarysMap, "message", buildMediaSummary)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUserConversationSummary.Chat, userConversationSummarysMap, "chat", buildMediaSummary)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUserConversationSummary.SocialExpression, userConversationSummarysMap, "social_expression", buildMediaSummary)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUserConversationSummary.Video, userConversationSummarysMap, "video", buildMediaSummary)

		userConversationSummarysSlice = append(userConversationSummarysSlice, sdkUserConversationSummary)
	}

	return &userConversationSummarysSlice
}

// buildOutOfOffices maps an []interface{} into a Genesys Cloud *[]platformclientv2.Outofoffice
func buildOutOfOffices(outOfOffices []interface{}) *[]platformclientv2.Outofoffice {
	outOfOfficesSlice := make([]platformclientv2.Outofoffice, 0)
	for _, outOfOffice := range outOfOffices {
		var sdkOutOfOffice platformclientv2.Outofoffice
		outOfOfficesMap, ok := outOfOffice.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkOutOfOffice.Name, outOfOfficesMap, "name")
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkOutOfOffice.User, outOfOfficesMap, "user", buildUser)
		resourcedata.BuildSDKStringValueIfNotNil(&sdkOutOfOffice.StartDate, outOfOfficesMap, "start_date")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkOutOfOffice.EndDate, outOfOfficesMap, "end_date")
		sdkOutOfOffice.Active = platformclientv2.Bool(outOfOfficesMap["active"].(bool))
		sdkOutOfOffice.Indefinite = platformclientv2.Bool(outOfOfficesMap["indefinite"].(bool))

		outOfOfficesSlice = append(outOfOfficesSlice, sdkOutOfOffice)
	}

	return &outOfOfficesSlice
}

// buildAddressableEntityRefs maps an []interface{} into a Genesys Cloud *[]platformclientv2.Addressableentityref
func buildAddressableEntityRefs(addressableEntityRefs []interface{}) *[]platformclientv2.Addressableentityref {
	addressableEntityRefsSlice := make([]platformclientv2.Addressableentityref, 0)
	for _, addressableEntityRef := range addressableEntityRefs {
		var sdkAddressableEntityRef platformclientv2.Addressableentityref
		addressableEntityRefsMap, ok := addressableEntityRef.(map[string]interface{})
		if !ok {
			continue
		}

		addressableEntityRefsSlice = append(addressableEntityRefsSlice, sdkAddressableEntityRef)
	}

	return &addressableEntityRefsSlice
}

// buildLocationEmergencyNumbers maps an []interface{} into a Genesys Cloud *[]platformclientv2.Locationemergencynumber
func buildLocationEmergencyNumbers(locationEmergencyNumbers []interface{}) *[]platformclientv2.Locationemergencynumber {
	locationEmergencyNumbersSlice := make([]platformclientv2.Locationemergencynumber, 0)
	for _, locationEmergencyNumber := range locationEmergencyNumbers {
		var sdkLocationEmergencyNumber platformclientv2.Locationemergencynumber
		locationEmergencyNumbersMap, ok := locationEmergencyNumber.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkLocationEmergencyNumber.E164, locationEmergencyNumbersMap, "e164")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkLocationEmergencyNumber.Number, locationEmergencyNumbersMap, "number")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkLocationEmergencyNumber.Type, locationEmergencyNumbersMap, "type")

		locationEmergencyNumbersSlice = append(locationEmergencyNumbersSlice, sdkLocationEmergencyNumber)
	}

	return &locationEmergencyNumbersSlice
}

// buildLocationAddresss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Locationaddress
func buildLocationAddresss(locationAddresss []interface{}) *[]platformclientv2.Locationaddress {
	locationAddresssSlice := make([]platformclientv2.Locationaddress, 0)
	for _, locationAddress := range locationAddresss {
		var sdkLocationAddress platformclientv2.Locationaddress
		locationAddresssMap, ok := locationAddress.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkLocationAddress.City, locationAddresssMap, "city")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkLocationAddress.Country, locationAddresssMap, "country")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkLocationAddress.CountryName, locationAddresssMap, "country_name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkLocationAddress.State, locationAddresssMap, "state")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkLocationAddress.Street1, locationAddresssMap, "street1")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkLocationAddress.Street2, locationAddresssMap, "street2")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkLocationAddress.Zipcode, locationAddresssMap, "zipcode")

		locationAddresssSlice = append(locationAddresssSlice, sdkLocationAddress)
	}

	return &locationAddresssSlice
}

// buildLocationImages maps an []interface{} into a Genesys Cloud *[]platformclientv2.Locationimage
func buildLocationImages(locationImages []interface{}) *[]platformclientv2.Locationimage {
	locationImagesSlice := make([]platformclientv2.Locationimage, 0)
	for _, locationImage := range locationImages {
		var sdkLocationImage platformclientv2.Locationimage
		locationImagesMap, ok := locationImage.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkLocationImage.Resolution, locationImagesMap, "resolution")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkLocationImage.ImageUri, locationImagesMap, "image_uri")

		locationImagesSlice = append(locationImagesSlice, sdkLocationImage)
	}

	return &locationImagesSlice
}

// buildLocationAddressVerificationDetailss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Locationaddressverificationdetails
func buildLocationAddressVerificationDetailss(locationAddressVerificationDetailss []interface{}) *[]platformclientv2.Locationaddressverificationdetails {
	locationAddressVerificationDetailssSlice := make([]platformclientv2.Locationaddressverificationdetails, 0)
	for _, locationAddressVerificationDetails := range locationAddressVerificationDetailss {
		var sdkLocationAddressVerificationDetails platformclientv2.Locationaddressverificationdetails
		locationAddressVerificationDetailssMap, ok := locationAddressVerificationDetails.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkLocationAddressVerificationDetails.Status, locationAddressVerificationDetailssMap, "status")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkLocationAddressVerificationDetails.DateFinished, locationAddressVerificationDetailssMap, "date_finished")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkLocationAddressVerificationDetails.DateStarted, locationAddressVerificationDetailssMap, "date_started")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkLocationAddressVerificationDetails.Service, locationAddressVerificationDetailssMap, "service")

		locationAddressVerificationDetailssSlice = append(locationAddressVerificationDetailssSlice, sdkLocationAddressVerificationDetails)
	}

	return &locationAddressVerificationDetailssSlice
}

// buildLocationDefinitions maps an []interface{} into a Genesys Cloud *[]platformclientv2.Locationdefinition
func buildLocationDefinitions(locationDefinitions []interface{}) *[]platformclientv2.Locationdefinition {
	locationDefinitionsSlice := make([]platformclientv2.Locationdefinition, 0)
	for _, locationDefinition := range locationDefinitions {
		var sdkLocationDefinition platformclientv2.Locationdefinition
		locationDefinitionsMap, ok := locationDefinition.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkLocationDefinition.Name, locationDefinitionsMap, "name")
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkLocationDefinition.ContactUser, locationDefinitionsMap, "contact_user", buildAddressableEntityRef)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkLocationDefinition.EmergencyNumber, locationDefinitionsMap, "emergency_number", buildLocationEmergencyNumber)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkLocationDefinition.Address, locationDefinitionsMap, "address", buildLocationAddress)
		resourcedata.BuildSDKStringValueIfNotNil(&sdkLocationDefinition.State, locationDefinitionsMap, "state")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkLocationDefinition.Notes, locationDefinitionsMap, "notes")
		resourcedata.BuildSDKStringArrayValueIfNotNil(&sdkLocationDefinition.Path, locationDefinitionsMap, "path")
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkLocationDefinition.ProfileImage, locationDefinitionsMap, "profile_image", buildLocationImages)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkLocationDefinition.FloorplanImage, locationDefinitionsMap, "floorplan_image", buildLocationImages)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkLocationDefinition.AddressVerificationDetails, locationDefinitionsMap, "address_verification_details", buildLocationAddressVerificationDetails)
		sdkLocationDefinition.AddressVerified = platformclientv2.Bool(locationDefinitionsMap["address_verified"].(bool))
		sdkLocationDefinition.AddressStored = platformclientv2.Bool(locationDefinitionsMap["address_stored"].(bool))
		resourcedata.BuildSDKStringValueIfNotNil(&sdkLocationDefinition.Images, locationDefinitionsMap, "images")

		locationDefinitionsSlice = append(locationDefinitionsSlice, sdkLocationDefinition)
	}

	return &locationDefinitionsSlice
}

// buildGeolocations maps an []interface{} into a Genesys Cloud *[]platformclientv2.Geolocation
func buildGeolocations(geolocations []interface{}) *[]platformclientv2.Geolocation {
	geolocationsSlice := make([]platformclientv2.Geolocation, 0)
	for _, geolocation := range geolocations {
		var sdkGeolocation platformclientv2.Geolocation
		geolocationsMap, ok := geolocation.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkGeolocation.Name, geolocationsMap, "name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkGeolocation.Type, geolocationsMap, "type")
		sdkGeolocation.Primary = platformclientv2.Bool(geolocationsMap["primary"].(bool))
		// TODO: Handle latitude property
		// TODO: Handle longitude property
		resourcedata.BuildSDKStringValueIfNotNil(&sdkGeolocation.Country, geolocationsMap, "country")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkGeolocation.Region, geolocationsMap, "region")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkGeolocation.City, geolocationsMap, "city")
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkGeolocation.Locations, geolocationsMap, "locations", buildLocationDefinitions)

		geolocationsSlice = append(geolocationsSlice, sdkGeolocation)
	}

	return &geolocationsSlice
}

// buildUserStations maps an []interface{} into a Genesys Cloud *[]platformclientv2.Userstation
func buildUserStations(userStations []interface{}) *[]platformclientv2.Userstation {
	userStationsSlice := make([]platformclientv2.Userstation, 0)
	for _, userStation := range userStations {
		var sdkUserStation platformclientv2.Userstation
		userStationsMap, ok := userStation.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkUserStation.Name, userStationsMap, "name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkUserStation.Type, userStationsMap, "type")
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUserStation.AssociatedUser, userStationsMap, "associated_user", buildUser)
		resourcedata.BuildSDKStringValueIfNotNil(&sdkUserStation.AssociatedDate, userStationsMap, "associated_date")
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUserStation.DefaultUser, userStationsMap, "default_user", buildUser)
		// TODO: Handle provider_info property
		sdkUserStation.WebRtcCallAppearances = platformclientv2.Int(userStationsMap["web_rtc_call_appearances"].(int))

		userStationsSlice = append(userStationsSlice, sdkUserStation)
	}

	return &userStationsSlice
}

// buildUserStationss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Userstations
func buildUserStationss(userStationss []interface{}) *[]platformclientv2.Userstations {
	userStationssSlice := make([]platformclientv2.Userstations, 0)
	for _, userStations := range userStationss {
		var sdkUserStations platformclientv2.Userstations
		userStationssMap, ok := userStations.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUserStations.AssociatedStation, userStationssMap, "associated_station", buildUserStation)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUserStations.EffectiveStation, userStationssMap, "effective_station", buildUserStation)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUserStations.DefaultStation, userStationssMap, "default_station", buildUserStation)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUserStations.LastAssociatedStation, userStationssMap, "last_associated_station", buildUserStation)

		userStationssSlice = append(userStationssSlice, sdkUserStations)
	}

	return &userStationssSlice
}

// buildDomainRoles maps an []interface{} into a Genesys Cloud *[]platformclientv2.Domainrole
func buildDomainRoles(domainRoles []interface{}) *[]platformclientv2.Domainrole {
	domainRolesSlice := make([]platformclientv2.Domainrole, 0)
	for _, domainRole := range domainRoles {
		var sdkDomainRole platformclientv2.Domainrole
		domainRolesMap, ok := domainRole.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkDomainRole.Name, domainRolesMap, "name")

		domainRolesSlice = append(domainRolesSlice, sdkDomainRole)
	}

	return &domainRolesSlice
}

// buildResourceConditionValues maps an []interface{} into a Genesys Cloud *[]platformclientv2.Resourceconditionvalue
func buildResourceConditionValues(resourceConditionValues []interface{}) *[]platformclientv2.Resourceconditionvalue {
	resourceConditionValuesSlice := make([]platformclientv2.Resourceconditionvalue, 0)
	for _, resourceConditionValue := range resourceConditionValues {
		var sdkResourceConditionValue platformclientv2.Resourceconditionvalue
		resourceConditionValuesMap, ok := resourceConditionValue.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkResourceConditionValue.Type, resourceConditionValuesMap, "type")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkResourceConditionValue.Value, resourceConditionValuesMap, "value")

		resourceConditionValuesSlice = append(resourceConditionValuesSlice, sdkResourceConditionValue)
	}

	return &resourceConditionValuesSlice
}

// buildResourceConditionNodes maps an []interface{} into a Genesys Cloud *[]platformclientv2.Resourceconditionnode
func buildResourceConditionNodes(resourceConditionNodes []interface{}) *[]platformclientv2.Resourceconditionnode {
	resourceConditionNodesSlice := make([]platformclientv2.Resourceconditionnode, 0)
	for _, resourceConditionNode := range resourceConditionNodes {
		var sdkResourceConditionNode platformclientv2.Resourceconditionnode
		resourceConditionNodesMap, ok := resourceConditionNode.(map[string]interface{})
		if !ok {
			continue
		}

		resourceConditionNodesSlice = append(resourceConditionNodesSlice, sdkResourceConditionNode)
	}

	return &resourceConditionNodesSlice
}

// buildResourceConditionNodes maps an []interface{} into a Genesys Cloud *[]platformclientv2.Resourceconditionnode
func buildResourceConditionNodes(resourceConditionNodes []interface{}) *[]platformclientv2.Resourceconditionnode {
	resourceConditionNodesSlice := make([]platformclientv2.Resourceconditionnode, 0)
	for _, resourceConditionNode := range resourceConditionNodes {
		var sdkResourceConditionNode platformclientv2.Resourceconditionnode
		resourceConditionNodesMap, ok := resourceConditionNode.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkResourceConditionNode.VariableName, resourceConditionNodesMap, "variable_name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkResourceConditionNode.Conjunction, resourceConditionNodesMap, "conjunction")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkResourceConditionNode.Operator, resourceConditionNodesMap, "operator")
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkResourceConditionNode.Operands, resourceConditionNodesMap, "operands", buildResourceConditionValues)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkResourceConditionNode.Terms, resourceConditionNodesMap, "terms", buildResourceConditionNodes)

		resourceConditionNodesSlice = append(resourceConditionNodesSlice, sdkResourceConditionNode)
	}

	return &resourceConditionNodesSlice
}

// buildResourcePermissionPolicys maps an []interface{} into a Genesys Cloud *[]platformclientv2.Resourcepermissionpolicy
func buildResourcePermissionPolicys(resourcePermissionPolicys []interface{}) *[]platformclientv2.Resourcepermissionpolicy {
	resourcePermissionPolicysSlice := make([]platformclientv2.Resourcepermissionpolicy, 0)
	for _, resourcePermissionPolicy := range resourcePermissionPolicys {
		var sdkResourcePermissionPolicy platformclientv2.Resourcepermissionpolicy
		resourcePermissionPolicysMap, ok := resourcePermissionPolicy.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkResourcePermissionPolicy.Domain, resourcePermissionPolicysMap, "domain")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkResourcePermissionPolicy.EntityName, resourcePermissionPolicysMap, "entity_name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkResourcePermissionPolicy.PolicyName, resourcePermissionPolicysMap, "policy_name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkResourcePermissionPolicy.PolicyDescription, resourcePermissionPolicysMap, "policy_description")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkResourcePermissionPolicy.ActionSetKey, resourcePermissionPolicysMap, "action_set_key")
		sdkResourcePermissionPolicy.AllowConditions = platformclientv2.Bool(resourcePermissionPolicysMap["allow_conditions"].(bool))
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkResourcePermissionPolicy.ResourceConditionNode, resourcePermissionPolicysMap, "resource_condition_node", buildResourceConditionNode)
		resourcedata.BuildSDKStringArrayValueIfNotNil(&sdkResourcePermissionPolicy.NamedResources, resourcePermissionPolicysMap, "named_resources")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkResourcePermissionPolicy.ResourceCondition, resourcePermissionPolicysMap, "resource_condition")
		resourcedata.BuildSDKStringArrayValueIfNotNil(&sdkResourcePermissionPolicy.ActionSet, resourcePermissionPolicysMap, "action_set")

		resourcePermissionPolicysSlice = append(resourcePermissionPolicysSlice, sdkResourcePermissionPolicy)
	}

	return &resourcePermissionPolicysSlice
}

// buildUserAuthorizations maps an []interface{} into a Genesys Cloud *[]platformclientv2.Userauthorization
func buildUserAuthorizations(userAuthorizations []interface{}) *[]platformclientv2.Userauthorization {
	userAuthorizationsSlice := make([]platformclientv2.Userauthorization, 0)
	for _, userAuthorization := range userAuthorizations {
		var sdkUserAuthorization platformclientv2.Userauthorization
		userAuthorizationsMap, ok := userAuthorization.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUserAuthorization.Roles, userAuthorizationsMap, "roles", buildDomainRoles)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUserAuthorization.UnusedRoles, userAuthorizationsMap, "unused_roles", buildDomainRoles)
		resourcedata.BuildSDKStringArrayValueIfNotNil(&sdkUserAuthorization.Permissions, userAuthorizationsMap, "permissions")
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUserAuthorization.PermissionPolicies, userAuthorizationsMap, "permission_policies", buildResourcePermissionPolicys)

		userAuthorizationsSlice = append(userAuthorizationsSlice, sdkUserAuthorization)
	}

	return &userAuthorizationsSlice
}

// buildLocations maps an []interface{} into a Genesys Cloud *[]platformclientv2.Location
func buildLocations(locations []interface{}) *[]platformclientv2.Location {
	locationsSlice := make([]platformclientv2.Location, 0)
	for _, location := range locations {
		var sdkLocation platformclientv2.Location
		locationsMap, ok := location.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkLocation.FloorplanId, locationsMap, "floorplan_id")
		// TODO: Handle coordinates property
		resourcedata.BuildSDKStringValueIfNotNil(&sdkLocation.Notes, locationsMap, "notes")
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkLocation.LocationDefinition, locationsMap, "location_definition", buildLocationDefinition)

		locationsSlice = append(locationsSlice, sdkLocation)
	}

	return &locationsSlice
}

// buildGroupContacts maps an []interface{} into a Genesys Cloud *[]platformclientv2.Groupcontact
func buildGroupContacts(groupContacts []interface{}) *[]platformclientv2.Groupcontact {
	groupContactsSlice := make([]platformclientv2.Groupcontact, 0)
	for _, groupContact := range groupContacts {
		var sdkGroupContact platformclientv2.Groupcontact
		groupContactsMap, ok := groupContact.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkGroupContact.Address, groupContactsMap, "address")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkGroupContact.Extension, groupContactsMap, "extension")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkGroupContact.Display, groupContactsMap, "display")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkGroupContact.Type, groupContactsMap, "type")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkGroupContact.MediaType, groupContactsMap, "media_type")

		groupContactsSlice = append(groupContactsSlice, sdkGroupContact)
	}

	return &groupContactsSlice
}

// buildGroups maps an []interface{} into a Genesys Cloud *[]platformclientv2.Group
func buildGroups(groups []interface{}) *[]platformclientv2.Group {
	groupsSlice := make([]platformclientv2.Group, 0)
	for _, group := range groups {
		var sdkGroup platformclientv2.Group
		groupsMap, ok := group.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkGroup.Name, groupsMap, "name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkGroup.Description, groupsMap, "description")
		sdkGroup.MemberCount = platformclientv2.Int(groupsMap["member_count"].(int))
		resourcedata.BuildSDKStringValueIfNotNil(&sdkGroup.State, groupsMap, "state")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkGroup.Type, groupsMap, "type")
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkGroup.Images, groupsMap, "images", buildUserImages)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkGroup.Addresses, groupsMap, "addresses", buildGroupContacts)
		sdkGroup.RulesVisible = platformclientv2.Bool(groupsMap["rules_visible"].(bool))
		resourcedata.BuildSDKStringValueIfNotNil(&sdkGroup.Visibility, groupsMap, "visibility")
		sdkGroup.RolesEnabled = platformclientv2.Bool(groupsMap["roles_enabled"].(bool))
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkGroup.Owners, groupsMap, "owners", buildUsers)

		groupsSlice = append(groupsSlice, sdkGroup)
	}

	return &groupsSlice
}

// buildTeams maps an []interface{} into a Genesys Cloud *[]platformclientv2.Team
func buildTeams(teams []interface{}) *[]platformclientv2.Team {
	teamsSlice := make([]platformclientv2.Team, 0)
	for _, team := range teams {
		var sdkTeam platformclientv2.Team
		teamsMap, ok := team.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkTeam.Name, teamsMap, "name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkTeam.DivisionId, teamsMap, "division_id")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkTeam.Description, teamsMap, "description")
		sdkTeam.MemberCount = platformclientv2.Int(teamsMap["member_count"].(int))

		teamsSlice = append(teamsSlice, sdkTeam)
	}

	return &teamsSlice
}

// buildUserRoutingSkills maps an []interface{} into a Genesys Cloud *[]platformclientv2.Userroutingskill
func buildUserRoutingSkills(userRoutingSkills []interface{}) *[]platformclientv2.Userroutingskill {
	userRoutingSkillsSlice := make([]platformclientv2.Userroutingskill, 0)
	for _, userRoutingSkill := range userRoutingSkills {
		var sdkUserRoutingSkill platformclientv2.Userroutingskill
		userRoutingSkillsMap, ok := userRoutingSkill.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkUserRoutingSkill.Name, userRoutingSkillsMap, "name")
		// TODO: Handle proficiency property
		resourcedata.BuildSDKStringValueIfNotNil(&sdkUserRoutingSkill.State, userRoutingSkillsMap, "state")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkUserRoutingSkill.SkillUri, userRoutingSkillsMap, "skill_uri")

		userRoutingSkillsSlice = append(userRoutingSkillsSlice, sdkUserRoutingSkill)
	}

	return &userRoutingSkillsSlice
}

// buildUserRoutingLanguages maps an []interface{} into a Genesys Cloud *[]platformclientv2.Userroutinglanguage
func buildUserRoutingLanguages(userRoutingLanguages []interface{}) *[]platformclientv2.Userroutinglanguage {
	userRoutingLanguagesSlice := make([]platformclientv2.Userroutinglanguage, 0)
	for _, userRoutingLanguage := range userRoutingLanguages {
		var sdkUserRoutingLanguage platformclientv2.Userroutinglanguage
		userRoutingLanguagesMap, ok := userRoutingLanguage.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkUserRoutingLanguage.Name, userRoutingLanguagesMap, "name")
		// TODO: Handle proficiency property
		resourcedata.BuildSDKStringValueIfNotNil(&sdkUserRoutingLanguage.State, userRoutingLanguagesMap, "state")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkUserRoutingLanguage.LanguageUri, userRoutingLanguagesMap, "language_uri")

		userRoutingLanguagesSlice = append(userRoutingLanguagesSlice, sdkUserRoutingLanguage)
	}

	return &userRoutingLanguagesSlice
}

// buildOAuthLastTokenIssueds maps an []interface{} into a Genesys Cloud *[]platformclientv2.Oauthlasttokenissued
func buildOAuthLastTokenIssueds(oAuthLastTokenIssueds []interface{}) *[]platformclientv2.Oauthlasttokenissued {
	oAuthLastTokenIssuedsSlice := make([]platformclientv2.Oauthlasttokenissued, 0)
	for _, oAuthLastTokenIssued := range oAuthLastTokenIssueds {
		var sdkOAuthLastTokenIssued platformclientv2.Oauthlasttokenissued
		oAuthLastTokenIssuedsMap, ok := oAuthLastTokenIssued.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkOAuthLastTokenIssued.DateIssued, oAuthLastTokenIssuedsMap, "date_issued")

		oAuthLastTokenIssuedsSlice = append(oAuthLastTokenIssuedsSlice, sdkOAuthLastTokenIssued)
	}

	return &oAuthLastTokenIssuedsSlice
}

// buildUsers maps an []interface{} into a Genesys Cloud *[]platformclientv2.User
func buildUsers(users []interface{}) *[]platformclientv2.User {
	usersSlice := make([]platformclientv2.User, 0)
	for _, user := range users {
		var sdkUser platformclientv2.User
		usersMap, ok := user.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkUser.Name, usersMap, "name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkUser.DivisionId, usersMap, "division_id")
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUser.Chat, usersMap, "chat", buildChat)
		resourcedata.BuildSDKStringValueIfNotNil(&sdkUser.Department, usersMap, "department")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkUser.Email, usersMap, "email")
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUser.PrimaryContactInfo, usersMap, "primary_contact_info", buildContacts)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUser.Addresses, usersMap, "addresses", buildContacts)
		resourcedata.BuildSDKStringValueIfNotNil(&sdkUser.State, usersMap, "state")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkUser.Title, usersMap, "title")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkUser.Username, usersMap, "username")
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUser.Manager, usersMap, "manager", buildUser)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUser.Images, usersMap, "images", buildUserImages)
		resourcedata.BuildSDKStringArrayValueIfNotNil(&sdkUser.Certifications, usersMap, "certifications")
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUser.Biography, usersMap, "biography", buildBiography)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUser.EmployerInfo, usersMap, "employer_info", buildEmployerInfo)
		resourcedata.BuildSDKStringValueIfNotNil(&sdkUser.PreferredName, usersMap, "preferred_name")
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUser.RoutingStatus, usersMap, "routing_status", buildRoutingStatus)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUser.Presence, usersMap, "presence", buildUserPresence)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUser.IntegrationPresence, usersMap, "integration_presence", buildUserPresence)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUser.ConversationSummary, usersMap, "conversation_summary", buildUserConversationSummary)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUser.OutOfOffice, usersMap, "out_of_office", buildOutOfOffice)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUser.Geolocation, usersMap, "geolocation", buildGeolocation)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUser.Station, usersMap, "station", buildUserStations)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUser.Authorization, usersMap, "authorization", buildUserAuthorization)
		resourcedata.BuildSDKStringArrayValueIfNotNil(&sdkUser.ProfileSkills, usersMap, "profile_skills")
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUser.Locations, usersMap, "locations", buildLocations)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUser.Groups, usersMap, "groups", buildGroups)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUser.Team, usersMap, "team", buildTeam)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUser.Skills, usersMap, "skills", buildUserRoutingSkills)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUser.Languages, usersMap, "languages", buildUserRoutingLanguages)
		sdkUser.AcdAutoAnswer = platformclientv2.Bool(usersMap["acd_auto_answer"].(bool))
		resourcedata.BuildSDKStringValueIfNotNil(&sdkUser.LanguagePreference, usersMap, "language_preference")
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkUser.LastTokenIssued, usersMap, "last_token_issued", buildOAuthLastTokenIssued)
		resourcedata.BuildSDKStringValueIfNotNil(&sdkUser.DateLastLogin, usersMap, "date_last_login")

		usersSlice = append(usersSlice, sdkUser)
	}

	return &usersSlice
}

// buildServiceLevels maps an []interface{} into a Genesys Cloud *[]platformclientv2.Servicelevel
func buildServiceLevels(serviceLevels []interface{}) *[]platformclientv2.Servicelevel {
	serviceLevelsSlice := make([]platformclientv2.Servicelevel, 0)
	for _, serviceLevel := range serviceLevels {
		var sdkServiceLevel platformclientv2.Servicelevel
		serviceLevelsMap, ok := serviceLevel.(map[string]interface{})
		if !ok {
			continue
		}

		// TODO: Handle percentage property
		sdkServiceLevel.DurationMs = platformclientv2.Int(serviceLevelsMap["duration_ms"].(int))

		serviceLevelsSlice = append(serviceLevelsSlice, sdkServiceLevel)
	}

	return &serviceLevelsSlice
}

// buildMediaSettingss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Mediasettings
func buildMediaSettingss(mediaSettingss []interface{}) *[]platformclientv2.Mediasettings {
	mediaSettingssSlice := make([]platformclientv2.Mediasettings, 0)
	for _, mediaSettings := range mediaSettingss {
		var sdkMediaSettings platformclientv2.Mediasettings
		mediaSettingssMap, ok := mediaSettings.(map[string]interface{})
		if !ok {
			continue
		}

		sdkMediaSettings.EnableAutoAnswer = platformclientv2.Bool(mediaSettingssMap["enable_auto_answer"].(bool))
		sdkMediaSettings.AlertingTimeoutSeconds = platformclientv2.Int(mediaSettingssMap["alerting_timeout_seconds"].(int))
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkMediaSettings.ServiceLevel, mediaSettingssMap, "service_level", buildServiceLevel)
		// TODO: Handle auto_answer_alert_tone_seconds property
		// TODO: Handle manual_answer_alert_tone_seconds property
		// TODO: Handle sub_type_settings property

		mediaSettingssSlice = append(mediaSettingssSlice, sdkMediaSettings)
	}

	return &mediaSettingssSlice
}

// buildCallbackMediaSettingss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Callbackmediasettings
func buildCallbackMediaSettingss(callbackMediaSettingss []interface{}) *[]platformclientv2.Callbackmediasettings {
	callbackMediaSettingssSlice := make([]platformclientv2.Callbackmediasettings, 0)
	for _, callbackMediaSettings := range callbackMediaSettingss {
		var sdkCallbackMediaSettings platformclientv2.Callbackmediasettings
		callbackMediaSettingssMap, ok := callbackMediaSettings.(map[string]interface{})
		if !ok {
			continue
		}

		sdkCallbackMediaSettings.EnableAutoAnswer = platformclientv2.Bool(callbackMediaSettingssMap["enable_auto_answer"].(bool))
		sdkCallbackMediaSettings.AlertingTimeoutSeconds = platformclientv2.Int(callbackMediaSettingssMap["alerting_timeout_seconds"].(int))
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkCallbackMediaSettings.ServiceLevel, callbackMediaSettingssMap, "service_level", buildServiceLevel)
		// TODO: Handle auto_answer_alert_tone_seconds property
		// TODO: Handle manual_answer_alert_tone_seconds property
		// TODO: Handle sub_type_settings property
		sdkCallbackMediaSettings.EnableAutoDialAndEnd = platformclientv2.Bool(callbackMediaSettingssMap["enable_auto_dial_and_end"].(bool))
		sdkCallbackMediaSettings.AutoDialDelaySeconds = platformclientv2.Int(callbackMediaSettingssMap["auto_dial_delay_seconds"].(int))
		sdkCallbackMediaSettings.AutoEndDelaySeconds = platformclientv2.Int(callbackMediaSettingssMap["auto_end_delay_seconds"].(int))

		callbackMediaSettingssSlice = append(callbackMediaSettingssSlice, sdkCallbackMediaSettings)
	}

	return &callbackMediaSettingssSlice
}

// buildQueueMediaSettingss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Queuemediasettings
func buildQueueMediaSettingss(queueMediaSettingss []interface{}) *[]platformclientv2.Queuemediasettings {
	queueMediaSettingssSlice := make([]platformclientv2.Queuemediasettings, 0)
	for _, queueMediaSettings := range queueMediaSettingss {
		var sdkQueueMediaSettings platformclientv2.Queuemediasettings
		queueMediaSettingssMap, ok := queueMediaSettings.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkQueueMediaSettings.Call, queueMediaSettingssMap, "call", buildMediaSettings)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkQueueMediaSettings.Callback, queueMediaSettingssMap, "callback", buildCallbackMediaSettings)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkQueueMediaSettings.Chat, queueMediaSettingssMap, "chat", buildMediaSettings)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkQueueMediaSettings.Email, queueMediaSettingssMap, "email", buildMediaSettings)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkQueueMediaSettings.Message, queueMediaSettingssMap, "message", buildMediaSettings)

		queueMediaSettingssSlice = append(queueMediaSettingssSlice, sdkQueueMediaSettings)
	}

	return &queueMediaSettingssSlice
}

// buildRoutingRules maps an []interface{} into a Genesys Cloud *[]platformclientv2.Routingrule
func buildRoutingRules(routingRules []interface{}) *[]platformclientv2.Routingrule {
	routingRulesSlice := make([]platformclientv2.Routingrule, 0)
	for _, routingRule := range routingRules {
		var sdkRoutingRule platformclientv2.Routingrule
		routingRulesMap, ok := routingRule.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkRoutingRule.Operator, routingRulesMap, "operator")
		sdkRoutingRule.Threshold = platformclientv2.Int(routingRulesMap["threshold"].(int))
		// TODO: Handle wait_seconds property

		routingRulesSlice = append(routingRulesSlice, sdkRoutingRule)
	}

	return &routingRulesSlice
}

// buildMemberGroups maps an []interface{} into a Genesys Cloud *[]platformclientv2.Membergroup
func buildMemberGroups(memberGroups []interface{}) *[]platformclientv2.Membergroup {
	memberGroupsSlice := make([]platformclientv2.Membergroup, 0)
	for _, memberGroup := range memberGroups {
		var sdkMemberGroup platformclientv2.Membergroup
		memberGroupsMap, ok := memberGroup.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkMemberGroup.Name, memberGroupsMap, "name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkMemberGroup.DivisionId, memberGroupsMap, "division_id")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkMemberGroup.Type, memberGroupsMap, "type")
		sdkMemberGroup.MemberCount = platformclientv2.Int(memberGroupsMap["member_count"].(int))

		memberGroupsSlice = append(memberGroupsSlice, sdkMemberGroup)
	}

	return &memberGroupsSlice
}

// buildConditionalGroupRoutingRules maps an []interface{} into a Genesys Cloud *[]platformclientv2.Conditionalgrouproutingrule
func buildConditionalGroupRoutingRules(conditionalGroupRoutingRules []interface{}) *[]platformclientv2.Conditionalgrouproutingrule {
	conditionalGroupRoutingRulesSlice := make([]platformclientv2.Conditionalgrouproutingrule, 0)
	for _, conditionalGroupRoutingRule := range conditionalGroupRoutingRules {
		var sdkConditionalGroupRoutingRule platformclientv2.Conditionalgrouproutingrule
		conditionalGroupRoutingRulesMap, ok := conditionalGroupRoutingRule.(map[string]interface{})
		if !ok {
			continue
		}

		sdkConditionalGroupRoutingRule.QueueId = &platformclientv2.Domainentityref{Id: platformclientv2.String(conditionalGroupRoutingRulesMap["queue_id"].(string))}
		resourcedata.BuildSDKStringValueIfNotNil(&sdkConditionalGroupRoutingRule.Metric, conditionalGroupRoutingRulesMap, "metric")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkConditionalGroupRoutingRule.Operator, conditionalGroupRoutingRulesMap, "operator")
		// TODO: Handle condition_value property
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkConditionalGroupRoutingRule.Groups, conditionalGroupRoutingRulesMap, "groups", buildMemberGroups)
		sdkConditionalGroupRoutingRule.WaitSeconds = platformclientv2.Int(conditionalGroupRoutingRulesMap["wait_seconds"].(int))

		conditionalGroupRoutingRulesSlice = append(conditionalGroupRoutingRulesSlice, sdkConditionalGroupRoutingRule)
	}

	return &conditionalGroupRoutingRulesSlice
}

// buildConditionalGroupRoutings maps an []interface{} into a Genesys Cloud *[]platformclientv2.Conditionalgrouprouting
func buildConditionalGroupRoutings(conditionalGroupRoutings []interface{}) *[]platformclientv2.Conditionalgrouprouting {
	conditionalGroupRoutingsSlice := make([]platformclientv2.Conditionalgrouprouting, 0)
	for _, conditionalGroupRouting := range conditionalGroupRoutings {
		var sdkConditionalGroupRouting platformclientv2.Conditionalgrouprouting
		conditionalGroupRoutingsMap, ok := conditionalGroupRouting.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkConditionalGroupRouting.Rules, conditionalGroupRoutingsMap, "rules", buildConditionalGroupRoutingRules)

		conditionalGroupRoutingsSlice = append(conditionalGroupRoutingsSlice, sdkConditionalGroupRouting)
	}

	return &conditionalGroupRoutingsSlice
}

// buildExpansionCriteriums maps an []interface{} into a Genesys Cloud *[]platformclientv2.Expansioncriterium
func buildExpansionCriteriums(expansionCriteriums []interface{}) *[]platformclientv2.Expansioncriterium {
	expansionCriteriumsSlice := make([]platformclientv2.Expansioncriterium, 0)
	for _, expansionCriterium := range expansionCriteriums {
		var sdkExpansionCriterium platformclientv2.Expansioncriterium
		expansionCriteriumsMap, ok := expansionCriterium.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkExpansionCriterium.Type, expansionCriteriumsMap, "type")
		// TODO: Handle threshold property

		expansionCriteriumsSlice = append(expansionCriteriumsSlice, sdkExpansionCriterium)
	}

	return &expansionCriteriumsSlice
}

// buildSkillsToRemoves maps an []interface{} into a Genesys Cloud *[]platformclientv2.Skillstoremove
func buildSkillsToRemoves(skillsToRemoves []interface{}) *[]platformclientv2.Skillstoremove {
	skillsToRemovesSlice := make([]platformclientv2.Skillstoremove, 0)
	for _, skillsToRemove := range skillsToRemoves {
		var sdkSkillsToRemove platformclientv2.Skillstoremove
		skillsToRemovesMap, ok := skillsToRemove.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkSkillsToRemove.Name, skillsToRemovesMap, "name")

		skillsToRemovesSlice = append(skillsToRemovesSlice, sdkSkillsToRemove)
	}

	return &skillsToRemovesSlice
}

// buildActionss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Actions
func buildActionss(actionss []interface{}) *[]platformclientv2.Actions {
	actionssSlice := make([]platformclientv2.Actions, 0)
	for _, actions := range actionss {
		var sdkActions platformclientv2.Actions
		actionssMap, ok := actions.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkActions.SkillsToRemove, actionssMap, "skills_to_remove", buildSkillsToRemoves)

		actionssSlice = append(actionssSlice, sdkActions)
	}

	return &actionssSlice
}

// buildRings maps an []interface{} into a Genesys Cloud *[]platformclientv2.Ring
func buildRings(rings []interface{}) *[]platformclientv2.Ring {
	ringsSlice := make([]platformclientv2.Ring, 0)
	for _, ring := range rings {
		var sdkRing platformclientv2.Ring
		ringsMap, ok := ring.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkRing.ExpansionCriteria, ringsMap, "expansion_criteria", buildExpansionCriteriums)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkRing.Actions, ringsMap, "actions", buildActions)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkRing.MemberGroups, ringsMap, "member_groups", buildMemberGroups)

		ringsSlice = append(ringsSlice, sdkRing)
	}

	return &ringsSlice
}

// buildBullseyes maps an []interface{} into a Genesys Cloud *[]platformclientv2.Bullseye
func buildBullseyes(bullseyes []interface{}) *[]platformclientv2.Bullseye {
	bullseyesSlice := make([]platformclientv2.Bullseye, 0)
	for _, bullseye := range bullseyes {
		var sdkBullseye platformclientv2.Bullseye
		bullseyesMap, ok := bullseye.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkBullseye.Rings, bullseyesMap, "rings", buildRings)

		bullseyesSlice = append(bullseyesSlice, sdkBullseye)
	}

	return &bullseyesSlice
}

// buildAcwSettingss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Acwsettings
func buildAcwSettingss(acwSettingss []interface{}) *[]platformclientv2.Acwsettings {
	acwSettingssSlice := make([]platformclientv2.Acwsettings, 0)
	for _, acwSettings := range acwSettingss {
		var sdkAcwSettings platformclientv2.Acwsettings
		acwSettingssMap, ok := acwSettings.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkAcwSettings.WrapupPrompt, acwSettingssMap, "wrapup_prompt")
		sdkAcwSettings.TimeoutMs = platformclientv2.Int(acwSettingssMap["timeout_ms"].(int))

		acwSettingssSlice = append(acwSettingssSlice, sdkAcwSettings)
	}

	return &acwSettingssSlice
}

// buildAgentOwnedRoutings maps an []interface{} into a Genesys Cloud *[]platformclientv2.Agentownedrouting
func buildAgentOwnedRoutings(agentOwnedRoutings []interface{}) *[]platformclientv2.Agentownedrouting {
	agentOwnedRoutingsSlice := make([]platformclientv2.Agentownedrouting, 0)
	for _, agentOwnedRouting := range agentOwnedRoutings {
		var sdkAgentOwnedRouting platformclientv2.Agentownedrouting
		agentOwnedRoutingsMap, ok := agentOwnedRouting.(map[string]interface{})
		if !ok {
			continue
		}

		sdkAgentOwnedRouting.EnableAgentOwnedCallbacks = platformclientv2.Bool(agentOwnedRoutingsMap["enable_agent_owned_callbacks"].(bool))
		sdkAgentOwnedRouting.MaxOwnedCallbackHours = platformclientv2.Int(agentOwnedRoutingsMap["max_owned_callback_hours"].(int))
		sdkAgentOwnedRouting.MaxOwnedCallbackDelayHours = platformclientv2.Int(agentOwnedRoutingsMap["max_owned_callback_delay_hours"].(int))

		agentOwnedRoutingsSlice = append(agentOwnedRoutingsSlice, sdkAgentOwnedRouting)
	}

	return &agentOwnedRoutingsSlice
}

// buildDirectRoutingMediaSettingss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Directroutingmediasettings
func buildDirectRoutingMediaSettingss(directRoutingMediaSettingss []interface{}) *[]platformclientv2.Directroutingmediasettings {
	directRoutingMediaSettingssSlice := make([]platformclientv2.Directroutingmediasettings, 0)
	for _, directRoutingMediaSettings := range directRoutingMediaSettingss {
		var sdkDirectRoutingMediaSettings platformclientv2.Directroutingmediasettings
		directRoutingMediaSettingssMap, ok := directRoutingMediaSettings.(map[string]interface{})
		if !ok {
			continue
		}

		sdkDirectRoutingMediaSettings.UseAgentAddressOutbound = platformclientv2.Bool(directRoutingMediaSettingssMap["use_agent_address_outbound"].(bool))

		directRoutingMediaSettingssSlice = append(directRoutingMediaSettingssSlice, sdkDirectRoutingMediaSettings)
	}

	return &directRoutingMediaSettingssSlice
}

// buildDirectRoutings maps an []interface{} into a Genesys Cloud *[]platformclientv2.Directrouting
func buildDirectRoutings(directRoutings []interface{}) *[]platformclientv2.Directrouting {
	directRoutingsSlice := make([]platformclientv2.Directrouting, 0)
	for _, directRouting := range directRoutings {
		var sdkDirectRouting platformclientv2.Directrouting
		directRoutingsMap, ok := directRouting.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkDirectRouting.CallMediaSettings, directRoutingsMap, "call_media_settings", buildDirectRoutingMediaSettings)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkDirectRouting.EmailMediaSettings, directRoutingsMap, "email_media_settings", buildDirectRoutingMediaSettings)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkDirectRouting.MessageMediaSettings, directRoutingsMap, "message_media_settings", buildDirectRoutingMediaSettings)
		resourcedata.BuildSDKStringValueIfNotNil(&sdkDirectRouting.BackupQueueId, directRoutingsMap, "backup_queue_id")
		sdkDirectRouting.WaitForAgent = platformclientv2.Bool(directRoutingsMap["wait_for_agent"].(bool))
		sdkDirectRouting.AgentWaitSeconds = platformclientv2.Int(directRoutingsMap["agent_wait_seconds"].(int))

		directRoutingsSlice = append(directRoutingsSlice, sdkDirectRouting)
	}

	return &directRoutingsSlice
}

// buildQueueMessagingAddressess maps an []interface{} into a Genesys Cloud *[]platformclientv2.Queuemessagingaddresses
func buildQueueMessagingAddressess(queueMessagingAddressess []interface{}) *[]platformclientv2.Queuemessagingaddresses {
	queueMessagingAddressessSlice := make([]platformclientv2.Queuemessagingaddresses, 0)
	for _, queueMessagingAddresses := range queueMessagingAddressess {
		var sdkQueueMessagingAddresses platformclientv2.Queuemessagingaddresses
		queueMessagingAddressessMap, ok := queueMessagingAddresses.(map[string]interface{})
		if !ok {
			continue
		}

		sdkQueueMessagingAddresses.SmsAddressId = &platformclientv2.Domainentityref{Id: platformclientv2.String(queueMessagingAddressessMap["sms_address_id"].(string))}
		sdkQueueMessagingAddresses.OpenMessagingRecipientId = &platformclientv2.Domainentityref{Id: platformclientv2.String(queueMessagingAddressessMap["open_messaging_recipient_id"].(string))}
		sdkQueueMessagingAddresses.WhatsAppRecipientId = &platformclientv2.Domainentityref{Id: platformclientv2.String(queueMessagingAddressessMap["whats_app_recipient_id"].(string))}

		queueMessagingAddressessSlice = append(queueMessagingAddressessSlice, sdkQueueMessagingAddresses)
	}

	return &queueMessagingAddressessSlice
}

// buildQueueEmailAddresss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Queueemailaddress
func buildQueueEmailAddresss(queueEmailAddresss []interface{}) *[]platformclientv2.Queueemailaddress {
	queueEmailAddresssSlice := make([]platformclientv2.Queueemailaddress, 0)
	for _, queueEmailAddress := range queueEmailAddresss {
		var sdkQueueEmailAddress platformclientv2.Queueemailaddress
		queueEmailAddresssMap, ok := queueEmailAddress.(map[string]interface{})
		if !ok {
			continue
		}

		queueEmailAddresssSlice = append(queueEmailAddresssSlice, sdkQueueEmailAddress)
	}

	return &queueEmailAddresssSlice
}

// buildEmailAddresss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Emailaddress
func buildEmailAddresss(emailAddresss []interface{}) *[]platformclientv2.Emailaddress {
	emailAddresssSlice := make([]platformclientv2.Emailaddress, 0)
	for _, emailAddress := range emailAddresss {
		var sdkEmailAddress platformclientv2.Emailaddress
		emailAddresssMap, ok := emailAddress.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkEmailAddress.Email, emailAddresssMap, "email")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkEmailAddress.Name, emailAddresssMap, "name")

		emailAddresssSlice = append(emailAddresssSlice, sdkEmailAddress)
	}

	return &emailAddresssSlice
}

// buildSignatures maps an []interface{} into a Genesys Cloud *[]platformclientv2.Signature
func buildSignatures(signatures []interface{}) *[]platformclientv2.Signature {
	signaturesSlice := make([]platformclientv2.Signature, 0)
	for _, signature := range signatures {
		var sdkSignature platformclientv2.Signature
		signaturesMap, ok := signature.(map[string]interface{})
		if !ok {
			continue
		}

		sdkSignature.Enabled = platformclientv2.Bool(signaturesMap["enabled"].(bool))
		resourcedata.BuildSDKStringValueIfNotNil(&sdkSignature.CannedResponseId, signaturesMap, "canned_response_id")
		sdkSignature.AlwaysIncluded = platformclientv2.Bool(signaturesMap["always_included"].(bool))
		resourcedata.BuildSDKStringValueIfNotNil(&sdkSignature.InclusionType, signaturesMap, "inclusion_type")

		signaturesSlice = append(signaturesSlice, sdkSignature)
	}

	return &signaturesSlice
}

// buildInboundRoutes maps an []interface{} into a Genesys Cloud *[]platformclientv2.Inboundroute
func buildInboundRoutes(inboundRoutes []interface{}) *[]platformclientv2.Inboundroute {
	inboundRoutesSlice := make([]platformclientv2.Inboundroute, 0)
	for _, inboundRoute := range inboundRoutes {
		var sdkInboundRoute platformclientv2.Inboundroute
		inboundRoutesMap, ok := inboundRoute.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkInboundRoute.Name, inboundRoutesMap, "name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkInboundRoute.Pattern, inboundRoutesMap, "pattern")
		sdkInboundRoute.QueueId = &platformclientv2.Domainentityref{Id: platformclientv2.String(inboundRoutesMap["queue_id"].(string))}
		sdkInboundRoute.Priority = platformclientv2.Int(inboundRoutesMap["priority"].(int))
		// TODO: Handle skills property
		sdkInboundRoute.LanguageId = &platformclientv2.Domainentityref{Id: platformclientv2.String(inboundRoutesMap["language_id"].(string))}
		resourcedata.BuildSDKStringValueIfNotNil(&sdkInboundRoute.FromName, inboundRoutesMap, "from_name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkInboundRoute.FromEmail, inboundRoutesMap, "from_email")
		sdkInboundRoute.FlowId = &platformclientv2.Domainentityref{Id: platformclientv2.String(inboundRoutesMap["flow_id"].(string))}
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkInboundRoute.ReplyEmailAddress, inboundRoutesMap, "reply_email_address", buildQueueEmailAddress)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkInboundRoute.AutoBcc, inboundRoutesMap, "auto_bcc", buildEmailAddresss)
		sdkInboundRoute.SpamFlowId = &platformclientv2.Domainentityref{Id: platformclientv2.String(inboundRoutesMap["spam_flow_id"].(string))}
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkInboundRoute.Signature, inboundRoutesMap, "signature", buildSignature)
		resourcedata.BuildSDKStringValueIfNotNil(&sdkInboundRoute.HistoryInclusion, inboundRoutesMap, "history_inclusion")
		sdkInboundRoute.AllowMultipleActions = platformclientv2.Bool(inboundRoutesMap["allow_multiple_actions"].(bool))

		inboundRoutesSlice = append(inboundRoutesSlice, sdkInboundRoute)
	}

	return &inboundRoutesSlice
}

// buildQueueEmailAddresss maps an []interface{} into a Genesys Cloud *[]platformclientv2.Queueemailaddress
func buildQueueEmailAddresss(queueEmailAddresss []interface{}) *[]platformclientv2.Queueemailaddress {
	queueEmailAddresssSlice := make([]platformclientv2.Queueemailaddress, 0)
	for _, queueEmailAddress := range queueEmailAddresss {
		var sdkQueueEmailAddress platformclientv2.Queueemailaddress
		queueEmailAddresssMap, ok := queueEmailAddress.(map[string]interface{})
		if !ok {
			continue
		}

		sdkQueueEmailAddress.DomainId = &platformclientv2.Domainentityref{Id: platformclientv2.String(queueEmailAddresssMap["domain_id"].(string))}
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkQueueEmailAddress.Route, queueEmailAddresssMap, "route", buildInboundRoute)

		queueEmailAddresssSlice = append(queueEmailAddresssSlice, sdkQueueEmailAddress)
	}

	return &queueEmailAddresssSlice
}

// buildQueues maps an []interface{} into a Genesys Cloud *[]platformclientv2.Queue
func buildQueues(queues []interface{}) *[]platformclientv2.Queue {
	queuesSlice := make([]platformclientv2.Queue, 0)
	for _, queue := range queues {
		var sdkQueue platformclientv2.Queue
		queuesMap, ok := queue.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkQueue.Name, queuesMap, "name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkQueue.DivisionId, queuesMap, "division_id")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkQueue.Description, queuesMap, "description")
		sdkQueue.MemberCount = platformclientv2.Int(queuesMap["member_count"].(int))
		sdkQueue.UserMemberCount = platformclientv2.Int(queuesMap["user_member_count"].(int))
		sdkQueue.JoinedMemberCount = platformclientv2.Int(queuesMap["joined_member_count"].(int))
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkQueue.MediaSettings, queuesMap, "media_settings", buildQueueMediaSettings)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkQueue.RoutingRules, queuesMap, "routing_rules", buildRoutingRules)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkQueue.ConditionalGroupRouting, queuesMap, "conditional_group_routing", buildConditionalGroupRouting)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkQueue.Bullseye, queuesMap, "bullseye", buildBullseye)
		resourcedata.BuildSDKStringValueIfNotNil(&sdkQueue.ScoringMethod, queuesMap, "scoring_method")
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkQueue.AcwSettings, queuesMap, "acw_settings", buildAcwSettings)
		resourcedata.BuildSDKStringValueIfNotNil(&sdkQueue.SkillEvaluationMethod, queuesMap, "skill_evaluation_method")
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkQueue.MemberGroups, queuesMap, "member_groups", buildMemberGroups)
		sdkQueue.QueueFlowId = &platformclientv2.Domainentityref{Id: platformclientv2.String(queuesMap["queue_flow_id"].(string))}
		sdkQueue.EmailInQueueFlowId = &platformclientv2.Domainentityref{Id: platformclientv2.String(queuesMap["email_in_queue_flow_id"].(string))}
		sdkQueue.MessageInQueueFlowId = &platformclientv2.Domainentityref{Id: platformclientv2.String(queuesMap["message_in_queue_flow_id"].(string))}
		sdkQueue.WhisperPromptId = &platformclientv2.Domainentityref{Id: platformclientv2.String(queuesMap["whisper_prompt_id"].(string))}
		sdkQueue.OnHoldPromptId = &platformclientv2.Domainentityref{Id: platformclientv2.String(queuesMap["on_hold_prompt_id"].(string))}
		sdkQueue.AutoAnswerOnly = platformclientv2.Bool(queuesMap["auto_answer_only"].(bool))
		sdkQueue.EnableTranscription = platformclientv2.Bool(queuesMap["enable_transcription"].(bool))
		sdkQueue.EnableAudioMonitoring = platformclientv2.Bool(queuesMap["enable_audio_monitoring"].(bool))
		sdkQueue.EnableManualAssignment = platformclientv2.Bool(queuesMap["enable_manual_assignment"].(bool))
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkQueue.AgentOwnedRouting, queuesMap, "agent_owned_routing", buildAgentOwnedRouting)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkQueue.DirectRouting, queuesMap, "direct_routing", buildDirectRouting)
		resourcedata.BuildSDKStringValueIfNotNil(&sdkQueue.CallingPartyName, queuesMap, "calling_party_name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkQueue.CallingPartyNumber, queuesMap, "calling_party_number")
		// TODO: Handle default_scripts property
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkQueue.OutboundMessagingAddresses, queuesMap, "outbound_messaging_addresses", buildQueueMessagingAddresses)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkQueue.OutboundEmailAddress, queuesMap, "outbound_email_address", buildQueueEmailAddress)
		resourcedata.BuildSDKStringValueIfNotNil(&sdkQueue.PeerId, queuesMap, "peer_id")
		sdkQueue.SuppressInQueueCallRecording = platformclientv2.Bool(queuesMap["suppress_in_queue_call_recording"].(bool))

		queuesSlice = append(queuesSlice, sdkQueue)
	}

	return &queuesSlice
}

// buildDomainResourceConditionValues maps an []interface{} into a Genesys Cloud *[]platformclientv2.Domainresourceconditionvalue
func buildDomainResourceConditionValues(domainResourceConditionValues []interface{}) *[]platformclientv2.Domainresourceconditionvalue {
	domainResourceConditionValuesSlice := make([]platformclientv2.Domainresourceconditionvalue, 0)
	for _, domainResourceConditionValue := range domainResourceConditionValues {
		var sdkDomainResourceConditionValue platformclientv2.Domainresourceconditionvalue
		domainResourceConditionValuesMap, ok := domainResourceConditionValue.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkDomainResourceConditionValue.User, domainResourceConditionValuesMap, "user", buildUser)
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkDomainResourceConditionValue.Queue, domainResourceConditionValuesMap, "queue", buildQueue)
		resourcedata.BuildSDKStringValueIfNotNil(&sdkDomainResourceConditionValue.Value, domainResourceConditionValuesMap, "value")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkDomainResourceConditionValue.Type, domainResourceConditionValuesMap, "type")

		domainResourceConditionValuesSlice = append(domainResourceConditionValuesSlice, sdkDomainResourceConditionValue)
	}

	return &domainResourceConditionValuesSlice
}

// buildDomainResourceConditionNodes maps an []interface{} into a Genesys Cloud *[]platformclientv2.Domainresourceconditionnode
func buildDomainResourceConditionNodes(domainResourceConditionNodes []interface{}) *[]platformclientv2.Domainresourceconditionnode {
	domainResourceConditionNodesSlice := make([]platformclientv2.Domainresourceconditionnode, 0)
	for _, domainResourceConditionNode := range domainResourceConditionNodes {
		var sdkDomainResourceConditionNode platformclientv2.Domainresourceconditionnode
		domainResourceConditionNodesMap, ok := domainResourceConditionNode.(map[string]interface{})
		if !ok {
			continue
		}

		domainResourceConditionNodesSlice = append(domainResourceConditionNodesSlice, sdkDomainResourceConditionNode)
	}

	return &domainResourceConditionNodesSlice
}

// buildDomainResourceConditionNodes maps an []interface{} into a Genesys Cloud *[]platformclientv2.Domainresourceconditionnode
func buildDomainResourceConditionNodes(domainResourceConditionNodes []interface{}) *[]platformclientv2.Domainresourceconditionnode {
	domainResourceConditionNodesSlice := make([]platformclientv2.Domainresourceconditionnode, 0)
	for _, domainResourceConditionNode := range domainResourceConditionNodes {
		var sdkDomainResourceConditionNode platformclientv2.Domainresourceconditionnode
		domainResourceConditionNodesMap, ok := domainResourceConditionNode.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkDomainResourceConditionNode.VariableName, domainResourceConditionNodesMap, "variable_name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkDomainResourceConditionNode.Operator, domainResourceConditionNodesMap, "operator")
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkDomainResourceConditionNode.Operands, domainResourceConditionNodesMap, "operands", buildDomainResourceConditionValues)
		resourcedata.BuildSDKStringValueIfNotNil(&sdkDomainResourceConditionNode.Conjunction, domainResourceConditionNodesMap, "conjunction")
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkDomainResourceConditionNode.Terms, domainResourceConditionNodesMap, "terms", buildDomainResourceConditionNodes)

		domainResourceConditionNodesSlice = append(domainResourceConditionNodesSlice, sdkDomainResourceConditionNode)
	}

	return &domainResourceConditionNodesSlice
}

// buildDomainPermissionPolicys maps an []interface{} into a Genesys Cloud *[]platformclientv2.Domainpermissionpolicy
func buildDomainPermissionPolicys(domainPermissionPolicys []interface{}) *[]platformclientv2.Domainpermissionpolicy {
	domainPermissionPolicysSlice := make([]platformclientv2.Domainpermissionpolicy, 0)
	for _, domainPermissionPolicy := range domainPermissionPolicys {
		var sdkDomainPermissionPolicy platformclientv2.Domainpermissionpolicy
		domainPermissionPolicysMap, ok := domainPermissionPolicy.(map[string]interface{})
		if !ok {
			continue
		}

		resourcedata.BuildSDKStringValueIfNotNil(&sdkDomainPermissionPolicy.Domain, domainPermissionPolicysMap, "domain")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkDomainPermissionPolicy.EntityName, domainPermissionPolicysMap, "entity_name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkDomainPermissionPolicy.PolicyName, domainPermissionPolicysMap, "policy_name")
		resourcedata.BuildSDKStringValueIfNotNil(&sdkDomainPermissionPolicy.PolicyDescription, domainPermissionPolicysMap, "policy_description")
		resourcedata.BuildSDKStringArrayValueIfNotNil(&sdkDomainPermissionPolicy.ActionSet, domainPermissionPolicysMap, "action_set")
		resourcedata.BuildSDKStringArrayValueIfNotNil(&sdkDomainPermissionPolicy.NamedResources, domainPermissionPolicysMap, "named_resources")
		sdkDomainPermissionPolicy.AllowConditions = platformclientv2.Bool(domainPermissionPolicysMap["allow_conditions"].(bool))
		resourcedata.BuildSDKInterfaceArrayValueIfNotNil(&sdkDomainPermissionPolicy.ResourceConditionNode, domainPermissionPolicysMap, "resource_condition_node", buildDomainResourceConditionNode)

		domainPermissionPolicysSlice = append(domainPermissionPolicysSlice, sdkDomainPermissionPolicy)
	}

	return &domainPermissionPolicysSlice
}

// flattenChats maps a Genesys Cloud *[]platformclientv2.Chat into a []interface{}
func flattenChats(chats *[]platformclientv2.Chat) []interface{} {
	if len(*chats) == 0 {
		return nil
	}

	var chatList []interface{}
	for _, chat := range *chats {
		chatMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(chatMap, "jabber_id", chat.JabberId)

		chatList = append(chatList, chatMap)
	}

	return chatList
}

// flattenContacts maps a Genesys Cloud *[]platformclientv2.Contact into a []interface{}
func flattenContacts(contacts *[]platformclientv2.Contact) []interface{} {
	if len(*contacts) == 0 {
		return nil
	}

	var contactList []interface{}
	for _, contact := range *contacts {
		contactMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(contactMap, "address", contact.Address)
		resourcedata.SetMapValueIfNotNil(contactMap, "display", contact.Display)
		resourcedata.SetMapValueIfNotNil(contactMap, "media_type", contact.MediaType)
		resourcedata.SetMapValueIfNotNil(contactMap, "type", contact.Type)
		resourcedata.SetMapValueIfNotNil(contactMap, "extension", contact.Extension)
		resourcedata.SetMapValueIfNotNil(contactMap, "country_code", contact.CountryCode)
		resourcedata.SetMapValueIfNotNil(contactMap, "integration", contact.Integration)

		contactList = append(contactList, contactMap)
	}

	return contactList
}

// flattenUsers maps a Genesys Cloud *[]platformclientv2.User into a []interface{}
func flattenUsers(users *[]platformclientv2.User) []interface{} {
	if len(*users) == 0 {
		return nil
	}

	var userList []interface{}
	for _, user := range *users {
		userMap := make(map[string]interface{})

		userList = append(userList, userMap)
	}

	return userList
}

// flattenUserImages maps a Genesys Cloud *[]platformclientv2.Userimage into a []interface{}
func flattenUserImages(userImages *[]platformclientv2.Userimage) []interface{} {
	if len(*userImages) == 0 {
		return nil
	}

	var userImageList []interface{}
	for _, userImage := range *userImages {
		userImageMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(userImageMap, "resolution", userImage.Resolution)
		resourcedata.SetMapValueIfNotNil(userImageMap, "image_uri", userImage.ImageUri)

		userImageList = append(userImageList, userImageMap)
	}

	return userImageList
}

// flattenEducations maps a Genesys Cloud *[]platformclientv2.Education into a []interface{}
func flattenEducations(educations *[]platformclientv2.Education) []interface{} {
	if len(*educations) == 0 {
		return nil
	}

	var educationList []interface{}
	for _, education := range *educations {
		educationMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(educationMap, "school", education.School)
		resourcedata.SetMapValueIfNotNil(educationMap, "field_of_study", education.FieldOfStudy)
		resourcedata.SetMapValueIfNotNil(educationMap, "notes", education.Notes)
		resourcedata.SetMapValueIfNotNil(educationMap, "date_start", education.DateStart)
		resourcedata.SetMapValueIfNotNil(educationMap, "date_end", education.DateEnd)

		educationList = append(educationList, educationMap)
	}

	return educationList
}

// flattenBiographys maps a Genesys Cloud *[]platformclientv2.Biography into a []interface{}
func flattenBiographys(biographys *[]platformclientv2.Biography) []interface{} {
	if len(*biographys) == 0 {
		return nil
	}

	var biographyList []interface{}
	for _, biography := range *biographys {
		biographyMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(biographyMap, "biography", biography.Biography)
		resourcedata.SetMapStringArrayValueIfNotNil(biographyMap, "interests", biography.Interests)
		resourcedata.SetMapStringArrayValueIfNotNil(biographyMap, "hobbies", biography.Hobbies)
		resourcedata.SetMapValueIfNotNil(biographyMap, "spouse", biography.Spouse)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(biographyMap, "education", biography.Education, flattenEducations)

		biographyList = append(biographyList, biographyMap)
	}

	return biographyList
}

// flattenEmployerInfos maps a Genesys Cloud *[]platformclientv2.Employerinfo into a []interface{}
func flattenEmployerInfos(employerInfos *[]platformclientv2.Employerinfo) []interface{} {
	if len(*employerInfos) == 0 {
		return nil
	}

	var employerInfoList []interface{}
	for _, employerInfo := range *employerInfos {
		employerInfoMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(employerInfoMap, "official_name", employerInfo.OfficialName)
		resourcedata.SetMapValueIfNotNil(employerInfoMap, "employee_id", employerInfo.EmployeeId)
		resourcedata.SetMapValueIfNotNil(employerInfoMap, "employee_type", employerInfo.EmployeeType)
		resourcedata.SetMapValueIfNotNil(employerInfoMap, "date_hire", employerInfo.DateHire)

		employerInfoList = append(employerInfoList, employerInfoMap)
	}

	return employerInfoList
}

// flattenRoutingStatuss maps a Genesys Cloud *[]platformclientv2.Routingstatus into a []interface{}
func flattenRoutingStatuss(routingStatuss *[]platformclientv2.Routingstatus) []interface{} {
	if len(*routingStatuss) == 0 {
		return nil
	}

	var routingStatusList []interface{}
	for _, routingStatus := range *routingStatuss {
		routingStatusMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(routingStatusMap, "user_id", routingStatus.UserId)
		resourcedata.SetMapValueIfNotNil(routingStatusMap, "status", routingStatus.Status)
		resourcedata.SetMapValueIfNotNil(routingStatusMap, "start_time", routingStatus.StartTime)

		routingStatusList = append(routingStatusList, routingStatusMap)
	}

	return routingStatusList
}

// flattenPresenceDefinitions maps a Genesys Cloud *[]platformclientv2.Presencedefinition into a []interface{}
func flattenPresenceDefinitions(presenceDefinitions *[]platformclientv2.Presencedefinition) []interface{} {
	if len(*presenceDefinitions) == 0 {
		return nil
	}

	var presenceDefinitionList []interface{}
	for _, presenceDefinition := range *presenceDefinitions {
		presenceDefinitionMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(presenceDefinitionMap, "system_presence", presenceDefinition.SystemPresence)

		presenceDefinitionList = append(presenceDefinitionList, presenceDefinitionMap)
	}

	return presenceDefinitionList
}

// flattenUserPresences maps a Genesys Cloud *[]platformclientv2.Userpresence into a []interface{}
func flattenUserPresences(userPresences *[]platformclientv2.Userpresence) []interface{} {
	if len(*userPresences) == 0 {
		return nil
	}

	var userPresenceList []interface{}
	for _, userPresence := range *userPresences {
		userPresenceMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(userPresenceMap, "name", userPresence.Name)
		resourcedata.SetMapValueIfNotNil(userPresenceMap, "source", userPresence.Source)
		resourcedata.SetMapValueIfNotNil(userPresenceMap, "source_id", userPresence.SourceId)
		resourcedata.SetMapValueIfNotNil(userPresenceMap, "primary", userPresence.Primary)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userPresenceMap, "presence_definition", userPresence.PresenceDefinition, flattenPresenceDefinition)
		resourcedata.SetMapValueIfNotNil(userPresenceMap, "message", userPresence.Message)
		resourcedata.SetMapValueIfNotNil(userPresenceMap, "modified_date", userPresence.ModifiedDate)

		userPresenceList = append(userPresenceList, userPresenceMap)
	}

	return userPresenceList
}

// flattenMediaSummaryDetails maps a Genesys Cloud *[]platformclientv2.Mediasummarydetail into a []interface{}
func flattenMediaSummaryDetails(mediaSummaryDetails *[]platformclientv2.Mediasummarydetail) []interface{} {
	if len(*mediaSummaryDetails) == 0 {
		return nil
	}

	var mediaSummaryDetailList []interface{}
	for _, mediaSummaryDetail := range *mediaSummaryDetails {
		mediaSummaryDetailMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(mediaSummaryDetailMap, "active", mediaSummaryDetail.Active)
		resourcedata.SetMapValueIfNotNil(mediaSummaryDetailMap, "acw", mediaSummaryDetail.Acw)

		mediaSummaryDetailList = append(mediaSummaryDetailList, mediaSummaryDetailMap)
	}

	return mediaSummaryDetailList
}

// flattenMediaSummarys maps a Genesys Cloud *[]platformclientv2.Mediasummary into a []interface{}
func flattenMediaSummarys(mediaSummarys *[]platformclientv2.Mediasummary) []interface{} {
	if len(*mediaSummarys) == 0 {
		return nil
	}

	var mediaSummaryList []interface{}
	for _, mediaSummary := range *mediaSummarys {
		mediaSummaryMap := make(map[string]interface{})

		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(mediaSummaryMap, "contact_center", mediaSummary.ContactCenter, flattenMediaSummaryDetail)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(mediaSummaryMap, "enterprise", mediaSummary.Enterprise, flattenMediaSummaryDetail)

		mediaSummaryList = append(mediaSummaryList, mediaSummaryMap)
	}

	return mediaSummaryList
}

// flattenUserConversationSummarys maps a Genesys Cloud *[]platformclientv2.Userconversationsummary into a []interface{}
func flattenUserConversationSummarys(userConversationSummarys *[]platformclientv2.Userconversationsummary) []interface{} {
	if len(*userConversationSummarys) == 0 {
		return nil
	}

	var userConversationSummaryList []interface{}
	for _, userConversationSummary := range *userConversationSummarys {
		userConversationSummaryMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(userConversationSummaryMap, "user_id", userConversationSummary.UserId)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userConversationSummaryMap, "call", userConversationSummary.Call, flattenMediaSummary)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userConversationSummaryMap, "callback", userConversationSummary.Callback, flattenMediaSummary)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userConversationSummaryMap, "email", userConversationSummary.Email, flattenMediaSummary)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userConversationSummaryMap, "message", userConversationSummary.Message, flattenMediaSummary)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userConversationSummaryMap, "chat", userConversationSummary.Chat, flattenMediaSummary)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userConversationSummaryMap, "social_expression", userConversationSummary.SocialExpression, flattenMediaSummary)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userConversationSummaryMap, "video", userConversationSummary.Video, flattenMediaSummary)

		userConversationSummaryList = append(userConversationSummaryList, userConversationSummaryMap)
	}

	return userConversationSummaryList
}

// flattenOutOfOffices maps a Genesys Cloud *[]platformclientv2.Outofoffice into a []interface{}
func flattenOutOfOffices(outOfOffices *[]platformclientv2.Outofoffice) []interface{} {
	if len(*outOfOffices) == 0 {
		return nil
	}

	var outOfOfficeList []interface{}
	for _, outOfOffice := range *outOfOffices {
		outOfOfficeMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(outOfOfficeMap, "name", outOfOffice.Name)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(outOfOfficeMap, "user", outOfOffice.User, flattenUser)
		resourcedata.SetMapValueIfNotNil(outOfOfficeMap, "start_date", outOfOffice.StartDate)
		resourcedata.SetMapValueIfNotNil(outOfOfficeMap, "end_date", outOfOffice.EndDate)
		resourcedata.SetMapValueIfNotNil(outOfOfficeMap, "active", outOfOffice.Active)
		resourcedata.SetMapValueIfNotNil(outOfOfficeMap, "indefinite", outOfOffice.Indefinite)

		outOfOfficeList = append(outOfOfficeList, outOfOfficeMap)
	}

	return outOfOfficeList
}

// flattenAddressableEntityRefs maps a Genesys Cloud *[]platformclientv2.Addressableentityref into a []interface{}
func flattenAddressableEntityRefs(addressableEntityRefs *[]platformclientv2.Addressableentityref) []interface{} {
	if len(*addressableEntityRefs) == 0 {
		return nil
	}

	var addressableEntityRefList []interface{}
	for _, addressableEntityRef := range *addressableEntityRefs {
		addressableEntityRefMap := make(map[string]interface{})

		addressableEntityRefList = append(addressableEntityRefList, addressableEntityRefMap)
	}

	return addressableEntityRefList
}

// flattenLocationEmergencyNumbers maps a Genesys Cloud *[]platformclientv2.Locationemergencynumber into a []interface{}
func flattenLocationEmergencyNumbers(locationEmergencyNumbers *[]platformclientv2.Locationemergencynumber) []interface{} {
	if len(*locationEmergencyNumbers) == 0 {
		return nil
	}

	var locationEmergencyNumberList []interface{}
	for _, locationEmergencyNumber := range *locationEmergencyNumbers {
		locationEmergencyNumberMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(locationEmergencyNumberMap, "e164", locationEmergencyNumber.E164)
		resourcedata.SetMapValueIfNotNil(locationEmergencyNumberMap, "number", locationEmergencyNumber.Number)
		resourcedata.SetMapValueIfNotNil(locationEmergencyNumberMap, "type", locationEmergencyNumber.Type)

		locationEmergencyNumberList = append(locationEmergencyNumberList, locationEmergencyNumberMap)
	}

	return locationEmergencyNumberList
}

// flattenLocationAddresss maps a Genesys Cloud *[]platformclientv2.Locationaddress into a []interface{}
func flattenLocationAddresss(locationAddresss *[]platformclientv2.Locationaddress) []interface{} {
	if len(*locationAddresss) == 0 {
		return nil
	}

	var locationAddressList []interface{}
	for _, locationAddress := range *locationAddresss {
		locationAddressMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(locationAddressMap, "city", locationAddress.City)
		resourcedata.SetMapValueIfNotNil(locationAddressMap, "country", locationAddress.Country)
		resourcedata.SetMapValueIfNotNil(locationAddressMap, "country_name", locationAddress.CountryName)
		resourcedata.SetMapValueIfNotNil(locationAddressMap, "state", locationAddress.State)
		resourcedata.SetMapValueIfNotNil(locationAddressMap, "street1", locationAddress.Street1)
		resourcedata.SetMapValueIfNotNil(locationAddressMap, "street2", locationAddress.Street2)
		resourcedata.SetMapValueIfNotNil(locationAddressMap, "zipcode", locationAddress.Zipcode)

		locationAddressList = append(locationAddressList, locationAddressMap)
	}

	return locationAddressList
}

// flattenLocationImages maps a Genesys Cloud *[]platformclientv2.Locationimage into a []interface{}
func flattenLocationImages(locationImages *[]platformclientv2.Locationimage) []interface{} {
	if len(*locationImages) == 0 {
		return nil
	}

	var locationImageList []interface{}
	for _, locationImage := range *locationImages {
		locationImageMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(locationImageMap, "resolution", locationImage.Resolution)
		resourcedata.SetMapValueIfNotNil(locationImageMap, "image_uri", locationImage.ImageUri)

		locationImageList = append(locationImageList, locationImageMap)
	}

	return locationImageList
}

// flattenLocationAddressVerificationDetailss maps a Genesys Cloud *[]platformclientv2.Locationaddressverificationdetails into a []interface{}
func flattenLocationAddressVerificationDetailss(locationAddressVerificationDetailss *[]platformclientv2.Locationaddressverificationdetails) []interface{} {
	if len(*locationAddressVerificationDetailss) == 0 {
		return nil
	}

	var locationAddressVerificationDetailsList []interface{}
	for _, locationAddressVerificationDetails := range *locationAddressVerificationDetailss {
		locationAddressVerificationDetailsMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(locationAddressVerificationDetailsMap, "status", locationAddressVerificationDetails.Status)
		resourcedata.SetMapValueIfNotNil(locationAddressVerificationDetailsMap, "date_finished", locationAddressVerificationDetails.DateFinished)
		resourcedata.SetMapValueIfNotNil(locationAddressVerificationDetailsMap, "date_started", locationAddressVerificationDetails.DateStarted)
		resourcedata.SetMapValueIfNotNil(locationAddressVerificationDetailsMap, "service", locationAddressVerificationDetails.Service)

		locationAddressVerificationDetailsList = append(locationAddressVerificationDetailsList, locationAddressVerificationDetailsMap)
	}

	return locationAddressVerificationDetailsList
}

// flattenLocationDefinitions maps a Genesys Cloud *[]platformclientv2.Locationdefinition into a []interface{}
func flattenLocationDefinitions(locationDefinitions *[]platformclientv2.Locationdefinition) []interface{} {
	if len(*locationDefinitions) == 0 {
		return nil
	}

	var locationDefinitionList []interface{}
	for _, locationDefinition := range *locationDefinitions {
		locationDefinitionMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(locationDefinitionMap, "name", locationDefinition.Name)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(locationDefinitionMap, "contact_user", locationDefinition.ContactUser, flattenAddressableEntityRef)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(locationDefinitionMap, "emergency_number", locationDefinition.EmergencyNumber, flattenLocationEmergencyNumber)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(locationDefinitionMap, "address", locationDefinition.Address, flattenLocationAddress)
		resourcedata.SetMapValueIfNotNil(locationDefinitionMap, "state", locationDefinition.State)
		resourcedata.SetMapValueIfNotNil(locationDefinitionMap, "notes", locationDefinition.Notes)
		resourcedata.SetMapStringArrayValueIfNotNil(locationDefinitionMap, "path", locationDefinition.Path)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(locationDefinitionMap, "profile_image", locationDefinition.ProfileImage, flattenLocationImages)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(locationDefinitionMap, "floorplan_image", locationDefinition.FloorplanImage, flattenLocationImages)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(locationDefinitionMap, "address_verification_details", locationDefinition.AddressVerificationDetails, flattenLocationAddressVerificationDetails)
		resourcedata.SetMapValueIfNotNil(locationDefinitionMap, "address_verified", locationDefinition.AddressVerified)
		resourcedata.SetMapValueIfNotNil(locationDefinitionMap, "address_stored", locationDefinition.AddressStored)
		resourcedata.SetMapValueIfNotNil(locationDefinitionMap, "images", locationDefinition.Images)

		locationDefinitionList = append(locationDefinitionList, locationDefinitionMap)
	}

	return locationDefinitionList
}

// flattenGeolocations maps a Genesys Cloud *[]platformclientv2.Geolocation into a []interface{}
func flattenGeolocations(geolocations *[]platformclientv2.Geolocation) []interface{} {
	if len(*geolocations) == 0 {
		return nil
	}

	var geolocationList []interface{}
	for _, geolocation := range *geolocations {
		geolocationMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(geolocationMap, "name", geolocation.Name)
		resourcedata.SetMapValueIfNotNil(geolocationMap, "type", geolocation.Type)
		resourcedata.SetMapValueIfNotNil(geolocationMap, "primary", geolocation.Primary)
		resourcedata.SetMapValueIfNotNil(geolocationMap, "latitude", geolocation.Latitude)
		resourcedata.SetMapValueIfNotNil(geolocationMap, "longitude", geolocation.Longitude)
		resourcedata.SetMapValueIfNotNil(geolocationMap, "country", geolocation.Country)
		resourcedata.SetMapValueIfNotNil(geolocationMap, "region", geolocation.Region)
		resourcedata.SetMapValueIfNotNil(geolocationMap, "city", geolocation.City)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(geolocationMap, "locations", geolocation.Locations, flattenLocationDefinitions)

		geolocationList = append(geolocationList, geolocationMap)
	}

	return geolocationList
}

// flattenUserStations maps a Genesys Cloud *[]platformclientv2.Userstation into a []interface{}
func flattenUserStations(userStations *[]platformclientv2.Userstation) []interface{} {
	if len(*userStations) == 0 {
		return nil
	}

	var userStationList []interface{}
	for _, userStation := range *userStations {
		userStationMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(userStationMap, "name", userStation.Name)
		resourcedata.SetMapValueIfNotNil(userStationMap, "type", userStation.Type)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userStationMap, "associated_user", userStation.AssociatedUser, flattenUser)
		resourcedata.SetMapValueIfNotNil(userStationMap, "associated_date", userStation.AssociatedDate)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userStationMap, "default_user", userStation.DefaultUser, flattenUser)
		// TODO: Handle provider_info property
		resourcedata.SetMapValueIfNotNil(userStationMap, "web_rtc_call_appearances", userStation.WebRtcCallAppearances)

		userStationList = append(userStationList, userStationMap)
	}

	return userStationList
}

// flattenUserStationss maps a Genesys Cloud *[]platformclientv2.Userstations into a []interface{}
func flattenUserStationss(userStationss *[]platformclientv2.Userstations) []interface{} {
	if len(*userStationss) == 0 {
		return nil
	}

	var userStationsList []interface{}
	for _, userStations := range *userStationss {
		userStationsMap := make(map[string]interface{})

		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userStationsMap, "associated_station", userStations.AssociatedStation, flattenUserStation)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userStationsMap, "effective_station", userStations.EffectiveStation, flattenUserStation)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userStationsMap, "default_station", userStations.DefaultStation, flattenUserStation)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userStationsMap, "last_associated_station", userStations.LastAssociatedStation, flattenUserStation)

		userStationsList = append(userStationsList, userStationsMap)
	}

	return userStationsList
}

// flattenDomainRoles maps a Genesys Cloud *[]platformclientv2.Domainrole into a []interface{}
func flattenDomainRoles(domainRoles *[]platformclientv2.Domainrole) []interface{} {
	if len(*domainRoles) == 0 {
		return nil
	}

	var domainRoleList []interface{}
	for _, domainRole := range *domainRoles {
		domainRoleMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(domainRoleMap, "name", domainRole.Name)

		domainRoleList = append(domainRoleList, domainRoleMap)
	}

	return domainRoleList
}

// flattenResourceConditionValues maps a Genesys Cloud *[]platformclientv2.Resourceconditionvalue into a []interface{}
func flattenResourceConditionValues(resourceConditionValues *[]platformclientv2.Resourceconditionvalue) []interface{} {
	if len(*resourceConditionValues) == 0 {
		return nil
	}

	var resourceConditionValueList []interface{}
	for _, resourceConditionValue := range *resourceConditionValues {
		resourceConditionValueMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(resourceConditionValueMap, "type", resourceConditionValue.Type)
		resourcedata.SetMapValueIfNotNil(resourceConditionValueMap, "value", resourceConditionValue.Value)

		resourceConditionValueList = append(resourceConditionValueList, resourceConditionValueMap)
	}

	return resourceConditionValueList
}

// flattenResourceConditionNodes maps a Genesys Cloud *[]platformclientv2.Resourceconditionnode into a []interface{}
func flattenResourceConditionNodes(resourceConditionNodes *[]platformclientv2.Resourceconditionnode) []interface{} {
	if len(*resourceConditionNodes) == 0 {
		return nil
	}

	var resourceConditionNodeList []interface{}
	for _, resourceConditionNode := range *resourceConditionNodes {
		resourceConditionNodeMap := make(map[string]interface{})

		resourceConditionNodeList = append(resourceConditionNodeList, resourceConditionNodeMap)
	}

	return resourceConditionNodeList
}

// flattenResourceConditionNodes maps a Genesys Cloud *[]platformclientv2.Resourceconditionnode into a []interface{}
func flattenResourceConditionNodes(resourceConditionNodes *[]platformclientv2.Resourceconditionnode) []interface{} {
	if len(*resourceConditionNodes) == 0 {
		return nil
	}

	var resourceConditionNodeList []interface{}
	for _, resourceConditionNode := range *resourceConditionNodes {
		resourceConditionNodeMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(resourceConditionNodeMap, "variable_name", resourceConditionNode.VariableName)
		resourcedata.SetMapValueIfNotNil(resourceConditionNodeMap, "conjunction", resourceConditionNode.Conjunction)
		resourcedata.SetMapValueIfNotNil(resourceConditionNodeMap, "operator", resourceConditionNode.Operator)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(resourceConditionNodeMap, "operands", resourceConditionNode.Operands, flattenResourceConditionValues)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(resourceConditionNodeMap, "terms", resourceConditionNode.Terms, flattenResourceConditionNodes)

		resourceConditionNodeList = append(resourceConditionNodeList, resourceConditionNodeMap)
	}

	return resourceConditionNodeList
}

// flattenResourcePermissionPolicys maps a Genesys Cloud *[]platformclientv2.Resourcepermissionpolicy into a []interface{}
func flattenResourcePermissionPolicys(resourcePermissionPolicys *[]platformclientv2.Resourcepermissionpolicy) []interface{} {
	if len(*resourcePermissionPolicys) == 0 {
		return nil
	}

	var resourcePermissionPolicyList []interface{}
	for _, resourcePermissionPolicy := range *resourcePermissionPolicys {
		resourcePermissionPolicyMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(resourcePermissionPolicyMap, "domain", resourcePermissionPolicy.Domain)
		resourcedata.SetMapValueIfNotNil(resourcePermissionPolicyMap, "entity_name", resourcePermissionPolicy.EntityName)
		resourcedata.SetMapValueIfNotNil(resourcePermissionPolicyMap, "policy_name", resourcePermissionPolicy.PolicyName)
		resourcedata.SetMapValueIfNotNil(resourcePermissionPolicyMap, "policy_description", resourcePermissionPolicy.PolicyDescription)
		resourcedata.SetMapValueIfNotNil(resourcePermissionPolicyMap, "action_set_key", resourcePermissionPolicy.ActionSetKey)
		resourcedata.SetMapValueIfNotNil(resourcePermissionPolicyMap, "allow_conditions", resourcePermissionPolicy.AllowConditions)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(resourcePermissionPolicyMap, "resource_condition_node", resourcePermissionPolicy.ResourceConditionNode, flattenResourceConditionNode)
		resourcedata.SetMapStringArrayValueIfNotNil(resourcePermissionPolicyMap, "named_resources", resourcePermissionPolicy.NamedResources)
		resourcedata.SetMapValueIfNotNil(resourcePermissionPolicyMap, "resource_condition", resourcePermissionPolicy.ResourceCondition)
		resourcedata.SetMapStringArrayValueIfNotNil(resourcePermissionPolicyMap, "action_set", resourcePermissionPolicy.ActionSet)

		resourcePermissionPolicyList = append(resourcePermissionPolicyList, resourcePermissionPolicyMap)
	}

	return resourcePermissionPolicyList
}

// flattenUserAuthorizations maps a Genesys Cloud *[]platformclientv2.Userauthorization into a []interface{}
func flattenUserAuthorizations(userAuthorizations *[]platformclientv2.Userauthorization) []interface{} {
	if len(*userAuthorizations) == 0 {
		return nil
	}

	var userAuthorizationList []interface{}
	for _, userAuthorization := range *userAuthorizations {
		userAuthorizationMap := make(map[string]interface{})

		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userAuthorizationMap, "roles", userAuthorization.Roles, flattenDomainRoles)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userAuthorizationMap, "unused_roles", userAuthorization.UnusedRoles, flattenDomainRoles)
		resourcedata.SetMapStringArrayValueIfNotNil(userAuthorizationMap, "permissions", userAuthorization.Permissions)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userAuthorizationMap, "permission_policies", userAuthorization.PermissionPolicies, flattenResourcePermissionPolicys)

		userAuthorizationList = append(userAuthorizationList, userAuthorizationMap)
	}

	return userAuthorizationList
}

// flattenLocations maps a Genesys Cloud *[]platformclientv2.Location into a []interface{}
func flattenLocations(locations *[]platformclientv2.Location) []interface{} {
	if len(*locations) == 0 {
		return nil
	}

	var locationList []interface{}
	for _, location := range *locations {
		locationMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(locationMap, "floorplan_id", location.FloorplanId)
		// TODO: Handle coordinates property
		resourcedata.SetMapValueIfNotNil(locationMap, "notes", location.Notes)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(locationMap, "location_definition", location.LocationDefinition, flattenLocationDefinition)

		locationList = append(locationList, locationMap)
	}

	return locationList
}

// flattenGroupContacts maps a Genesys Cloud *[]platformclientv2.Groupcontact into a []interface{}
func flattenGroupContacts(groupContacts *[]platformclientv2.Groupcontact) []interface{} {
	if len(*groupContacts) == 0 {
		return nil
	}

	var groupContactList []interface{}
	for _, groupContact := range *groupContacts {
		groupContactMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(groupContactMap, "address", groupContact.Address)
		resourcedata.SetMapValueIfNotNil(groupContactMap, "extension", groupContact.Extension)
		resourcedata.SetMapValueIfNotNil(groupContactMap, "display", groupContact.Display)
		resourcedata.SetMapValueIfNotNil(groupContactMap, "type", groupContact.Type)
		resourcedata.SetMapValueIfNotNil(groupContactMap, "media_type", groupContact.MediaType)

		groupContactList = append(groupContactList, groupContactMap)
	}

	return groupContactList
}

// flattenGroups maps a Genesys Cloud *[]platformclientv2.Group into a []interface{}
func flattenGroups(groups *[]platformclientv2.Group) []interface{} {
	if len(*groups) == 0 {
		return nil
	}

	var groupList []interface{}
	for _, group := range *groups {
		groupMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(groupMap, "name", group.Name)
		resourcedata.SetMapValueIfNotNil(groupMap, "description", group.Description)
		resourcedata.SetMapValueIfNotNil(groupMap, "member_count", group.MemberCount)
		resourcedata.SetMapValueIfNotNil(groupMap, "state", group.State)
		resourcedata.SetMapValueIfNotNil(groupMap, "type", group.Type)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(groupMap, "images", group.Images, flattenUserImages)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(groupMap, "addresses", group.Addresses, flattenGroupContacts)
		resourcedata.SetMapValueIfNotNil(groupMap, "rules_visible", group.RulesVisible)
		resourcedata.SetMapValueIfNotNil(groupMap, "visibility", group.Visibility)
		resourcedata.SetMapValueIfNotNil(groupMap, "roles_enabled", group.RolesEnabled)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(groupMap, "owners", group.Owners, flattenUsers)

		groupList = append(groupList, groupMap)
	}

	return groupList
}

// flattenTeams maps a Genesys Cloud *[]platformclientv2.Team into a []interface{}
func flattenTeams(teams *[]platformclientv2.Team) []interface{} {
	if len(*teams) == 0 {
		return nil
	}

	var teamList []interface{}
	for _, team := range *teams {
		teamMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(teamMap, "name", team.Name)
		resourcedata.SetMapValueIfNotNil(teamMap, "division_id", team.DivisionId)
		resourcedata.SetMapValueIfNotNil(teamMap, "description", team.Description)
		resourcedata.SetMapValueIfNotNil(teamMap, "member_count", team.MemberCount)

		teamList = append(teamList, teamMap)
	}

	return teamList
}

// flattenUserRoutingSkills maps a Genesys Cloud *[]platformclientv2.Userroutingskill into a []interface{}
func flattenUserRoutingSkills(userRoutingSkills *[]platformclientv2.Userroutingskill) []interface{} {
	if len(*userRoutingSkills) == 0 {
		return nil
	}

	var userRoutingSkillList []interface{}
	for _, userRoutingSkill := range *userRoutingSkills {
		userRoutingSkillMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(userRoutingSkillMap, "name", userRoutingSkill.Name)
		resourcedata.SetMapValueIfNotNil(userRoutingSkillMap, "proficiency", userRoutingSkill.Proficiency)
		resourcedata.SetMapValueIfNotNil(userRoutingSkillMap, "state", userRoutingSkill.State)
		resourcedata.SetMapValueIfNotNil(userRoutingSkillMap, "skill_uri", userRoutingSkill.SkillUri)

		userRoutingSkillList = append(userRoutingSkillList, userRoutingSkillMap)
	}

	return userRoutingSkillList
}

// flattenUserRoutingLanguages maps a Genesys Cloud *[]platformclientv2.Userroutinglanguage into a []interface{}
func flattenUserRoutingLanguages(userRoutingLanguages *[]platformclientv2.Userroutinglanguage) []interface{} {
	if len(*userRoutingLanguages) == 0 {
		return nil
	}

	var userRoutingLanguageList []interface{}
	for _, userRoutingLanguage := range *userRoutingLanguages {
		userRoutingLanguageMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(userRoutingLanguageMap, "name", userRoutingLanguage.Name)
		resourcedata.SetMapValueIfNotNil(userRoutingLanguageMap, "proficiency", userRoutingLanguage.Proficiency)
		resourcedata.SetMapValueIfNotNil(userRoutingLanguageMap, "state", userRoutingLanguage.State)
		resourcedata.SetMapValueIfNotNil(userRoutingLanguageMap, "language_uri", userRoutingLanguage.LanguageUri)

		userRoutingLanguageList = append(userRoutingLanguageList, userRoutingLanguageMap)
	}

	return userRoutingLanguageList
}

// flattenOAuthLastTokenIssueds maps a Genesys Cloud *[]platformclientv2.Oauthlasttokenissued into a []interface{}
func flattenOAuthLastTokenIssueds(oAuthLastTokenIssueds *[]platformclientv2.Oauthlasttokenissued) []interface{} {
	if len(*oAuthLastTokenIssueds) == 0 {
		return nil
	}

	var oAuthLastTokenIssuedList []interface{}
	for _, oAuthLastTokenIssued := range *oAuthLastTokenIssueds {
		oAuthLastTokenIssuedMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(oAuthLastTokenIssuedMap, "date_issued", oAuthLastTokenIssued.DateIssued)

		oAuthLastTokenIssuedList = append(oAuthLastTokenIssuedList, oAuthLastTokenIssuedMap)
	}

	return oAuthLastTokenIssuedList
}

// flattenUsers maps a Genesys Cloud *[]platformclientv2.User into a []interface{}
func flattenUsers(users *[]platformclientv2.User) []interface{} {
	if len(*users) == 0 {
		return nil
	}

	var userList []interface{}
	for _, user := range *users {
		userMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(userMap, "name", user.Name)
		resourcedata.SetMapValueIfNotNil(userMap, "division_id", user.DivisionId)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userMap, "chat", user.Chat, flattenChat)
		resourcedata.SetMapValueIfNotNil(userMap, "department", user.Department)
		resourcedata.SetMapValueIfNotNil(userMap, "email", user.Email)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userMap, "primary_contact_info", user.PrimaryContactInfo, flattenContacts)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userMap, "addresses", user.Addresses, flattenContacts)
		resourcedata.SetMapValueIfNotNil(userMap, "state", user.State)
		resourcedata.SetMapValueIfNotNil(userMap, "title", user.Title)
		resourcedata.SetMapValueIfNotNil(userMap, "username", user.Username)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userMap, "manager", user.Manager, flattenUser)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userMap, "images", user.Images, flattenUserImages)
		resourcedata.SetMapStringArrayValueIfNotNil(userMap, "certifications", user.Certifications)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userMap, "biography", user.Biography, flattenBiography)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userMap, "employer_info", user.EmployerInfo, flattenEmployerInfo)
		resourcedata.SetMapValueIfNotNil(userMap, "preferred_name", user.PreferredName)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userMap, "routing_status", user.RoutingStatus, flattenRoutingStatus)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userMap, "presence", user.Presence, flattenUserPresence)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userMap, "integration_presence", user.IntegrationPresence, flattenUserPresence)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userMap, "conversation_summary", user.ConversationSummary, flattenUserConversationSummary)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userMap, "out_of_office", user.OutOfOffice, flattenOutOfOffice)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userMap, "geolocation", user.Geolocation, flattenGeolocation)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userMap, "station", user.Station, flattenUserStations)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userMap, "authorization", user.Authorization, flattenUserAuthorization)
		resourcedata.SetMapStringArrayValueIfNotNil(userMap, "profile_skills", user.ProfileSkills)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userMap, "locations", user.Locations, flattenLocations)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userMap, "groups", user.Groups, flattenGroups)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userMap, "team", user.Team, flattenTeam)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userMap, "skills", user.Skills, flattenUserRoutingSkills)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userMap, "languages", user.Languages, flattenUserRoutingLanguages)
		resourcedata.SetMapValueIfNotNil(userMap, "acd_auto_answer", user.AcdAutoAnswer)
		resourcedata.SetMapValueIfNotNil(userMap, "language_preference", user.LanguagePreference)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(userMap, "last_token_issued", user.LastTokenIssued, flattenOAuthLastTokenIssued)
		resourcedata.SetMapValueIfNotNil(userMap, "date_last_login", user.DateLastLogin)

		userList = append(userList, userMap)
	}

	return userList
}

// flattenServiceLevels maps a Genesys Cloud *[]platformclientv2.Servicelevel into a []interface{}
func flattenServiceLevels(serviceLevels *[]platformclientv2.Servicelevel) []interface{} {
	if len(*serviceLevels) == 0 {
		return nil
	}

	var serviceLevelList []interface{}
	for _, serviceLevel := range *serviceLevels {
		serviceLevelMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(serviceLevelMap, "percentage", serviceLevel.Percentage)
		resourcedata.SetMapValueIfNotNil(serviceLevelMap, "duration_ms", serviceLevel.DurationMs)

		serviceLevelList = append(serviceLevelList, serviceLevelMap)
	}

	return serviceLevelList
}

// flattenMediaSettingss maps a Genesys Cloud *[]platformclientv2.Mediasettings into a []interface{}
func flattenMediaSettingss(mediaSettingss *[]platformclientv2.Mediasettings) []interface{} {
	if len(*mediaSettingss) == 0 {
		return nil
	}

	var mediaSettingsList []interface{}
	for _, mediaSettings := range *mediaSettingss {
		mediaSettingsMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(mediaSettingsMap, "enable_auto_answer", mediaSettings.EnableAutoAnswer)
		resourcedata.SetMapValueIfNotNil(mediaSettingsMap, "alerting_timeout_seconds", mediaSettings.AlertingTimeoutSeconds)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(mediaSettingsMap, "service_level", mediaSettings.ServiceLevel, flattenServiceLevel)
		resourcedata.SetMapValueIfNotNil(mediaSettingsMap, "auto_answer_alert_tone_seconds", mediaSettings.AutoAnswerAlertToneSeconds)
		resourcedata.SetMapValueIfNotNil(mediaSettingsMap, "manual_answer_alert_tone_seconds", mediaSettings.ManualAnswerAlertToneSeconds)
		// TODO: Handle sub_type_settings property

		mediaSettingsList = append(mediaSettingsList, mediaSettingsMap)
	}

	return mediaSettingsList
}

// flattenCallbackMediaSettingss maps a Genesys Cloud *[]platformclientv2.Callbackmediasettings into a []interface{}
func flattenCallbackMediaSettingss(callbackMediaSettingss *[]platformclientv2.Callbackmediasettings) []interface{} {
	if len(*callbackMediaSettingss) == 0 {
		return nil
	}

	var callbackMediaSettingsList []interface{}
	for _, callbackMediaSettings := range *callbackMediaSettingss {
		callbackMediaSettingsMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(callbackMediaSettingsMap, "enable_auto_answer", callbackMediaSettings.EnableAutoAnswer)
		resourcedata.SetMapValueIfNotNil(callbackMediaSettingsMap, "alerting_timeout_seconds", callbackMediaSettings.AlertingTimeoutSeconds)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(callbackMediaSettingsMap, "service_level", callbackMediaSettings.ServiceLevel, flattenServiceLevel)
		resourcedata.SetMapValueIfNotNil(callbackMediaSettingsMap, "auto_answer_alert_tone_seconds", callbackMediaSettings.AutoAnswerAlertToneSeconds)
		resourcedata.SetMapValueIfNotNil(callbackMediaSettingsMap, "manual_answer_alert_tone_seconds", callbackMediaSettings.ManualAnswerAlertToneSeconds)
		// TODO: Handle sub_type_settings property
		resourcedata.SetMapValueIfNotNil(callbackMediaSettingsMap, "enable_auto_dial_and_end", callbackMediaSettings.EnableAutoDialAndEnd)
		resourcedata.SetMapValueIfNotNil(callbackMediaSettingsMap, "auto_dial_delay_seconds", callbackMediaSettings.AutoDialDelaySeconds)
		resourcedata.SetMapValueIfNotNil(callbackMediaSettingsMap, "auto_end_delay_seconds", callbackMediaSettings.AutoEndDelaySeconds)

		callbackMediaSettingsList = append(callbackMediaSettingsList, callbackMediaSettingsMap)
	}

	return callbackMediaSettingsList
}

// flattenQueueMediaSettingss maps a Genesys Cloud *[]platformclientv2.Queuemediasettings into a []interface{}
func flattenQueueMediaSettingss(queueMediaSettingss *[]platformclientv2.Queuemediasettings) []interface{} {
	if len(*queueMediaSettingss) == 0 {
		return nil
	}

	var queueMediaSettingsList []interface{}
	for _, queueMediaSettings := range *queueMediaSettingss {
		queueMediaSettingsMap := make(map[string]interface{})

		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(queueMediaSettingsMap, "call", queueMediaSettings.Call, flattenMediaSettings)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(queueMediaSettingsMap, "callback", queueMediaSettings.Callback, flattenCallbackMediaSettings)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(queueMediaSettingsMap, "chat", queueMediaSettings.Chat, flattenMediaSettings)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(queueMediaSettingsMap, "email", queueMediaSettings.Email, flattenMediaSettings)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(queueMediaSettingsMap, "message", queueMediaSettings.Message, flattenMediaSettings)

		queueMediaSettingsList = append(queueMediaSettingsList, queueMediaSettingsMap)
	}

	return queueMediaSettingsList
}

// flattenRoutingRules maps a Genesys Cloud *[]platformclientv2.Routingrule into a []interface{}
func flattenRoutingRules(routingRules *[]platformclientv2.Routingrule) []interface{} {
	if len(*routingRules) == 0 {
		return nil
	}

	var routingRuleList []interface{}
	for _, routingRule := range *routingRules {
		routingRuleMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(routingRuleMap, "operator", routingRule.Operator)
		resourcedata.SetMapValueIfNotNil(routingRuleMap, "threshold", routingRule.Threshold)
		resourcedata.SetMapValueIfNotNil(routingRuleMap, "wait_seconds", routingRule.WaitSeconds)

		routingRuleList = append(routingRuleList, routingRuleMap)
	}

	return routingRuleList
}

// flattenMemberGroups maps a Genesys Cloud *[]platformclientv2.Membergroup into a []interface{}
func flattenMemberGroups(memberGroups *[]platformclientv2.Membergroup) []interface{} {
	if len(*memberGroups) == 0 {
		return nil
	}

	var memberGroupList []interface{}
	for _, memberGroup := range *memberGroups {
		memberGroupMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(memberGroupMap, "name", memberGroup.Name)
		resourcedata.SetMapValueIfNotNil(memberGroupMap, "division_id", memberGroup.DivisionId)
		resourcedata.SetMapValueIfNotNil(memberGroupMap, "type", memberGroup.Type)
		resourcedata.SetMapValueIfNotNil(memberGroupMap, "member_count", memberGroup.MemberCount)

		memberGroupList = append(memberGroupList, memberGroupMap)
	}

	return memberGroupList
}

// flattenConditionalGroupRoutingRules maps a Genesys Cloud *[]platformclientv2.Conditionalgrouproutingrule into a []interface{}
func flattenConditionalGroupRoutingRules(conditionalGroupRoutingRules *[]platformclientv2.Conditionalgrouproutingrule) []interface{} {
	if len(*conditionalGroupRoutingRules) == 0 {
		return nil
	}

	var conditionalGroupRoutingRuleList []interface{}
	for _, conditionalGroupRoutingRule := range *conditionalGroupRoutingRules {
		conditionalGroupRoutingRuleMap := make(map[string]interface{})

		resourcedata.SetMapReferenceValueIfNotNil(conditionalGroupRoutingRuleMap, "queue_id", conditionalGroupRoutingRule.QueueId)
		resourcedata.SetMapValueIfNotNil(conditionalGroupRoutingRuleMap, "metric", conditionalGroupRoutingRule.Metric)
		resourcedata.SetMapValueIfNotNil(conditionalGroupRoutingRuleMap, "operator", conditionalGroupRoutingRule.Operator)
		resourcedata.SetMapValueIfNotNil(conditionalGroupRoutingRuleMap, "condition_value", conditionalGroupRoutingRule.ConditionValue)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(conditionalGroupRoutingRuleMap, "groups", conditionalGroupRoutingRule.Groups, flattenMemberGroups)
		resourcedata.SetMapValueIfNotNil(conditionalGroupRoutingRuleMap, "wait_seconds", conditionalGroupRoutingRule.WaitSeconds)

		conditionalGroupRoutingRuleList = append(conditionalGroupRoutingRuleList, conditionalGroupRoutingRuleMap)
	}

	return conditionalGroupRoutingRuleList
}

// flattenConditionalGroupRoutings maps a Genesys Cloud *[]platformclientv2.Conditionalgrouprouting into a []interface{}
func flattenConditionalGroupRoutings(conditionalGroupRoutings *[]platformclientv2.Conditionalgrouprouting) []interface{} {
	if len(*conditionalGroupRoutings) == 0 {
		return nil
	}

	var conditionalGroupRoutingList []interface{}
	for _, conditionalGroupRouting := range *conditionalGroupRoutings {
		conditionalGroupRoutingMap := make(map[string]interface{})

		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(conditionalGroupRoutingMap, "rules", conditionalGroupRouting.Rules, flattenConditionalGroupRoutingRules)

		conditionalGroupRoutingList = append(conditionalGroupRoutingList, conditionalGroupRoutingMap)
	}

	return conditionalGroupRoutingList
}

// flattenExpansionCriteriums maps a Genesys Cloud *[]platformclientv2.Expansioncriterium into a []interface{}
func flattenExpansionCriteriums(expansionCriteriums *[]platformclientv2.Expansioncriterium) []interface{} {
	if len(*expansionCriteriums) == 0 {
		return nil
	}

	var expansionCriteriumList []interface{}
	for _, expansionCriterium := range *expansionCriteriums {
		expansionCriteriumMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(expansionCriteriumMap, "type", expansionCriterium.Type)
		resourcedata.SetMapValueIfNotNil(expansionCriteriumMap, "threshold", expansionCriterium.Threshold)

		expansionCriteriumList = append(expansionCriteriumList, expansionCriteriumMap)
	}

	return expansionCriteriumList
}

// flattenSkillsToRemoves maps a Genesys Cloud *[]platformclientv2.Skillstoremove into a []interface{}
func flattenSkillsToRemoves(skillsToRemoves *[]platformclientv2.Skillstoremove) []interface{} {
	if len(*skillsToRemoves) == 0 {
		return nil
	}

	var skillsToRemoveList []interface{}
	for _, skillsToRemove := range *skillsToRemoves {
		skillsToRemoveMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(skillsToRemoveMap, "name", skillsToRemove.Name)

		skillsToRemoveList = append(skillsToRemoveList, skillsToRemoveMap)
	}

	return skillsToRemoveList
}

// flattenActionss maps a Genesys Cloud *[]platformclientv2.Actions into a []interface{}
func flattenActionss(actionss *[]platformclientv2.Actions) []interface{} {
	if len(*actionss) == 0 {
		return nil
	}

	var actionsList []interface{}
	for _, actions := range *actionss {
		actionsMap := make(map[string]interface{})

		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(actionsMap, "skills_to_remove", actions.SkillsToRemove, flattenSkillsToRemoves)

		actionsList = append(actionsList, actionsMap)
	}

	return actionsList
}

// flattenRings maps a Genesys Cloud *[]platformclientv2.Ring into a []interface{}
func flattenRings(rings *[]platformclientv2.Ring) []interface{} {
	if len(*rings) == 0 {
		return nil
	}

	var ringList []interface{}
	for _, ring := range *rings {
		ringMap := make(map[string]interface{})

		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(ringMap, "expansion_criteria", ring.ExpansionCriteria, flattenExpansionCriteriums)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(ringMap, "actions", ring.Actions, flattenActions)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(ringMap, "member_groups", ring.MemberGroups, flattenMemberGroups)

		ringList = append(ringList, ringMap)
	}

	return ringList
}

// flattenBullseyes maps a Genesys Cloud *[]platformclientv2.Bullseye into a []interface{}
func flattenBullseyes(bullseyes *[]platformclientv2.Bullseye) []interface{} {
	if len(*bullseyes) == 0 {
		return nil
	}

	var bullseyeList []interface{}
	for _, bullseye := range *bullseyes {
		bullseyeMap := make(map[string]interface{})

		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(bullseyeMap, "rings", bullseye.Rings, flattenRings)

		bullseyeList = append(bullseyeList, bullseyeMap)
	}

	return bullseyeList
}

// flattenAcwSettingss maps a Genesys Cloud *[]platformclientv2.Acwsettings into a []interface{}
func flattenAcwSettingss(acwSettingss *[]platformclientv2.Acwsettings) []interface{} {
	if len(*acwSettingss) == 0 {
		return nil
	}

	var acwSettingsList []interface{}
	for _, acwSettings := range *acwSettingss {
		acwSettingsMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(acwSettingsMap, "wrapup_prompt", acwSettings.WrapupPrompt)
		resourcedata.SetMapValueIfNotNil(acwSettingsMap, "timeout_ms", acwSettings.TimeoutMs)

		acwSettingsList = append(acwSettingsList, acwSettingsMap)
	}

	return acwSettingsList
}

// flattenAgentOwnedRoutings maps a Genesys Cloud *[]platformclientv2.Agentownedrouting into a []interface{}
func flattenAgentOwnedRoutings(agentOwnedRoutings *[]platformclientv2.Agentownedrouting) []interface{} {
	if len(*agentOwnedRoutings) == 0 {
		return nil
	}

	var agentOwnedRoutingList []interface{}
	for _, agentOwnedRouting := range *agentOwnedRoutings {
		agentOwnedRoutingMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(agentOwnedRoutingMap, "enable_agent_owned_callbacks", agentOwnedRouting.EnableAgentOwnedCallbacks)
		resourcedata.SetMapValueIfNotNil(agentOwnedRoutingMap, "max_owned_callback_hours", agentOwnedRouting.MaxOwnedCallbackHours)
		resourcedata.SetMapValueIfNotNil(agentOwnedRoutingMap, "max_owned_callback_delay_hours", agentOwnedRouting.MaxOwnedCallbackDelayHours)

		agentOwnedRoutingList = append(agentOwnedRoutingList, agentOwnedRoutingMap)
	}

	return agentOwnedRoutingList
}

// flattenDirectRoutingMediaSettingss maps a Genesys Cloud *[]platformclientv2.Directroutingmediasettings into a []interface{}
func flattenDirectRoutingMediaSettingss(directRoutingMediaSettingss *[]platformclientv2.Directroutingmediasettings) []interface{} {
	if len(*directRoutingMediaSettingss) == 0 {
		return nil
	}

	var directRoutingMediaSettingsList []interface{}
	for _, directRoutingMediaSettings := range *directRoutingMediaSettingss {
		directRoutingMediaSettingsMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(directRoutingMediaSettingsMap, "use_agent_address_outbound", directRoutingMediaSettings.UseAgentAddressOutbound)

		directRoutingMediaSettingsList = append(directRoutingMediaSettingsList, directRoutingMediaSettingsMap)
	}

	return directRoutingMediaSettingsList
}

// flattenDirectRoutings maps a Genesys Cloud *[]platformclientv2.Directrouting into a []interface{}
func flattenDirectRoutings(directRoutings *[]platformclientv2.Directrouting) []interface{} {
	if len(*directRoutings) == 0 {
		return nil
	}

	var directRoutingList []interface{}
	for _, directRouting := range *directRoutings {
		directRoutingMap := make(map[string]interface{})

		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(directRoutingMap, "call_media_settings", directRouting.CallMediaSettings, flattenDirectRoutingMediaSettings)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(directRoutingMap, "email_media_settings", directRouting.EmailMediaSettings, flattenDirectRoutingMediaSettings)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(directRoutingMap, "message_media_settings", directRouting.MessageMediaSettings, flattenDirectRoutingMediaSettings)
		resourcedata.SetMapValueIfNotNil(directRoutingMap, "backup_queue_id", directRouting.BackupQueueId)
		resourcedata.SetMapValueIfNotNil(directRoutingMap, "wait_for_agent", directRouting.WaitForAgent)
		resourcedata.SetMapValueIfNotNil(directRoutingMap, "agent_wait_seconds", directRouting.AgentWaitSeconds)

		directRoutingList = append(directRoutingList, directRoutingMap)
	}

	return directRoutingList
}

// flattenQueueMessagingAddressess maps a Genesys Cloud *[]platformclientv2.Queuemessagingaddresses into a []interface{}
func flattenQueueMessagingAddressess(queueMessagingAddressess *[]platformclientv2.Queuemessagingaddresses) []interface{} {
	if len(*queueMessagingAddressess) == 0 {
		return nil
	}

	var queueMessagingAddressesList []interface{}
	for _, queueMessagingAddresses := range *queueMessagingAddressess {
		queueMessagingAddressesMap := make(map[string]interface{})

		resourcedata.SetMapReferenceValueIfNotNil(queueMessagingAddressesMap, "sms_address_id", queueMessagingAddresses.SmsAddressId)
		resourcedata.SetMapReferenceValueIfNotNil(queueMessagingAddressesMap, "open_messaging_recipient_id", queueMessagingAddresses.OpenMessagingRecipientId)
		resourcedata.SetMapReferenceValueIfNotNil(queueMessagingAddressesMap, "whats_app_recipient_id", queueMessagingAddresses.WhatsAppRecipientId)

		queueMessagingAddressesList = append(queueMessagingAddressesList, queueMessagingAddressesMap)
	}

	return queueMessagingAddressesList
}

// flattenQueueEmailAddresss maps a Genesys Cloud *[]platformclientv2.Queueemailaddress into a []interface{}
func flattenQueueEmailAddresss(queueEmailAddresss *[]platformclientv2.Queueemailaddress) []interface{} {
	if len(*queueEmailAddresss) == 0 {
		return nil
	}

	var queueEmailAddressList []interface{}
	for _, queueEmailAddress := range *queueEmailAddresss {
		queueEmailAddressMap := make(map[string]interface{})

		queueEmailAddressList = append(queueEmailAddressList, queueEmailAddressMap)
	}

	return queueEmailAddressList
}

// flattenEmailAddresss maps a Genesys Cloud *[]platformclientv2.Emailaddress into a []interface{}
func flattenEmailAddresss(emailAddresss *[]platformclientv2.Emailaddress) []interface{} {
	if len(*emailAddresss) == 0 {
		return nil
	}

	var emailAddressList []interface{}
	for _, emailAddress := range *emailAddresss {
		emailAddressMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(emailAddressMap, "email", emailAddress.Email)
		resourcedata.SetMapValueIfNotNil(emailAddressMap, "name", emailAddress.Name)

		emailAddressList = append(emailAddressList, emailAddressMap)
	}

	return emailAddressList
}

// flattenSignatures maps a Genesys Cloud *[]platformclientv2.Signature into a []interface{}
func flattenSignatures(signatures *[]platformclientv2.Signature) []interface{} {
	if len(*signatures) == 0 {
		return nil
	}

	var signatureList []interface{}
	for _, signature := range *signatures {
		signatureMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(signatureMap, "enabled", signature.Enabled)
		resourcedata.SetMapValueIfNotNil(signatureMap, "canned_response_id", signature.CannedResponseId)
		resourcedata.SetMapValueIfNotNil(signatureMap, "always_included", signature.AlwaysIncluded)
		resourcedata.SetMapValueIfNotNil(signatureMap, "inclusion_type", signature.InclusionType)

		signatureList = append(signatureList, signatureMap)
	}

	return signatureList
}

// flattenInboundRoutes maps a Genesys Cloud *[]platformclientv2.Inboundroute into a []interface{}
func flattenInboundRoutes(inboundRoutes *[]platformclientv2.Inboundroute) []interface{} {
	if len(*inboundRoutes) == 0 {
		return nil
	}

	var inboundRouteList []interface{}
	for _, inboundRoute := range *inboundRoutes {
		inboundRouteMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(inboundRouteMap, "name", inboundRoute.Name)
		resourcedata.SetMapValueIfNotNil(inboundRouteMap, "pattern", inboundRoute.Pattern)
		resourcedata.SetMapReferenceValueIfNotNil(inboundRouteMap, "queue_id", inboundRoute.QueueId)
		resourcedata.SetMapValueIfNotNil(inboundRouteMap, "priority", inboundRoute.Priority)
		// TODO: Handle skills property
		resourcedata.SetMapReferenceValueIfNotNil(inboundRouteMap, "language_id", inboundRoute.LanguageId)
		resourcedata.SetMapValueIfNotNil(inboundRouteMap, "from_name", inboundRoute.FromName)
		resourcedata.SetMapValueIfNotNil(inboundRouteMap, "from_email", inboundRoute.FromEmail)
		resourcedata.SetMapReferenceValueIfNotNil(inboundRouteMap, "flow_id", inboundRoute.FlowId)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(inboundRouteMap, "reply_email_address", inboundRoute.ReplyEmailAddress, flattenQueueEmailAddress)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(inboundRouteMap, "auto_bcc", inboundRoute.AutoBcc, flattenEmailAddresss)
		resourcedata.SetMapReferenceValueIfNotNil(inboundRouteMap, "spam_flow_id", inboundRoute.SpamFlowId)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(inboundRouteMap, "signature", inboundRoute.Signature, flattenSignature)
		resourcedata.SetMapValueIfNotNil(inboundRouteMap, "history_inclusion", inboundRoute.HistoryInclusion)
		resourcedata.SetMapValueIfNotNil(inboundRouteMap, "allow_multiple_actions", inboundRoute.AllowMultipleActions)

		inboundRouteList = append(inboundRouteList, inboundRouteMap)
	}

	return inboundRouteList
}

// flattenQueueEmailAddresss maps a Genesys Cloud *[]platformclientv2.Queueemailaddress into a []interface{}
func flattenQueueEmailAddresss(queueEmailAddresss *[]platformclientv2.Queueemailaddress) []interface{} {
	if len(*queueEmailAddresss) == 0 {
		return nil
	}

	var queueEmailAddressList []interface{}
	for _, queueEmailAddress := range *queueEmailAddresss {
		queueEmailAddressMap := make(map[string]interface{})

		resourcedata.SetMapReferenceValueIfNotNil(queueEmailAddressMap, "domain_id", queueEmailAddress.DomainId)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(queueEmailAddressMap, "route", queueEmailAddress.Route, flattenInboundRoute)

		queueEmailAddressList = append(queueEmailAddressList, queueEmailAddressMap)
	}

	return queueEmailAddressList
}

// flattenQueues maps a Genesys Cloud *[]platformclientv2.Queue into a []interface{}
func flattenQueues(queues *[]platformclientv2.Queue) []interface{} {
	if len(*queues) == 0 {
		return nil
	}

	var queueList []interface{}
	for _, queue := range *queues {
		queueMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(queueMap, "name", queue.Name)
		resourcedata.SetMapValueIfNotNil(queueMap, "division_id", queue.DivisionId)
		resourcedata.SetMapValueIfNotNil(queueMap, "description", queue.Description)
		resourcedata.SetMapValueIfNotNil(queueMap, "member_count", queue.MemberCount)
		resourcedata.SetMapValueIfNotNil(queueMap, "user_member_count", queue.UserMemberCount)
		resourcedata.SetMapValueIfNotNil(queueMap, "joined_member_count", queue.JoinedMemberCount)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(queueMap, "media_settings", queue.MediaSettings, flattenQueueMediaSettings)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(queueMap, "routing_rules", queue.RoutingRules, flattenRoutingRules)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(queueMap, "conditional_group_routing", queue.ConditionalGroupRouting, flattenConditionalGroupRouting)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(queueMap, "bullseye", queue.Bullseye, flattenBullseye)
		resourcedata.SetMapValueIfNotNil(queueMap, "scoring_method", queue.ScoringMethod)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(queueMap, "acw_settings", queue.AcwSettings, flattenAcwSettings)
		resourcedata.SetMapValueIfNotNil(queueMap, "skill_evaluation_method", queue.SkillEvaluationMethod)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(queueMap, "member_groups", queue.MemberGroups, flattenMemberGroups)
		resourcedata.SetMapReferenceValueIfNotNil(queueMap, "queue_flow_id", queue.QueueFlowId)
		resourcedata.SetMapReferenceValueIfNotNil(queueMap, "email_in_queue_flow_id", queue.EmailInQueueFlowId)
		resourcedata.SetMapReferenceValueIfNotNil(queueMap, "message_in_queue_flow_id", queue.MessageInQueueFlowId)
		resourcedata.SetMapReferenceValueIfNotNil(queueMap, "whisper_prompt_id", queue.WhisperPromptId)
		resourcedata.SetMapReferenceValueIfNotNil(queueMap, "on_hold_prompt_id", queue.OnHoldPromptId)
		resourcedata.SetMapValueIfNotNil(queueMap, "auto_answer_only", queue.AutoAnswerOnly)
		resourcedata.SetMapValueIfNotNil(queueMap, "enable_transcription", queue.EnableTranscription)
		resourcedata.SetMapValueIfNotNil(queueMap, "enable_audio_monitoring", queue.EnableAudioMonitoring)
		resourcedata.SetMapValueIfNotNil(queueMap, "enable_manual_assignment", queue.EnableManualAssignment)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(queueMap, "agent_owned_routing", queue.AgentOwnedRouting, flattenAgentOwnedRouting)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(queueMap, "direct_routing", queue.DirectRouting, flattenDirectRouting)
		resourcedata.SetMapValueIfNotNil(queueMap, "calling_party_name", queue.CallingPartyName)
		resourcedata.SetMapValueIfNotNil(queueMap, "calling_party_number", queue.CallingPartyNumber)
		// TODO: Handle default_scripts property
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(queueMap, "outbound_messaging_addresses", queue.OutboundMessagingAddresses, flattenQueueMessagingAddresses)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(queueMap, "outbound_email_address", queue.OutboundEmailAddress, flattenQueueEmailAddress)
		resourcedata.SetMapValueIfNotNil(queueMap, "peer_id", queue.PeerId)
		resourcedata.SetMapValueIfNotNil(queueMap, "suppress_in_queue_call_recording", queue.SuppressInQueueCallRecording)

		queueList = append(queueList, queueMap)
	}

	return queueList
}

// flattenDomainResourceConditionValues maps a Genesys Cloud *[]platformclientv2.Domainresourceconditionvalue into a []interface{}
func flattenDomainResourceConditionValues(domainResourceConditionValues *[]platformclientv2.Domainresourceconditionvalue) []interface{} {
	if len(*domainResourceConditionValues) == 0 {
		return nil
	}

	var domainResourceConditionValueList []interface{}
	for _, domainResourceConditionValue := range *domainResourceConditionValues {
		domainResourceConditionValueMap := make(map[string]interface{})

		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(domainResourceConditionValueMap, "user", domainResourceConditionValue.User, flattenUser)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(domainResourceConditionValueMap, "queue", domainResourceConditionValue.Queue, flattenQueue)
		resourcedata.SetMapValueIfNotNil(domainResourceConditionValueMap, "value", domainResourceConditionValue.Value)
		resourcedata.SetMapValueIfNotNil(domainResourceConditionValueMap, "type", domainResourceConditionValue.Type)

		domainResourceConditionValueList = append(domainResourceConditionValueList, domainResourceConditionValueMap)
	}

	return domainResourceConditionValueList
}

// flattenDomainResourceConditionNodes maps a Genesys Cloud *[]platformclientv2.Domainresourceconditionnode into a []interface{}
func flattenDomainResourceConditionNodes(domainResourceConditionNodes *[]platformclientv2.Domainresourceconditionnode) []interface{} {
	if len(*domainResourceConditionNodes) == 0 {
		return nil
	}

	var domainResourceConditionNodeList []interface{}
	for _, domainResourceConditionNode := range *domainResourceConditionNodes {
		domainResourceConditionNodeMap := make(map[string]interface{})

		domainResourceConditionNodeList = append(domainResourceConditionNodeList, domainResourceConditionNodeMap)
	}

	return domainResourceConditionNodeList
}

// flattenDomainResourceConditionNodes maps a Genesys Cloud *[]platformclientv2.Domainresourceconditionnode into a []interface{}
func flattenDomainResourceConditionNodes(domainResourceConditionNodes *[]platformclientv2.Domainresourceconditionnode) []interface{} {
	if len(*domainResourceConditionNodes) == 0 {
		return nil
	}

	var domainResourceConditionNodeList []interface{}
	for _, domainResourceConditionNode := range *domainResourceConditionNodes {
		domainResourceConditionNodeMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(domainResourceConditionNodeMap, "variable_name", domainResourceConditionNode.VariableName)
		resourcedata.SetMapValueIfNotNil(domainResourceConditionNodeMap, "operator", domainResourceConditionNode.Operator)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(domainResourceConditionNodeMap, "operands", domainResourceConditionNode.Operands, flattenDomainResourceConditionValues)
		resourcedata.SetMapValueIfNotNil(domainResourceConditionNodeMap, "conjunction", domainResourceConditionNode.Conjunction)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(domainResourceConditionNodeMap, "terms", domainResourceConditionNode.Terms, flattenDomainResourceConditionNodes)

		domainResourceConditionNodeList = append(domainResourceConditionNodeList, domainResourceConditionNodeMap)
	}

	return domainResourceConditionNodeList
}

// flattenDomainPermissionPolicys maps a Genesys Cloud *[]platformclientv2.Domainpermissionpolicy into a []interface{}
func flattenDomainPermissionPolicys(domainPermissionPolicys *[]platformclientv2.Domainpermissionpolicy) []interface{} {
	if len(*domainPermissionPolicys) == 0 {
		return nil
	}

	var domainPermissionPolicyList []interface{}
	for _, domainPermissionPolicy := range *domainPermissionPolicys {
		domainPermissionPolicyMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(domainPermissionPolicyMap, "domain", domainPermissionPolicy.Domain)
		resourcedata.SetMapValueIfNotNil(domainPermissionPolicyMap, "entity_name", domainPermissionPolicy.EntityName)
		resourcedata.SetMapValueIfNotNil(domainPermissionPolicyMap, "policy_name", domainPermissionPolicy.PolicyName)
		resourcedata.SetMapValueIfNotNil(domainPermissionPolicyMap, "policy_description", domainPermissionPolicy.PolicyDescription)
		resourcedata.SetMapStringArrayValueIfNotNil(domainPermissionPolicyMap, "action_set", domainPermissionPolicy.ActionSet)
		resourcedata.SetMapStringArrayValueIfNotNil(domainPermissionPolicyMap, "named_resources", domainPermissionPolicy.NamedResources)
		resourcedata.SetMapValueIfNotNil(domainPermissionPolicyMap, "allow_conditions", domainPermissionPolicy.AllowConditions)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(domainPermissionPolicyMap, "resource_condition_node", domainPermissionPolicy.ResourceConditionNode, flattenDomainResourceConditionNode)

		domainPermissionPolicyList = append(domainPermissionPolicyList, domainPermissionPolicyMap)
	}

	return domainPermissionPolicyList
}
