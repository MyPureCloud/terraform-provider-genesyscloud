package architect_datatable

import (
	"fmt"
	"log"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/provider"
	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v162/platformclientv2"
)

// DatatableBuilder handles the construction of Datatable objects
type DatatableBuilder struct {
	datatable *Datatable
}

// NewDatatableBuilder creates a new DatatableBuilder
func NewDatatableBuilder() *DatatableBuilder {
	return &DatatableBuilder{
		datatable: &Datatable{},
	}
}

// WithName sets the name of the datatable
func (b *DatatableBuilder) WithName(name string) *DatatableBuilder {
	b.datatable.Name = &name
	return b
}

// WithID sets the ID of the datatable
func (b *DatatableBuilder) WithID(id string) *DatatableBuilder {
	b.datatable.Id = &id
	return b
}

// WithDescription sets the description if not empty
func (b *DatatableBuilder) WithDescription(description string) *DatatableBuilder {
	if description != "" {
		b.datatable.Description = &description
	}
	return b
}

// WithDivision sets the division if not empty
func (b *DatatableBuilder) WithDivision(divisionID string) *DatatableBuilder {
	if divisionID != "" {
		b.datatable.Division = &platformclientv2.Writabledivision{Id: &divisionID}
	}
	return b
}

// WithSchema sets the schema
func (b *DatatableBuilder) WithSchema(schema *Jsonschemadocument) *DatatableBuilder {
	b.datatable.Schema = schema
	return b
}

// Build returns the constructed Datatable
func (b *DatatableBuilder) Build() *Datatable {
	return b.datatable
}

// DatatableFieldSetter handles setting fields in the schema.ResourceData
type DatatableFieldSetter struct {
	d *schema.ResourceData
}

// NewDatatableFieldSetter creates a new DatatableFieldSetter
func NewDatatableFieldSetter(d *schema.ResourceData) *DatatableFieldSetter {
	return &DatatableFieldSetter{d: d}
}

// SetName sets the name field
func (s *DatatableFieldSetter) SetName(name *string) {
	if name != nil {
		_ = s.d.Set("name", *name)
	}
}

// SetDivisionID sets the division_id field
func (s *DatatableFieldSetter) SetDivisionID(division *platformclientv2.Writabledivision) {
	if division != nil && division.Id != nil {
		_ = s.d.Set("division_id", *division.Id)
	}
}

// SetDescription sets the description field
func (s *DatatableFieldSetter) SetDescription(description *string) {
	if description != nil {
		_ = s.d.Set("description", *description)
	} else {
		_ = s.d.Set("description", nil)
	}
}

// SetProperties sets the properties field
func (s *DatatableFieldSetter) SetProperties(schema *Jsonschemadocument) {
	if schema != nil && schema.Properties != nil {
		_ = s.d.Set("properties", flattenDatatableProperties(*schema.Properties))
	} else {
		_ = s.d.Set("properties", nil)
	}
}

// SetAllFields sets all fields from a Datatable
func (s *DatatableFieldSetter) SetAllFields(datatable *Datatable) {
	s.SetName(datatable.Name)
	s.SetDivisionID(datatable.Division)
	s.SetDescription(datatable.Description)
	s.SetProperties(datatable.Schema)
}

// ErrorHandler provides consistent error handling
type ErrorHandler struct {
	resourceType string
}

// NewErrorHandler creates a new ErrorHandler
func NewErrorHandler(resourceType string) *ErrorHandler {
	return &ErrorHandler{resourceType: resourceType}
}

// HandleCreateError handles create operation errors
func (h *ErrorHandler) HandleCreateError(datatable *Datatable, err error, resp *platformclientv2.APIResponse) diag.Diagnostics {
	datatableName := "unknown"
	if datatable != nil && datatable.Name != nil {
		datatableName = *datatable.Name
	}
	return util.BuildAPIDiagnosticError(h.resourceType, fmt.Sprintf("Failed to create architect_datatable %s error: %s", datatableName, err), resp)
}

// HandleReadError handles read operation errors
func (h *ErrorHandler) HandleReadError(id string, err error, resp *platformclientv2.APIResponse) diag.Diagnostics {
	return util.BuildAPIDiagnosticError(h.resourceType, fmt.Sprintf("Failed to read architect_datatable %s | error: %s", id, err), resp)
}

