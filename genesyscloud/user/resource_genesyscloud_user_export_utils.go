// Package user contains temporary export utilities for Plugin Framework user resource.
//
// IMPORTANT: This file contains migration scaffolding that converts SDK types to flat
// attribute maps for the legacy exporter's dependency resolution logic.
//
// TODO: Remove this entire file once all resources are migrated to Plugin Framework
// and the exporter is updated to work natively with Framework types (Phase 2).
// This is Phase 1 temporary code - resource-specific implementation.
//
// File: genesyscloud/user/resource_genesyscloud_user_export_utils.go

package user

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
)

// buildUserAttributes creates a flat attribute map from SDK user object for export.
// This function fetches ALL user attributes including voicemail and routing utilization
// via separate API calls, matching SDKv2 readUser behavior.
//
// Parameters:
//   - ctx: Context for API calls
//   - user: User object from API (must include expansions: skills, languages, locations, profileSkills, certifications, employerInfo)
//   - proxy: User proxy for additional API calls (voicemail, utilization, extension pools)
//
// Returns:
//   - map[string]string: Flat attribute map with all user attributes
//   - error: Error if any fetch operation fails (caller should skip this user)
//
// Attribute Map Format (matching SDKv2 InstanceState):
//   - Basic: "name", "email", "division_id", "manager"
//   - Nested: "addresses.0.phone_numbers.0.extension_pool_id"
//   - Arrays: "routing_skills.#" = count, "routing_skills.0.skill_id" = value
//
// Error Handling:
//   - Returns error if voicemail or utilization fetch fails (matching SDKv2 behavior)
//   - Caller should skip user and continue with others
//   - Logs warnings for non-critical issues (e.g., extension pool not found)
func buildUserAttributes(ctx context.Context, user *platformclientv2.User, proxy *userProxy) (map[string]string, error) {
	attributes := make(map[string]string)

	// Basic attributes
	if user.Id != nil {
		attributes["id"] = *user.Id
	}
	if user.Name != nil {
		attributes["name"] = *user.Name
	}
	if user.Email != nil {
		attributes["email"] = *user.Email
	}
	if user.State != nil {
		attributes["state"] = *user.State
	}
	if user.Department != nil {
		attributes["department"] = *user.Department
	}
	if user.Title != nil {
		attributes["title"] = *user.Title
	}
	if user.AcdAutoAnswer != nil {
		attributes["acd_auto_answer"] = strconv.FormatBool(*user.AcdAutoAnswer)
	}

	// ⭐ CRITICAL: Dependency references (used by exporter for dependency resolution)
	if user.Division != nil && user.Division.Id != nil {
		attributes["division_id"] = *user.Division.Id
	}
	if user.Manager != nil && (*user.Manager).Id != nil {
		attributes["manager"] = *(*user.Manager).Id
	}

	// Complex nested attributes
	if user.Addresses != nil {
		if err := flattenSDKAddressesToAttributes(ctx, *user.Addresses, attributes, proxy); err != nil {
			return nil, fmt.Errorf("failed to flatten addresses: %w", err)
		}
	}

	if user.Skills != nil {
		flattenSDKSkillsToAttributes(*user.Skills, attributes)
	}

	if user.Languages != nil {
		flattenSDKLanguagesToAttributes(*user.Languages, attributes)
	}

	if user.Locations != nil {
		flattenSDKLocationsToAttributes(*user.Locations, attributes)
	}

	if user.ProfileSkills != nil {
		flattenSDKProfileSkillsToAttributes(*user.ProfileSkills, attributes)
	}

	if user.Certifications != nil {
		flattenSDKCertificationsToAttributes(*user.Certifications, attributes)
	}

	if user.EmployerInfo != nil {
		flattenSDKEmployerInfoToAttributes(user.EmployerInfo, attributes)
	}

	// Fetch voicemail policies (separate API call, matching SDKv2)
	// Return error if fetch fails (matching SDKv2 behavior - skip user)
	voicemail, _, err := proxy.getVoicemailUserpoliciesById(ctx, *user.Id)
	if err != nil {
		log.Printf("Failed to fetch voicemail policies for user %s: %v", *user.Id, err)
		return nil, fmt.Errorf("failed to fetch voicemail policies: %w", err)
	}
	if voicemail != nil {
		flattenSDKVoicemailToAttributes(voicemail, attributes)
	}

	// Fetch routing utilization (separate API call, matching SDKv2)
	// Return error if fetch fails (matching SDKv2 behavior - skip user)
	if err := flattenSDKRoutingUtilizationToAttributes(ctx, *user.Id, proxy, attributes); err != nil {
		log.Printf("Failed to fetch routing utilization for user %s: %v", *user.Id, err)
		return nil, fmt.Errorf("failed to fetch routing utilization: %w", err)
	}

	return attributes, nil
}

