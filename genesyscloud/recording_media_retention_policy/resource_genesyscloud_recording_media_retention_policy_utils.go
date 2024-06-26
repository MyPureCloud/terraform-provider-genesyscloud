package recording_media_retention_policy

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"time"

	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v133/platformclientv2"
)

/*
The resource_genesyscloud_recording_media_retention_policy_utils.go file contains various helper methods to marshal
and unmarshal data into formats consumable by Terraform and/or Genesys Cloud.

Note:  Look for opportunities to minimize boilerplate code using functions and Generics
*/

type VisibilityConditionStruct struct {
	CombiningOperation string
	Predicates         []string
}

func buildEvaluationAssignments(evaluations []interface{}, pp *policyProxy, ctx context.Context) *[]platformclientv2.Evaluationassignment {
	assignEvaluations := make([]platformclientv2.Evaluationassignment, 0)

	for _, assignEvaluation := range evaluations {
		assignEvaluationMap, ok := assignEvaluation.(map[string]interface{})
		if !ok {
			continue
		}
		evaluationFormId := assignEvaluationMap["evaluation_form_id"].(string)
		userId := assignEvaluationMap["user_id"].(string)
		assignment := platformclientv2.Evaluationassignment{}

		// if evaluation form id is present, get the context id and build the evaluation form
		if evaluationFormId != "" {
			form, resp, err := pp.getFormsEvaluation(ctx, evaluationFormId)
			if err != nil {
				log.Fatalf("failed to read evaluation form %s: %s %v", evaluationFormId, err, resp)
			} else {
				evaluationFormContextId := form.ContextId
				assignment.EvaluationForm = &platformclientv2.Evaluationform{Id: &evaluationFormId, ContextId: evaluationFormContextId}
			}
		}
		if userId != "" {
			assignment.User = &platformclientv2.User{Id: &userId}
		}
		assignEvaluations = append(assignEvaluations, assignment)
	}
	return &assignEvaluations
}

func flattenEvaluationAssignments(assignments *[]platformclientv2.Evaluationassignment, pp *policyProxy, ctx context.Context) []interface{} {
	if assignments == nil {
		return nil
	}

	evaluationAssignments := make([]interface{}, 0)
	for _, assignment := range *assignments {
		assignmentMap := make(map[string]interface{})

		// if form is present in the response, assign the most recent unpublished version id to align with evaluation form resource behavior for export purposes.
		if assignment.EvaluationForm != nil {
			formId := *assignment.EvaluationForm.Id
			formVersionId, resp, err := pp.getEvaluationFormRecentVerId(ctx, formId)
			if err != nil {
				log.Fatalf("Failed to get evaluation form versions %s %v", *assignment.EvaluationForm.Name, resp)
			} else {
				formId = formVersionId
			}

			assignmentMap["evaluation_form_id"] = formId
		}
		if assignment.User != nil {
			assignmentMap["user_id"] = *assignment.User.Id
		}
		evaluationAssignments = append(evaluationAssignments, assignmentMap)
	}
	return evaluationAssignments
}

func buildMeteredEvaluationsTimeInterval(interval []interface{}) *platformclientv2.Timeinterval {
	var timeInterval platformclientv2.Timeinterval

	if interval == nil || len(interval) <= 0 || (len(interval) == 1 && interval[0] == nil) {
		return nil
	}

	timeIntervalMap, ok := interval[0].(map[string]interface{})
	if !ok {
		return nil
	}

	if days, ok := timeIntervalMap["days"].(int); ok && days != 0 {
		timeInterval.Days = &days
	}
	if hours, ok := timeIntervalMap["hours"].(int); ok && hours != 0 {
		timeInterval.Hours = &hours
	}
	return &timeInterval
}

func buildMeteredAssignmentByAgentTimeInterval(interval []interface{}) *platformclientv2.Timeinterval {
	var timeInterval platformclientv2.Timeinterval

	if interval == nil || len(interval) <= 0 || (len(interval) == 1 && interval[0] == nil) {
		return nil
	}

	timeIntervalMap, ok := interval[0].(map[string]interface{})
	if !ok {
		return nil
	}

	if months, ok := timeIntervalMap["months"].(int); ok && months != 0 {
		timeInterval.Months = &months
	}
	if weeks, ok := timeIntervalMap["weeks"].(int); ok && weeks != 0 {
		timeInterval.Weeks = &weeks
	}
	if days, ok := timeIntervalMap["days"].(int); ok && days != 0 {
		timeInterval.Days = &days
	}

	return &timeInterval
}

func flattenEvalTimeInterval(timeInterval *platformclientv2.Timeinterval) []interface{} {
	if timeInterval == nil {
		return nil
	}

	timeIntervalMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(timeIntervalMap, "days", timeInterval.Days)
	resourcedata.SetMapValueIfNotNil(timeIntervalMap, "hours", timeInterval.Hours)

	return []interface{}{timeIntervalMap}
}

func flattenAgentTimeInterval(timeInterval *platformclientv2.Timeinterval) []interface{} {
	if timeInterval == nil {
		return nil
	}

	timeIntervalMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(timeIntervalMap, "months", timeInterval.Months)
	resourcedata.SetMapValueIfNotNil(timeIntervalMap, "weeks", timeInterval.Weeks)
	resourcedata.SetMapValueIfNotNil(timeIntervalMap, "days", timeInterval.Days)

	return []interface{}{timeIntervalMap}
}

func buildAssignMeteredEvaluations(assignments []interface{}, pp *policyProxy, ctx context.Context) *[]platformclientv2.Meteredevaluationassignment {
	meteredAssignments := make([]platformclientv2.Meteredevaluationassignment, 0)

	for _, assignment := range assignments {
		assignmentMap, ok := assignment.(map[string]interface{})
		if !ok {
			continue
		}
		maxNumberEvaluations := assignmentMap["max_number_evaluations"].(int)
		assignToActiveUser := assignmentMap["assign_to_active_user"].(bool)
		evaluationFormId := assignmentMap["evaluation_form_id"].(string)
		evaluatorIds := assignmentMap["evaluator_ids"].([]interface{})

		idStrings := make([]string, 0)
		for _, evaluatorId := range evaluatorIds {
			idStrings = append(idStrings, fmt.Sprintf("%v", evaluatorId))
		}

		evaluators := make([]platformclientv2.User, 0)
		for _, evaluatorId := range idStrings {
			evaluator := evaluatorId
			evaluators = append(evaluators, platformclientv2.User{Id: &evaluator})
		}

		timeInterval := buildMeteredEvaluationsTimeInterval(assignmentMap["time_interval"].([]interface{}))

		temp := platformclientv2.Meteredevaluationassignment{
			Evaluators:           &evaluators,
			MaxNumberEvaluations: &maxNumberEvaluations,
			AssignToActiveUser:   &assignToActiveUser,
			TimeInterval:         timeInterval,
		}

		// if evaluation form id is present, get the context id and build the evaluation form
		if evaluationFormId != "" {
			form, resp, err := pp.getFormsEvaluation(ctx, evaluationFormId)
			if err != nil {
				log.Fatalf("failed to read media evaluation form %s: %s %v", evaluationFormId, err, resp)
			} else {
				evaluationFormContextId := form.ContextId
				temp.EvaluationForm = &platformclientv2.Evaluationform{Id: &evaluationFormId, ContextId: evaluationFormContextId}
			}
		}
		meteredAssignments = append(meteredAssignments, temp)
	}
	return &meteredAssignments
}

func flattenAssignMeteredEvaluations(assignments *[]platformclientv2.Meteredevaluationassignment, pp *policyProxy, ctx context.Context) []interface{} {
	if assignments == nil {
		return nil
	}

	meteredAssignments := make([]interface{}, 0)
	for _, assignment := range *assignments {
		assignmentMap := make(map[string]interface{})
		if assignment.Evaluators != nil {
			evaluatorIds := make([]string, 0)
			for _, evaluator := range *assignment.Evaluators {
				evaluatorIds = append(evaluatorIds, *evaluator.Id)
			}
			assignmentMap["evaluator_ids"] = evaluatorIds
		}

		resourcedata.SetMapValueIfNotNil(assignmentMap, "max_number_evaluations", assignment.MaxNumberEvaluations)

		// if form is present in the response, assign the most recent unpublished version id to align with evaluation form resource behavior for export purposes.
		if assignment.EvaluationForm != nil {
			formId := *assignment.EvaluationForm.Id
			formVersionId, resp, err := pp.getEvaluationFormRecentVerId(ctx, formId)
			if err != nil {
				log.Fatalf("Failed to get evaluation form versions %s %v", *assignment.EvaluationForm.Name, resp)
			} else {
				formId = formVersionId
			}

			assignmentMap["evaluation_form_id"] = formId
		}

		resourcedata.SetMapValueIfNotNil(assignmentMap, "assign_to_active_user", assignment.AssignToActiveUser)
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(assignmentMap, "time_interval", assignment.TimeInterval, flattenEvalTimeInterval)

		meteredAssignments = append(meteredAssignments, assignmentMap)
	}
	return meteredAssignments
}

