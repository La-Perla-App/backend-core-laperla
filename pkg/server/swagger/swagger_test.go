package swagger

import (
	"strings"
	"testing"
)

func TestModifyAndWrapResponsesYAML_HttpBodyStarContent(t *testing.T) {
	// QrLoginGenerate-style: 200 with */* and no application/json.
	in := []byte(`
openapi: 3.0.3
paths:
  /api/auth/v1/qr-login/generate:
    get:
      responses:
        "200":
          description: OK
          content:
            '*/*': {}
  /api/auth/v1/login:
    post:
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/LoginResponse'
components:
  schemas:
    LoginResponse:
      type: object
`)
	out, err := modifyAndWrapResponsesYAML(in)
	if err != nil {
		t.Fatalf("modify: %v", err)
	}
	if !strings.Contains(string(out), "ApiResponse_LoginResponse") {
		t.Fatalf("expected wrapped login schema, got:\n%s", out)
	}
	if !strings.Contains(string(out), "qr-login/generate") {
		t.Fatalf("expected generate path preserved")
	}
}

func TestFindMappingValueNil(t *testing.T) {
	if findMappingValue(nil, "x") != nil {
		t.Fatal("nil node should return nil")
	}
}
