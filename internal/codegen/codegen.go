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
	return &c.PortalDefinitions[0]
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
