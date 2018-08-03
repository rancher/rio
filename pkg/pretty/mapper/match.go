package mapper

func NewMatch(field string) ObjectsToSlice {
	return ObjectsToSlice{
		Field: field,
		NewObject: func() MaybeStringer {
			return &PrettyMatch{}
		},
		ToObject: func(obj interface{}) (interface{}, error) {
			if s, ok := obj.(string); ok {
				return map[string]interface{}{
					"path": s,
				}, nil
			}
			return obj, nil
		},
	}
}

type PrettyMatch struct {
	Cookies map[string]string `json:"cookies,omitempty"`
	From    string            `json:"from,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Method  string            `json:"method,omitempty"`
	Path    string            `json:"path,omitempty"`
	Port    int64             `json:"port,omitempty"`
	Scheme  string            `json:"scheme,omitempty"`
}

func (p *PrettyMatch) MaybeString() interface{} {
	if p.From != "" ||
		p.Method != "" ||
		p.Port > 0 ||
		p.Scheme != "" ||
		len(p.Cookies) > 0 ||
		len(p.Headers) > 0 {
		return p
	}

	return p.Path
}
