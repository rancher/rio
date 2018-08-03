package yaml

func IsYAMLFile(check, prefix string) bool {
	return check == prefix+".yaml" ||
		check == prefix+".yml"
}
