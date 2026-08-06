package framework

import (
	"sort"
	"sync"
)

// Scope identifies the application runtime in which a module participates.
// A module may be registered for more than one scope.
type Scope string

const (
	ScopeHTTP      Scope = "http"
	ScopeScheduler Scope = "scheduler"
)

type moduleEntry struct {
	module Module
	scopes []Scope
}

var (
	registryMu sync.RWMutex
	entries    []moduleEntry
)

// Register adds a module to the process-wide compile-time registry.
// Modules without an explicit scope participate in the HTTP runtime.
func Register(module Module, scopes ...Scope) {
	if module == nil {
		return
	}
	if len(scopes) == 0 {
		scopes = []Scope{ScopeHTTP}
	}

	registryMu.Lock()
	defer registryMu.Unlock()
	entries = append(entries, moduleEntry{
		module: module,
		scopes: append([]Scope(nil), scopes...),
	})
}

// Modules returns all registered modules matching at least one requested
// scope. Results are de-duplicated by module name and sorted for stable startup
// order.
func Modules(scopes ...Scope) []Module {
	if len(scopes) == 0 {
		scopes = []Scope{ScopeHTTP}
	}

	registryMu.RLock()
	defer registryMu.RUnlock()

	modules := make([]Module, 0, len(entries))
	seen := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		if !entry.hasAnyScope(scopes...) {
			continue
		}
		if _, ok := seen[entry.module.Name()]; ok {
			continue
		}
		seen[entry.module.Name()] = struct{}{}
		modules = append(modules, entry.module)
	}

	sort.SliceStable(modules, func(i, j int) bool {
		return modules[i].Name() < modules[j].Name()
	})
	return modules
}

func (entry moduleEntry) hasAnyScope(scopes ...Scope) bool {
	for _, registeredScope := range entry.scopes {
		for _, requestedScope := range scopes {
			if registeredScope == requestedScope {
				return true
			}
		}
	}
	return false
}
