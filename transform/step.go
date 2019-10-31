// Package transform implements document transforms, which are used by the
// editor to treat changes as first-class values, which can be saved, shared,
// and reasoned about.
package transform

import "github.com/cozy/prosemirror-go/model"

// Step objects represent an atomic change. It generally applies only to the
// document it was created for, since the positions stored in it will only make
// sense for that document.
//
// New steps are defined by creating classes that extend Step, overriding the
// apply, invert, map, getMap and fromJSON methods, and registering your class
// with a unique JSON-serialization identifier using Step.jsonID.
type Step interface {
	// Applies this step to the given document, returning a result
	// object that either indicates failure, if the step can not be
	// applied to this document, or indicates success by containing a
	// transformed document.
	Apply(doc *model.Node) StepResult

	// GetMap gets the step map that represents the changes made by this step,
	// and which can be used to transform between positions in the old and the
	// new document.
	GetMap() *StepMap

	// Invert creates an inverted version of this step. Needs the document as
	// it was before the step as argument.
	Invert(doc *model.Node) Step

	// Map this step through a mappable thing, returning either a version of
	// that step with its positions adjusted, or null if the step was entirely
	// deleted by the mapping.
	Map(mapping Mappable) Step

	// Merge tries to merge this step with another one, to be applied directly
	// after it. Returns the merged step when possible, null if the steps can't
	// be merged.
	Merge(other Step) (Step, bool)
}

// StepResult is the result of applying a step. Contains either a new document
// or a failure value.
type StepResult struct {
	// :: ?Node The transformed document.
	Doc *model.Node
	// :: ?string Text providing information about a failed step.
	Failed string
}

// OK creates a successful step result.
func OK(doc *model.Node) StepResult {
	return StepResult{Doc: doc}
}

// Fail creates a failed step result.
func Fail(message string) StepResult {
	return StepResult{Failed: message}
}

// FromReplace calls Node.replace with the given arguments. Create a successful
// result if it succeeds, and a failed one if it throws a `ReplaceError`.
func FromReplace(doc *model.Node, from, to int, slice *model.Slice) StepResult {
	replaced, err := doc.Replace(from, to, slice)
	if err != nil {
		return Fail(err.Error())
	}
	return OK(replaced)
}
