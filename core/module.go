package core

type Module interface {
	// Load configuration, validate configuration (test connection etc..)
	Init(id string, c *Context) error

	// Start working.
	// Start scheduled working. But first run should be performed under Start to check error.
	Start(c *Context) error

	// Stop module
	Close() error

	// Return module error, to aggregate system status.
	Status() ModuleStatus
}

type ModuleStatus struct {
	Error     error
	Available bool
	Warning   string
}

var Status = DefaultModuleStatus{}

type DefaultModuleStatus struct {
}

func (m DefaultModuleStatus) OK() ModuleStatus {
	return ModuleStatus{
		Available: true,
	}
}

// The module is no longer working on.
func (m DefaultModuleStatus) Error(err error) ModuleStatus {
	return ModuleStatus{
		Available: false,
		Error:     err,
	}
}

// The module still working, but has some warning.
func (m DefaultModuleStatus) Warn(msg string) ModuleStatus {
	return ModuleStatus{
		Available: false,
		Warning:   msg,
	}
}
