package api

import (
	"bufio"
	"bytes"
	"code.google.com/p/go-uuid/uuid"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
	"time"
)

type GitContentAPI struct {
	dir string
}

func (gci GitContentAPI) ByUUID(uuid string) (bool, Content) {
	var content Content
	bytes, err := ioutil.ReadFile(path.Join(gci.dir, fmt.Sprintf("%s.json", uuid)))
	if err != nil {
		panic(err)
		return false, Content{}
	}
	err = json.Unmarshal(bytes, &content)
	if err != nil {
		panic(err)
		return false, Content{}
	}

	return true, content
}

func (gci GitContentAPI) ByUUIDAndDate(id string, dateTime time.Time) (bool, Content) {
	filename := fmt.Sprintf("%s.json", id)
	cmd := exec.Command("git", "log", "--date=iso", "--pretty=format:%ad%x08%x08%x08%x08%x08%x08 %H", filename) //git log --date=iso --pretty=format:'%ad%x08%x08%x08%x08%x08%x08'
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	cmd.Dir = gci.dir
	err := cmd.Run()
	if err != nil {
		log.Fatalf("git log command failed with error %v:\n%s\n", err, out.String())
		return false, Content{}
	}

	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		line := scanner.Text()
		if containsDateTime(line, dateTime) {
			gitData := strings.Split(line, " ")
			log.Printf("Git Data is %v\n", gitData)
			return gci.Version(id, gitData[3]) // Why gitData[3] and why not gitData[2]
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return gci.ByUUID(id)
}

func (gci GitContentAPI) Version(id string, hash string) (bool, Content) {
	log.Printf("Returning content:%s @ version: %s\n", id, hash)

	filename := fmt.Sprintf("%s.json", id)
	cmd := exec.Command("git", "show", hash+":"+filename) //git show 0718a08eea5480ce0eed731a5d3157086548e3d8:c5a98003-0a8c-4fd0-9707-df880e1627b5.json
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	cmd.Dir = gci.dir
	err := cmd.Run()
	if err != nil {
		log.Printf("git show command failed with error %v:\n%s\n", err, out.String())
		return false, Content{}
	}

	var content Content
	err = json.Unmarshal(out.Bytes(), &content)
	if err != nil {
		log.Fatalf("git show command failed with error %v:\n%s\n", err, out.String())
	}
	return true, content

}

func containsDateTime(item string, dateTime time.Time) bool {
	items := strings.Split(strings.Trim(item, "'"), " ")
	commitDateTime, err := time.Parse(time.RFC3339, items[0]+"T"+items[1]+"Z")
	if err != nil {
		log.Printf("Error when parsing date time %v\n", err)
	}

	return dateTime.After(commitDateTime) || dateTime.Equal(commitDateTime)

}

func (gci GitContentAPI) Versions(id string) (versions []Version) {
	filename := fmt.Sprintf("%s.json", id)
	cmd := exec.Command("git", "log", "--pretty=format:%ai %H", filename)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	cmd.Dir = gci.dir
	err := cmd.Run()
	if err != nil {
		log.Fatalf("git log command failed with error %v:\n%s\n", err, out.String())
		return nil
	}

	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		line := scanner.Text()
		gitData := strings.Split(line, " ")

		dateString := gitData[0] + " " + gitData[1] + " " + gitData[2]
		pubDate, err := time.Parse("2006-01-02 15:04:05 -0700", dateString)
		if err != nil {
			log.Fatal(err)
		}
		version := Version{
			UUID:          id,
			Version:       gitData[3],
			PublishedDate: pubDate,
			PDString: pubDate.Local().Format("2006-01-02 15:04:05"),
		}
		versions = append(versions, version)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	sort.Sort(Versions(versions))

	return versions
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

	// add to git index
	cmd := exec.Command("git", "add", filename)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Dir = gci.dir
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("git add command failed with error %v:\n%s\n", err, out.String())
	}

	// do commit
	cmd = exec.Command("git", "commit", "-m", fmt.Sprintf("adding or updating %s", filename),
		fmt.Sprintf("--date=%s", c.PublishedDate.UTC().Format(time.RFC3339)))
	out = bytes.Buffer{}
	cmd.Stdout = &out
	cmd.Stderr = &out
	cmd.Dir = gci.dir
	err = cmd.Run()
	if err != nil {
		if strings.Contains(string(out.Bytes()), "nothing to commit, working directory clean") {
			return nil
		}
		return fmt.Errorf("git commit command failed with error %v:\n%s\n", err, out.String())
	}
	return nil
}

func (gci GitContentAPI) Close() {
}

func (gci GitContentAPI) Count() int {
	count := 0
	files, _ := ioutil.ReadDir(gci.dir)
	for _, f := range files {
		if !f.IsDir() {
			count++
		}
	}

	return count
}

func (gci GitContentAPI) init() error {

	_, err := os.Stat(gci.dir)
	if err == nil {
		// already set up
		return nil
	}
	if os.IsNotExist(err) {
		return gci.doInit()
	}
	return err

}

func (gci GitContentAPI) doInit() error {
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

func (gci GitContentAPI) All() ([]Content, error) {
	cmd := exec.Command("ls")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	cmd.Dir = gci.dir
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	var cont []Content

	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		line := scanner.Text()
		uuid := strings.Split(line, ".")[0]
		found, c := gci.ByUUID(uuid)
		if !found {
			return nil, fmt.Errorf("failed to find content %s", uuid)
		}
		cont = append(cont, c)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	sort.Sort(sort.Reverse(Contents(cont)))
	return cont, nil
}

func NewGitContentAPI() (ContentAPI, error) {
	api := GitContentAPI{"/tmp/gitapi/"}
	err := api.init()
	if err != nil {
		return nil, err
	}
	return api, nil
}
