package namespace

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
)

func StackToNamespace(stack *v1beta1.Stack) string {
	return StackNamespace(stack.Namespace, stack.Name)
}

func StackNamespace(spaceName, stackName string) string {
	parts := strings.Split(stackName, ":")
	stackName = parts[len(parts)-1]

	id := fmt.Sprintf("%s:%s", spaceName, stackName)
	h := sha256.New()
	h.Write([]byte(id))
	hash := hex.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("%s-%s", stackName, string(hash)[:8])
}
