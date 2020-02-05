package client

import (
	"cf-tool/client/api"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func (c *Client) Status(username string) ([]api.Submission, error) {
	resp, err := c.client.Get(fmt.Sprintf(c.Host+"/api/user.status?handle=%v", username))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(resp.Status)
	}

	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var data map[string]interface{}
	if err = decoder.Decode(&data); err != nil {
		return nil, err
	}

	if status, ok := data["status"].(string); !ok || status != "OK" {
		return nil, fmt.Errorf("Cannot get any submission")
	}

	submissions := data["result"].([]interface{})
	var status []api.Submission

	for _, _submission := range submissions {
		submission := _submission.(map[string]interface{})

		verdict := submission["verdict"].(string)
		problemsetName, ok := submission["problem"].(map[string]interface{})["problemsetName"]
		var contestID float64 = -1
		if !ok || (ok && problemsetName.(string) != "acmsguru") {
			contestID = submission["contestId"].(float64)
		}
		submissionID := submission["id"].(float64)
		lang := submission["programmingLanguage"].(string)
		timestamp := submission["creationTimeSeconds"].(float64)
		problemID := strings.ToLower(submission["problem"].(map[string]interface{})["index"].(string))

		status = append(status, *api.NewSubmission(contestID, submissionID, problemID, verdict, lang, timestamp))
	}

	return status, nil
}

func (c *Client) SaveStatus(username, path string) error {
	status, err := c.Status(username)
	if err != nil {
		return err
	}
	path = filepath.Join(path, "status")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.Mkdir(path, 0755); err != nil {
			return err
		}
	}

	path = filepath.Join(path, fmt.Sprintf("%s.json", username))

	b, err := json.Marshal(status)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, b, 0644)
}

func (c *Client) SaveUserStatuses(handlesPath, savePath string) (err error) {
	b, err := ioutil.ReadFile(handlesPath)
	if err != nil {
		return
	}

	var handles *[]Handle = &[]Handle{}
	err = json.Unmarshal(b, handles)
	if err != nil {
		return
	}
	if handles == nil {
		return fmt.Errorf("handles are empty")
	}

	threadNumber := 32

	ch := make(chan Handle, threadNumber)
	again := make(chan Handle, threadNumber)

	wg := sync.WaitGroup{}
	wg.Add(threadNumber + 1)
	mu := sync.Mutex{}

	count := 0
	total := len(*handles)

	go func() {
		for {
			s, ok := <-again
			if !ok {
				wg.Done()
				return
			}
			ch <- s
		}
	}()

	for gid := 0; gid < threadNumber; gid++ {
		go func() {
			for {
				handle, ok := <-ch
				if !ok {
					wg.Done()
					return
				}
				err := c.SaveStatus(handle.Handle, savePath)
				if err == nil {
					mu.Lock()
					count++
					color.Green(fmt.Sprintf(`%v/%v Saved %v`, count, total, handle.Handle))
					mu.Unlock()
				} else {
					color.Red("%v", err.Error())
					err = fmt.Errorf("Too many requests")

					if err.Error() == "Too many requests" {
						mu.Lock()
						count++
						const WAIT int = 5
						color.Red(fmt.Sprintf(`%v/%v Error in %v: %v. Waiting for %v seconds to continue.`,
							count, total, handle.Handle, err.Error(), WAIT))
						mu.Unlock()
						time.Sleep(time.Duration(WAIT) * time.Second)
						mu.Lock()
						count--
						mu.Unlock()
						again <- handle
					}
				}
			}
		}()
	}

	// We don't want to download submissions of unrated users
	var filteredHandles []Handle

	for _, handle := range *handles {
		if handle.Color != "black" {
			filteredHandles = append(filteredHandles, handle)
		}
	}

	for _, handle := range filteredHandles {
		ch <- handle
	}

	close(ch)
	close(again)
	wg.Wait()
	return

}
