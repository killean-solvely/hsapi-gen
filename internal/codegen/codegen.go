package codegen

import (
	"github.com/killean-solvely/hsapi-gen/internal/codegen/typescriptgen"
)

type Codegen struct {
	portals map[string]string
	debug   bool
}

func New(debug bool) *Codegen {
	return &Codegen{
		portals: map[string]string{},
		debug:   debug,
	}
}

// Adds the portal to the list of portals to be processed
func (c *Codegen) AddPortal(portalName, token string) {
	c.portals[portalName] = token
}

func (c Codegen) GenerateTypescript(outfolder string) error {
	gen := typescriptgen.New(c.debug)

	for name, token := range c.portals {
		gen.AddPortal(name, token)
	}

	return gen.GenerateCode(outfolder)
}
