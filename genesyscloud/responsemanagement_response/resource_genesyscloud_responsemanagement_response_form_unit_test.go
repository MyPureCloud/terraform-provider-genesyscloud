package responsemanagement_response

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
Unit tests for the `form` attribute of genesyscloud_responsemanagement_response.

The fixture below mirrors the Response Management API contract example for a Form response
(POST /api/v2/responsemanagement/responses with responseType "Form"), covering all four
component types: ListPicker, DatePicker, Input and WheelPicker.
*/

// testFormConfig returns a fully populated `form` block as Terraform would hand it to us.
func testFormConfig() []interface{} {
	return []interface{}{
		map[string]interface{}{
			"form_description": "A test form for customer feedback",
			"show_summary":     true,
			"received_message": []interface{}{
				map[string]interface{}{
					"title":     "Thank you for your feedback",
					"subtitle":  "We appreciate your input",
					"image_url": "https://example.com/received.png",
				},
			},
			"reply_message": []interface{}{
				map[string]interface{}{
					"title":    "Your response has been received",
					"subtitle": "We will review your feedback",
				},
			},
			"introduction": []interface{}{
				map[string]interface{}{
					"title":       "Customer Feedback Form",
					"subtitle":    "Please help us improve our service",
					"button_text": "Start Survey",
				},
			},
			"form_pages": []interface{}{
				map[string]interface{}{
					"title":    "Service Rating",
					"subtitle": "How would you rate our service?",
					"page_components": []interface{}{
						map[string]interface{}{
							"form_component_type": "ListPicker",
							"list_picker": []interface{}{
								map[string]interface{}{
									"sections": []interface{}{
										map[string]interface{}{
											"title":              "Select your rating",
											"multiple_selection": false,
											"items": []interface{}{
												map[string]interface{}{"title": "Excellent"},
												map[string]interface{}{"title": "Good"},
												map[string]interface{}{"title": "Poor"},
											},
										},
									},
								},
							},
						},
					},
				},
				map[string]interface{}{
					"title":    "Visit Date",
					"subtitle": "When did you visit us?",
					"page_components": []interface{}{
						map[string]interface{}{
							"form_component_type": "DatePicker",
							"date_picker": []interface{}{
								map[string]interface{}{
									"title":               "Select Date",
									"subtitle":            "Choose the date of your visit",
									"date_display_format": "dayMonthYear",
								},
							},
						},
					},
				},
				map[string]interface{}{
					"title":    "Additional Comments",
					"subtitle": "Please share any additional feedback",
					"page_components": []interface{}{
						map[string]interface{}{
							"form_component_type": "Input",
							"input": []interface{}{
								map[string]interface{}{
									"title":              "Comments",
									"subtitle":           "Your feedback helps us improve",
									"placeholder_text":   "Enter your comments here...",
									"is_multiple_line":   true,
									"is_required":        false,
									"keyboard_type":      "Default",
									"auto_complete_type": "Name",
								},
							},
						},
					},
				},
				map[string]interface{}{
					"title":    "Recommendation",
					"subtitle": "Would you recommend us to others?",
					"page_components": []interface{}{
						map[string]interface{}{
							"form_component_type": "WheelPicker",
							"wheel_picker": []interface{}{
								map[string]interface{}{
									"items": []interface{}{
										map[string]interface{}{"title": "Definitely", "value": "definitely"},
										map[string]interface{}{"title": "Maybe", "value": "maybe"},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// TestUnitBuildFormMapsAllComponentTypes verifies buildForm maps every part of the `form` block
// onto the SDK Form object, including ordering and booleans that are explicitly false.
func TestUnitBuildFormMapsAllComponentTypes(t *testing.T) {
	sdkForm := buildForm(testFormConfig())
	if sdkForm == nil {
		t.Fatal("expected buildForm to return a Form, got nil")
	}

	if sdkForm.FormDescription == nil || *sdkForm.FormDescription != "A test form for customer feedback" {
		t.Errorf("form_description not mapped, got %v", sdkForm.FormDescription)
	}
	if sdkForm.ShowSummary == nil || *sdkForm.ShowSummary != true {
		t.Errorf("show_summary not mapped, got %v", sdkForm.ShowSummary)
	}

	if sdkForm.ReceivedMessage == nil {
		t.Fatal("received_message not mapped")
	}
	if *sdkForm.ReceivedMessage.Title != "Thank you for your feedback" {
		t.Errorf("received_message.title = %q", *sdkForm.ReceivedMessage.Title)
	}
	if *sdkForm.ReceivedMessage.ImageUrl != "https://example.com/received.png" {
		t.Errorf("received_message.image_url = %q", *sdkForm.ReceivedMessage.ImageUrl)
	}
	if sdkForm.ReplyMessage == nil || *sdkForm.ReplyMessage.Subtitle != "We will review your feedback" {
		t.Error("reply_message not mapped")
	}
	if sdkForm.Introduction == nil || *sdkForm.Introduction.ButtonText != "Start Survey" {
		t.Error("introduction not mapped")
	}

	if sdkForm.FormPages == nil || len(*sdkForm.FormPages) != 4 {
		t.Fatalf("expected 4 form pages, got %v", sdkForm.FormPages)
	}
	pages := *sdkForm.FormPages

	if *pages[0].Title != "Service Rating" || *pages[0].Subtitle != "How would you rate our service?" {
		t.Error("page title/subtitle not mapped")
	}

	// ListPicker
	listPickerComponent := (*pages[0].PageComponents)[0]
	if *listPickerComponent.FormComponentType != "ListPicker" {
		t.Errorf("form_component_type = %q", *listPickerComponent.FormComponentType)
	}
	if listPickerComponent.ListPicker == nil {
		t.Fatal("list_picker not mapped")
	}
	sections := *listPickerComponent.ListPicker.Sections
	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}
	if sections[0].MultipleSelection == nil || *sections[0].MultipleSelection != false {
		t.Error("multiple_selection=false must be sent, not dropped as a zero value")
	}
	items := *sections[0].Items
	if len(items) != 3 || *items[0].Title != "Excellent" || *items[2].Title != "Poor" {
		t.Errorf("list picker items not mapped in order: %v", items)
	}
	// Components the config did not set must stay nil so they are omitted from the payload.
	if listPickerComponent.DatePicker != nil || listPickerComponent.Input != nil || listPickerComponent.WheelPicker != nil {
		t.Error("unset component blocks should remain nil")
	}

	// DatePicker
	datePickerComponent := (*pages[1].PageComponents)[0]
	if datePickerComponent.DatePicker == nil {
		t.Fatal("date_picker not mapped")
	}
	if *datePickerComponent.DatePicker.DateDisplayFormat != "dayMonthYear" {
		t.Errorf("date_display_format = %q", *datePickerComponent.DatePicker.DateDisplayFormat)
	}

	// Input
	inputComponent := (*pages[2].PageComponents)[0]
	if inputComponent.Input == nil {
		t.Fatal("input not mapped")
	}
	if inputComponent.Input.IsMultipleLine == nil || *inputComponent.Input.IsMultipleLine != true {
		t.Error("is_multiple_line not mapped")
	}
	if inputComponent.Input.IsRequired == nil || *inputComponent.Input.IsRequired != false {
		t.Error("is_required=false must be sent, not dropped as a zero value")
	}
	if *inputComponent.Input.KeyboardType != "Default" || *inputComponent.Input.AutoCompleteType != "Name" {
		t.Error("input keyboard_type/auto_complete_type not mapped")
	}
	if *inputComponent.Input.PlaceholderText != "Enter your comments here..." {
		t.Error("placeholder_text not mapped")
	}

	// WheelPicker
	wheelPickerComponent := (*pages[3].PageComponents)[0]
	if wheelPickerComponent.WheelPicker == nil {
		t.Fatal("wheel_picker not mapped")
	}
	wheelItems := *wheelPickerComponent.WheelPicker.Items
	if len(wheelItems) != 2 || *wheelItems[0].Value != "definitely" || *wheelItems[1].Title != "Maybe" {
		t.Errorf("wheel picker items not mapped in order: %v", wheelItems)
	}
}

// TestUnitFormRoundTripProducesNoDiff guards against perpetual Terraform diffs: flattening what we
// built must reproduce the original configuration exactly.
func TestUnitFormRoundTripProducesNoDiff(t *testing.T) {
	original := testFormConfig()

	roundTripped := flattenForm(buildForm(original))

	if !reflect.DeepEqual(original, roundTripped) {
		t.Errorf("form did not survive a build/flatten round trip.\noriginal:     %#v\nroundTripped: %#v", original, roundTripped)
	}
}

// TestUnitFormOptionalAttributesOmitted verifies a minimal form (no introduction, no optional
// leaf values) does not fabricate empty attributes.
func TestUnitFormOptionalAttributesOmitted(t *testing.T) {
	minimal := []interface{}{
		map[string]interface{}{
			"form_description": "Minimal",
			"show_summary":     false,
			"received_message": []interface{}{map[string]interface{}{"title": "Got it"}},
			"reply_message":    []interface{}{map[string]interface{}{"title": "Thanks"}},
			"form_pages": []interface{}{
				map[string]interface{}{
					"title":    "Page",
					"subtitle": "Sub",
					"page_components": []interface{}{
						map[string]interface{}{
							"form_component_type": "Input",
							"input": []interface{}{
								map[string]interface{}{
									"is_multiple_line": false,
									"is_required":      true,
								},
							},
						},
					},
				},
			},
		},
	}

	sdkForm := buildForm(minimal)
	if sdkForm == nil {
		t.Fatal("expected a Form, got nil")
	}
	if sdkForm.Introduction != nil {
		t.Error("introduction should be nil when the block is absent")
	}
	if sdkForm.ShowSummary == nil || *sdkForm.ShowSummary != false {
		t.Error("show_summary=false must be preserved")
	}
	if sdkForm.ReceivedMessage.Subtitle != nil || sdkForm.ReceivedMessage.ImageUrl != nil {
		t.Error("unset optional strings should be nil, not empty strings")
	}

	input := (*(*sdkForm.FormPages)[0].PageComponents)[0].Input
	if input.Title != nil || input.PlaceholderText != nil || input.KeyboardType != nil {
		t.Error("unset optional input strings should be nil")
	}
	if *input.IsRequired != true || *input.IsMultipleLine != false {
		t.Error("input booleans not mapped")
	}

	if !reflect.DeepEqual(minimal, flattenForm(sdkForm)) {
		t.Errorf("minimal form did not round trip: %#v", flattenForm(sdkForm))
	}
}

// TestUnitFormAbsentYieldsNil confirms an omitted form block is not sent to the API as an empty object.
func TestUnitFormAbsentYieldsNil(t *testing.T) {
	if got := buildForm(nil); got != nil {
		t.Errorf("buildForm(nil) = %#v, want nil", got)
	}
	if got := buildForm([]interface{}{}); got != nil {
		t.Errorf("buildForm(empty) = %#v, want nil", got)
	}
	if got := flattenForm(nil); got != nil {
		t.Errorf("flattenForm(nil) = %#v, want nil", got)
	}
}

// TestUnitFormBuildersToleratePartialMaps ensures the builders never panic on maps that are missing
// keys, which is what callers outside of a full ResourceData hand us.
func TestUnitFormBuildersToleratePartialMaps(t *testing.T) {
	partial := []interface{}{
		map[string]interface{}{
			"form_pages": []interface{}{
				map[string]interface{}{
					"page_components": []interface{}{
						map[string]interface{}{
							"list_picker":  []interface{}{map[string]interface{}{"sections": []interface{}{map[string]interface{}{}}}},
							"date_picker":  []interface{}{map[string]interface{}{}},
							"input":        []interface{}{map[string]interface{}{}},
							"wheel_picker": []interface{}{map[string]interface{}{"items": []interface{}{map[string]interface{}{}}}},
						},
					},
				},
			},
		},
	}

	sdkForm := buildForm(partial) // must not panic
	if sdkForm == nil {
		t.Fatal("expected a Form, got nil")
	}
	if sdkForm.FormDescription != nil {
		t.Error("absent form_description should be nil")
	}
	component := (*(*sdkForm.FormPages)[0].PageComponents)[0]
	if component.FormComponentType != nil {
		t.Error("absent form_component_type should be nil")
	}
	if component.DatePicker == nil || component.DatePicker.DateDisplayFormat != nil {
		t.Error("absent date_display_format should be nil")
	}
}

// TestUnitGetResponseFromResourceDataSetsForm exercises the real ResourceData path so the schema and
// the builders are verified together.
func TestUnitGetResponseFromResourceDataSetsForm(t *testing.T) {
	d := schema.TestResourceDataRaw(t, ResourceResponsemanagementResponse().Schema, map[string]interface{}{
		"name":          "Test Form Response",
		"library_ids":   []interface{}{"lib-1"},
		"response_type": "Form",
		"form":          testFormConfig(),
	})

	response := getResponseFromResourceData(d)

	if response.ResponseType == nil || *response.ResponseType != "Form" {
		t.Errorf("response_type not set, got %v", response.ResponseType)
	}
	if response.Form == nil {
		t.Fatal("expected Response.Form to be populated")
	}
	if response.Form.FormPages == nil || len(*response.Form.FormPages) != 4 {
		t.Fatalf("expected 4 form pages through ResourceData, got %v", response.Form.FormPages)
	}
	if *response.Form.FormDescription != "A test form for customer feedback" {
		t.Errorf("form_description = %q", *response.Form.FormDescription)
	}
}

// TestUnitGetResponseFromResourceDataOmitsFormForOtherTypes confirms non-Form responses do not send a form.
func TestUnitGetResponseFromResourceDataOmitsFormForOtherTypes(t *testing.T) {
	d := schema.TestResourceDataRaw(t, ResourceResponsemanagementResponse().Schema, map[string]interface{}{
		"name":          "Footer response",
		"library_ids":   []interface{}{"lib-1"},
		"response_type": "Footer",
	})

	response := getResponseFromResourceData(d)

	if response.Form != nil {
		t.Errorf("Form should be nil when the block is absent, got %#v", response.Form)
	}
}

// TestUnitResponseResourceSchemaIsValid validates the resource schema, including the new form block,
// against the plugin SDK's own rules.
func TestUnitResponseResourceSchemaIsValid(t *testing.T) {
	resourceSchema := ResourceResponsemanagementResponse()
	if err := resourceSchema.InternalValidate(nil, true); err != nil {
		t.Fatalf("resource schema failed InternalValidate: %s", err)
	}

	formAttr, ok := resourceSchema.Schema["form"]
	if !ok {
		t.Fatal("expected a `form` attribute on the resource schema")
	}
	if formAttr.Type != schema.TypeList || formAttr.MaxItems != 1 {
		t.Errorf("form should be a TypeList with MaxItems 1, got type=%v maxItems=%d", formAttr.Type, formAttr.MaxItems)
	}

	// `Form` must be an accepted response_type or the form block is unusable.
	if _, errs := resourceSchema.Schema["response_type"].ValidateFunc("Form", "response_type"); len(errs) > 0 {
		t.Errorf("response_type should accept \"Form\": %v", errs)
	}
}
