package ipset

import "os/exec"

var bin = ""

func init() {
	bin, _ = exec.LookPath("ipset")
}

func Create(name string, setType string) error {
	cmd := exec.Command(bin, "create", name, setType)
	return cmd.Run()

}
func Destroy(name string) error {
	cmd := exec.Command(bin, "create", name)
	return cmd.Run()

}
