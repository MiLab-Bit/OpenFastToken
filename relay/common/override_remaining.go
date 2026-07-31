package common

func (e *ParamOverrideReturnError) Error() string {
	if e == nil {
		return "param override return error"
	}
	if e.Message == "" {
		return "param override return error"
	}
	return e.Message
}
