package flows

// step is an abstract wrapper around step of proxy service plug-in
// install/update etc.
type step struct {
	name string
	do   func(*ProxyInstallation) error
	undo func(*ProxyInstallation) error
}

// Name returns the steps name.
func (o step) Name() string {
	return o.name
}

// Do executes the step.
func (o step) Do(v interface{}) error {
	return o.do(v.(*ProxyInstallation))
}

// Undo executes the steps cancel function.
func (o step) Undo(v interface{}) error {
	return o.undo(v.(*ProxyInstallation))
}

func newStep(name string, do func(*ProxyInstallation) error,
	undo func(*ProxyInstallation) error) step {
	blank := func(*ProxyInstallation) error { return nil }
	if do == nil {
		do = blank
	}
	if undo == nil {
		undo = blank
	}
	return step{name: name, do: do, undo: undo}
}
