/* Pulp CLI
 *
 * - Version 1.0 - 2021/07/31
 */
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func pulpSyncRepo(repo string) error {

	repoInfo, err := pulpRepositoryInfo(repo)
	if err != nil {
		return err
	}
	if repoInfo.Count == 0 {
		return fmt.Errorf("repository %s does not exist", repo)
	}
	if repoInfo.Results[0].Remote == "" {
		return fmt.Errorf("remote is not set for repo %s", repo)
	}
	content := SyncSet{
		Mirror:     true,
		Skip_types: []string{"srpm"},
		Optimize:   true,
	}
	body, err := json.Marshal(content)
	if err != nil {
		return err
	}
	data := bytes.NewReader(body)
	req, err := http.NewRequest("POST", apiSrv+repoInfo.Results[0].Pulp_href+"sync/", data)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	result, status, err := pulpExec(req)
	if err != nil {
		return err
	}
	if status != http.StatusAccepted {
		return fmt.Errorf("HTTP response: %d, body: %s", status, string(result))
	}
	decoder := json.NewDecoder(bytes.NewReader(result))
	task := Task{}
	err = decoder.Decode(&task)
	if err != nil {
		return err
	}
	_, err = pulpWaitForTask(task)
	if err != nil {
		return err
	}
	fmt.Printf("Repository %s synced with remote.\n", repo)
	return nil
}