func buildAssignMeteredAssignmentByAgent(assignments []interface{}, pp *policyProxy, ctx context.Context) *[]platformclientv2.Meteredassignmentbyagent {
	meteredAssignments := make([]platformclientv2.Meteredassignmentbyagent, 0)
	for _, assignment := range assignments {
		assignmentMap, ok := assignment.(map[string]interface{})
		if !ok {
			continue
		}
		maxNumberEvaluations := assignmentMap["max_number_evaluations"].(int)
		timeZone := assignmentMap["time_zone"].(string)
		evaluationFormId := assignmentMap["evaluation_form_id"].(string)
		evaluatorIds := assignmentMap["evaluator_ids"].([]interface{})

		idStrings := make([]string, 0)
		for _, evaluatorId := range evaluatorIds {
			idStrings = append(idStrings, fmt.Sprintf("%v", evaluatorId))
		}

		evaluators := make([]platformclientv2.User, 0)
		for _, evaluatorId := range idStrings {
			evaluator := evaluatorId
			evaluators = append(evaluators, platformclientv2.User{Id: &evaluator})
		}

		timeInterval := buildMeteredAssignmentByAgentTimeInterval(assignmentMap["time_interval"].([]interface{}))

		temp := platformclientv2.Meteredassignmentbyagent{
			Evaluators:           &evaluators,
			MaxNumberEvaluations: &maxNumberEvaluations,
			TimeInterval:         timeInterval,
			TimeZone:             &timeZone,
		}

		// if evaluation form id is present, get the context id and build the evaluation form
		if evaluationFormId != "" {
			form, resp, err := pp.getFormsEvaluation(ctx, evaluationFormId)
			if err != nil {
				log.Fatalf("failed to read evaluation form %s: %s %v", evaluationFormId, err, resp)
			} else {
				evaluationFormContextId := form.ContextId
				temp.EvaluationForm = &platformclientv2.Evaluationform{Id: &evaluationFormId, ContextId: evaluationFormContextId}
			}
		}
		meteredAssignments = append(meteredAssignments, temp)
	}
	return &meteredAssignments
}

func flattenAssignMeteredAssignmentByAgent(assignments *[]platformclientv2.Meteredassignmentbyagent, pp *policyProxy, ctx context.Context) []interface{} {
	if assignments == nil {
		return nil
	}

	meteredAssignments := make([]interface{}, 0)
	for _, assignment := range *assignments {
		assignmentMap := make(map[string]interface{})
		if assignment.Evaluators != nil {
			evaluatorIds := make([]string, 0)
			for _, evaluator := range *assignment.Evaluators {
				evaluatorIds = append(evaluatorIds, *evaluator.Id)
			}
			assignmentMap["evaluator_ids"] = evaluatorIds
		}

		resourcedata.SetMapValueIfNotNil(assignmentMap, "max_number_evaluations", assignment.MaxNumberEvaluations)

		// if form is present in the response, assign the most recent unpublished version id to align with evaluation form resource behavior for export purposes.
		if assignment.EvaluationForm != nil {
			formId := *assignment.EvaluationForm.Id
			formVersionId, resp, err := pp.getEvaluationFormRecentVerId(ctx, formId)
			if err != nil {
				log.Fatalf("Failed to get evaluation form versions %s %v", *assignment.EvaluationForm.Name, resp)
			} else {
				formId = formVersionId
			}

			assignmentMap["evaluation_form_id"] = formId
		}

		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(assignmentMap, "time_interval", assignment.TimeInterval, flattenAgentTimeInterval)
		resourcedata.SetMapValueIfNotNil(assignmentMap, "time_zone", assignment.TimeZone)

		meteredAssignments = append(meteredAssignments, assignmentMap)
	}
	return meteredAssignments
}

func buildAssignCalibrations(assignments []interface{}, pp *policyProxy, ctx context.Context) *[]platformclientv2.Calibrationassignment {
	calibrationAssignments := make([]platformclientv2.Calibrationassignment, 0)

	for _, assignment := range assignments {
		assignmentMap, ok := assignment.(map[string]interface{})
		if !ok {
			continue
		}
		evaluationFormId := assignmentMap["evaluation_form_id"].(string)
		calibratorId := assignmentMap["calibrator_id"].(string)
		expertEvaluatorId := assignmentMap["expert_evaluator_id"].(string)
		evaluatorIds := assignmentMap["evaluator_ids"].([]interface{})

		idStrings := make([]string, 0)
		for _, evaluatorId := range evaluatorIds {
			idStrings = append(idStrings, fmt.Sprintf("%v", evaluatorId))
		}

		evaluators := make([]platformclientv2.User, 0)
		for _, evaluatorId := range idStrings {
			id := evaluatorId
			evaluators = append(evaluators, platformclientv2.User{Id: &id})
		}

		temp := platformclientv2.Calibrationassignment{
			Evaluators: &evaluators,
		}

		// if evaluation form id is present, get the context id and build the evaluation form
		if evaluationFormId != "" {
			form, resp, err := pp.getFormsEvaluation(ctx, evaluationFormId)
			if err != nil {
				log.Fatalf("failed to read evaluation form %s: %s %v", evaluationFormId, err, resp)
			} else {
				evaluationFormContextId := form.ContextId
				temp.EvaluationForm = &platformclientv2.Evaluationform{Id: &evaluationFormId, ContextId: evaluationFormContextId}
			}
		}

		if calibratorId != "" {
			temp.Calibrator = &platformclientv2.User{Id: &calibratorId}
		}
		if expertEvaluatorId != "" {
			temp.ExpertEvaluator = &platformclientv2.User{Id: &expertEvaluatorId}
		}

		calibrationAssignments = append(calibrationAssignments, temp)
	}

	return &calibrationAssignments
}

func flattenAssignCalibrations(assignments *[]platformclientv2.Calibrationassignment, pp *policyProxy, ctx context.Context) []interface{} {
	if assignments == nil {
		return nil
	}

	calibrationAssignments := make([]interface{}, 0)
	for _, assignment := range *assignments {
		assignmentMap := make(map[string]interface{})
		if assignment.Calibrator != nil {
			assignmentMap["calibrator_id"] = *assignment.Calibrator.Id
		}
		if assignment.Evaluators != nil {
			evaluatorIds := make([]string, 0)
			for _, evaluator := range *assignment.Evaluators {
				evaluatorIds = append(evaluatorIds, *evaluator.Id)
			}
			assignmentMap["evaluator_ids"] = evaluatorIds
		}
		// if form is present in the response, assign the most recent unpublished version id to align with evaluation form resource behavior for export purposes.
		if assignment.EvaluationForm != nil {
			formId := *assignment.EvaluationForm.Id
			formVersionId, resp, err := pp.getEvaluationFormRecentVerId(ctx, formId)
			if err != nil {
				log.Fatalf("Failed to get evaluation form versions %s %v", *assignment.EvaluationForm.Name, resp)
			} else {
				formId = formVersionId
			}

			assignmentMap["evaluation_form_id"] = formId
		}
		if assignment.ExpertEvaluator != nil {
			assignmentMap["expert_evaluator_id"] = *assignment.ExpertEvaluator.Id
		}

		calibrationAssignments = append(calibrationAssignments, assignmentMap)
	}
	return calibrationAssignments
}

func buildDomainEntityRef(idVal string) *platformclientv2.Domainentityref {
	if idVal == "nil" {
		return nil
	}

	return &platformclientv2.Domainentityref{
		Id: &idVal,
	}
}

func buildAssignSurveys(assignments []interface{}, pp *policyProxy, ctx context.Context) *[]platformclientv2.Surveyassignment {
	surveyAssignments := make([]platformclientv2.Surveyassignment, 0)

	for _, assignment := range assignments {
		assignmentMap, ok := assignment.(map[string]interface{})
		if !ok {
			continue
		}
		sendingUser := assignmentMap["sending_user"].(string)
		sendingDomain := assignmentMap["sending_domain"].(string)
		inviteTimeInterval := assignmentMap["invite_time_interval"].(string)
		surveyFormName := assignmentMap["survey_form_name"].(string)

		temp := platformclientv2.Surveyassignment{
			Flow:               buildDomainEntityRef(assignmentMap["flow_id"].(string)),
			InviteTimeInterval: &inviteTimeInterval,
			SendingUser:        &sendingUser,
			SendingDomain:      &sendingDomain,
		}

		// If a survey form name is provided, get the context id and build the published survey form reference
		if surveyFormName != "" {
			form, resp, err := pp.getQualityFormsSurveyByName(ctx, surveyFormName)
			if err != nil {
				log.Fatalf("Error requesting survey forms %s: %s %v", surveyFormName, err, resp)
			} else {
				surveyFormReference := platformclientv2.Publishedsurveyformreference{Name: &surveyFormName, ContextId: form.ContextId}
				temp.SurveyForm = &surveyFormReference
			}
		}

		surveyAssignments = append(surveyAssignments, temp)
	}

	return &surveyAssignments
}

