package errors

func InitializationError(err error, obj string) error {
	return Wrapf(err, "unable to initialize %s", obj)
}
