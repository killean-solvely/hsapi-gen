package typescriptgen

import "github.com/killean-solvely/hsapi-gen/internal/portal"

// Helper function to find intersecting ObjectNameToType entries across all portals
func intersectObjectNameToTypeAcrossPortals(
	portals []portal.PortalDefinition,
	codeDefinitions map[string]PortalCodeDefinition,
) map[string]SchemaData {
	typeMap := make(map[string]SchemaData)

	// Start with the ObjectNameToType of the first portal
	for name, schemaData := range codeDefinitions[portals[0].PortalName].ObjectNameToType {
		typeMap[name] = schemaData
	}

	// Intersect ObjectNameToType entries across all portals
	for _, portalDef := range portals[1:] {
		for typeName := range typeMap {
			if _, exists := codeDefinitions[portalDef.PortalName].ObjectNameToType[typeName]; !exists {
				delete(typeMap, typeName)
			}
		}
	}

	return typeMap
}

// Helper function to intersect AssociationTypes across all portals
func intersectAssociationTypesAcrossPortals(
	portals []portal.PortalDefinition,
) map[string]map[string]map[string]portal.Association {
	assocTypeMap := make(map[string]map[string]map[string]portal.Association)

	// Initialize with the AssociationTypes from the first portal
	for primaryType, subMap := range portals[0].AssociationTypes {
		assocTypeMap[primaryType] = make(map[string]map[string]portal.Association)
		for secondaryType, innerMap := range subMap {
			assocTypeMap[primaryType][secondaryType] = make(map[string]portal.Association)
			for assocName, association := range innerMap {
				assocTypeMap[primaryType][secondaryType][assocName] = association
			}
		}
	}

	// Intersect AssociationTypes across all portals
	for _, portalDef := range portals[1:] {
		for primaryType, subMap := range assocTypeMap {
			for secondaryType, innerMap := range subMap {
				for assocName := range innerMap {
					// Check if the association exists in the current portal
					if _, exists := portalDef.AssociationTypes[primaryType][secondaryType][assocName]; !exists {
						delete(assocTypeMap[primaryType][secondaryType], assocName)
					}
				}
				// Remove secondaryType if it has no associations left
				if len(assocTypeMap[primaryType][secondaryType]) == 0 {
					delete(assocTypeMap[primaryType], secondaryType)
				}
			}
			// Remove primaryType if it has no secondary types left
			if len(assocTypeMap[primaryType]) == 0 {
				delete(assocTypeMap, primaryType)
			}
		}
	}

	return assocTypeMap
}

// Helper function to find intersecting properties across all portals for a given object
func intersectPropertiesAcrossPortals(
	objectName string,
	portals []portal.PortalDefinition,
	codeDefinitions map[string]PortalCodeDefinition,
) []Property {
	propertyMap := make(map[string]Property)

	// Start with the properties of the object in the first portal
	for _, obj := range codeDefinitions[portals[0].PortalName].Objects {
		if obj.InternalName == objectName {
			for _, prop := range obj.Properties {
				propertyMap[prop.Name] = prop
			}
			break
		}
	}

	// Compare properties with other portals
	for _, portalDef := range portals[1:] {
		for propName := range propertyMap {
			found := false
			for _, obj := range codeDefinitions[portalDef.PortalName].Objects {
				if obj.InternalName == objectName {
					for _, prop := range obj.Properties {
						if prop.Name == propName {
							found = true
							break
						}
					}
				}
			}
			if !found {
				delete(propertyMap, propName)
			}
		}
	}

	// Convert map to slice
	intersectingProps := []Property{}
	for _, prop := range propertyMap {
		intersectingProps = append(intersectingProps, prop)
	}
	return intersectingProps
}

// Helper function to find intersecting Enums across all portals
func intersectEnumsAcrossPortals(
	portals []portal.PortalDefinition,
	codeDefinitions map[string]PortalCodeDefinition,
) []Enum {
	enumMap := make(map[string]Enum)

	// Initialize enum map with first portal enums
	for _, enum := range codeDefinitions[portals[0].PortalName].Enums {
		enumMap[enum.Name] = enum
	}

	// Intersect enums across all portals
	for _, portalDef := range portals[1:] {
		for enumName := range enumMap {
			found := false
			for _, enum := range codeDefinitions[portalDef.PortalName].Enums {
				if enum.Name == enumName {
					found = true
					break
				}
			}
			if !found {
				delete(enumMap, enumName)
			}
		}
	}

	// Convert map to slice
	intersectingEnums := []Enum{}
	for _, enum := range enumMap {
		intersectingEnums = append(intersectingEnums, enum)
	}
	return intersectingEnums
}