func flattenAssignSurveys(assignments *[]platformclientv2.Surveyassignment) []interface{} {
	if assignments == nil {
		return nil
	}

	var surveyAssignments []interface{}

	for _, assignment := range *assignments {
		assignmentMap := make(map[string]interface{}, 0)
		if assignment.SurveyForm != nil && assignment.SurveyForm.Name != nil {
			assignmentMap["survey_form_name"] = *assignment.SurveyForm.Name
		}
		if assignment.Flow != nil && assignment.Flow.Id != nil {
			assignmentMap["flow_id"] = *assignment.Flow.Id
		}

		resourcedata.SetMapValueIfNotNil(assignmentMap, "invite_time_interval", assignment.InviteTimeInterval)
		resourcedata.SetMapValueIfNotNil(assignmentMap, "sending_user", assignment.SendingUser)
		resourcedata.SetMapValueIfNotNil(assignmentMap, "sending_domain", assignment.SendingDomain)

		surveyAssignments = append(surveyAssignments, assignmentMap)
	}
	return surveyAssignments
}

func buildArchiveRetention(archiveRetention []interface{}) *platformclientv2.Archiveretention {
	if archiveRetention == nil || len(archiveRetention) <= 0 {
		return nil
	}

	archiveRetentionMap, ok := archiveRetention[0].(map[string]interface{})
	if !ok {
		return nil
	}

	days := archiveRetentionMap["days"].(int)
	storageMedium := archiveRetentionMap["storage_medium"].(string)

	return &platformclientv2.Archiveretention{
		Days:          &days,
		StorageMedium: &storageMedium,
	}
}

func flattenArchiveRetention(archiveRetention *platformclientv2.Archiveretention) []interface{} {
	if archiveRetention == nil {
		return nil
	}

	archiveRetentionMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(archiveRetentionMap, "days", archiveRetention.Days)
	resourcedata.SetMapValueIfNotNil(archiveRetentionMap, "storage_medium", archiveRetention.StorageMedium)

	return []interface{}{archiveRetentionMap}
}

func buildDeleteRetention(deleteRetention []interface{}) *platformclientv2.Deleteretention {
	if deleteRetention == nil || len(deleteRetention) <= 0 {
		return nil
	}

	deleteRetentionMap, ok := deleteRetention[0].(map[string]interface{})
	if !ok {
		return nil
	}

	days := deleteRetentionMap["days"].(int)

	return &platformclientv2.Deleteretention{
		Days: &days,
	}
}

func flattenDeleteRetention(deleteRetention *platformclientv2.Deleteretention) []interface{} {
	if deleteRetention == nil {
		return nil
	}

	deleteRetentionMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(deleteRetentionMap, "days", deleteRetention.Days)

	return []interface{}{deleteRetentionMap}
}

func buildRetentionDuration(retentionDuration []interface{}) *platformclientv2.Retentionduration {
	if retentionDuration == nil || len(retentionDuration) <= 0 {
		return nil
	}

	retentionDurationMap, ok := retentionDuration[0].(map[string]interface{})
	if !ok {
		return nil
	}

	return &platformclientv2.Retentionduration{
		ArchiveRetention: buildArchiveRetention(retentionDurationMap["archive_retention"].([]interface{})),
		DeleteRetention:  buildDeleteRetention(retentionDurationMap["delete_retention"].([]interface{})),
	}
}

func flattenRetentionDuration(retentionDuration *platformclientv2.Retentionduration) []interface{} {
	if retentionDuration == nil {
		return nil
	}

	retentionDurationMap := make(map[string]interface{})

	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(retentionDurationMap, "archive_retention", retentionDuration.ArchiveRetention, flattenArchiveRetention)
	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(retentionDurationMap, "delete_retention", retentionDuration.DeleteRetention, flattenDeleteRetention)

	return []interface{}{retentionDurationMap}
}

func buildInitiateScreenRecording(initiateScreenRecording []interface{}) *platformclientv2.Initiatescreenrecording {
	if initiateScreenRecording == nil || len(initiateScreenRecording) <= 0 {
		return nil
	}

	initiateScreenRecordingMap, ok := initiateScreenRecording[0].(map[string]interface{})
	if !ok {
		return nil
	}
	recordACW := initiateScreenRecordingMap["record_acw"].(bool)

	return &platformclientv2.Initiatescreenrecording{
		RecordACW:        &recordACW,
		ArchiveRetention: buildArchiveRetention(initiateScreenRecordingMap["archive_retention"].([]interface{})),
		DeleteRetention:  buildDeleteRetention(initiateScreenRecordingMap["delete_retention"].([]interface{})),
	}
}

func flattenInitiateScreenRecording(recording *platformclientv2.Initiatescreenrecording) []interface{} {
	if recording == nil {
		return nil
	}

	recordingMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(recordingMap, "record_acw", recording.RecordACW)
	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(recordingMap, "archive_retention", recording.ArchiveRetention, flattenArchiveRetention)
	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(recordingMap, "delete_retention", recording.DeleteRetention, flattenDeleteRetention)

	return []interface{}{recordingMap}
}

func buildMediaTranscriptions(transcriptions []interface{}) *[]platformclientv2.Mediatranscription {
	mediaTranscriptions := make([]platformclientv2.Mediatranscription, 0)

	for _, transcription := range transcriptions {
		transcriptionMap, ok := transcription.(map[string]interface{})
		if !ok {
			continue
		}
		displayName := transcriptionMap["display_name"].(string)
		transcriptionProvider := transcriptionMap["transcription_provider"].(string)
		integrationId := transcriptionMap["integration_id"].(string)

		mediaTranscriptions = append(mediaTranscriptions, platformclientv2.Mediatranscription{
			DisplayName:           &displayName,
			TranscriptionProvider: &transcriptionProvider,
			IntegrationId:         &integrationId,
		})
	}

	return &mediaTranscriptions
}

func flattenMediaTranscriptions(transcriptions *[]platformclientv2.Mediatranscription) []interface{} {
	if transcriptions == nil {
		return nil
	}

	mediaTranscriptions := make([]interface{}, 0)

	for _, transcription := range *transcriptions {
		transcriptionMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(transcriptionMap, "display_name", transcription.DisplayName)
		resourcedata.SetMapValueIfNotNil(transcriptionMap, "transcription_provider", transcription.TranscriptionProvider)
		resourcedata.SetMapValueIfNotNil(transcriptionMap, "integration_id", transcription.IntegrationId)

		mediaTranscriptions = append(mediaTranscriptions, transcriptionMap)
	}

	return mediaTranscriptions
}

func buildIntegrationExport(integrationExport []interface{}) *platformclientv2.Integrationexport {
	if integrationExport == nil || len(integrationExport) <= 0 {
		return nil
	}

	integrationExportMap, ok := integrationExport[0].(map[string]interface{})
	if !ok {
		return nil
	}
	shouldExportScreenRecordings := integrationExportMap["should_export_screen_recordings"].(bool)

	return &platformclientv2.Integrationexport{
		Integration:                  buildDomainEntityRef(integrationExportMap["integration_id"].(string)),
		ShouldExportScreenRecordings: &shouldExportScreenRecordings,
	}
}

func flattenIntegrationExport(integrationExport *platformclientv2.Integrationexport) []interface{} {
	if integrationExport == nil {
		return nil
	}

	integrationExportMap := make(map[string]interface{})
	if integrationExport.Integration != nil {
		integrationExportMap["integration_id"] = *integrationExport.Integration.Id
	}
	resourcedata.SetMapValueIfNotNil(integrationExportMap, "should_export_screen_recordings", integrationExport.ShouldExportScreenRecordings)

	return []interface{}{integrationExportMap}
}

