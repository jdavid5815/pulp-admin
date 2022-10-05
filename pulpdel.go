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

func pulpDelPackage(repo, pack string) error {

	var (
		pcr = PulpContentResults{
			Count:    0,
			Next:     "",
			Previous: "",
			Results:  []PulpContent{},
		}
		remove []string = make([]string, 1)
	)

	cinfo, err := pulpContentInfo(deconstructPackage(pack))
	if err != nil {
		return err
	}
	if cinfo.Count == 0 {
		return fmt.Errorf("contents of %s not found", pack)
	}
	rinfo, err := pulpRepositoryInfo(repo)
	if err != nil {
		return err
	}
	if rinfo.Count == 0 {
		return fmt.Errorf("repository %s not found", repo)
	}
	// Check if package is actually in repo
	requestString := apiEnd + "/content/rpm/packages/"
	requestString += "?repository_version=" + rinfo.Results[0].Latest_version_href
	requestString += "&pkgId=" + cinfo.Results[0].PkgId
	req, err := http.NewRequest("GET", requestString, nil)
	if err != nil {
		return err
	}
	body, status, err := pulpExec(req)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("HTTP response: %d", status)
	}
	err = json.Unmarshal(body, &pcr)
	if err != nil {
		return err
	}
	if pcr.Count == 0 {
		return fmt.Errorf("content of package %s is not linked with repo %s", pack, repo)
	}
	// Remove content
	remove[0] = cinfo.Results[0].Pulp_href
	content := RemoveContentUnits{
		Remove_content_units: remove,
	}
	body, err = json.Marshal(content)
	if err != nil {
		return err
	}
	data := bytes.NewReader(body)
	requestString = apiSrv + rinfo.Results[0].Pulp_href
	requestString += "modify/"
	req, err = http.NewRequest("POST", requestString, data)
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
	fmt.Printf("Content removed from repository.\n")
	return nil
}

func pulpDelPublication(pub PulpPublish) error {

	req, err := http.NewRequest("DELETE", apiSrv+pub.Pulp_href, nil)
	if err != nil {
		return err
	}
	_, status, err := pulpExec(req)
	if err != nil {
		return err
	}
	if status != http.StatusNoContent {
		return fmt.Errorf("HTTP response: %d", status)
	}
	return nil
}
