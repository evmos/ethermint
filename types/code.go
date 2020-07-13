package types

// ----------------------------------------------------------------------------
// Code
// ----------------------------------------------------------------------------

// Code is account Code type alias
type Code []byte

func (c Code) String() string {
	return string(c)
}
