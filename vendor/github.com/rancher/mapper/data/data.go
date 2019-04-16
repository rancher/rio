package data

type Object interface {
	Delete()
	DeleteField(name string)
	Field(name string) Object
	Set(val interface{}) Object
	SetField(name string, val interface{}) Object
	Exists() bool

	Map() map[string]interface{}
	String() string
	Interface() interface{}
}

type Map interface {
}
