package custd

import (
	"os"
	"path/filepath"
	"testing"
)

// readContractFixture loads a shared contract fixture, resolving both layouts it
// can run under: the standalone custd-sdk-go module (the release split vendors
// the fixtures at the module root) and the monorepo (the shared fixtures live
// one directory up from sdk-go).
func readContractFixture(t *testing.T, name string) []byte {
	t.Helper()
	candidates := []string{
		filepath.Join("contract-fixtures", name),       // standalone custd-sdk-go module
		filepath.Join("..", "contract-fixtures", name), // monorepo layout
	}
	for _, path := range candidates {
		if data, err := os.ReadFile(path); err == nil {
			return data
		}
	}
	t.Fatalf("contract fixture %s not found (looked in %v)", name, candidates)
	return nil
}

// readLifecycleFixture loads a shared lifecycle contract fixture under
// contract-fixtures/lifecycle/{namespace}/{name}. It resolves the same two
// layouts as readContractFixture (standalone vs monorepo) so lifecycle tests
// work both in the vendored release split and in the monorepo layout.
func readLifecycleFixture(t *testing.T, namespace, name string) []byte {
	t.Helper()
	rel := filepath.Join("lifecycle", namespace, name)
	return readContractFixture(t, rel)
}
