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
	"path"
	"strconv"
)

func pulpSetPubVersion(repository string, distribution string, environment string, version int) error {

	var available, publication string

	publications, err := pulpPublicationList(repository)
	if err != nil {
		return err
	}
	for _, pub := range publications {
		available = path.Base(pub.Repository_version)
		i, err := strconv.Atoi(available)
		if err != nil {
			return err
		}
		if version == i {
			publication = pub.Pulp_href
			break
		}
	}
	distInfo, err := pulpDistributionInfo(distribution)
	if err != nil {
		return err
	}
	info, err := deconstructRepository(repository)
	if err != nil {
		return err
	}
	content := DistroSet{
		Base_path:     info.Distribution + "/" + info.Release + "/" + info.Architecture + "/" + info.Name + "/" + environment,
		Content_guard: "",
		Name:          distribution,
		Publication:   publication,
	}
	body, err := json.Marshal(content)
	if err != nil {
		return err
	}
	data := bytes.NewReader(body)
	req, err := http.NewRequest("PATCH", apiSrv+distInfo.Results[0].Pulp_href, data)
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
	fmt.Printf("Distribution %s set to version %d.\n", distribution, version)
	return nil
}
