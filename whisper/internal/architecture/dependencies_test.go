package architecture

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const internalPath = "github.com/wonjinsin/ai-toybox/whisper/internal/"

func TestDomainHasNoInfrastructureDependencies(t *testing.T) {
	assertBoundary(t, "domain", func(dependency string) bool {
		return !strings.Contains(dependency, "/internal/") && !infrastructure(dependency)
	})
}

func TestUseCasesDependOnlyOnDomainAndPorts(t *testing.T) {
	assertBoundary(t, "usecase", func(dependency string) bool {
		return allowedInternal(dependency, "domain", "port") && !infrastructure(dependency)
	})
}

func TestPortsHaveNoImplementationDependencies(t *testing.T) {
	assertBoundary(t, "port", func(dependency string) bool {
		return allowedInternal(dependency, "domain", "port") && !infrastructure(dependency)
	})
	for _, direction := range []string{"in", "out"} {
		t.Run(direction, func(t *testing.T) {
			assertBoundary(t, "port/"+direction, func(dependency string) bool {
				return (dependency == internalPath+"port" || allowedInternal(dependency, "domain", "port/"+direction)) &&
					!infrastructure(dependency)
			})
		})
	}
}

func TestInputAdaptersDependOnlyOnInputContracts(t *testing.T) {
	assertBoundary(t, "adapter/in", func(dependency string) bool {
		return dependency == internalPath+"port" || allowedInternal(dependency, "domain", "port/in", "adapter/in")
	})
}

func TestOutputAdaptersDependOnlyOnOutputContracts(t *testing.T) {
	assertBoundary(t, "adapter/out", func(dependency string) bool {
		return dependency == internalPath+"port" || allowedInternal(dependency, "domain", "port/out", "adapter/out")
	})
}

func TestBootstrapOwnsConcreteWiring(t *testing.T) {
	assertBoundary(t, "bootstrap", func(dependency string) bool {
		return allowedInternal(dependency, "domain", "port", "usecase", "adapter")
	})
	imports := assertBoundary(t, "../cmd/whisper-local", func(dependency string) bool {
		return allowedInternal(dependency, "bootstrap")
	})
	if !imports[internalPath+"bootstrap"] {
		t.Fatal("command must import bootstrap to wire the application")
	}
}

func TestLegacyPackagesHaveBeenRemoved(t *testing.T) {
	for _, layer := range []string{"app", "application", "adapters"} {
		paths, err := filepath.Glob(filepath.Join("..", layer))
		if err != nil {
			t.Fatal(err)
		}
		if len(paths) != 0 {
			t.Errorf("legacy internal/%s must be fully migrated without compatibility aliases", layer)
		}
	}
}

func allowedInternal(dependency string, layers ...string) bool {
	if !strings.HasPrefix(dependency, internalPath) {
		return true
	}
	for _, layer := range layers {
		path := internalPath + layer
		if dependency == path || strings.HasPrefix(dependency, path+"/") {
			return true
		}
	}
	return false
}

func infrastructure(dependency string) bool {
	for _, name := range []string{"os", "flag", "syscall", "net", "encoding/json"} {
		if dependency == name || strings.HasPrefix(dependency, name+"/") {
			return true
		}
	}
	return false
}

func assertBoundary(t *testing.T, layer string, allowed func(string) bool) map[string]bool {
	t.Helper()
	productionFiles := 0
	imports := make(map[string]bool)
	err := filepath.WalkDir(filepath.Join("..", layer), func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		productionFiles++
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, spec := range file.Imports {
			dependency, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return err
			}
			imports[dependency] = true
			if !allowed(dependency) {
				t.Errorf("%s imports disallowed dependency %q", path, dependency)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if productionFiles == 0 {
		t.Fatalf("%s must contain production code, not only tests", layer)
	}
	return imports
}
