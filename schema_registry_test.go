package custd

import (
	"context"
	"net/http"
	"testing"
)

// The schema registry answers its lists as bare JSON arrays and a schema
// version as the raw document; the client returns named DTOs for both.
func TestSchemaRegistryDecodesServerShapes(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `[{"slug":"order.created","name":"Order created","enabled":true}]`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	list, err := client.Schemas.List(context.Background())
	if err != nil || len(list.EventTypes) != 1 || list.EventTypes[0].Slug != "order.created" {
		t.Fatalf("list = %+v, %v", list, err)
	}
	doer.body = `null`
	versions, err := client.Schemas.ListVersions(context.Background(), "order.created")
	if err != nil || versions.Versions == nil || len(versions.Versions) != 0 {
		t.Fatalf("null versions = %+v, %v", versions, err)
	}
	doer.body = `{"type":"object"}`
	document, err := client.Schemas.Get(context.Background(), "order.created", "1.0.0")
	if err != nil || string(document.Schema) != `{"type":"object"}` {
		t.Fatalf("document = %s, %v", document.Schema, err)
	}
}
