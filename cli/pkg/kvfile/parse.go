package kvfile

import "os"

// Copied from github.com/docker/cli/opts a23c5d157b5265520ae41c133e988a85ac1b4606

func ReadKVEnvStrings(files []string, override []string) ([]string, error) {
	return readKVStrings(files, override, os.Getenv)
}

func readKVStrings(files []string, override []string, emptyFn func(string) string) ([]string, error) {
	variables := []string{}
	for _, ef := range files {
		parsedVars, err := parseKeyValueFile(ef, emptyFn)
		if err != nil {
			return nil, err
		}
		variables = append(variables, parsedVars...)
	}
	// parse the '-e' and '--env' after, to allow override
	variables = append(variables, override...)

	return variables, nil
}

func ReadKVStrings(files []string, override []string) ([]string, error) {
	return readKVStrings(files, override, nil)
}
