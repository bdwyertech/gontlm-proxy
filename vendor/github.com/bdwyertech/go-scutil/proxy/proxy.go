package proxy

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type ProxyConfig struct {
	ExceptionsList           []string `json:"ExceptionsList,omitempty"`
	ExcludeSimpleHostnames   string   `json:"ExcludeSimpleHostnames,omitempty"`
	FTPPassive               string   `json:"FTPPassive,omitempty"`
	HTTPEnable               string   `json:"HTTPEnable,omitempty"`
	HTTPPort                 string   `json:"HTTPPort,omitempty"`
	HTTPProxy                string   `json:"HTTPProxy,omitempty"`
	HTTPUser                 string   `json:"HTTPUser,omitempty"`
	HTTPSEnable              string   `json:"HTTPSEnable,omitempty"`
	HTTPSPort                string   `json:"HTTPSPort,omitempty"`
	HTTPSProxy               string   `json:"HTTPSProxy,omitempty"`
	HTTPSUser                string   `json:"HTTPSUser,omitempty"`
	ProxyAutoConfigEnable    string   `json:"ProxyAutoConfigEnable,omitempty"`
	ProxyAutoConfigURLString string   `json:"ProxyAutoConfigURLString,omitempty"`
	ProxyAutoDiscoveryEnable string   `json:"ProxyAutoDiscoveryEnable,omitempty"`
	RTSPEnable               string   `json:"RTSPEnable,omitempty"`
	RTSPPort                 string   `json:"RTSPPort,omitempty"`
	RTSPProxy                string   `json:"RTSPProxy,omitempty"`
	RTSPUser                 string   `json:"RTSPUser,omitempty"`
	SOCKSEnable              string   `json:"SOCKSEnable,omitempty"`
	SOCKSPort                string   `json:"SOCKSPort,omitempty"`
	SOCKSProxy               string   `json:"SOCKSProxy,omitempty"`
	SOCKSUser                string   `json:"SOCKSUser,omitempty"`
}

func Get() (p ProxyConfig, err error) {
	cmd := exec.Command("scutil", "--proxy")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		return
	}
	br := bufio.NewReader(&out)
	firstLine, err := br.ReadString('\n')
	if err != nil {
		return
	}
	firstLine = strings.TrimSpace(firstLine)
	if firstLine != "<dictionary> {" {
		err = fmt.Errorf("unexpected format: %s", firstLine)
		return
	}

	res, err := parseDict(bufio.NewScanner(br))
	if err != nil {
		return
	}

	jsonBytes, err := json.Marshal(res)
	if err != nil {
		return
	}

	err = json.Unmarshal(jsonBytes, &p)

	return
}

func (p *ProxyConfig) ToJSON() string {
	prettyJSON, _ := json.MarshalIndent(p, "", "  ")

	return string(prettyJSON)
}

var errEndOfBlock = errors.New("eob")

func parseDict(scanner *bufio.Scanner) (map[string]interface{}, error) {
	res := make(map[string]interface{})

	for scanner.Scan() {
		line := scanner.Text()
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		key, value, err := parseLine(scanner, line)
		if err != nil {
			if err != errEndOfBlock {
				return nil, err
			}
			return res, nil
		}
		res[key] = value
	}
	return res, nil
}

func parseArray(scanner *bufio.Scanner) ([]interface{}, error) {
	res := make([]interface{}, 0)

	for scanner.Scan() {
		line := scanner.Text()
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		_, value, err := parseLine(scanner, line)
		if err != nil {
			if err != errEndOfBlock {
				return nil, err
			}
			return res, nil
		}
		res = append(res, value)
	}

	return res, nil
}

func parseLine(scanner *bufio.Scanner, line string) (key string, value interface{}, err error) {
	parts := strings.SplitN(line, " : ", 2)
	if len(parts) == 1 {
		if line[len(line)-1] == '}' {
			return "", nil, errEndOfBlock
		}
		return "", nil, fmt.Errorf("do not know how to parse: %s", line)
	}

	switch p := parts[1]; p {
	case "<dictionary> {":
		value, err = parseDict(scanner)
	case "<array> {":
		value, err = parseArray(scanner)
	default:
		value = p
	}

	return strings.TrimSpace(parts[0]), value, err
}
