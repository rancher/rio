package namespace

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	v1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
)

const (
	CloudNamespace = "rio-cloud"
)

func StackToNamespace(stack *v1.Stack) string {
	return StackNamespace(stack.Namespace, stack.Name)
}

func StackNamespace(projectName, stackName string) string {
	parts := strings.Split(stackName, ":")
	stackName = parts[len(parts)-1]

	id := fmt.Sprintf("%s:%s", projectName, stackName)
	h := sha256.New()
	h.Write([]byte(id))
	hash := hex.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("%s-%s", stackName, string(hash)[:8])
}

func StackNamespaceOnlyHash(projectName, stackName string) string {
	parts := strings.Split(stackName, ":")
	stackName = parts[len(parts)-1]

	id := fmt.Sprintf("%s:%s", projectName, stackName)
	h := sha256.New()
	h.Write([]byte(id))
	hash := hex.EncodeToString(h.Sum(nil))
	return string(hash)[:8]
}

func HashIfNeed(name, stackName, projectName string) string {
	fullPath := fmt.Sprintf("%s-%s", name, StackNamespaceOnlyHash(projectName, stackName))
	if len(fullPath) > 63 {
		digest := sha256.Sum256([]byte(fullPath))
		return fullPath[0:57] + "-" + string(digest[:])[0:5]
	}
	return fullPath
}

func NameRefWithNamespace(name string, stack *v1.Stack) (string, string) {
	return CloudNamespace, HashIfNeed(name, stack.Name, stack.Namespace)
}

func NameRef(name string, stack *v1.Stack) string {
	return HashIfNeed(name, stack.Name, stack.Namespace)
}
