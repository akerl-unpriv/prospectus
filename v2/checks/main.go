package checks

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/akerl/prospectus/expectations"
)

const (
	prospectusDirName = ".prospectus.d"
)

// TODO: add timber logging
// TODO: add parallelization

// Check defines a single check that is ready for execution
type Check struct {
	Dir      string            `json:"dir"`
	File     string            `json:"file"`
	Name     string            `json:"name"`
	Metadata map[string]string `json:"metadata"`
}

// CheckSet defines a group of Checks
type CheckSet []Check

// Result defines the results of executing a Check
type Result struct {
	Actual   string               `json:"actual"`
	Expected expectations.Wrapper `json:"expected"`
	Check    Check                `json:"check"`
}

// ResultSet defines a group of Results
type ResultSet []Result

type loadCheckInput struct {
	Dir string `json:"dir"`
}

// NewSet returns a CheckSet based on a provided list of directories
func NewSet(relativeDirs []string) (CheckSet, error) {
	var err error

	dirs := make([]string, len(relativeDirs))
	for index, item := range relativeDirs {
		dirs[index], err = filepath.Abs(item)
		if err != nil {
			return CheckSet{}, err
		}
	}

	var cs CheckSet
	for _, item := range dirs {
		newSet, err := newSetFromDir(item)
		if err != nil {
			return CheckSet{}, err
		}
		cs = append(cs, newSet...)
	}

	return cs, nil
}

func newSetFromDir(absoluteDir string) (CheckSet, error) {
	prospectusDir := filepath.Join(absoluteDir, prospectusDirName)

	fileObjs, err := ioutil.ReadDir(prospectusDir)
	if err != nil {
		return CheckSet{}, err
	}

	var cs CheckSet
	for _, fileObj := range fileObjs {
		file := filepath.Join(prospectusDir, fileObj.Name())
		newSet, err := newSetFromFile(absoluteDir, file)
		if err != nil {
			return CheckSet{}, err
		}
		cs = append(cs, newSet...)
	}

	return cs, nil
}

func newSetFromFile(dir, file string) (CheckSet, error) {
	cs := CheckSet{}
	input := loadCheckInput{Dir: dir}
	err := execProspectusFile(file, "load", input, &cs)
	if err != nil {
		return CheckSet{}, fmt.Errorf("Failed loading %s: %s", file, err)
	}
	for index := range cs {
		cs[index].Dir = dir
		cs[index].File = file
	}
	return cs, nil
}

func execProspectusFile(file, command string, input interface{}, output interface{}) error {
	cmd := exec.Command(file, command)

	inputBytes, err := json.Marshal(input)
	if err != nil {
		return err
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdin.Write(inputBytes)
	stdin.Close()

	stdout, err := cmd.Output()
	if err != nil {
		return err
	}

	return json.Unmarshal(stdout, output)
}

// Execute returns the Results from a CheckSet by calling Execute on each Check
func (cs CheckSet) Execute() ResultSet {
	resultSet := make(ResultSet, len(cs))
	for index, item := range cs {
		resultSet[index] = item.Execute()
	}
	return resultSet
}

// Execute runs the Check and returns Results
func (c Check) Execute() Result {
	return execProspectusForResult("execute", c, c)
}

func execProspectusForResult(method string, c Check, input interface{}) Result {
	r := Result{}
	err := execProspectusFile(c.File, method, input, &r)
	if err != nil {
		return NewErrorResult(fmt.Sprintf("%s error: %s", method, err), c)
	}
	r.Check = c
	return r
}

// String returns the Result as a human-readable string
func (c Check) String() string {
	return fmt.Sprintf(
		"%s::%s",
		c.Dir,
		c.Name,
	)
}

// Changed filters a ResultSet to only Results which do not match
func (rs ResultSet) Changed() ResultSet {
	var newResultSet ResultSet
	for _, item := range rs {
		if !item.Matches() {
			newResultSet = append(newResultSet, item)
		}
	}
	return newResultSet
}

// Matches returns true if the Expected and Actual values of the Result match
func (r Result) Matches() bool {
	return r.Expected.Matches(r.Actual)
}

// JSON returns the ResultsSet as a marshalled JSON string
func (rs ResultSet) JSON() (string, error) {
	data, err := json.MarshalIndent(rs, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// String returns the ResultsSet as a human-readable string
func (rs ResultSet) String() string {
	var b strings.Builder
	for _, item := range rs {
		b.WriteString(item.String())
		b.WriteString("\n")
	}
	return b.String()
}

// String returns the Result as a human-readable string
func (r Result) String() string {
	return fmt.Sprintf(
		"%s: %s / %s",
		r.Check,
		r.Actual,
		r.Expected.String(),
	)
}

// Fix attempts to resolve a mismatched expectation
func (r Result) Fix() Result {
	return execProspectusForResult("fix", r.Check, r)
}

// NewErrorResult creates an error result from a given string
func NewErrorResult(msg string, c Check) Result {
	return Result{
		Actual: "error",
		Expected: expectations.Wrapper{
			Type: "error",
			Data: map[string]string{"msg": msg},
		},
		Check: c,
	}
}
