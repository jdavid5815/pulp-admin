/* Pulp CLI
 *
 * - Version 1.1.3 - 2021/09/21
 */
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func pulpExec(request *http.Request) ([]byte, int, error) {

	request.SetBasicAuth(apiUser, apiPass)
	result, err := apiClt.Do(request)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	defer result.Body.Close()
	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return nil, http.StatusNoContent, nil
	}
	return body, result.StatusCode, nil
}

func pulpWaitForTask(task Task) (TaskQuery, error) {

	taskQuery := TaskQuery{}
	taskQuery.State = "running"
	status := http.StatusOK
	var result []byte = nil

	req, err := http.NewRequest("GET", apiSrv+task.Task, nil)
	if err != nil {
		return taskQuery, err
	}
	for status == http.StatusOK && (taskQuery.State == "running" || taskQuery.State == "waiting") {
		result, status, err = pulpExec(req)
		if err != nil {
			return taskQuery, err
		}
		if status != http.StatusOK {
			return taskQuery, fmt.Errorf("HTTP response: %d, expected: %d", status, http.StatusOK)
		}
		decoder := json.NewDecoder(bytes.NewReader(result))
		err = decoder.Decode(&taskQuery)
		if err != nil {
			return taskQuery, err
		}
		fmt.Printf("Waiting for task %s to finish...\n", taskQuery.Pulp_href)
		time.Sleep(2 * time.Second)
	}
	if taskQuery.State != "completed" {
		if taskQuery.Error.Description != "" {
			return taskQuery, fmt.Errorf("%s", taskQuery.Error.Description)
		} else {
			return taskQuery, fmt.Errorf("TaskQuery State = %s, expected \"completed\".\n", taskQuery.State)
		}
	}
	return taskQuery, nil
}

func pulpStatus() (int, error) {

	req, err := http.NewRequest("GET", apiSrv+"/auth/login/?next=/pulp/api/v3/status/", nil)
	if err != nil {
		return http.StatusTeapot, err
	}
	_, status, err := pulpExec(req)
	if err != nil {
		return http.StatusTeapot, err
	}
	return status, nil
}

func pulpVerifyRepo(repository string) error {

	res, err := pulpRepositoryInfo(repository)
	if err != nil {
		return err
	}
	if res.Count == 0 {
		return fmt.Errorf("repository %s does not exist", repository)
	}
	return nil
}

func pulpOrphanClean() ([]ProgressReport, error) {

	req, err := http.NewRequest("DELETE", apiEnd+"/orphans/", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	result, status, err := pulpExec(req)
	if err != nil {
		return nil, err
	}
	if status != http.StatusAccepted {
		return nil, fmt.Errorf("HTTP response: %d, body: %s", status, string(result))
	}
	decoder := json.NewDecoder(bytes.NewReader(result))
	task := Task{}
	err = decoder.Decode(&task)
	if err != nil {
		return nil, err
	}
	taskResults, err := pulpWaitForTask(task)
	if err != nil {
		return nil, err
	}
	return taskResults.Progress_reports, nil
}
