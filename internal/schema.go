package internal

import (
	"fmt"
	"os"

	"github.com/enix/graphql-codegen-go/internal/readers"
	"github.com/pkg/errors"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/formatter"
	"github.com/vektah/gqlparser/v2/parser"
	"github.com/vektah/gqlparser/v2/validator"
)

type InputSchema struct {
	Data       string
	SourcePath string
}

func ReadSchemas(schemaPaths []string) ([]InputSchema, error) {
	var outs []InputSchema
	for _, s := range schemaPaths {
		r := readers.DiscoverReader(s)
		o, err := r.Read()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read from %s", s)
		}
		outs = append(outs, InputSchema{
			Data:       string(o),
			SourcePath: s,
		})
	}
	return outs, nil
}

func LoadSchemas(inputSchemas []InputSchema) (*ast.SchemaDocument, error) {
	sourceSchemas := []*ast.Source{validator.Prelude} // include types
	for _, inputSchema := range inputSchemas {
		sourceSchemas = append(sourceSchemas, &ast.Source{
			Name:    inputSchema.SourcePath,
			Input:   inputSchema.Data,
			BuiltIn: false,
		})
	}
	doc, gqlErr := parser.ParseSchemas(sourceSchemas...)
	if gqlErr != nil {
		return nil, gqlErr
	}

	err := inheritInterfaces(doc)
	if err != nil {
		return nil, err
	}

	if _, err := validator.ValidateSchemaDocument(doc); err != nil {
		f := formatter.NewFormatter(os.Stderr)
		os.Stderr.WriteString("Parsed schema:\n")
		f.FormatSchemaDocument(doc)
		return nil, fmt.Errorf(err.Message)
	}

	return doc, nil
}

func inheritInterfaces(doc *ast.SchemaDocument) error {
	for _, definition := range doc.Definitions {
		if definition.Kind != ast.Object {
			continue
		}

		for _, interfaceName := range definition.Interfaces {
			iface := doc.Definitions.ForName(interfaceName)
			if iface == nil {
				return fmt.Errorf("no such interface: %s", interfaceName)
			}

			for _, field := range iface.Fields {
				definition.Fields = append(definition.Fields, field)
			}
		}
	}

	return nil
}
