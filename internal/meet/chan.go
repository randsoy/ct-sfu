package meet

// Chan chan
type Chan interface {
	MID() string
	Notify(method string, params interface{}) error
}
