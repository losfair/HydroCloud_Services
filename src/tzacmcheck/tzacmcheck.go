package tzacmcheck

import "os/exec"

func DoCheck(user string,pw string) string {
	cmd := exec.Command("phantomjs","js/tzacmcheck_login.js",user,pw)
	out,_ := cmd.Output()
	return string(out)
}
