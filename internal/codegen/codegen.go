package codegen

import (
	"fmt"
	"os"
	"path"

	"github.com/killean-solvely/hsapi-gen/internal/codegen/portal"
	"github.com/killean-solvely/hsapi-gen/internal/codegen/templates"
)

type Codegen struct {
	PortalDefinitions []portal.PortalDefinition
}

func NewCodegen() *Codegen {
	return &Codegen{
		PortalDefinitions: []portal.PortalDefinition{},
	}
}

// Adds the portal to the list of portals to be processed
func (c *Codegen) AddPortal(portalName, token string) {
	c.PortalDefinitions = append(
		c.PortalDefinitions,
		*portal.NewPortalDefinition(portalName, token),
	)
}

func (c Codegen) GenerateCode(outfolder string) error {
	err := c.loadPortals()
	if err != nil {
		return err
	}

	sharedPD := c.createSharedPortalDefinition()

	return c.generateFiles(outfolder, sharedPD)
}

// Loads the portal definitions from the HubSpot API, or file if available
func (c *Codegen) loadPortals() error {
	for i := range c.PortalDefinitions {
		err := c.PortalDefinitions[i].LoadPortalDefinition()
		if err != nil {
			return err
		}
	}

	return nil
}

// Prepares a combined portal definition for the shared template
func (c Codegen) createSharedPortalDefinition() *portal.PortalDefinition {
	// Create a new portal definition
	sharedPD := portal.NewPortalDefinition("shared", "")

	// Ensure we have at least one portal to compare against
	if len(c.PortalDefinitions) == 0 {
		return sharedPD
	}

	// Initialize the intersecting objects map
	objectMap := make(map[string]portal.Object)

	// Iterate over all objects in the first portal as a base for comparison
	for _, obj := range c.PortalDefinitions[0].Objects {
		objectMap[obj.InternalName] = obj
	}

	// Filter out objects that don’t exist in all portals
	for _, portalDef := range c.PortalDefinitions[1:] {
		for objName := range objectMap {
			found := false
			for _, obj := range portalDef.Objects {
				if obj.InternalName == objName {
					found = true
					break
				}
			}
			if !found {
				delete(objectMap, objName)
			}
		}
	}

	// Intersect properties within shared objects
	intersectingObjects := []portal.Object{}
	for _, obj := range objectMap {
		intersectingProps := intersectPropertiesAcrossPortals(obj.InternalName, c.PortalDefinitions)
		obj.Properties = intersectingProps
		intersectingObjects = append(intersectingObjects, obj)
	}

	// Assign the intersected objects to the shared portal definition
	sharedPD.Objects = intersectingObjects

	// Intersect Enums across all portals
	sharedPD.Enums = intersectEnumsAcrossPortals(c.PortalDefinitions)

	// Intersect ObjectNameToType across all portals
	sharedPD.ObjectNameToType = intersectObjectNameToTypeAcrossPortals(c.PortalDefinitions)

	// Intersect AssociationTypes across all portals
	sharedPD.AssociationTypes = intersectAssociationTypesAcrossPortals(c.PortalDefinitions)

	return sharedPD
}

