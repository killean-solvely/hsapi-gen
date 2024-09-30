# hsapi-gen

Generates a .ts file with all of the relevant types for your Hubspot portal.

## Usage

### Install

You can download the latest binary from the releases page, run it through go, or compile it from the source.

### Run

Locally
`hsapi-gen -token <your-hubspot-api-token> -path <path-to-output-file>`

Through Go
`go run github.com/killean-solvely/hsapi-gen/cmd/codegen -token <your-hubspot-api-token> -path <path-to-output-file>`

## TODO

Currently only covers the base and custom object interactions for getting, creating, and updating, as well as associations.
Future state should cover all of the other endpoints.