// flattenSDKAddressesToAttributes converts SDK addresses to flat attribute map.
// Handles both phone numbers (with extension pool lookup) and other emails.
//
// Attribute Format (matching SDKv2 InstanceState):
//   - "addresses.#" = "1"
//   - "addresses.0.phone_numbers.#" = count
//   - "addresses.0.phone_numbers.0.number" = "+13175559001"
//   - "addresses.0.phone_numbers.0.extension" = "8701"
//   - "addresses.0.phone_numbers.0.extension_pool_id" = "pool-guid" ⭐ DEPENDENCY
//   - "addresses.0.phone_numbers.0.media_type" = "PHONE"
//   - "addresses.0.phone_numbers.0.type" = "WORK"
//   - "addresses.0.other_emails.#" = count
//   - "addresses.0.other_emails.0.address" = "alt@example.com"
//   - "addresses.0.other_emails.0.type" = "WORK"
//
// Extension Pool Lookup:
//   - For phone numbers with extensions, fetches extension_pool_id via API
//   - Logs warning if pool not found (edge case: extension exists but pool deleted)
//   - Sets empty string if lookup fails (exporter will handle gracefully)
func flattenSDKAddressesToAttributes(ctx context.Context, addresses []platformclientv2.Contact, attributes map[string]string, proxy *userProxy) error {
	if len(addresses) == 0 {
		return nil
	}

	phoneIndex := 0
	emailIndex := 0

	for _, address := range addresses {
		if address.MediaType == nil {
			continue
		}

		switch *address.MediaType {
		case "PHONE", "SMS":
			// Build attribute key prefix
			prefix := fmt.Sprintf("addresses.0.phone_numbers.%d", phoneIndex)

			if address.Address != nil {
				attributes[prefix+".number"] = *address.Address
			}
			if address.MediaType != nil {
				attributes[prefix+".media_type"] = *address.MediaType
			}
			if address.VarType != nil {
				attributes[prefix+".type"] = *address.VarType
			}
			if address.Extension != nil {
				attributes[prefix+".extension"] = *address.Extension

				// ⭐ CRITICAL: Fetch extension pool ID (dependency reference)
				// This is required for exporter to resolve extension pool dependency
				// Reuses existing fetchExtensionPoolId function from resource_genesyscloud_user_utils.go
				poolId := fetchExtensionPoolId(ctx, *address.Extension, proxy)
				if poolId != "" {
					attributes[prefix+".extension_pool_id"] = poolId
				} else {
					// Log warning but continue (edge case: pool deleted)
					log.Printf("Warning: Extension pool not found for extension %s", *address.Extension)
				}
			}

			phoneIndex++

		case "EMAIL":
			// Skip primary email (already in email attribute)
			if address.VarType != nil && *address.VarType == "PRIMARY" {
				continue
			}

			prefix := fmt.Sprintf("addresses.0.other_emails.%d", emailIndex)

			if address.Address != nil {
				attributes[prefix+".address"] = *address.Address
			}
			if address.VarType != nil {
				attributes[prefix+".type"] = *address.VarType
			}

			emailIndex++
		}
	}

	// Set counts (matching SDKv2 format)
	attributes["addresses.#"] = "1"
	attributes["addresses.0.phone_numbers.#"] = strconv.Itoa(phoneIndex)
	attributes["addresses.0.other_emails.#"] = strconv.Itoa(emailIndex)

	return nil
}

// flattenSDKSkillsToAttributes converts SDK skills to flat attribute map.
// Each skill includes skill_id (dependency reference) and proficiency.
//
// Attribute Format:
//   - "routing_skills.#" = count
//   - "routing_skills.0.skill_id" = "skill-guid" ⭐ DEPENDENCY
//   - "routing_skills.0.proficiency" = "4.5"
func flattenSDKSkillsToAttributes(skills []platformclientv2.Userroutingskill, attributes map[string]string) {
	if len(skills) == 0 {
		return
	}

	for i, skill := range skills {
		prefix := fmt.Sprintf("routing_skills.%d", i)

		// ⭐ CRITICAL: skill_id is the dependency reference
		if skill.Id != nil {
			attributes[prefix+".skill_id"] = *skill.Id
		}

		if skill.Proficiency != nil {
			attributes[prefix+".proficiency"] = fmt.Sprintf("%.1f", *skill.Proficiency)
		}
	}

	attributes["routing_skills.#"] = strconv.Itoa(len(skills))
}

