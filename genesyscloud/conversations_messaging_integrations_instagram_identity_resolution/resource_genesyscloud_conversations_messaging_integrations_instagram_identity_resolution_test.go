package conversations_messaging_integrations_instagram_identity_resolution

// Acceptance tests for messaging IR resources require an Active parent integration.
//
// only Open IR Acc is implemented in conversations_messaging_integrations_open_identity_resolution
// (mock outbound webhook URL + poll until status Active, then IR lifecycle).
//
// Instagram / WhatsApp / Apple IR Acc is not enabled in this repo's default Acc org because:
//   - Instagram: parent needs a valid Meta page access token (fake UUID → not Active → IR PUT 409).
//   - WhatsApp: parent typically needs real activate_whatsapp (phone/pin) / Meta setup, not Open webhook mock.
//   - Apple: parent needs Apple Messages for Business credentials; no equivalent internal mock URL.
//
// Coverage: unit tests + examples/docs.