func buildPolicyActionsFromMediaPolicy(actions []interface{}, pp *policyProxy, ctx context.Context) *platformclientv2.Policyactions {
	if actions == nil || len(actions) <= 0 {
		return nil
	}

	actionsMap, ok := actions[0].(map[string]interface{})
	if !ok {
		return nil
	}

	retainRecording := actionsMap["retain_recording"].(bool)
	deleteRecording := actionsMap["delete_recording"].(bool)
	alwaysDelete := actionsMap["always_delete"].(bool)

	assignMeteredAssignmentByAgent := buildAssignMeteredAssignmentByAgent(actionsMap["assign_metered_assignment_by_agent"].([]interface{}), pp, ctx)

	assignMeteredEvaluations := buildAssignMeteredEvaluations(actionsMap["assign_metered_evaluations"].([]interface{}), pp, ctx)

	return &platformclientv2.Policyactions{
		RetainRecording:                &retainRecording,
		DeleteRecording:                &deleteRecording,
		AlwaysDelete:                   &alwaysDelete,
		AssignEvaluations:              buildEvaluationAssignments(actionsMap["assign_evaluations"].([]interface{}), pp, ctx),
		AssignMeteredEvaluations:       assignMeteredEvaluations,
		AssignMeteredAssignmentByAgent: assignMeteredAssignmentByAgent,
		AssignCalibrations:             buildAssignCalibrations(actionsMap["assign_calibrations"].([]interface{}), pp, ctx),
		AssignSurveys:                  buildAssignSurveys(actionsMap["assign_surveys"].([]interface{}), pp, ctx),
		RetentionDuration:              buildRetentionDuration(actionsMap["retention_duration"].([]interface{})),
		InitiateScreenRecording:        buildInitiateScreenRecording(actionsMap["initiate_screen_recording"].([]interface{})),
		MediaTranscriptions:            buildMediaTranscriptions(actionsMap["media_transcriptions"].([]interface{})),
		IntegrationExport:              buildIntegrationExport(actionsMap["integration_export"].([]interface{})),
	}
}

func flattenPolicyActions(actions *platformclientv2.Policyactions, pp *policyProxy, ctx context.Context) []interface{} {
	if actions == nil || reflect.DeepEqual(platformclientv2.Policyactions{}, *actions) {
		return nil
	}

	actionsMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(actionsMap, "retain_recording", actions.RetainRecording)
	resourcedata.SetMapValueIfNotNil(actionsMap, "delete_recording", actions.DeleteRecording)
	resourcedata.SetMapValueIfNotNil(actionsMap, "always_delete", actions.AlwaysDelete)

	if actions.AssignEvaluations != nil {
		actionsMap["assign_evaluations"] = flattenEvaluationAssignments(actions.AssignEvaluations, pp, ctx)
	}
	if actions.AssignMeteredEvaluations != nil {
		actionsMap["assign_metered_evaluations"] = flattenAssignMeteredEvaluations(actions.AssignMeteredEvaluations, pp, ctx)
	}
	if actions.AssignMeteredAssignmentByAgent != nil {
		actionsMap["assign_metered_assignment_by_agent"] = flattenAssignMeteredAssignmentByAgent(actions.AssignMeteredAssignmentByAgent, pp, ctx)
	}
	if actions.AssignCalibrations != nil {
		actionsMap["assign_calibrations"] = flattenAssignCalibrations(actions.AssignCalibrations, pp, ctx)
	}

	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(actionsMap, "assign_surveys", actions.AssignSurveys, flattenAssignSurveys)
	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(actionsMap, "retention_duration", actions.RetentionDuration, flattenRetentionDuration)
	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(actionsMap, "initiate_screen_recording", actions.InitiateScreenRecording, flattenInitiateScreenRecording)
	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(actionsMap, "media_transcriptions", actions.MediaTranscriptions, flattenMediaTranscriptions)
	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(actionsMap, "integration_export", actions.IntegrationExport, flattenIntegrationExport)

	return []interface{}{actionsMap}
}

func buildTimeSlots(slots []interface{}) *[]platformclientv2.Timeslot {
	timeSlots := make([]platformclientv2.Timeslot, 0)

	for _, slot := range slots {
		slotMap, ok := slot.(map[string]interface{})
		if !ok {
			continue
		}
		startTime := slotMap["start_time"].(string)
		stopTime := slotMap["stop_time"].(string)
		day := slotMap["day"].(int)

		timeSlots = append(timeSlots, platformclientv2.Timeslot{
			StartTime: &startTime,
			StopTime:  &stopTime,
			Day:       &day,
		})
	}

	return &timeSlots
}

func flattenTimeSlots(slots *[]platformclientv2.Timeslot) []interface{} {
	if slots == nil {
		return nil
	}

	slotList := make([]interface{}, 0)

	for _, slot := range *slots {
		slotMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(slotMap, "start_time", slot.StartTime)
		resourcedata.SetMapValueIfNotNil(slotMap, "stop_time", slot.StopTime)
		resourcedata.SetMapValueIfNotNil(slotMap, "day", slot.Day)

		slotList = append(slotList, slotMap)
	}

	return slotList
}

func buildTimeAllowed(timeAllowed []interface{}) *platformclientv2.Timeallowed {
	if timeAllowed == nil || len(timeAllowed) <= 0 {
		return nil
	}

	timeAllowedMap, ok := timeAllowed[0].(map[string]interface{})
	if !ok {
		return nil
	}

	timeZoneId := timeAllowedMap["time_zone_id"].(string)
	empty := timeAllowedMap["empty"].(bool)

	return &platformclientv2.Timeallowed{
		TimeSlots:  buildTimeSlots(timeAllowedMap["time_slots"].([]interface{})),
		TimeZoneId: &timeZoneId,
		Empty:      &empty,
	}
}

func flattenTimeAllowed(timeAllowed *platformclientv2.Timeallowed) []interface{} {
	if timeAllowed == nil {
		return nil
	}

	timeAllowedMap := make(map[string]interface{})

	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(timeAllowedMap, "time_slots", timeAllowed.TimeSlots, flattenTimeSlots)
	resourcedata.SetMapValueIfNotNil(timeAllowedMap, "time_zone_id", timeAllowed.TimeZoneId)
	resourcedata.SetMapValueIfNotNil(timeAllowedMap, "empty", timeAllowed.Empty)

	return []interface{}{timeAllowedMap}
}

func buildDurationCondition(durationCondition []interface{}) *platformclientv2.Durationcondition {
	if durationCondition == nil || len(durationCondition) <= 0 {
		return nil
	}

	durationConditionMap, ok := durationCondition[0].(map[string]interface{})
	if !ok {
		return nil
	}

	durationTarget := durationConditionMap["duration_target"].(string)
	durationOperator := durationConditionMap["duration_operator"].(string)
	durationRange := durationConditionMap["duration_range"].(string)
	durationMode := durationConditionMap["duration_mode"].(string)

	return &platformclientv2.Durationcondition{
		DurationTarget:   &durationTarget,
		DurationOperator: &durationOperator,
		DurationRange:    &durationRange,
		DurationMode:     &durationMode,
	}
}

func flattenDurationCondition(durationCondition *platformclientv2.Durationcondition) []interface{} {
	if durationCondition == nil {
		return nil
	}

	durationConditionMap := make(map[string]interface{})

	resourcedata.SetMapValueIfNotNil(durationConditionMap, "duration_target", durationCondition.DurationTarget)
	resourcedata.SetMapValueIfNotNil(durationConditionMap, "duration_operator", durationCondition.DurationOperator)
	resourcedata.SetMapValueIfNotNil(durationConditionMap, "duration_range", durationCondition.DurationRange)
	resourcedata.SetMapValueIfNotNil(durationConditionMap, "duration_mode", durationCondition.DurationMode)

	return []interface{}{durationConditionMap}
}

func buildCallMediaPolicyConditions(callMediaPolicyConditions []interface{}) *platformclientv2.Callmediapolicyconditions {
	if callMediaPolicyConditions == nil || len(callMediaPolicyConditions) <= 0 {
		return nil
	}

	conditionsMap, ok := callMediaPolicyConditions[0].(map[string]interface{})
	if !ok {
		return nil
	}

	directions := make([]string, 0)
	for _, v := range conditionsMap["directions"].([]interface{}) {
		direction := fmt.Sprintf("%v", v)
		directions = append(directions, direction)
	}

	dateRanges := make([]string, 0)
	for _, v := range conditionsMap["date_ranges"].([]interface{}) {
		dateRange := fmt.Sprintf("%v", v)
		dateRanges = append(dateRanges, dateRange)
	}

	forUserIds := conditionsMap["for_user_ids"].([]interface{})
	idStrings := make([]string, 0)
	for _, id := range forUserIds {
		idStrings = append(idStrings, fmt.Sprintf("%v", id))
	}

	forUsers := make([]platformclientv2.User, 0)
	for _, id := range idStrings {
		userId := id
		forUsers = append(forUsers, platformclientv2.User{Id: &userId})
	}

	wrapupCodeIds := conditionsMap["wrapup_code_ids"].([]interface{})
	wrapupCodeIdStrings := make([]string, 0)
	for _, id := range wrapupCodeIds {
		wrapupCodeIdStrings = append(wrapupCodeIdStrings, fmt.Sprintf("%v", id))
	}

	wrapupCodes := make([]platformclientv2.Wrapupcode, 0)
	for _, id := range wrapupCodeIdStrings {
		wrapupId := id
		wrapupCodes = append(wrapupCodes, platformclientv2.Wrapupcode{Id: &wrapupId})
	}

	languageIds := conditionsMap["language_ids"].([]interface{})
	languageIdStrings := make([]string, 0)
	for _, id := range languageIds {
		languageIdStrings = append(languageIdStrings, fmt.Sprintf("%v", id))
	}

	languages := make([]platformclientv2.Language, 0)
	for _, id := range languageIdStrings {
		languageId := id
		languages = append(languages, platformclientv2.Language{Id: &languageId})
	}

	forQueueIds := conditionsMap["for_queue_ids"].([]interface{})
	queueIdStrings := make([]string, 0)
	for _, id := range forQueueIds {
		queueIdStrings = append(queueIdStrings, fmt.Sprintf("%v", id))
	}

	forQueues := make([]platformclientv2.Queue, 0)
	for _, id := range queueIdStrings {
		queueId := id
		forQueues = append(forQueues, platformclientv2.Queue{Id: &queueId})
	}

	return &platformclientv2.Callmediapolicyconditions{
		ForUsers:    &forUsers,
		DateRanges:  &dateRanges,
		ForQueues:   &forQueues,
		WrapupCodes: &wrapupCodes,
		Languages:   &languages,
		TimeAllowed: buildTimeAllowed(conditionsMap["time_allowed"].([]interface{})),
		Directions:  &directions,
		Duration:    buildDurationCondition(conditionsMap["duration"].([]interface{})),
	}
}

