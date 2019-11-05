package util

func ToInt32(v int) *int32 {
	r := int32(v)
	return &r
}

func ToInt(v *int32) *int {
	s := *v
	b := int(s)
	return &b
}
