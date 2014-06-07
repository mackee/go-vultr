package main

import (
	"bytes"
	"code.google.com/p/go.crypto/ssh"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

func makeCreateCmd() *commander.Command {
	var createCmd = &commander.Command{
		Run:       runCreateCmd,
		UsageLine: "create [options]",
		Flag:      *flag.NewFlagSet("vultr-create", flag.ExitOnError),
	}

	createCmd.Flag.String("osid", "", "os id (* required)")
	createCmd.Flag.String("dcid", "", "regions id (* required)")
	createCmd.Flag.String("vpsplanid", "", "plan id (* required)")
	return createCmd
}

var listCmd = &commander.Command{
	Run:       runListCmd,
	UsageLine: "list",
}

var regionsCmd = &commander.Command{
	Run:       runGetListHandlesubidCmd,
	UsageLine: "regions",
}

var osCmd = &commander.Command{
	Run:       runGetListHandlesubidCmd,
	UsageLine: "os",
}

var plansCmd = &commander.Command{
	Run:       runGetListHandlesubidCmd,
	UsageLine: "plans",
}

var startCmd = &commander.Command{
	Run:       runPostHandlesubidCmd,
	UsageLine: "start <subid>",
}

var haltCmd = &commander.Command{
	Run:       runPostHandlesubidCmd,
	UsageLine: "halt <subid>",
}

var rebootCmd = &commander.Command{
	Run:       runPostHandlesubidCmd,
	UsageLine: "reboot <subid>",
}

func makeDestroyCmd() *commander.Command {
	var removeCmd = &commander.Command{
		Run:       runDestroyCmd,
		UsageLine: "destroy <subid>",
		Flag:      *flag.NewFlagSet("vultr-destroy", flag.ExitOnError),
	}
	removeCmd.Flag.Bool("yes", false, "")
	return removeCmd
}

var sshCmd = &commander.Command{
	Run:       runSshCmd,
	UsageLine: "ssh <subid> ...",
}

func runListCmd(cmd *commander.Command, args []string) error {
	if err := validateCmdArgs(cmd, args); err != nil {
		return err
	}
	req, err := client.NewRequest("GET", "v1/server/list", nil)

	if err != nil {
		return err
	}
	if _, err = pp(client.Do(req)); err != nil {
		return err
	}
	return nil
}

func runPostHandlesubidCmd(cmd *commander.Command, args []string) error {
	if err := validateCmdArgs(cmd, args); err != nil {
		return err
	}

	subid := args[0]
	form := url.Values{}
	form.Set("SUBID", subid)
	r := strings.NewReader(form.Encode())
	req, err := client.NewRequest("POST", "v1/server/"+cmd.Name(), r)
	if err != nil {
		return err
	}

	if _, err := pp(client.Do(req)); err != nil {
		return err
	}
	return nil
}

func runGetListHandlesubidCmd(cmd *commander.Command, args []string) error {
	if err := validateCmdArgs(cmd, args); err != nil {
		return err
	}

	req, err := client.NewRequest("GET", "v1/"+cmd.Name()+"/list", nil)
	if err != nil {
		return err
	}
	if _, err := pp(client.Do(req)); err != nil {
		return err
	}
	return nil
}

func runDestroyCmd(cmd *commander.Command, args []string) error {
	if err := validateCmdArgs(cmd, args); err != nil {
		return err
	}
	yes := cmd.Flag.Lookup("yes").Value.Get().(bool)
	subid := args[0]

	if !yes {
		fmt.Printf("Really destroy %s? [y/N]", subid)
		var y string
		fmt.Scanf("%s", &y)
		if y != "y" && y != "Y" {
			os.Exit(1)
		}
	}
	return runPostHandlesubidCmd(cmd, args)
}

func runCreateCmd(cmd *commander.Command, args []string) error {
	if err := validateCmdArgs(cmd, args); err != nil {
		return err
	}
	osid := cmd.Flag.Lookup("osid").Value.Get().(string)
	dcid := cmd.Flag.Lookup("dcid").Value.Get().(string)
	vpsplanid := cmd.Flag.Lookup("vpsplanid").Value.Get().(string)
	if osid == "" || dcid == "" || vpsplanid == "" {
		return errors.New("Usage: " + cmd.UsageLine)
	}

	form := url.Values{}
	form.Set("OSID", osid)
	form.Set("DCID", dcid)
	form.Set("VPSPLANID", vpsplanid)

	r := strings.NewReader(form.Encode())

	req, err := client.NewRequest("POST", "v1/server/create", r)
	if err != nil {
		return err
	}

	if _, err := pp(client.Do(req)); err != nil {
		return err
	}

	return nil
}

type serverInfo struct {
	Os                 string `json:"os"`
	Ram                string `json:"ram"`
	Disk               string `json:"disk"`
	MainIp             string `json:"main_ip"`
	VcpuCount          string `json:"vcpu_count"`
	Location           string `json:"location"`
	DefaultPassword    string `json:"default_password"`
	DateCreated        string `json:"date_created"`
	PendingCharges     string `json:"pending_charges"`
	Status             string `json:"status"`
	CostPerMonth       string `json:"cost_per_month"`
	CurrentBandwidthGb int    `json:"current_bandwidth_gb"`
}

func runSshCmd(cmd *commander.Command, args []string) error {
	if err := validateCmdArgs(cmd, args); err != nil {
		return err
	}

	req, err := client.NewRequest("GET", "v1/server/list", nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		pp(resp, err)
		return err
	}

	dec := json.NewDecoder(resp.Body)
	defer resp.Body.Close()

	m := make(map[string]serverInfo)
	if err := dec.Decode(&m); err != nil {
		return err
	}
	target := m[args[0]]
	log.Println("IP:", target.MainIp)
	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password(target.DefaultPassword),
		},
	}
	client, err := ssh.Dial("tcp", target.MainIp+":22", config)
	if err != nil {
		panic("Failed to dial: " + err.Error())
	}

	session, err := client.NewSession()
	if err != nil {
		panic("Failed to create session: " + err.Error())
	}
	defer session.Close()

	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	// Set up terminal modes
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	// Request pseudo terminal
	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		log.Fatalf("request for pseudo terminal failed: %s", err)
	}
	// Start remote shell
	if err := session.Shell(); err != nil {
		log.Fatalf("failed to start shell: %s", err)
	}

	session.Run("bash")
	session.Wait()

	return nil
}