// HandleUpdateError handles update operation errors
func (h *ErrorHandler) HandleUpdateError(name string, err error, resp *platformclientv2.APIResponse) diag.Diagnostics {
	return util.BuildAPIDiagnosticError(h.resourceType, fmt.Sprintf("Failed to update architect_datatable %s, error: %s", name, err), resp)
}

// HandleDeleteError handles delete operation errors
func (h *ErrorHandler) HandleDeleteError(name string, err error, resp *platformclientv2.APIResponse) diag.Diagnostics {
	return util.BuildAPIDiagnosticError(h.resourceType, fmt.Sprintf("Failed to delete architect_datatable %s error: %s", name, err), resp)
}

// DatatableValidator provides validation for datatable operations
type DatatableValidator struct{}

// NewDatatableValidator creates a new DatatableValidator
func NewDatatableValidator() *DatatableValidator {
	return &DatatableValidator{}
}

// ValidateDatatable validates a datatable response
func (v *DatatableValidator) ValidateDatatable(datatable *Datatable) error {
	if datatable == nil {
		return fmt.Errorf("received nil datatable from API")
	}
	if datatable.Name == nil {
		return fmt.Errorf("received datatable with nil name from API")
	}
	return nil
}

// ValidateDatatableWithID validates a datatable response with ID requirement
func (v *DatatableValidator) ValidateDatatableWithID(datatable *Datatable) error {
	if err := v.ValidateDatatable(datatable); err != nil {
		return err
	}
	if datatable.Id == nil {
		return fmt.Errorf("received datatable with nil ID from API")
	}
	return nil
}

// DatatableLogger provides consistent logging
type DatatableLogger struct{}

// NewDatatableLogger creates a new DatatableLogger
func NewDatatableLogger() *DatatableLogger {
	return &DatatableLogger{}
}

// LogCreate logs create operations
func (l *DatatableLogger) LogCreate(name string) {
	log.Printf("Creating architect_datatable %s", name)
}

// LogCreateSuccess logs successful create operations
func (l *DatatableLogger) LogCreateSuccess(name, id string) {
	log.Printf("Created architect_datatable %s %s", name, id)
}

// LogRead logs read operations
func (l *DatatableLogger) LogRead(id string) {
	log.Printf("Reading architect_datatable %s", id)
}

// LogReadSuccess logs successful read operations
func (l *DatatableLogger) LogReadSuccess(id, name string) {
	log.Printf("Read architect_datatable %s %s", id, name)
}

// LogUpdate logs update operations
func (l *DatatableLogger) LogUpdate(name string) {
	log.Printf("Updating architect_datatable %s", name)
}

// LogUpdateSuccess logs successful update operations
func (l *DatatableLogger) LogUpdateSuccess(name string) {
	log.Printf("Updated architect_datatable %s", name)
}

// LogDelete logs delete operations
func (l *DatatableLogger) LogDelete(name string) {
	log.Printf("Deleting architect_datatable %s", name)
}

// LogDeleteSuccess logs successful delete operations
func (l *DatatableLogger) LogDeleteSuccess(name string) {
	log.Printf("Deleted architect_datatable row %s", name)
}

// getDatatableFromSchema extracts datatable data from schema.ResourceData
func getDatatableFromSchema(d *schema.ResourceData) (name, divisionID, description string) {
	if d == nil {
		return "", "", ""
	}

	if nameVal := d.Get("name"); nameVal != nil {
		name = nameVal.(string)
	}
	if divisionVal := d.Get("division_id"); divisionVal != nil {
		divisionID = divisionVal.(string)
	}
	if descVal := d.Get("description"); descVal != nil {
		description = descVal.(string)
	}
	return
}

// getProxyAndConfig extracts proxy and config from meta
func getProxyAndConfig(meta interface{}) (*architectDatatableProxy, *platformclientv2.Configuration, diag.Diagnostics) {
	if meta == nil {
		return nil, nil, diag.Errorf("meta parameter cannot be nil")
	}

	providerMeta, ok := meta.(*provider.ProviderMeta)
	if !ok {
		return nil, nil, diag.Errorf("meta parameter is not of type *provider.ProviderMeta")
	}

	if providerMeta.ClientConfig == nil {
		return nil, nil, diag.Errorf("ClientConfig cannot be nil")
	}

	sdkConfig := providerMeta.ClientConfig
	archProxy := getArchitectDatatableProxy(sdkConfig)
	return archProxy, sdkConfig, nil
}
