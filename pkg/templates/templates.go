package templates

import (
	"bytes"
	"embed"
	"text/template"
)

func createTemplate(name, t string) *template.Template {
	return template.Must(template.New(name).Parse(t))
}

func GenerateCode[T any](embededFiles embed.FS, templatePath string, input T) (string, error) {
	f, err := embededFiles.ReadFile(templatePath)
	if err != nil {
		return "", err
	}

	t := createTemplate("temp", string(f))

	var buf bytes.Buffer
	err = t.Execute(&buf, input)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
