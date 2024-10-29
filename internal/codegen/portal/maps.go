package portal

var categoryMap = map[string]string{
	"HUBSPOT_DEFINED":    "HubspotDefined",
	"USER_DEFINED":       "UserDefined",
	"INTEGRATOR_DEFINED": "IntegratorDefined",
}

var typeConversionMap = map[string]string{
	"date":               "string",
	"datetime":           "string",
	"bool":               "boolean",
	"object_coordinates": "string",
	"json":               "string",
	"phone_number":       "string",
}
