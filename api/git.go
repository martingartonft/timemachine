package api

import (
	"bytes"
	"code.google.com/p/go-uuid/uuid"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
)

type GitContentAPI struct {
	dir string
}

func (gci GitContentAPI) ByUUID(uuid string) (bool, Content) {
	panic("")
}

func (gci GitContentAPI) Write(c Content) error {
	// validate uuid
	u := uuid.Parse(c.UUID)
	if u.String() != c.UUID {
		return fmt.Errorf("invalid uuid: %v\n", c.UUID)
	}

	// marshal
	json, err := json.Marshal(c)
	if err != nil {
		return err
	}

	// write it to disk
	filename := fmt.Sprintf("%s.json", u.String())
	err = ioutil.WriteFile(path.Join(gci.dir, filename), json, 0644)
	if err != nil {
		return err
	}

	// commit it
	cmd := exec.Command("git", "add", filename)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Dir = gci.dir
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("git add command failed with error %v:\n%s\n", err, out.String())
	}
	cmd = exec.Command("git", "commit", "-m", fmt.Sprintf("adding or updating %s", filename))
	out = bytes.Buffer{}
	cmd.Stdout = &out
	cmd.Dir = gci.dir
	err = cmd.Run()
	if err != nil {
		if strings.Contains(string(out.Bytes()), "nothing to commit, working directory clean") {
			return nil
		}
		return fmt.Errorf("command failed with error %v:\n%s\n", err, out.String())
	}
	return nil
}

func (gci GitContentAPI) Close() {
}

func (gci GitContentAPI) Count() int {
	panic("")
}

func (gci GitContentAPI) init() error {

	if err := os.Mkdir(gci.dir, 0755); err != nil {
		return err
	}

	cmd := exec.Command("git", "init")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Dir = gci.dir
	return cmd.Run()
}

func (gci GitContentAPI) Drop() error {
	if err := os.RemoveAll(gci.dir); err != nil {
		return err
	}
	return gci.init()
}

func (gci GitContentAPI) Recent(stop chan struct{}, limit int) (chan Content, error) {
	panic("")
}

func (gci GitContentAPI) All(stop chan struct{}) (chan Content, error) {
	panic("")
}

func NewGitContentAPI() (ContentAPI, error) {
	api := GitContentAPI{"/tmp/gitapi/"}
	err := api.Drop() //TODO: don't drop every time once things are stable.
	if err != nil {
		return nil, err
	}
	return api, nil
}
