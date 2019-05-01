package condition

import (
	"reflect"
	"time"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

type Cond string

func (c Cond) GetStatus(obj runtime.Object) string {
	return getStatus(obj, string(c))
}

func (c Cond) SetError(obj runtime.Object, reason string, err error) {
	if err == nil {
		c.True(obj)
		c.Message(obj, "")
		c.Reason(obj, reason)
		return
	}
	if reason == "" {
		reason = "Error"
	}
	c.False(obj)
	c.Message(obj, err.Error())
	c.Reason(obj, reason)
}

func (c Cond) MatchesError(obj runtime.Object, reason string, err error) bool {
	if err == nil {
		return c.IsTrue(obj) &&
			c.GetMessage(obj) == "" &&
			c.GetReason(obj) == reason
	}
	if reason == "" {
		reason = "Error"
	}
	return c.IsFalse(obj) &&
		c.GetMessage(obj) == err.Error() &&
		c.GetReason(obj) == reason
}

func (c Cond) SetStatus(obj runtime.Object, status string) {
	setStatus(obj, string(c), status)
}

func (c Cond) SetStatusBool(obj runtime.Object, val bool) {
	if val {
		setStatus(obj, string(c), "True")
	} else {
		setStatus(obj, string(c), "False")
	}
}

func (c Cond) True(obj runtime.Object) {
	setStatus(obj, string(c), "True")
}

func (c Cond) IsTrue(obj runtime.Object) bool {
	return getStatus(obj, string(c)) == "True"
}

func (c Cond) False(obj runtime.Object) {
	setStatus(obj, string(c), "False")
}

func (c Cond) IsFalse(obj runtime.Object) bool {
	return getStatus(obj, string(c)) == "False"
}

func (c Cond) Unknown(obj runtime.Object) {
	setStatus(obj, string(c), "Unknown")
}

func (c Cond) IsUnknown(obj runtime.Object) bool {
	return getStatus(obj, string(c)) == "Unknown"
}

func (c Cond) LastUpdated(obj runtime.Object, ts string) {
	setTS(obj, string(c), ts)
}

func (c Cond) GetLastUpdated(obj runtime.Object) string {
	return getTS(obj, string(c))
}

func (c Cond) CreateUnknownIfNotExists(obj runtime.Object) {
	condSlice := getValue(obj, "Status", "Conditions")
	cond := findCond(condSlice, string(c))
	if cond == nil {
		c.Unknown(obj)
	}
}

func (c Cond) Reason(obj runtime.Object, reason string) {
	cond := findOrCreateCond(obj, string(c))
	getFieldValue(cond, "Reason").SetString(reason)
}

func (c Cond) GetReason(obj runtime.Object) string {
	cond := findOrNotCreateCond(obj, string(c))
	if cond == nil {
		return ""
	}
	return getFieldValue(*cond, "Reason").String()
}

func (c Cond) SetMessageIfBlank(obj runtime.Object, message string) {
	if c.GetMessage(obj) == "" {
		c.Message(obj, message)
	}
}

func (c Cond) Message(obj runtime.Object, message string) {
	cond := findOrCreateCond(obj, string(c))
	setValue(cond, "Message", message)
}

func (c Cond) GetMessage(obj runtime.Object) string {
	cond := findOrNotCreateCond(obj, string(c))
	if cond == nil {
		return ""
	}
	return getFieldValue(*cond, "Message").String()
}

func (c Cond) Once(obj runtime.Object, f func() (runtime.Object, error)) error {
	if c.IsFalse(obj) {
		return errors.New(c.GetReason(obj))
	}

	return c.DoUntilTrue(obj, f)
}

func (c Cond) DoUntilTrue(obj runtime.Object, f func() (runtime.Object, error)) error {
	if c.IsTrue(obj) {
		return nil
	}

	return c.Do(f)
}

func messageAndReason(err error) (string, string) {
	if err == nil {
		return "", ""
	}

	switch ce := err.(type) {
	case *conditionError:
		return err.Error(), ce.reason
	default:
		return err.Error(), "Error"
	}
}

func (c Cond) Do(f func() (runtime.Object, error)) error {
	obj, err := f()

	if apierrors.IsConflict(err) {
		// Don't update condition state on conflicts
		return err
	}

	message, reason := messageAndReason(err)
	c.SetStatusBool(obj, err == nil)
	c.Message(obj, message)
	c.Reason(obj, reason)

	return err
}

func touchTS(value reflect.Value) {
	now := time.Now().Format(time.RFC3339)
	getFieldValue(value, "LastUpdateTime").SetString(now)
}

func getStatus(obj interface{}, condName string) string {
	cond := findOrNotCreateCond(obj, condName)
	if cond == nil {
		return ""
	}
	return getFieldValue(*cond, "Status").String()
}

func setTS(obj interface{}, condName, ts string) {
	cond := findOrCreateCond(obj, condName)
	getFieldValue(cond, "LastUpdateTime").SetString(ts)
}

func getTS(obj interface{}, condName string) string {
	cond := findOrNotCreateCond(obj, condName)
	if cond == nil {
		return ""
	}
	return getFieldValue(*cond, "LastUpdateTime").String()
}

func setStatus(obj interface{}, condName, status string) {
	cond := findOrCreateCond(obj, condName)
	setValue(cond, "Status", status)
}

func setValue(cond reflect.Value, fieldName, newValue string) {
	value := getFieldValue(cond, fieldName)
	if value.String() != newValue {
		value.SetString(newValue)
		touchTS(cond)
	}
}

func findOrNotCreateCond(obj interface{}, condName string) *reflect.Value {
	condSlice := getValue(obj, "Status", "Conditions")
	return findCond(condSlice, condName)
}

func findOrCreateCond(obj interface{}, condName string) reflect.Value {
	condSlice := getValue(obj, "Status", "Conditions")
	cond := findCond(condSlice, condName)
	if cond != nil {
		return *cond
	}

	newCond := reflect.New(condSlice.Type().Elem()).Elem()
	newCond.FieldByName("Type").SetString(condName)
	newCond.FieldByName("Status").SetString("Unknown")
	condSlice.Set(reflect.Append(condSlice, newCond))
	return *findCond(condSlice, condName)
}

func findCond(val reflect.Value, name string) *reflect.Value {
	for i := 0; i < val.Len(); i++ {
		cond := val.Index(i)
		typeVal := getFieldValue(cond, "Type")
		if typeVal.String() == name {
			return &cond
		}
	}

	return nil
}

func getValue(obj interface{}, name ...string) reflect.Value {
	if obj == nil {
		return reflect.Value{}
	}
	v := reflect.ValueOf(obj)
	t := v.Type()
	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = v.Type()
	}

	field := v.FieldByName(name[0])
	if len(name) == 1 {
		return field
	}
	return getFieldValue(field, name[1:]...)
}

func getFieldValue(v reflect.Value, name ...string) reflect.Value {
	field := v.FieldByName(name[0])
	if len(name) == 1 {
		return field
	}
	return getFieldValue(field, name[1:]...)
}

func Error(reason string, err error) error {
	return &conditionError{
		reason:  reason,
		message: err.Error(),
	}
}

type conditionError struct {
	reason  string
	message string
}

func (e *conditionError) Error() string {
	return e.message
}
