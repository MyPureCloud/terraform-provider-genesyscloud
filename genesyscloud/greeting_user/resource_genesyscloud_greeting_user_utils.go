package greeting_user

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v176/platformclientv2"
)

func getUserGreetingFromResourceData(d *schema.ResourceData) platformclientv2.Greeting {
	greeting := platformclientv2.Greeting{
		Name:      platformclientv2.String(d.Get("name").(string)),
		VarType:   platformclientv2.String(d.Get("type").(string)),
		OwnerType: platformclientv2.String(d.Get("owner_type").(string)),
		AudioFile: buildAudioFile(d.Get("audio_file").([]interface{})),
		Owner: &platformclientv2.Domainentity{
			Id: platformclientv2.String(d.Get("user_id").(string)),
		},
	}

	// audio_tts is optional, only set if provided
	if audioTts, ok := d.GetOk("audio_tts"); ok {
		greeting.AudioTTS = platformclientv2.String(audioTts.(string))
	}

	return greeting
}

func buildAudioFile(body []interface{}) *platformclientv2.Greetingaudiofile {
	var audioFile platformclientv2.Greetingaudiofile

	if len(body) == 0 {
		return nil
	}

	audioFileMap := body[0].(map[string]interface{})
	if durationMilliseconds, ok := audioFileMap["duration_milliseconds"].(int); ok {
		audioFile.DurationMilliseconds = platformclientv2.Int(durationMilliseconds)
	}
	if sizeBytes, ok := audioFileMap["size_bytes"].(int); ok {
		audioFile.SizeBytes = platformclientv2.Int(sizeBytes)
	}
	if selfUri, ok := audioFileMap["self_uri"].(string); ok {
		audioFile.SelfUri = platformclientv2.String(selfUri)
	}

	return &audioFile
}

func flattenAudioFile(audioFile *platformclientv2.Greetingaudiofile) []map[string]interface{} {
	if audioFile == nil {
		return []map[string]interface{}{}
	}

	audioFileMap := make(map[string]interface{})
	if audioFile.DurationMilliseconds != nil {
		audioFileMap["duration_milliseconds"] = *audioFile.DurationMilliseconds
	}
	if audioFile.SizeBytes != nil {
		audioFileMap["size_bytes"] = *audioFile.SizeBytes
	}
	if audioFile.SelfUri != nil {
		audioFileMap["self_uri"] = *audioFile.SelfUri
	}

	return []map[string]interface{}{audioFileMap}
}

func timeToString(t *time.Time) *string {
	if t == nil {
		return nil
	}
	str := t.Format(time.RFC3339)
	return &str
}

func parseStringToTime(dateStr string) *time.Time {
	if dateStr == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return &t
	}
	return nil
}

func extractDomainEntityName(entity *platformclientv2.Domainentityref) *string {
	if entity == nil {
		return nil
	}
	return entity.Name
}

func GenerateGreetingUser(
	resourceLabel string,
	name string,
	greetingType string,
	ownerType string,
	user string,
	audioTts string,
) string {
	return fmt.Sprintf(`resource "genesyscloud_greeting_user" "%s" {
  name        = "%s"
  type        = "%s"
  owner_type  = "%s"
  user_id    = %s
  audio_tts   = "%s"
}`, resourceLabel, name, greetingType, ownerType, user, audioTts)
}
