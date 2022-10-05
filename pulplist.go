/* Pulp CLI
 *
 * - Version 1.1.4 - 2021/11/09
 */
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func pulpPublicationList(repo string) ([]PulpPublish, error) {

	var (
		err      error
		url      string
		pubInfo  PulpPublishResults
		repoInfo PulpRepositoryResults
		results  []PulpPublish
	)

	repoInfo, err = pulpRepositoryInfo(repo)
	if err != nil {
		return nil, err
	}
	reference := repoInfo.Results[0].Pulp_href
	url = apiEnd + "/publications/rpm/rpm/"
	for {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		body, status, err := pulpExec(req)
		if err != nil {
			return nil, err
		}
		if status != http.StatusOK {
			return nil, fmt.Errorf("HTTP response: %d", status)
		}
		err = json.Unmarshal(body, &pubInfo)
		if err != nil {
			return nil, err
		}
		for _, pub := range pubInfo.Results {
			if pub.Repository == reference {
				results = append(results, pub)
			}
		}
		if url != pubInfo.Next {
			url = pubInfo.Next
		} else {
			break
		}
	}
	return results, nil
}

func pulpDistributionList(repo string) ([]PulpDistActive, error) {

	var (
		result    PulpDistActive
		resultSet []PulpDistActive
	)

	publications, err := pulpPublicationList(repo)
	if err != nil {
		return nil, err
	}
	for _, env := range apiEnv {
		distInfo, err := pulpDistributionInfo(repo + "-" + env)
		if err != nil {
			return nil, err
		}
		for _, pub := range publications {
			if pub.Pulp_href == distInfo.Results[0].Publication {
				result.Distribution = repo + "-" + env
				result.ActivePublication = pub
				resultSet = append(resultSet, result)
			}
		}
	}
	return resultSet, nil
}
