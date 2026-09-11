package swagger

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path"

	"github.com/go-chi/chi/v5"
	"gopkg.in/yaml.v3"
)

//go:embed dist
var swaggerFS embed.FS

var OpenAPISchemaFileName = "openapi.yaml"

func RegisterSwaggerAssets(prefix string, openAPISchema []byte, r chi.Router) {
	modifiedSchema, err := modifyAndWrapResponsesYAML(openAPISchema)
	if err != nil {
		log.Println(err)
		modifiedSchema = openAPISchema
	}
	subFS, _ := fs.Sub(swaggerFS, "dist")
	httpFS := http.FileServer(http.FS(subFS))

	r.Handle(prefix, http.RedirectHandler(path.Join(prefix, "index.html"), http.StatusMovedPermanently))
	r.Handle(path.Join(prefix, "*"), http.StripPrefix(prefix, httpFS))

	r.Get(path.Join(prefix, OpenAPISchemaFileName), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		w.Write(modifiedSchema)
	})
}

// findMappingValue es una función auxiliar para encontrar un nodo de valor
// dado un nodo de mapeo (un objeto) y una clave.
func findMappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

// modifyAndWrapResponsesYAML procesa el schema OpenAPI usando un árbol de nodos YAML.
// Este método es más robusto que usar map[string]interface{}.
func modifyAndWrapResponsesYAML(originalSchema []byte) ([]byte, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(originalSchema, &root); err != nil {
		return nil, fmt.Errorf("error al decodificar el yaml original: %w", err)
	}

	// El contenido de la raíz es el nodo del documento, su contenido es el mapeo principal.
	specNode := root.Content[0]

	// 1. Añadir el schema genérico de la envoltura en components.schemas
	componentsNode := findMappingValue(specNode, "components")
	if componentsNode == nil {
		// Si 'components' no existe, lo creamos y lo añadimos
		keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: "components"}
		valueNode := &yaml.Node{Kind: yaml.MappingNode}
		specNode.Content = append(specNode.Content, keyNode, valueNode)
		componentsNode = valueNode
	}

	schemasNode := findMappingValue(componentsNode, "schemas")
	if schemasNode == nil {
		keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: "schemas"}
		valueNode := &yaml.Node{Kind: yaml.MappingNode}
		componentsNode.Content = append(componentsNode.Content, keyNode, valueNode)
		schemasNode = valueNode
	}

	wrapperSchemaName := "GenericApiResponse"
	var wrapperSchemaNode yaml.Node
	wrapperSchemaYAML := `
type: object
properties:
  statusCode:
    type: integer
    format: int32
    example: 200
  message:
    type: string
    example: "success"
`
	_ = yaml.Unmarshal([]byte(wrapperSchemaYAML), &wrapperSchemaNode)

	schemasNode.Content = append(schemasNode.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: wrapperSchemaName},
		wrapperSchemaNode.Content[0],
	)

	// 2. Iterar sobre todos los paths para modificar las respuestas
	pathsNode := findMappingValue(specNode, "paths")
	if pathsNode == nil {
		return originalSchema, nil // No hay paths, no hay nada que hacer.
	}

	// Iteramos sobre los path items (ej. /api/chat/v1/delete)
	for i := 0; i < len(pathsNode.Content); i += 2 {
		pathItemNode := pathsNode.Content[i+1] // El valor del path
		// Iteramos sobre los métodos (ej. post, get)
		for j := 0; j < len(pathItemNode.Content); j += 2 {
			operationNode := pathItemNode.Content[j+1]
			responsesNode := findMappingValue(operationNode, "responses")
			if responsesNode == nil {
				continue
			}

			response200Node := findMappingValue(responsesNode, "200")
			if response200Node == nil {
				continue
			}

			// Navegar hasta el schema: content -> application/json -> schema
			// (HttpBody / image/* no tienen application/json: no envolver.)
			contentNode := findMappingValue(response200Node, "content")
			if contentNode == nil {
				continue
			}
			appJSONNode := findMappingValue(contentNode, "application/json")
			if appJSONNode == nil {
				continue
			}
			schemaRefNode := findMappingValue(appJSONNode, "schema")
			if schemaRefNode == nil {
				continue
			}

			// Obtener la referencia original, ej. "#/components/schemas/PingResponse"
			originalRefValueNode := findMappingValue(schemaRefNode, "$ref")
			if originalRefValueNode == nil {
				continue
			}
			originalRef := originalRefValueNode.Value
			originalSchemaName := path.Base(originalRef)

			// 3. Crear el nuevo schema combinado y añadirlo a components.schemas
			newWrappedSchemaName := fmt.Sprintf("ApiResponse_%s", originalSchemaName)

			// Solo lo creamos si no existe
			if findMappingValue(schemasNode, newWrappedSchemaName) == nil {
				var newSchemaNode yaml.Node
				newSchemaYAML := fmt.Sprintf(`
allOf:
  - $ref: '#/components/schemas/%s'
  - type: object
    properties:
      data:
        $ref: '%s'
`, wrapperSchemaName, originalRef)
				_ = yaml.Unmarshal([]byte(newSchemaYAML), &newSchemaNode)

				schemasNode.Content = append(schemasNode.Content,
					&yaml.Node{Kind: yaml.ScalarNode, Value: newWrappedSchemaName},
					newSchemaNode.Content[0],
				)
			}

			// 4. Actualizar la referencia en la respuesta para que apunte al nuevo schema
			newRef := fmt.Sprintf("#/components/schemas/%s", newWrappedSchemaName)
			originalRefValueNode.Value = newRef
		}
	}

	// 5. Codificar el árbol de nodos modificado de vuelta a un []byte
	var b bytes.Buffer
	encoder := yaml.NewEncoder(&b)
	encoder.SetIndent(2)
	if err := encoder.Encode(&root); err != nil {
		return nil, fmt.Errorf("error al codificar el yaml modificado: %w", err)
	}

	// El encoder añade "---" al inicio, lo quitamos si no es parte del original
	finalYAML := b.Bytes()
	if !bytes.HasPrefix(originalSchema, []byte("---")) {
		finalYAML = bytes.TrimPrefix(finalYAML, []byte("---\n"))
	}

	return finalYAML, nil
}
