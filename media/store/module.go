package store

import (
	"sync"
)

// Module contains attributes which describe a storage module.
type Module struct {
	// ServiceConfigSkeleton returns an instance of config used to initialize
	// the service. This skeleton contains a config structure.
	ServiceConfigSkeleton func() ServiceConfig

	// NewService create a storage service backend connection. This is
	// usually initialize the client of the object storage.
	NewService func(config ServiceConfig) (Service, error)
}

var (
	modules   = map[string]Module{}
	modulesMu sync.RWMutex
)

func ModuleNames() []string {
	modulesMu.Lock()
	defer modulesMu.Unlock()

	var names []string
	for name := range modules {
		names = append(names, name)
	}

	return names
}

func NewServiceClient(
	serviceName string,
	config interface{},
) (Service, error) {
	if serviceName == "" {
		return nil, nil
	}

	var module Module
	modulesMu.RLock()
	module, _ = modules[serviceName]
	modulesMu.RUnlock()

	return module.NewService(config)
}

func RegisterModule(
	serviceName string,
	module Module,
) {
	modulesMu.Lock()
	defer modulesMu.Unlock()

	if _, dup := modules[serviceName]; dup {
		panic("called twice for service " + serviceName)
	}

	modules[serviceName] = module
}

func ModuleConfigSkeletons() map[string]interface{} {
	modulesMu.RLock()
	defer modulesMu.RUnlock()

	configs := map[string]interface{}{}
	for serviceName, mod := range modules {
		if mod.ServiceConfigSkeleton != nil {
			configs[serviceName] = mod.ServiceConfigSkeleton()
		}
	}

	return configs
}