func flattenCallMediaPolicyConditions(conditions *platformclientv2.Callmediapolicyconditions) []interface{} {
	if conditions == nil {
		return nil
	}

	conditionsMap := make(map[string]interface{})
	if conditions.ForUsers != nil {
		userIds := make([]string, 0)
		for _, user := range *conditions.ForUsers {
			userIds = append(userIds, *user.Id)
		}
		conditionsMap["for_user_ids"] = userIds
	}

	resourcedata.SetMapValueIfNotNil(conditionsMap, "date_ranges", conditions.DateRanges)
	resourcedata.SetMapValueIfNotNil(conditionsMap, "directions", conditions.Directions)

	if conditions.ForQueues != nil {
		queueIds := make([]string, 0)
		for _, queue := range *conditions.ForQueues {
			queueIds = append(queueIds, *queue.Id)
		}
		conditionsMap["for_queue_ids"] = queueIds
	}
	if conditions.WrapupCodes != nil {
		wrapupCodeIds := make([]string, 0)
		for _, code := range *conditions.WrapupCodes {
			wrapupCodeIds = append(wrapupCodeIds, *code.Id)
		}
		conditionsMap["wrapup_code_ids"] = wrapupCodeIds
	}
	if conditions.Languages != nil {
		languageIds := make([]string, 0)
		for _, code := range *conditions.Languages {
			languageIds = append(languageIds, *code.Id)
		}
		conditionsMap["language_ids"] = languageIds
	}

	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(conditionsMap, "time_allowed", conditions.TimeAllowed, flattenTimeAllowed)

	return []interface{}{conditionsMap}
}

func buildChatMediaPolicyConditions(chatMediaPolicyConditions []interface{}) *platformclientv2.Chatmediapolicyconditions {
	if chatMediaPolicyConditions == nil || len(chatMediaPolicyConditions) <= 0 {
		return nil
	}

	conditionsMap, ok := chatMediaPolicyConditions[0].(map[string]interface{})
	if !ok {
		return nil
	}

	dateRanges := make([]string, 0)
	for _, v := range conditionsMap["date_ranges"].([]interface{}) {
		dateRange := fmt.Sprintf("%v", v)
		dateRanges = append(dateRanges, dateRange)
	}

	forUserIds := conditionsMap["for_user_ids"].([]interface{})
	idStrings := make([]string, 0)
	for _, id := range forUserIds {
		idStrings = append(idStrings, fmt.Sprintf("%v", id))
	}

	forUsers := make([]platformclientv2.User, 0)
	for _, id := range idStrings {
		userId := id
		forUsers = append(forUsers, platformclientv2.User{Id: &userId})
	}

	wrapupCodeIds := conditionsMap["wrapup_code_ids"].([]interface{})
	wrapupCodeIdStrings := make([]string, 0)
	for _, id := range wrapupCodeIds {
		wrapupCodeIdStrings = append(wrapupCodeIdStrings, fmt.Sprintf("%v", id))
	}

	wrapupCodes := make([]platformclientv2.Wrapupcode, 0)
	for _, id := range wrapupCodeIdStrings {
		wrapupId := id
		wrapupCodes = append(wrapupCodes, platformclientv2.Wrapupcode{Id: &wrapupId})
	}

	languageIds := conditionsMap["language_ids"].([]interface{})
	languageIdStrings := make([]string, 0)
	for _, id := range languageIds {
		languageIdStrings = append(languageIdStrings, fmt.Sprintf("%v", id))
	}

	languages := make([]platformclientv2.Language, 0)
	for _, id := range languageIdStrings {
		languageId := id
		languages = append(languages, platformclientv2.Language{Id: &languageId})
	}

	forQueueIds := conditionsMap["for_queue_ids"].([]interface{})
	queueIdStrings := make([]string, 0)
	for _, id := range forQueueIds {
		queueIdStrings = append(queueIdStrings, fmt.Sprintf("%v", id))
	}

	forQueues := make([]platformclientv2.Queue, 0)
	for _, id := range queueIdStrings {
		queueId := id
		forQueues = append(forQueues, platformclientv2.Queue{Id: &queueId})
	}

	return &platformclientv2.Chatmediapolicyconditions{
		ForUsers:    &forUsers,
		DateRanges:  &dateRanges,
		ForQueues:   &forQueues,
		WrapupCodes: &wrapupCodes,
		Languages:   &languages,
		TimeAllowed: buildTimeAllowed(conditionsMap["time_allowed"].([]interface{})),
		Duration:    buildDurationCondition(conditionsMap["duration"].([]interface{})),
	}
}

func flattenChatMediaPolicyConditions(conditions *platformclientv2.Chatmediapolicyconditions) []interface{} {
	if conditions == nil {
		return nil
	}

	conditionsMap := make(map[string]interface{})
	if conditions.ForUsers != nil {
		userIds := make([]string, 0)
		for _, user := range *conditions.ForUsers {
			userIds = append(userIds, *user.Id)
		}
		conditionsMap["for_user_ids"] = userIds
	}

	resourcedata.SetMapValueIfNotNil(conditionsMap, "date_ranges", conditions.DateRanges)

	if conditions.ForQueues != nil {
		queueIds := make([]string, 0)
		for _, queue := range *conditions.ForQueues {
			queueIds = append(queueIds, *queue.Id)
		}
		conditionsMap["for_queue_ids"] = queueIds
	}
	if conditions.WrapupCodes != nil {
		wrapupCodeIds := make([]string, 0)
		for _, code := range *conditions.WrapupCodes {
			wrapupCodeIds = append(wrapupCodeIds, *code.Id)
		}
		conditionsMap["wrapup_code_ids"] = wrapupCodeIds
	}
	if conditions.Languages != nil {
		languageIds := make([]string, 0)
		for _, code := range *conditions.Languages {
			languageIds = append(languageIds, *code.Id)
		}
		conditionsMap["language_ids"] = languageIds
	}

	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(conditionsMap, "time_allowed", conditions.TimeAllowed, flattenTimeAllowed)
	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(conditionsMap, "duration", conditions.Duration, flattenDurationCondition)

	return []interface{}{conditionsMap}
}

