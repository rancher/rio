package namespace

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/rancher/rio/types/apis/rio.cattle.io/v1"
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
