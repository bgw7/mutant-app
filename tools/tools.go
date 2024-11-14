//go:build tools
// +build tools

package main

import (
	"fmt"
	"os"
	"strings"

	_ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
	"github.com/pb33f/libopenapi"
	_ "github.com/sqlc-dev/sqlc/cmd/sqlc"
)

func loadOpenAPISpec(filePath string) (libopenapi.Document, error) {
	spec, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read OpenAPI spec: %w", err)
	}

	doc, err := libopenapi.NewDocument(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI spec: %w", err)
	}

	return doc, nil
}

func generateClientCode(doc libopenapi.Document) string {
	clientCode := "package apiClient\n\nimport (\n\t\"net/http\"\n\t\"context\"\n)\n\n"

	paths := doc.Paths()
	for path, item := range paths {
		for method, op := range item.Operations() {
			funcName := fmt.Sprintf("%s%s", methodToFuncName(method), pathToFuncName(path))
			clientCode += generateEndpointFunction(funcName, path, method, op)
		}
	}
	return clientCode
}

func generateEndpointFunction(funcName, path, method string, op *libopenapi.Operation) string {
	return fmt.Sprintf(`
// %s calls the %s endpoint at %s
func (c *Client) %s(ctx context.Context, req *http.Request) (*http.Response, error) {
	req, err := http.NewRequest("%s", "%s", nil)
	if err != nil {
		return nil, err
	}
	// Set additional headers or parameters as needed
	return c.httpClient.Do(req.WithContext(ctx))
}
`, funcName, method, path, funcName, method, path)
}

func methodToFuncName(method string) string {
	return strings.Title(strings.ToLower(method))
}

func pathToFuncName(path string) string {
	// Convert `/users/{id}` to `UsersByID` or similar format
	path = strings.ReplaceAll(path, "/", "")
	path = strings.ReplaceAll(path, "{", "By")
	path = strings.ReplaceAll(path, "}", "")
	return strings.Title(path)
}

func saveGeneratedCode(filename, code string) error {
	return os.WriteFile(filename, []byte(code), 0644)
}

func main() {
	doc, err := loadOpenAPISpec("path/to/openapi.yaml")
	if err != nil {
		fmt.Println("Error loading spec:", err)
		return
	}

	clientCode := generateClientCode(doc)
	if err := saveGeneratedCode("client_gen.go", clientCode); err != nil {
		fmt.Println("Error saving generated code:", err)
		return
	}

	fmt.Println("Client code generated successfully in client_gen.go")
}