// flattenSDKLanguagesToAttributes converts SDK languages to flat attribute map.
//
// Attribute Format:
//   - "routing_languages.#" = count
//   - "routing_languages.0.language_id" = "lang-guid" ⭐ DEPENDENCY
//   - "routing_languages.0.proficiency" = "5"
func flattenSDKLanguagesToAttributes(languages []platformclientv2.Userroutinglanguage, attributes map[string]string) {
	if len(languages) == 0 {
		return
	}

	for i, language := range languages {
		prefix := fmt.Sprintf("routing_languages.%d", i)

		// ⭐ CRITICAL: language_id is the dependency reference
		if language.Id != nil {
			attributes[prefix+".language_id"] = *language.Id
		}

		if language.Proficiency != nil {
			attributes[prefix+".proficiency"] = strconv.Itoa(int(*language.Proficiency))
		}
	}

	attributes["routing_languages.#"] = strconv.Itoa(len(languages))
}

// flattenSDKLocationsToAttributes converts SDK locations to flat attribute map.
//
// Attribute Format:
//   - "locations.#" = count
//   - "locations.0.location_id" = "location-guid" ⭐ DEPENDENCY
//   - "locations.0.notes" = "Primary office"
func flattenSDKLocationsToAttributes(locations []platformclientv2.Location, attributes map[string]string) {
	if len(locations) == 0 {
		return
	}

	for i, location := range locations {
		prefix := fmt.Sprintf("locations.%d", i)

		// ⭐ CRITICAL: location_id is the dependency reference
		if location.Id != nil {
			attributes[prefix+".location_id"] = *location.Id
		}

		if location.Notes != nil {
			attributes[prefix+".notes"] = *location.Notes
		}
	}

	attributes["locations.#"] = strconv.Itoa(len(locations))
}

// flattenSDKProfileSkillsToAttributes converts profile skills to flat attribute map.
//
// Attribute Format:
//   - "profile_skills.#" = count
//   - "profile_skills.0" = "Java"
//   - "profile_skills.1" = "Python"
func flattenSDKProfileSkillsToAttributes(skills []string, attributes map[string]string) {
	if len(skills) == 0 {
		return
	}

	for i, skill := range skills {
		attributes[fmt.Sprintf("profile_skills.%d", i)] = skill
	}

	attributes["profile_skills.#"] = strconv.Itoa(len(skills))
}

// flattenSDKCertificationsToAttributes converts certifications to flat attribute map.
//
// Attribute Format:
//   - "certifications.#" = count
//   - "certifications.0" = "AWS Certified"
func flattenSDKCertificationsToAttributes(certifications []string, attributes map[string]string) {
	if len(certifications) == 0 {
		return
	}

	for i, cert := range certifications {
		attributes[fmt.Sprintf("certifications.%d", i)] = cert
	}

	attributes["certifications.#"] = strconv.Itoa(len(certifications))
}

// flattenSDKEmployerInfoToAttributes converts employer info to flat attribute map.
//
// Attribute Format:
//   - "employer_info.#" = "1"
//   - "employer_info.0.official_name" = "John Smith"
//   - "employer_info.0.employee_id" = "EMP-12345"
//   - "employer_info.0.employee_type" = "Full-time"
//   - "employer_info.0.date_hire" = "2020-01-15"
func flattenSDKEmployerInfoToAttributes(employerInfo *platformclientv2.Employerinfo, attributes map[string]string) {
	if employerInfo == nil {
		return
	}

	prefix := "employer_info.0"

	if employerInfo.OfficialName != nil {
		attributes[prefix+".official_name"] = *employerInfo.OfficialName
	}
	if employerInfo.EmployeeId != nil {
		attributes[prefix+".employee_id"] = *employerInfo.EmployeeId
	}
	if employerInfo.EmployeeType != nil {
		attributes[prefix+".employee_type"] = *employerInfo.EmployeeType
	}
	if employerInfo.DateHire != nil {
		attributes[prefix+".date_hire"] = *employerInfo.DateHire
	}

	attributes["employer_info.#"] = "1"
}

// flattenSDKVoicemailToAttributes converts voicemail policies to flat attribute map.
//
// Attribute Format:
//   - "voicemail_userpolicies.#" = "1"
//   - "voicemail_userpolicies.0.alert_timeout_seconds" = "30"
//   - "voicemail_userpolicies.0.send_email_notifications" = "true"
func flattenSDKVoicemailToAttributes(voicemail *platformclientv2.Voicemailuserpolicy, attributes map[string]string) {
	if voicemail == nil {
		return
	}

	prefix := "voicemail_userpolicies.0"

	if voicemail.AlertTimeoutSeconds != nil {
		attributes[prefix+".alert_timeout_seconds"] = strconv.Itoa(*voicemail.AlertTimeoutSeconds)
	}
	if voicemail.SendEmailNotifications != nil {
		attributes[prefix+".send_email_notifications"] = strconv.FormatBool(*voicemail.SendEmailNotifications)
	}

	attributes["voicemail_userpolicies.#"] = "1"
}

