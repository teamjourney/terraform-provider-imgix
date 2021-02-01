package imgix

func String(v interface{}) *string {
	vp := v.(string)
	return &vp
}

func StringNilIfEmpty(v interface{}) *string {
	vp := v.(string)
	if vp == "" {
		return nil
	}
	return &vp
}

func Int(v interface{}) *int {
	vp := v.(int)
	return &vp
}

func Bool(v interface{}) *bool {
	switch i := v.(type) {
	case *bool:
		return i
	case bool:
		return &i
	}
	return nil
}

func SliceString(v interface{}) []string {
	rs := v.([]interface{})
	s := make([]string, len(rs))
	for i, v := range rs {
		s[i] = v.(string)
	}
	return s
}
