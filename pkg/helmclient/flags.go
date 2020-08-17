package helmclient

import "fmt"

// FlagMap acts as a generic flag type. This type is used as a shortform in the rest
// of the application to represent flags
type FlagMap map[string]interface{}

// flagError is a generic error struct that has a Error defined on it.
// This can be embedded in concrete error structs to satisfy the error contract
type flagError struct {
	error string
}

// Error method on flagError prints the error member on the error struct.
// Concrete members (those that have embedded the flagError type) can
// override this by defining their own implementation on the concrete structs
func (err *flagError) Error() string {
	return err.error
}

// UnsupportedFlagError represents an error returned because of passing
// an unsupported flag
type UnsupportedFlagError struct {
	FlagName string
}

func (err *UnsupportedFlagError) Error() string {
	return fmt.Sprintf("Unsupported Flag: %s", err.FlagName)
}

// InvalidFlagValueError represents an invalid flag value format
// The error field contains the detailed error information
type InvalidFlagValueError struct {
	FlagName string
}

func (err *InvalidFlagValueError) Error() string {
	return fmt.Sprintf("Invalid value for Flag: %s", err.FlagName)
}

// ActionFlag defines the contract for a actionflags stuct.
// The idea is that while initializing a Flags struct, individual actions
// can provide an Action Flag struct that implements a method that
// returns a list of flagSetterFunc for that struct. These flatSetterFuncs
// are merged with the global ones to initialize the flags.
// Since the Flags embeds the ActionFlag, the member on concrete ActionFlags
// are accessible.
type ActionFlag interface {
	// SetLocal sets local action flags
	SetLocal(flagmap FlagMap) error
}

// GlobalFlags is an inventory of all supported flags.
// It exposes methods to validate and get specific flag values.
// It embeds an ActionFlag struct that provides all action specific flags
type GlobalFlags struct {
	KubeCtx       string
	KubeToken     string
	KubeAPIServer string
	Namespace     string

	flagmap FlagMap
}

// flagSetterFunc is a flag setter type that that each flag setter func should conform with
type flagSetterFunc func(flagmap FlagMap, flags *GlobalFlags) error

// namespaceFlagSetter is a setter for namespace flag
func namespaceFlagSetter(flagmap FlagMap, flags *GlobalFlags) error {
	namespace, ok := flagmap["namespace"]

	if ok {
		namespace, ok := namespace.(string)
		if !ok {
			return &InvalidFlagValueError{
				FlagName: "namespace",
			}
		}

		flags.Namespace = namespace
	}

	return nil
}

// flagSetters is list of flag setter funcs. These are executed serially,
// and each setter modifies the globalflag instance passed to it or returns an error
var flagSetters = []flagSetterFunc{
	namespaceFlagSetter,
}

// Returns a new GlobalFlags instance
func NewGlobalFlags(flagmap FlagMap) (*GlobalFlags, error) {
	flags := new(GlobalFlags)
	flags.flagmap = flagmap

	for _, setter := range flagSetters {
		err := setter(flagmap, flags)

		if err != nil {
			return nil, err
		}
	}

	return flags, nil
}