// flattenSDKRoutingUtilizationToAttributes fetches and converts routing utilization to flat map.
// Makes separate API call to get utilization settings (matching SDKv2 behavior).
//
// Attribute Format:
//   - "routing_utilization.#" = "1"
//   - "routing_utilization.0.call.#" = "1"
//   - "routing_utilization.0.call.0.maximum_capacity" = "3"
//   - "routing_utilization.0.call.0.include_non_acd" = "false"
//   - "routing_utilization.0.call.0.interruptible_media_types.#" = count
//   - "routing_utilization.0.call.0.interruptible_media_types.0" = "email"
//   - ... similar for callback, message, email, chat
//   - "routing_utilization.0.label_utilizations.#" = count
//   - "routing_utilization.0.label_utilizations.0.label_id" = "label-guid"
//   - "routing_utilization.0.label_utilizations.0.maximum_capacity" = "5"
//
// Returns error if API call fails (matching SDKv2 behavior - caller should skip user).
func flattenSDKRoutingUtilizationToAttributes(ctx context.Context, userId string, proxy *userProxy, attributes map[string]string) error {
	// Make API call to get routing utilization
	// Reuses existing buildHeaderParams function from resource_genesyscloud_user_utils.go
	apiClient := &proxy.routingApi.Configuration.APIClient
	path := fmt.Sprintf("%s/api/v2/routing/users/%s/utilization",
		proxy.routingApi.Configuration.BasePath, userId)

	response, err := apiClient.CallAPI(path, "GET", nil, buildHeaderParams(proxy.routingApi),
		nil, nil, "", nil, "")
	if err != nil {
		return fmt.Errorf("failed to fetch routing utilization: %w", err)
	}

	// Unmarshal response
	// Reuses existing agentUtilizationWithLabels type from resource_genesyscloud_user.go
	var agentUtilization agentUtilizationWithLabels
	if err := json.Unmarshal(response.RawBody, &agentUtilization); err != nil {
		return fmt.Errorf("failed to unmarshal routing utilization: %w", err)
	}

	// If organization-level settings, don't export (matching SDKv2 behavior)
	if agentUtilization.Level == "Organization" {
		return nil
	}

	// Flatten media utilization settings
	// Reuses existing getUtilizationMediaTypes function from resource_genesyscloud_user_utils.go
	for sdkType, schemaType := range getUtilizationMediaTypes() {
		if mediaSettings, ok := agentUtilization.Utilization[sdkType]; ok {
			prefix := fmt.Sprintf("routing_utilization.0.%s.0", schemaType)

			attributes[prefix+".maximum_capacity"] = strconv.Itoa(int(mediaSettings.MaximumCapacity))
			attributes[prefix+".include_non_acd"] = strconv.FormatBool(mediaSettings.IncludeNonAcd)

			// Interruptible media types
			if len(mediaSettings.InterruptableMediaTypes) > 0 {
				for i, mediaType := range mediaSettings.InterruptableMediaTypes {
					attributes[fmt.Sprintf("%s.interruptible_media_types.%d", prefix, i)] = mediaType
				}
				attributes[prefix+".interruptible_media_types.#"] = strconv.Itoa(len(mediaSettings.InterruptableMediaTypes))
			}

			attributes[fmt.Sprintf("routing_utilization.0.%s.#", schemaType)] = "1"
		}
	}

	// Flatten label utilizations
	if len(agentUtilization.LabelUtilizations) > 0 {
		labelIndex := 0
		// LabelUtilizations is a map where the key is the label ID
		for labelId, labelUtil := range agentUtilization.LabelUtilizations {
			prefix := fmt.Sprintf("routing_utilization.0.label_utilizations.%d", labelIndex)

			attributes[prefix+".label_id"] = labelId
			attributes[prefix+".maximum_capacity"] = strconv.Itoa(int(labelUtil.MaximumCapacity))

			if len(labelUtil.InterruptingLabelIds) > 0 {
				for i, interruptingLabelId := range labelUtil.InterruptingLabelIds {
					attributes[fmt.Sprintf("%s.interrupting_label_ids.%d", prefix, i)] = interruptingLabelId
				}
				attributes[prefix+".interrupting_label_ids.#"] = strconv.Itoa(len(labelUtil.InterruptingLabelIds))
			}

			labelIndex++
		}
		attributes["routing_utilization.0.label_utilizations.#"] = strconv.Itoa(labelIndex)
	}

	attributes["routing_utilization.#"] = "1"
	return nil
}
