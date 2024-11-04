package typescriptgen

import (
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/killean-solvely/hsapi-gen/internal/portal"
)

type Codegen struct {
	portalDefinitions []portal.PortalDefinition
	codeDefinitions   map[string]PortalCodeDefinition
	sharedDefinition  *SharedDefinition

	debug bool
}

func New(debug bool) *Codegen {
	return &Codegen{
		portalDefinitions: []portal.PortalDefinition{},
		codeDefinitions:   map[string]PortalCodeDefinition{},
		debug:             debug,
	}
}

// Adds the portal to the list of portals to be generated.
func (c *Codegen) AddPortal(portalName, token string) {
	c.portalDefinitions = append(
		c.portalDefinitions,
		*portal.NewPortalDefinition(portalName, token, c.debug),
	)
}

// Generates the TypeScript code for the portals added to the Codegen instance.
func (c Codegen) GenerateCode(outfolder string) error {
	// Load the portals from hubspot
	err := c.loadPortals()
	if err != nil {
		return err
	}

	// Get the typescript code definitions for each portal
	c.parsePortals()

	// Generate the shared portal definition
	c.sharedDefinition = c.createSharedPortalDefinition()

	return nil
}

// Loads the portal definitions from the HubSpot API, or file if available
func (c *Codegen) loadPortals() error {
	var wg sync.WaitGroup
	for i := range c.portalDefinitions {
		wg.Add(1)
		go func(pd *portal.PortalDefinition) {
			err := pd.LoadPortalDefinition()
			if err != nil {
				fmt.Printf("Error loading portal definition for %s: %s\n", pd.PortalName, err)
			}
			wg.Done()
		}(&c.portalDefinitions[i])
	}
	wg.Wait()

	return nil
}

// Gets the typescript code definitions for each portal
func (c *Codegen) parsePortals() {
	for i := range c.portalDefinitions {
		pd := &c.portalDefinitions[i]
		pcd := PortalCodeDefinition{
			OriginalPortal: pd,
		}
		pcd.parsePortalDefinition()
		c.codeDefinitions[pd.PortalName] = pcd
	}
}

// Prepares a combined portal definition for the shared template
func (c Codegen) createSharedPortalDefinition() *SharedDefinition {
	// Create a new portal definition
	sharedPD := &SharedDefinition{
		portal:         portal.NewPortalDefinition("shared", "", c.debug),
		codeDefinition: nil,
	}

	// Ensure we have at least one portal to compare against
	if len(c.portalDefinitions) == 0 {
		return sharedPD
	}

	if len(c.portalDefinitions) == 1 {
		codeDef := c.codeDefinitions[c.portalDefinitions[0].PortalName]
		return &SharedDefinition{
			portal:         &c.portalDefinitions[0],
			codeDefinition: &codeDef,
		}
	}

	// Initialize the intersecting objects map
	objectMap := make(map[string]Object)

	// Iterate over all objects in the first portal as a base for comparison
	firstPortal := c.portalDefinitions[0]
	for _, obj := range c.codeDefinitions[firstPortal.PortalName].Objects {
		objectMap[obj.InternalName] = obj
	}

	// Filter out objects that don’t exist in all portals
	for _, portalDef := range c.portalDefinitions[1:] {
		for objName := range objectMap {
			found := false

			codeDef := c.codeDefinitions[portalDef.PortalName]
			for _, obj := range codeDef.Objects {
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
	intersectingObjects := []Object{}
	for _, obj := range objectMap {
		intersectingProps := intersectPropertiesAcrossPortals(
			obj.InternalName,
			c.portalDefinitions,
			c.codeDefinitions,
		)
		obj.Properties = intersectingProps
		intersectingObjects = append(intersectingObjects, obj)
	}

	// Assign the intersected objects to the shared portal definition
	sharedPD.codeDefinition.Objects = intersectingObjects

	// Intersect Enums across all portals
	sharedPD.codeDefinition.Enums = intersectEnumsAcrossPortals(
		c.portalDefinitions,
		c.codeDefinitions,
	)

	// Intersect ObjectNameToType across all portals
	sharedPD.codeDefinition.ObjectNameToType = intersectObjectNameToTypeAcrossPortals(
		c.portalDefinitions,
		c.codeDefinitions,
	)

	// Intersect AssociationTypes across all portals
	sharedPD.portal.AssociationTypes = intersectAssociationTypesAcrossPortals(
		c.portalDefinitions,
	)

	return sharedPD
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
	clientCode, err := c.generateClientCode()
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
	for i := range c.portalDefinitions {
		fmt.Printf("Processing portal %s...\n", c.portalDefinitions[i].PortalName)
		codeDef := c.codeDefinitions[c.portalDefinitions[i].PortalName]
		portalCode, err := c.generatePortalCode(
			&c.portalDefinitions[i],
			&codeDef,
		)
		if err != nil {
			return err
		}

		// Write the portal code to a file
		err = os.WriteFile(
			path.Clean(outfolder+"/"+c.portalDefinitions[i].PortalName+".ts"),
			[]byte(portalCode),
			0644,
		)
		if err != nil {
			return err
		}
	}

	// Generate the code for the shared types
	fmt.Println("Generating Shared Code...")
	sharedCode, err := c.generateSharedCode()
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

func (c Codegen) generateClientCode() (string, error) {
	portalNames := map[string]string{}
	for _, pd := range c.portalDefinitions {
		portalNames[pd.PortalName] = pd.PortalName
	}

	fileData, err := GenerateTSClient(HubspotTSClientTemplateInput{
		PortalNames:      portalNames,
		ObjectNameToType: c.sharedDefinition.codeDefinition.ObjectNameToType,
		AssociationTypes: c.sharedDefinition.portal.AssociationTypes,
	})
	if err != nil {
		return "", err
	}

	return fileData, nil
}

func (c Codegen) generatePortalCode(
	p *portal.PortalDefinition,
	cd *PortalCodeDefinition,
) (string, error) {
	objectMap := map[string]string{}
	for _, obj := range cd.Objects {
		objectMap[obj.InternalName] = obj.ID
	}

	fileData, err := GenerateTSPortal(TSPortalTemplateInput{
		PortalName:       p.PortalName,
		AssociationTypes: p.AssociationTypes,
		Objects:          objectMap,
	})
	if err != nil {
		return "", err
	}

	return fileData, nil
}

func (c Codegen) generateSharedCode() (string, error) {
	fileData, err := GenerateTSShared(TSSharedTemplateInput{
		AssociationTypes: c.sharedDefinition.portal.AssociationTypes,
		Enums:            c.sharedDefinition.codeDefinition.Enums,
		Objects:          c.sharedDefinition.codeDefinition.Objects,
	})
	if err != nil {
		return "", err
	}

	return fileData, nil
}
