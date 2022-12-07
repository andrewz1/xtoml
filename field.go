package xtoml

type fieldTag struct {
	tag string // struct field tag
	req bool   // is field required
}

func (f *fieldTag) isRequired() bool {
	return f.req
}

func (f *fieldTag) getTag() string {
	return f.tag
}
