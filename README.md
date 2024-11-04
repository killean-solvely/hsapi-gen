# hsapi-gen

Generates a .ts file with all of the relevant types for your Hubspot portals.

## Usage

### Install

You can download the latest binary from the releases page, run it through go, or compile it from the source.

### Setup

Create a config file with the following structure:

```json
{
  "outfolder": "./generated/",
  "lang": "ts",
  "schemas": [
    {
      "name": "production",
      "token": "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
    },
    {
      "name": "staging",
      "token": "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
    }
  ]
}
```

- `outfolder` is the folder where the generated files will be saved.
- `lang` is the language you want the types to be generated in. Currently only `ts` is supported.
- `schemas` is an array of objects that represent the different Hubspot portals you want to generate types for.
  - `name` is the name of the portal.
  - `token` is the API key for the portal.

### Run

Locally after downloading the binary
`hsapi-gen -config path-to-your-config.json`

Through Go
`go install github.com/killean-solvely/hsapi-gen/cmd/hsapi-gen`
`hsapi-gen -config path-to-your-config.json`

## TODO

Currently only covers the base and custom object interactions for getting, creating, and updating, as well as associations.
Future state should cover all of the other endpoints.