func buildEmailMediaPolicyConditions(emailMediaPolicyConditions []interface{}) *platformclientv2.Emailmediapolicyconditions {
	if emailMediaPolicyConditions == nil || len(emailMediaPolicyConditions) <= 0 {
		return nil
	}

	conditionsMap, ok := emailMediaPolicyConditions[0].(map[string]interface{})
	if !ok {
		return nil
	}

	dateRanges := make([]string, 0)
	for _, v := range conditionsMap["date_ranges"].([]interface{}) {
		dateRange := fmt.Sprintf("%v", v)
		dateRanges = append(dateRanges, dateRange)
	}

	forUserIds := conditionsMap["for_user_ids"].([]interface{})
	idStrings := make([]string, 0)
	for _, id := range forUserIds {
		idStrings = append(idStrings, fmt.Sprintf("%v", id))
	}

	forUsers := make([]platformclientv2.User, 0)
	for _, id := range idStrings {
		userId := id
		forUsers = append(forUsers, platformclientv2.User{Id: &userId})
	}

	wrapupCodeIds := conditionsMap["wrapup_code_ids"].([]interface{})
	wrapupCodeIdStrings := make([]string, 0)
	for _, id := range wrapupCodeIds {
		wrapupCodeIdStrings = append(wrapupCodeIdStrings, fmt.Sprintf("%v", id))
	}

	wrapupCodes := make([]platformclientv2.Wrapupcode, 0)
	for _, id := range wrapupCodeIdStrings {
		wrapupId := id
		wrapupCodes = append(wrapupCodes, platformclientv2.Wrapupcode{Id: &wrapupId})
	}

	languageIds := conditionsMap["language_ids"].([]interface{})
	languageIdStrings := make([]string, 0)
	for _, id := range languageIds {
		languageIdStrings = append(languageIdStrings, fmt.Sprintf("%v", id))
	}

	languages := make([]platformclientv2.Language, 0)
	for _, id := range languageIdStrings {
		languageId := id
		languages = append(languages, platformclientv2.Language{Id: &languageId})
	}

	forQueueIds := conditionsMap["for_queue_ids"].([]interface{})
	queueIdStrings := make([]string, 0)
	for _, id := range forQueueIds {
		queueIdStrings = append(queueIdStrings, fmt.Sprintf("%v", id))
	}

	forQueues := make([]platformclientv2.Queue, 0)
	for _, id := range queueIdStrings {
		queueId := id
		forQueues = append(forQueues, platformclientv2.Queue{Id: &queueId})
	}

	return &platformclientv2.Emailmediapolicyconditions{
		ForUsers:    &forUsers,
		DateRanges:  &dateRanges,
		ForQueues:   &forQueues,
		WrapupCodes: &wrapupCodes,
		Languages:   &languages,
		TimeAllowed: buildTimeAllowed(conditionsMap["time_allowed"].([]interface{})),
	}
}

func flattenEmailMediaPolicyConditions(conditions *platformclientv2.Emailmediapolicyconditions) []interface{} {
	if conditions == nil {
		return nil
	}

	conditionsMap := make(map[string]interface{})
	if conditions.ForUsers != nil {
		userIds := make([]string, 0)
		for _, user := range *conditions.ForUsers {
			userIds = append(userIds, *user.Id)
		}
		conditionsMap["for_user_ids"] = userIds
	}

	resourcedata.SetMapValueIfNotNil(conditionsMap, "date_ranges", conditions.DateRanges)

	if conditions.ForQueues != nil {
		queueIds := make([]string, 0)
		for _, queue := range *conditions.ForQueues {
			queueIds = append(queueIds, *queue.Id)
		}
		conditionsMap["for_queue_ids"] = queueIds
	}
	if conditions.WrapupCodes != nil {
		wrapupCodeIds := make([]string, 0)
		for _, code := range *conditions.WrapupCodes {
			wrapupCodeIds = append(wrapupCodeIds, *code.Id)
		}
		conditionsMap["wrapup_code_ids"] = wrapupCodeIds
	}
	if conditions.Languages != nil {
		languageIds := make([]string, 0)
		for _, code := range *conditions.Languages {
			languageIds = append(languageIds, *code.Id)
		}
		conditionsMap["language_ids"] = languageIds
	}

	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(conditionsMap, "time_allowed", conditions.TimeAllowed, flattenTimeAllowed)

	return []interface{}{conditionsMap}
}

func buildMessageMediaPolicyConditions(messageMediaPolicyConditions []interface{}) *platformclientv2.Messagemediapolicyconditions {
	if messageMediaPolicyConditions == nil || len(messageMediaPolicyConditions) <= 0 {
		return nil
	}

	conditionsMap, ok := messageMediaPolicyConditions[0].(map[string]interface{})
	if !ok {
		return nil
	}

	dateRanges := make([]string, 0)
	for _, v := range conditionsMap["date_ranges"].([]interface{}) {
		dateRange := fmt.Sprintf("%v", v)
		dateRanges = append(dateRanges, dateRange)
	}

	forUserIds := conditionsMap["for_user_ids"].([]interface{})
	idStrings := make([]string, 0)
	for _, id := range forUserIds {
		idStrings = append(idStrings, fmt.Sprintf("%v", id))
	}

	forUsers := make([]platformclientv2.User, 0)
	for _, id := range idStrings {
		userId := id
		forUsers = append(forUsers, platformclientv2.User{Id: &userId})
	}

	wrapupCodeIds := conditionsMap["wrapup_code_ids"].([]interface{})
	wrapupCodeIdStrings := make([]string, 0)
	for _, id := range wrapupCodeIds {
		wrapupCodeIdStrings = append(wrapupCodeIdStrings, fmt.Sprintf("%v", id))
	}

	wrapupCodes := make([]platformclientv2.Wrapupcode, 0)
	for _, id := range wrapupCodeIdStrings {
		wrapupId := id
		wrapupCodes = append(wrapupCodes, platformclientv2.Wrapupcode{Id: &wrapupId})
	}

	languageIds := conditionsMap["language_ids"].([]interface{})
	languageIdStrings := make([]string, 0)
	for _, id := range languageIds {
		languageIdStrings = append(languageIdStrings, fmt.Sprintf("%v", id))
	}

	languages := make([]platformclientv2.Language, 0)
	for _, id := range languageIdStrings {
		languageId := id
		languages = append(languages, platformclientv2.Language{Id: &languageId})
	}

	forQueueIds := conditionsMap["for_queue_ids"].([]interface{})
	queueIdStrings := make([]string, 0)
	for _, id := range forQueueIds {
		queueIdStrings = append(queueIdStrings, fmt.Sprintf("%v", id))
	}

	forQueues := make([]platformclientv2.Queue, 0)
	for _, id := range queueIdStrings {
		queueId := id
		forQueues = append(forQueues, platformclientv2.Queue{Id: &queueId})
	}

	return &platformclientv2.Messagemediapolicyconditions{
		ForUsers:    &forUsers,
		DateRanges:  &dateRanges,
		ForQueues:   &forQueues,
		WrapupCodes: &wrapupCodes,
		Languages:   &languages,
		TimeAllowed: buildTimeAllowed(conditionsMap["time_allowed"].([]interface{})),
	}
}

func flattenMessageMediaPolicyConditions(conditions *platformclientv2.Messagemediapolicyconditions) []interface{} {
	if conditions == nil {
		return nil
	}

	conditionsMap := make(map[string]interface{})
	if conditions.ForUsers != nil {
		userIds := make([]string, 0)
		for _, user := range *conditions.ForUsers {
			userIds = append(userIds, *user.Id)
		}
		conditionsMap["for_user_ids"] = userIds
	}

	resourcedata.SetMapValueIfNotNil(conditionsMap, "date_ranges", conditions.DateRanges)

	if conditions.ForQueues != nil {
		queueIds := make([]string, 0)
		for _, queue := range *conditions.ForQueues {
			queueIds = append(queueIds, *queue.Id)
		}
		conditionsMap["for_queue_ids"] = queueIds
	}
	if conditions.WrapupCodes != nil {
		wrapupCodeIds := make([]string, 0)
		for _, code := range *conditions.WrapupCodes {
			wrapupCodeIds = append(wrapupCodeIds, *code.Id)
		}
		conditionsMap["wrapup_code_ids"] = wrapupCodeIds
	}
	if conditions.Languages != nil {
		languageIds := make([]string, 0)
		for _, code := range *conditions.Languages {
			languageIds = append(languageIds, *code.Id)
		}
		conditionsMap["language_ids"] = languageIds
	}

	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(conditionsMap, "time_allowed", conditions.TimeAllowed, flattenTimeAllowed)

	return []interface{}{conditionsMap}
}

func buildCallMediaPolicy(callMediaPolicy []interface{}, pp *policyProxy, ctx context.Context) *platformclientv2.Callmediapolicy {
	if callMediaPolicy == nil || len(callMediaPolicy) <= 0 {
		return nil
	}

	policyMap, ok := callMediaPolicy[0].(map[string]interface{})
	if !ok {
		return nil
	}
	actions := buildPolicyActionsFromMediaPolicy(policyMap["actions"].([]interface{}), pp, ctx)

	return &platformclientv2.Callmediapolicy{
		Actions:    actions,
		Conditions: buildCallMediaPolicyConditions(policyMap["conditions"].([]interface{})),
	}
}

func flattenCallMediaPolicy(chatMediaPolicy *platformclientv2.Callmediapolicy, pp *policyProxy, ctx context.Context) []interface{} {
	if chatMediaPolicy == nil {
		return nil
	}

	chatMediaPolicyMap := make(map[string]interface{})
	if chatMediaPolicy.Actions != nil {
		chatMediaPolicyMap["actions"] = flattenPolicyActions(chatMediaPolicy.Actions, pp, ctx)
	}

	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(chatMediaPolicyMap, "conditions", chatMediaPolicy.Conditions, flattenCallMediaPolicyConditions)

	return []interface{}{chatMediaPolicyMap}
}

