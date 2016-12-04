package diffmp

// Op is a diff operation.
type Op int8

// All types of diff operations.
const (
	Delete Op = -1
	Insert Op = 1
	Noop   Op = 0
)

func opStr(op Op) string {
	switch op {
	case Insert:
		return "+"
	case Delete:
		return "-"
	case Noop:
		return " "
	default:
		return "?"
	}
}
