package debug

import (
	"os/exec"
	"strings"
)

func Init() {
	//os.Args = append(os.Args, "--kubeconfig", "/Users/acejilam/.kube/vmip.config")
	exec.Command("k", strings.Split("k -n kruise-system delete configmap kruise-manager && k -n kruise-system delete  lease kruise-manager", " ")...).Start()
}