func buildChatMediaPolicy(chatMediaPolicy []interface{}, pp *policyProxy, ctx context.Context) *platformclientv2.Chatmediapolicy {
	if chatMediaPolicy == nil || len(chatMediaPolicy) <= 0 {
		return nil
	}

	policyMap, ok := chatMediaPolicy[0].(map[string]interface{})
	if !ok {
		return nil
	}

	actions := buildPolicyActionsFromMediaPolicy(policyMap["actions"].([]interface{}), pp, ctx)

	return &platformclientv2.Chatmediapolicy{
		Actions:    actions,
		Conditions: buildChatMediaPolicyConditions(policyMap["conditions"].([]interface{})),
	}
}

func flattenChatMediaPolicy(chatMediaPolicy *platformclientv2.Chatmediapolicy, pp *policyProxy, ctx context.Context) []interface{} {
	if chatMediaPolicy == nil {
		return nil
	}

	chatMediaPolicyMap := make(map[string]interface{})
	if chatMediaPolicy.Actions != nil {
		chatMediaPolicyMap["actions"] = flattenPolicyActions(chatMediaPolicy.Actions, pp, ctx)
	}

	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(chatMediaPolicyMap, "conditions", chatMediaPolicy.Conditions, flattenChatMediaPolicyConditions)

	return []interface{}{chatMediaPolicyMap}
}

func buildEmailMediaPolicy(emailMediaPolicy []interface{}, pp *policyProxy, ctx context.Context) *platformclientv2.Emailmediapolicy {
	if emailMediaPolicy == nil || len(emailMediaPolicy) <= 0 {
		return nil
	}

	policyMap, ok := emailMediaPolicy[0].(map[string]interface{})
	if !ok {
		return nil
	}

	actions := buildPolicyActionsFromMediaPolicy(policyMap["actions"].([]interface{}), pp, ctx)

	return &platformclientv2.Emailmediapolicy{
		Actions:    actions,
		Conditions: buildEmailMediaPolicyConditions(policyMap["conditions"].([]interface{})),
	}
}

func flattenEmailMediaPolicy(emailMediaPolicy *platformclientv2.Emailmediapolicy, pp *policyProxy, ctx context.Context) []interface{} {
	if emailMediaPolicy == nil {
		return nil
	}

	emailMediaPolicyMap := make(map[string]interface{})
	if emailMediaPolicy.Actions != nil {
		emailMediaPolicyMap["actions"] = flattenPolicyActions(emailMediaPolicy.Actions, pp, ctx)
	}

	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(emailMediaPolicyMap, "conditions", emailMediaPolicy.Conditions, flattenEmailMediaPolicyConditions)

	return []interface{}{emailMediaPolicyMap}
}

func buildMessageMediaPolicy(messageMediaPolicy []interface{}, pp *policyProxy, ctx context.Context) *platformclientv2.Messagemediapolicy {
	if messageMediaPolicy == nil || len(messageMediaPolicy) <= 0 {
		return nil
	}

	policyMap, ok := messageMediaPolicy[0].(map[string]interface{})
	if !ok {
		return nil
	}

	actions := buildPolicyActionsFromMediaPolicy(policyMap["actions"].([]interface{}), pp, ctx)

	return &platformclientv2.Messagemediapolicy{
		Actions:    actions,
		Conditions: buildMessageMediaPolicyConditions(policyMap["conditions"].([]interface{})),
	}
}

func flattenMessageMediaPolicy(messageMediaPolicy *platformclientv2.Messagemediapolicy, pp *policyProxy, ctx context.Context) []interface{} {
	if messageMediaPolicy == nil {
		return nil
	}

	messageMediaPolicyMap := make(map[string]interface{})
	if messageMediaPolicy.Actions != nil {
		messageMediaPolicyMap["actions"] = flattenPolicyActions(messageMediaPolicy.Actions, pp, ctx)
	}

	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(messageMediaPolicyMap, "conditions", messageMediaPolicy.Conditions, flattenMessageMediaPolicyConditions)

	return []interface{}{messageMediaPolicyMap}
}

func buildMediaPolicies(d *schema.ResourceData, pp *policyProxy, ctx context.Context) *platformclientv2.Mediapolicies {
	sdkMediaPolicies := platformclientv2.Mediapolicies{}

	if mediaPolicies, ok := d.Get("media_policies").([]interface{}); ok && len(mediaPolicies) > 0 {
		mediaPoliciesMap, ok := mediaPolicies[0].(map[string]interface{})
		if !ok {
			return nil
		}

		if callPolicy := mediaPoliciesMap["call_policy"]; callPolicy != nil {
			sdkMediaPolicies.CallPolicy = buildCallMediaPolicy(callPolicy.([]interface{}), pp, ctx)
		}

		if chatPolicy := mediaPoliciesMap["chat_policy"]; chatPolicy != nil {
			sdkMediaPolicies.ChatPolicy = buildChatMediaPolicy(chatPolicy.([]interface{}), pp, ctx)
		}

		if emailPolicy := mediaPoliciesMap["email_policy"]; emailPolicy != nil {
			sdkMediaPolicies.EmailPolicy = buildEmailMediaPolicy(emailPolicy.([]interface{}), pp, ctx)
		}

		if messagePolicy := mediaPoliciesMap["message_policy"]; messagePolicy != nil {
			sdkMediaPolicies.MessagePolicy = buildMessageMediaPolicy(messagePolicy.([]interface{}), pp, ctx)
		}
	}
	return &sdkMediaPolicies
}

func flattenMediaPolicies(mediaPolicies *platformclientv2.Mediapolicies, pp *policyProxy, ctx context.Context) []interface{} {
	if mediaPolicies == nil {
		return nil
	}

	mediaPoliciesMap := make(map[string]interface{})
	if mediaPolicies.CallPolicy != nil {
		mediaPoliciesMap["call_policy"] = flattenCallMediaPolicy(mediaPolicies.CallPolicy, pp, ctx)
	}
	if mediaPolicies.ChatPolicy != nil {
		mediaPoliciesMap["chat_policy"] = flattenChatMediaPolicy(mediaPolicies.ChatPolicy, pp, ctx)
	}
	if mediaPolicies.EmailPolicy != nil {
		mediaPoliciesMap["email_policy"] = flattenEmailMediaPolicy(mediaPolicies.EmailPolicy, pp, ctx)
	}
	if mediaPolicies.MessagePolicy != nil {
		mediaPoliciesMap["message_policy"] = flattenMessageMediaPolicy(mediaPolicies.MessagePolicy, pp, ctx)
	}

	return []interface{}{mediaPoliciesMap}
}

func buildConditions(d *schema.ResourceData) *platformclientv2.Policyconditions {
	if conditions, ok := d.Get("conditions").([]interface{}); ok && len(conditions) > 0 {
		conditionsMap, ok := conditions[0].(map[string]interface{})
		if !ok {
			return nil
		}

		directions := make([]string, 0)
		for _, v := range conditionsMap["directions"].([]interface{}) {
			direction := fmt.Sprintf("%v", v)
			directions = append(directions, direction)
		}

		dateRanges := make([]string, 0)
		for _, v := range conditionsMap["date_ranges"].([]interface{}) {
			dateRange := fmt.Sprintf("%v", v)
			dateRanges = append(dateRanges, dateRange)
		}

		mediaTypes := make([]string, 0)
		for _, v := range conditionsMap["media_types"].([]interface{}) {
			mediaType := fmt.Sprintf("%v", v)
			mediaTypes = append(mediaTypes, mediaType)
		}

		forUserIds := conditionsMap["for_user_ids"].([]interface{})
		idStrings := make([]string, 0)
		for _, id := range forUserIds {
			idStrings = append(idStrings, fmt.Sprintf("%v", id))
		}

		forUsers := make([]platformclientv2.User, 0)
		for _, id := range idStrings {
			userId := id
			forUsers = append(forUsers, platformclientv2.User{Id: &userId})
		}

		wrapupCodeIds := conditionsMap["wrapup_code_ids"].([]interface{})
		wrapupCodeIdStrings := make([]string, 0)
		for _, id := range wrapupCodeIds {
			wrapupCodeIdStrings = append(wrapupCodeIdStrings, fmt.Sprintf("%v", id))
		}

		wrapupCodes := make([]platformclientv2.Wrapupcode, 0)
		for _, id := range wrapupCodeIdStrings {
			wrapupId := id
			wrapupCodes = append(wrapupCodes, platformclientv2.Wrapupcode{Id: &wrapupId})
		}

		forQueueIds := conditionsMap["for_queue_ids"].([]interface{})
		queueIdStrings := make([]string, 0)
		for _, id := range forQueueIds {
			queueIdStrings = append(queueIdStrings, fmt.Sprintf("%v", id))
		}

		forQueues := make([]platformclientv2.Queue, 0)
		for _, id := range queueIdStrings {
			queueId := id
			forQueues = append(forQueues, platformclientv2.Queue{Id: &queueId})
		}

		return &platformclientv2.Policyconditions{
			ForUsers:    &forUsers,
			Directions:  &directions,
			DateRanges:  &dateRanges,
			MediaTypes:  &mediaTypes,
			ForQueues:   &forQueues,
			Duration:    buildDurationCondition(conditionsMap["duration"].([]interface{})),
			WrapupCodes: &wrapupCodes,
			TimeAllowed: buildTimeAllowed(conditionsMap["time_allowed"].([]interface{})),
		}
	}

	return nil
}

