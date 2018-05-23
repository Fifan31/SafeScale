package nfs

import (
	"bytes"
	"fmt"
	"html/template"
	"os/exec"
	"syscall"

	rice "github.com/GeertJohan/go.rice"
	"github.com/SafeScale/system"
)

//go:generate rice embed-go

//templateProvider is the instance of TemplateProvider used by package nfs
var tmplBox *rice.box

//getTemplateProvider returns the instance of TemplateProvider
func getTemplateBox() (*rice.Box, error) {
	if tmplBox == nil {
		tmplBox, err := rice.FindBox("../nfs/scripts")
		if err != nil {
			return nil, err
		}
	}
	return tmplBox, nil
}

//executeScript executes a script template with parameters in data map
// Returns retcode, stdout, stderr, error
// If error == nil && retcode != 0, the script ran but failed.
func executeScript(sshconfig system.SSHConfig, name string, data map[string]interface{}) (int, string, string, error) {
	commonTools, err := system.RealizeCommonTools()
	if err != nil {
		return 255, "", "", err
	}
	data["CommonTools"] = commonTools

	tmplBox, err := getTemplateBox()

	// get file content as string
	tmplContent, err := tmplBox.String(name)
	if err != nil {
		return 255, "", nil, err
	}

	// Prepare the template for execution
	tmplPrepared, err := template.New(name).Parse(tmplContent)
	if err != nil {
		return 255, "", "", err
	}

	var buffer bytes.Buffer
	if err := tmplPrepared.Execute(&buffer, data); err != nil {
		return 255, "", "", fmt.Errorf("failed to execute template: %s", err.Error())
	}
	tmplResult := buffer.String()

	sshCmd, err := sshconfig.Command(tmplResult)
	if err != nil {
		return 255, "", "", err
	}

	cmdResult, err := sshCmd.Output()
	retcode := 0
	stdout := string(cmdResult)
	stderr := ""
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			if status, ok := ee.Sys().(syscall.WaitStatus); ok {
				retcode = int(status)
			}
			stderr = string(ee.Stderr)
		} else {
			return 255, "", "", fmt.Errorf("failed to execute script '%s': %s", name, err.Error())
		}
	}
	return retcode, stdout, stderr, nil
}