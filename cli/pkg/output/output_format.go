package output

import "fmt"

func Format(format string) (string, error) {
	switch format {
	case "json":
		return "application/json", nil
	case "yaml":
		return "application/yaml", nil
	case "":
		return "application/yaml", nil
	default:
		return "", fmt.Errorf("invalid format %s", format)
	}
}