func flattenConditions(conditions *platformclientv2.Policyconditions) []interface{} {
	if conditions == nil || reflect.DeepEqual(platformclientv2.Policyconditions{}, *conditions) {
		return nil
	}

	conditionsMap := make(map[string]interface{})
	if conditions.ForUsers != nil {
		userIds := make([]string, 0)
		for _, user := range *conditions.ForUsers {
			userIds = append(userIds, *user.Id)
		}
		conditionsMap["for_user_ids"] = userIds
	}
	resourcedata.SetMapValueIfNotNil(conditionsMap, "directions", conditions.Directions)
	resourcedata.SetMapValueIfNotNil(conditionsMap, "date_ranges", conditions.DateRanges)
	resourcedata.SetMapValueIfNotNil(conditionsMap, "media_types", conditions.MediaTypes)

	if conditions.ForQueues != nil {
		queueIds := make([]string, 0)
		for _, queue := range *conditions.ForQueues {
			queueIds = append(queueIds, *queue.Id)
		}
		conditionsMap["for_queue_ids"] = queueIds
	}

	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(conditionsMap, "duration", conditions.Duration, flattenDurationCondition)

	if conditions.WrapupCodes != nil {
		wrapupCodeIds := make([]string, 0)
		for _, code := range *conditions.WrapupCodes {
			wrapupCodeIds = append(wrapupCodeIds, *code.Id)
		}
		conditionsMap["wrapup_code_ids"] = wrapupCodeIds
	}

	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(conditionsMap, "time_allowed", conditions.TimeAllowed, flattenTimeAllowed)

	return []interface{}{conditionsMap}
}

func buildPolicyActionsFromResource(d *schema.ResourceData, pp *policyProxy, ctx context.Context) *platformclientv2.Policyactions {
	if actions, ok := d.Get("actions").([]interface{}); ok && len(actions) > 0 {
		actionsMap, ok := actions[0].(map[string]interface{})
		if !ok {
			return nil
		}

		retainRecording := actionsMap["retain_recording"].(bool)
		deleteRecording := actionsMap["delete_recording"].(bool)
		alwaysDelete := actionsMap["always_delete"].(bool)

		meteredAssignmentByAgent := buildAssignMeteredAssignmentByAgent(actionsMap["assign_metered_assignment_by_agent"].([]interface{}), pp, ctx)

		assignMeteredEvaluations := buildAssignMeteredEvaluations(actionsMap["assign_metered_evaluations"].([]interface{}), pp, ctx)

		return &platformclientv2.Policyactions{
			RetainRecording:                &retainRecording,
			DeleteRecording:                &deleteRecording,
			AlwaysDelete:                   &alwaysDelete,
			AssignEvaluations:              buildEvaluationAssignments(actionsMap["assign_evaluations"].([]interface{}), pp, ctx),
			AssignMeteredEvaluations:       assignMeteredEvaluations,
			AssignMeteredAssignmentByAgent: meteredAssignmentByAgent,
			AssignCalibrations:             buildAssignCalibrations(actionsMap["assign_calibrations"].([]interface{}), pp, ctx),
			AssignSurveys:                  buildAssignSurveys(actionsMap["assign_surveys"].([]interface{}), pp, ctx),
			RetentionDuration:              buildRetentionDuration(actionsMap["retention_duration"].([]interface{})),
			InitiateScreenRecording:        buildInitiateScreenRecording(actionsMap["initiate_screen_recording"].([]interface{})),
			MediaTranscriptions:            buildMediaTranscriptions(actionsMap["media_transcriptions"].([]interface{})),
			IntegrationExport:              buildIntegrationExport(actionsMap["integration_export"].([]interface{})),
		}
	}
	return nil
}

func buildUserParams(params []interface{}) *[]platformclientv2.Userparam {
	userParams := make([]platformclientv2.Userparam, 0)

	for _, param := range params {
		paramMap, ok := param.(map[string]interface{})
		if !ok {
			continue
		}
		key := paramMap["key"].(string)
		value := paramMap["value"].(string)

		userParams = append(userParams, platformclientv2.Userparam{
			Key:   &key,
			Value: &value,
		})
	}

	return &userParams
}

func flattenUserParams(params *[]platformclientv2.Userparam) []interface{} {
	if params == nil {
		return nil
	}

	paramList := make([]interface{}, 0)

	for _, param := range *params {
		paramMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(paramMap, "key", param.Key)
		resourcedata.SetMapValueIfNotNil(paramMap, "value", param.Value)

		paramList = append(paramList, paramMap)
	}

	return paramList
}

func buildPolicyErrorMessages(messages []interface{}) *[]platformclientv2.Policyerrormessage {
	policyErrorMessages := make([]platformclientv2.Policyerrormessage, 0)

	for _, message := range messages {
		messageMap, ok := message.(map[string]interface{})
		if !ok {
			continue
		}
		statusCode := messageMap["status_code"].(int)
		userMessage := messageMap["user_message"]
		userParamsMessage := messageMap["user_params_message"].(string)
		errorCode := messageMap["error_code"].(string)
		correlationId := messageMap["correlation_id"].(string)
		insertDateString := messageMap["insert_date"].(string)

		temp := platformclientv2.Policyerrormessage{
			StatusCode:        &statusCode,
			UserMessage:       &userMessage,
			UserParamsMessage: &userParamsMessage,
			ErrorCode:         &errorCode,
			CorrelationId:     &correlationId,
			UserParams:        buildUserParams(messageMap["user_params"].([]interface{})),
		}

		insertDate, insertErr := time.Parse("2006-01-02T15:04:05-0700", insertDateString)
		if insertErr == nil {
			temp.InsertDate = &insertDate
		}

		policyErrorMessages = append(policyErrorMessages, temp)
	}

	return &policyErrorMessages
}

func flattenPolicyErrorMessages(errorMessages *[]platformclientv2.Policyerrormessage) []interface{} {
	if errorMessages == nil {
		return nil
	}

	errorMessageList := make([]interface{}, 0)

	for _, errorMessage := range *errorMessages {
		errorMessageMap := make(map[string]interface{})

		resourcedata.SetMapValueIfNotNil(errorMessageMap, "status_code", errorMessage.StatusCode)
		resourcedata.SetMapValueIfNotNil(errorMessageMap, "user_message", errorMessage.UserMessage)
		resourcedata.SetMapValueIfNotNil(errorMessageMap, "user_params_message", errorMessage.UserParamsMessage)
		resourcedata.SetMapValueIfNotNil(errorMessageMap, "error_code", errorMessage.ErrorCode)
		resourcedata.SetMapValueIfNotNil(errorMessageMap, "correlation_id", errorMessage.CorrelationId)
		if errorMessage.InsertDate != nil && len(errorMessage.InsertDate.String()) > 0 {
			temp := *errorMessage.InsertDate
			errorMessageMap["insert_date"] = temp.String()
		}
		resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(errorMessageMap, "user_params", errorMessage.UserParams, flattenUserParams)

		errorMessageList = append(errorMessageList, errorMessageMap)
	}

	return errorMessageList
}

func buildPolicyErrors(d *schema.ResourceData) *platformclientv2.Policyerrors {
	if errors, ok := d.GetOk("policy_errors"); ok {
		if errorsList, ok := errors.([]interface{}); ok || len(errorsList) > 0 {
			errorsMap, ok := errorsList[0].(map[string]interface{})
			if !ok {
				return nil
			}
			return &platformclientv2.Policyerrors{
				PolicyErrorMessages: buildPolicyErrorMessages(errorsMap["policy_error_messages"].([]interface{})),
			}
		}
	}

	return nil
}

func flattenPolicyErrors(policyErrors *platformclientv2.Policyerrors) []interface{} {
	if policyErrors == nil {
		return nil
	}

	policyErrorsMap := make(map[string]interface{})

	resourcedata.SetMapInterfaceArrayWithFuncIfNotNil(policyErrorsMap, "policy_error_messages", policyErrors.PolicyErrorMessages, flattenPolicyErrorMessages)

	return []interface{}{policyErrorsMap}
}
