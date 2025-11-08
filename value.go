package bson

// Value represents a BSON value with a type and raw bytes.
type Value struct {
	Type Type
	Data []byte
}

func (v Value) IsNumber() bool {
	switch v.Type {
	case TypeDouble, TypeInt32, TypeInt64, TypeDecimal128:
		return true
	default:
		return false
	}
}