// Helper function to find intersecting ObjectNameToType entries across all portals
func intersectObjectNameToTypeAcrossPortals(
	portals []portal.PortalDefinition,
) map[string]portal.SchemaData {
	typeMap := make(map[string]portal.SchemaData)

	// Start with the ObjectNameToType of the first portal
	for name, schemaData := range portals[0].ObjectNameToType {
		typeMap[name] = schemaData
	}

	// Intersect ObjectNameToType entries across all portals
	for _, portalDef := range portals[1:] {
		for typeName := range typeMap {
			found := false
			if schemaData, exists := portalDef.ObjectNameToType[typeName]; exists &&
				schemaData == typeMap[typeName] {
				found = true
			}
			if !found {
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
					if assoc, exists := portalDef.AssociationTypes[primaryType][secondaryType][assocName]; !exists ||
						assoc != assocTypeMap[primaryType][secondaryType][assocName] {
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
) []portal.Property {
	propertyMap := make(map[string]portal.Property)

	// Start with the properties of the object in the first portal
	for _, obj := range portals[0].Objects {
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
			for _, obj := range portalDef.Objects {
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
	intersectingProps := []portal.Property{}
	for _, prop := range propertyMap {
		intersectingProps = append(intersectingProps, prop)
	}
	return intersectingProps
}

// Helper function to find intersecting Enums across all portals
func intersectEnumsAcrossPortals(portals []portal.PortalDefinition) []portal.Enum {
	enumMap := make(map[string]portal.Enum)

	// Initialize enum map with first portal enums
	for _, enum := range portals[0].Enums {
		enumMap[enum.Name] = enum
	}

	// Intersect enums across all portals
	for _, portalDef := range portals[1:] {
		for enumName := range enumMap {
			found := false
			for _, enum := range portalDef.Enums {
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
	intersectingEnums := []portal.Enum{}
	for _, enum := range enumMap {
		intersectingEnums = append(intersectingEnums, enum)
	}
	return intersectingEnums
}

// Generates the code for the portals
func (c Codegen) generateFiles(outfolder string, sharedPD *portal.PortalDefinition) error {
	// Check to see if the output folder exists
	if _, err := os.Stat(outfolder); os.IsNotExist(err) {
		// Create the output folder
		err := os.Mkdir(outfolder, 0755)
		if err != nil {
			return err
		}
	}

	// Generate the client code
	fmt.Println("Generating Client Code...")
	clientCode, err := c.generateClientCode(sharedPD)
	if err != nil {
		return err
	}

	// Write the client code to a file
	err = os.WriteFile(path.Clean(outfolder+"/"+"client.ts"), []byte(clientCode), 0644)
	if err != nil {
		return err
	}

	// Generate the code for the portals
	fmt.Println("Generating Portal Code...")
	for i := range c.PortalDefinitions {
		fmt.Printf("Processing portal %s...\n", c.PortalDefinitions[i].PortalName)
		portalCode, err := c.generatePortalCode(&c.PortalDefinitions[i])
		if err != nil {
			return err
		}

		// Write the portal code to a file
		err = os.WriteFile(
			path.Clean(outfolder+"/"+c.PortalDefinitions[i].PortalName+".ts"),
			[]byte(portalCode),
			0644,
		)
		if err != nil {
			return err
		}
	}

	// Generate the code for the shared types
	fmt.Println("Generating Shared Code...")
	sharedCode, err := c.generateSharedCode(sharedPD)
	if err != nil {
		return err
	}

	// Write the shared code to a file
	err = os.WriteFile(path.Clean(outfolder+"/"+"shared.ts"), []byte(sharedCode), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (c Codegen) generateClientCode(sharedPD *portal.PortalDefinition) (string, error) {
	portalNames := map[string]string{}
	for _, pd := range c.PortalDefinitions {
		portalNames[pd.PortalName] = pd.PortalName
	}

	fileData, err := templates.GenerateClient(templates.HubspotClientTemplateInput{
		PortalNames:      portalNames,
		ObjectNameToType: sharedPD.ObjectNameToType,
		AssociationTypes: sharedPD.AssociationTypes,
	})
	if err != nil {
		return "", err
	}

	return fileData, nil
}

func (c Codegen) generatePortalCode(p *portal.PortalDefinition) (string, error) {
	objectMap := map[string]string{}
	for _, obj := range p.Objects {
		objectMap[obj.InternalName] = obj.ID
	}

	fileData, err := templates.GeneratePortal(templates.PortalTemplateInput{
		PortalName:       p.PortalName,
		AssociationTypes: p.AssociationTypes,
		Objects:          objectMap,
	})
	if err != nil {
		return "", err
	}

	return fileData, nil
}

func (c Codegen) generateSharedCode(sharedPD *portal.PortalDefinition) (string, error) {
	fileData, err := templates.GenerateShared(templates.SharedTemplateInput{
		AssociationTypes: sharedPD.AssociationTypes,
		Enums:            sharedPD.Enums,
		Objects:          sharedPD.Objects,
	})
	if err != nil {
		return "", err
	}

	return fileData, nil
}
