// Package parser provides functionality for parsing .env files and annotations.
package parser

// VariableType represents the type of a variable value.
type VariableType int

const (
	// TypeString represents a string variable type.
	TypeString VariableType = iota
	// TypeInt represents an integer variable type.
	TypeInt
	// TypeNumeric represents a numeric (float) variable type.
	TypeNumeric
	// TypeBoolean represents a boolean variable type.
	TypeBoolean
	// TypeEnum represents an enum variable type with predefined options.
	TypeEnum
	// TypeObject represents a structured object type (JSON/YAML).
	TypeObject
)

// String returns the string representation of a VariableType.
func (t VariableType) String() string {
	switch t {
	case TypeString:
		return "string"
	case TypeInt:
		return "int"
	case TypeNumeric:
		return "numeric"
	case TypeBoolean:
		return "boolean"
	case TypeEnum:
		return "enum"
	case TypeObject:
		return "object"
	default:
		return "unknown"
	}
}

// ParseVariableType parses a string into a VariableType.
func ParseVariableType(s string) VariableType {
	switch s {
	case "string":
		return TypeString
	case "int":
		return TypeInt
	case "numeric":
		return TypeNumeric
	case "boolean":
		return TypeBoolean
	case "enum":
		return TypeEnum
	case "object":
		return TypeObject
	default:
		return TypeString // Default to string if unknown
	}
}

// Constraint represents a validation constraint attached to an annotation.
type Constraint struct {
	Name  string // "min", "max", "minlen", "maxlen", "pattern", "options", "format", "encoding"
	Value string // Raw string value; parsed per constraint type
}

// Annotation represents metadata extracted from an inline comment on a variable line.
type Annotation struct {
	PromptText  string       // The question to ask the user
	Type        VariableType // The type of variable
	Constraints []Constraint // Validation constraints
	IsOptional  bool         // Whether the variable is optional
	IsSecret    bool         // Whether to hide input/output
}

// GetConstraint returns the constraint value for a given name, or empty string if not found.
func (a *Annotation) GetConstraint(name string) string {
	for _, c := range a.Constraints {
		if c.Name == name {
			return c.Value
		}
	}
	return ""
}

// HasConstraint checks if the annotation has a specific constraint.
func (a *Annotation) HasConstraint(name string) bool {
	for _, c := range a.Constraints {
		if c.Name == name {
			return true
		}
	}
	return false
}

// Variable represents a single environment variable with optional annotation.
type Variable struct {
	Name       string      // Variable name (e.g., "DB_HOST")
	Value      string      // Variable value (may be empty string)
	Annotation *Annotation // nil if no annotation present
	LineNumber int         // 1-indexed line number in source file
	IsSet      bool        // true if value was explicitly set (vs undefined)
}

// Comment represents a standalone comment line (not attached to a variable).
type Comment struct {
	Text       string // The comment text (without #)
	LineNumber int    // 1-indexed line number in source file
}

// EnvFile represents a parsed .env or .env.dist file.
type EnvFile struct {
	Path      string         // File path
	Variables []Variable     // Variables in order of appearance
	Config    *KrakenvConfig // Krakenv configuration (nil if not a distributable)
	Comments  []Comment      // Standalone comments
}

// GetVariable returns a variable by name, or nil if not found.
func (f *EnvFile) GetVariable(name string) *Variable {
	for i := range f.Variables {
		if f.Variables[i].Name == name {
			return &f.Variables[i]
		}
	}
	return nil
}

// HasVariable checks if a variable exists in the file.
func (f *EnvFile) HasVariable(name string) bool {
	return f.GetVariable(name) != nil
}

// KrakenvConfig represents project-level krakenv configuration extracted from distributable.
type KrakenvConfig struct {
	Environments []string // e.g., ["local", "testing", "production"]
	Strict       bool     // If true, unannotated variables are errors
	DistPath     string   // Override default .env.dist path
}

// DefaultKrakenvConfig returns a KrakenvConfig with default values.
func DefaultKrakenvConfig() *KrakenvConfig {
	return &KrakenvConfig{
		Environments: []string{"local"},
		Strict:       false,
		DistPath:     ".env.dist",
	}
}
