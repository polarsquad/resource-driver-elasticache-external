package messages

// ResourceData is the payload that is returned to the deployment service containing all information to complete usage of the reource.
type ResourceData struct {
	Type       string        `json:"type"`
	Data       ValuesSecrets `json:"data"`
	DriverType string        `json:"driver_type"`
	DriverData ValuesSecrets `json:"driver_data"`
}

// ValuesSecrets respresents data that should be passed around for a resource split by sensitivity.
type ValuesSecrets struct {
	Values  map[string]interface{} `json:"values"`
	Secrets map[string]interface{} `json:"secrets"`
}

// DriverResourceDefinition holds a description of how the driver operates.
type DriverResourceDefinition struct {
	// the id for this resource
	// required: true
	// pattern: ^[a-z0-9][a-z0-9-]+[a-z0-9]$
	ID string `json:"id"`

	// the type of resource to generate.
	// required: true
	Type string `json:"type"`

	// The parameters passed in from the deployment set
	ResourceParams map[string]interface{} `json:"resource_params"`

	// The parameters passed in from the Dynamic Resource.
	DriverParams map[string]interface{} `json:"driver_params"`

	// Secret parameters passed in from the Dynamic Resource.
	DriverSecrets map[string]interface{} `json:"driver_secrets,omitempty"`
}
