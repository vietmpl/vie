package analysis

// TypeVar represents an identifier whose concrete type cannot be directly
// inferred in the current context. It serves as a placeholder until all
// usages are analyzed, at which point its type is inferred from context.
type TypeVar string

func (tv TypeVar) String() string {
	return string(tv)
}
