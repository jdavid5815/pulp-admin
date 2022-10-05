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

func pulpPublishPackage(repo string) ([]string, error) {

	// Get info on the state of the repository
	repoinfo, err := pulpRepositoryInfo(repo)
	if err != nil {
		return nil, err
	}
	if repoinfo.Count == 0 {
		return nil, fmt.Errorf("repository %s does not exist", repo)
	}
	pubinfo, err := pulpPublishAll()
	if err != nil {
		return nil, err
	}
	for _, pub := range pubinfo.Results {
		if pub.Repository_version == repoinfo.Results[0].Latest_version_href {
			return nil, fmt.Errorf("publication for repository %s already exists", repo)
		}
	}
	// Create new publication
	content := RepoSet{
		Repository_version: repoinfo.Results[0].Latest_version_href,
	}
	body, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}
	data := bytes.NewReader(body)
	req, err := http.NewRequest("POST", apiEnd+"/publications/rpm/rpm/", data)
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
	fmt.Printf("Repository publication created.\n")
	return taskResults.Created_resources, nil
}

func pulpDistributePackage(repo string, publication []string) error {

	var (
		req *http.Request
		err error
	)

	info, err := deconstructRepository(repo)
	if err != nil {
		return err
	}
	distInfo, err := pulpDistributionInfo(repo + "-" + apiEnv[0]) // Element 0 is the default
	if err != nil {
		return err
	}
	if distInfo.Count > 0 {
		// Existing distribution. Only update the default distribution.
		content := DistroSet{
			Base_path:     info.Distribution + "/" + info.Release + "/" + info.Architecture + "/" + info.Name + "/" + apiEnv[0],
			Content_guard: "",
			Name:          repo + "-" + apiEnv[0],
			Publication:   publication[0],
		}
		body, err := json.Marshal(content)
		if err != nil {
			return err
		}
		data := bytes.NewReader(body)
		req, err = http.NewRequest("PATCH", apiSrv+distInfo.Results[0].Pulp_href, data)
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
		fmt.Printf("Default repository distribution updated.\n")
	} else {
		// New distribution. Create all the distributions.
		var (
			content      DistroSet
			body, result []byte
			data         *bytes.Reader
			decoder      *json.Decoder
			status       int
			task         Task
		)

		for _, value := range apiEnv {
			content = DistroSet{
				Base_path:     info.Distribution + "/" + info.Release + "/" + info.Architecture + "/" + info.Name + "/" + value,
				Content_guard: "",
				Name:          repo + "-" + value,
				Publication:   publication[0],
			}
			body, err = json.Marshal(content)
			if err != nil {
				return err
			}
			data = bytes.NewReader(body)
			req, err = http.NewRequest("POST", apiEnd+"/distributions/rpm/rpm/", data)
			if err != nil {
				return err
			}
			req.Header.Set("Content-Type", "application/json")
			result, status, err = pulpExec(req)
			if err != nil {
				return err
			}
			if status != http.StatusAccepted {
				return fmt.Errorf("HTTP response: %d, body: %s", status, string(result))
			}
			decoder = json.NewDecoder(bytes.NewReader(result))
			task = Task{}
			err = decoder.Decode(&task)
			if err != nil {
				return err
			}
			_, err = pulpWaitForTask(task)
			if err != nil {
				return err
			}
			fmt.Printf("%s repository distribution created.\n", value)
		}
	}
	return nil
}