func validateCmdArgs(cmd *commander.Command, args []string) error {
	mustVal := 0
	optionVal := 0
	must := regexp.MustCompile(`<\w+>`)
	option := regexp.MustCompile(`\[\w+\]`)

	for _, action := range strings.Split(cmd.UsageLine, " ") {
		if must.MatchString(action) {
			mustVal++
		}
		if option.MatchString(action) {
			optionVal++
		}
		// skip arguments validation
		if action == "..." {
			if mustVal <= len(args) {
				return nil
			}
		}
	}
	if mustVal <= len(args) && len(args) <= mustVal+optionVal {
		return nil
	}
	return errors.New("invalid argument\nUsage: " + cmd.UsageLine)
}

func writeFileField(w *multipart.Writer, fieldsubid string, filePath string) error {
	ff, err := w.CreateFormField(fieldsubid)
	if err != nil {
		return err
	}
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	if _, err := io.Copy(ff, file); err != nil {
		return err
	}
	return nil
}

func pp(resp *http.Response, err error) (*http.Response, error) {
	if err != nil {
		return resp, err
	}
	log.Println(resp.Status)

	defer resp.Body.Close()
	src, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp, err
	}

	index := strings.Index(resp.Header.Get("Content-Type"), "json")

	if index > 0 {
		var b bytes.Buffer
		err := json.Indent(&b, src, "", "    ")
		if err != nil {
			return resp, err
		}
		fmt.Println(b.String())
	} else {
		fmt.Println(string(src))
	}
	return resp, err
}
