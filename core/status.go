package core

import "fmt"

// System stauts:
// Green :: No error/warning detected.
// Yellow :: Some module has warning.
// Orange :: Some module has error and not working, but others is still working.
// Red :: Entire system is not working.
type SystemStatus int

const (
	// Green :: No error/warning detected.
	SystemGreen SystemStatus = iota
	// Yellow :: Some module has warning.
	SystemYellow
	// Orange :: Some module has error and not working, but others is still working.
	SystemOrange
	// Red :: Entire system is not working.
	SystemRed
)

func (s SystemStatus) String() string {
	switch s {
	case SystemGreen:
		return "Green"
	case SystemYellow:
		return "Yellow"
	case SystemOrange:
		return "Orange"
	case SystemRed:
		return "Red"
	default:
		return fmt.Sprintf("Status(%d)", int(s))
	}
}
