package main

import (
	"github.com/gonuts/commander"
	"testing"
)

func Test_validateCmdArgs(t *testing.T) {
	var testCmd = &commander.Command{
		Run:       runListCmd,
		UsageLine: "list <hoge> <fuga> [yap] [dap]",
	}
	var err error
	err = validateCmdArgs(testCmd, []string{"hoge"})
	if err == nil {
		t.Error("should error")
	}

	err = validateCmdArgs(testCmd, []string{"hoge", "fuga"})
	if err != nil {
		t.Error("should not error")
	}

	err = validateCmdArgs(testCmd, []string{"hoge", "fuga", "opt", "opt"})
	if err != nil {
		t.Error("options are allowed")
	}

	err = validateCmdArgs(testCmd, []string{"hoge", "fuga", "opt", "opt", "opt"})
	if err == nil {
		t.Error("too many options")
	}

	testCmd = &commander.Command{
		Run:       runListCmd,
		UsageLine: "list <hoge> ...",
	}

	err = validateCmdArgs(testCmd, []string{"hoge", "fuga", "yap"})
	if err != nil {
		t.Error("... is stop validation")
	}
}
