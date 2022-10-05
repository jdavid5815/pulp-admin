/* Pulp CLI
 *
 * - Version 1.1.0 - 2021/08/19
 */
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func pulpRepositoryAll() (PulpRepositoryResults, error) {

	var r = PulpRepositoryResults{
		Count:    0,
		Next:     "",
		Previous: "",
		Results:  []PulpRepository{},
	}

	req, err := http.NewRequest("GET", apiEnd+"/repositories/rpm/rpm/", nil)
	if err != nil {
		return r, err
	}
	body, status, err := pulpExec(req)
	if err != nil {
		return r, err
	}
	if status != http.StatusOK {
		return r, fmt.Errorf("HTTP response: %d", status)
	}
	err = json.Unmarshal(body, &r)
	if err != nil {
		return r, err
	}
	return r, nil
}

func pulpRepositoryInfo(repo string) (PulpRepositoryResults, error) {

	var r = PulpRepositoryResults{
		Count:    0,
		Next:     "",
		Previous: "",
		Results:  []PulpRepository{},
	}

	req, err := http.NewRequest("GET", apiEnd+"/repositories/rpm/rpm/?name="+repo, nil)
	if err != nil {
		return r, err
	}
	body, status, err := pulpExec(req)
	if err != nil {
		return r, err
	}
	if status != http.StatusOK {
		return r, fmt.Errorf("HTTP response: %d", status)
	}
	err = json.Unmarshal(body, &r)
	if err != nil {
		return r, err
	}
	return r, nil
}

func pulpPublishAll() (PulpPublishResults, error) {

	var r = PulpPublishResults{
		Count:    0,
		Next:     "",
		Previous: "",
		Results:  []PulpPublish{},
	}

	req, err := http.NewRequest("GET", apiEnd+"/publications/rpm/rpm/", nil)
	if err != nil {
		return r, err
	}
	body, status, err := pulpExec(req)
	if err != nil {
		return r, err
	}
	if status != http.StatusOK && status != http.StatusNotFound {
		return r, fmt.Errorf("HTTP response: %d", status)
	}
	err = json.Unmarshal(body, &r)
	if err != nil {
		return r, err
	}
	return r, nil
}

func pulpDistributionInfo(distribution string) (PulpDistributionResults, error) {

	var r = PulpDistributionResults{
		Count:    0,
		Next:     "",
		Previous: "",
		Results:  []PulpDistribution{},
	}

	req, err := http.NewRequest("GET", apiEnd+"/distributions/rpm/rpm/?name="+distribution, nil)
	if err != nil {
		return r, err
	}
	body, status, err := pulpExec(req)
	if err != nil {
		return r, err
	}
	if status != http.StatusOK {
		return r, fmt.Errorf("HTTP response: %d", status)
	}
	err = json.Unmarshal(body, &r)
	if err != nil {
		return r, err
	}
	return r, nil
}

func pulpPackageInfo(pack string) (PulpCreateResults, error) {

	var r = PulpCreateResults{
		Count:    0,
		Next:     "",
		Previous: "",
		Results:  []PulpCreate{},
	}

	sha256, err := calcSHA256(pack)
	if err != nil {
		return r, err
	}
	req, err := http.NewRequest("GET", apiEnd+"/artifacts/?sha256="+sha256, nil)
	if err != nil {
		return r, err
	}
	body, status, err := pulpExec(req)
	if err != nil {
		return r, err
	}
	if status != http.StatusOK {
		return r, fmt.Errorf("HTTP response: %d", status)
	}
	err = json.Unmarshal(body, &r)
	if err != nil {
		return r, err
	}
	return r, nil
}

func pulpContentInfo(pack PackageDetails) (PulpContentResults, error) {

	var r = PulpContentResults{
		Count:    0,
		Next:     "",
		Previous: "",
		Results:  []PulpContent{},
	}

	requestString := apiEnd + "/content/rpm/packages/"
	requestString += "?name=" + pack.Name
	requestString += "&version=" + pack.Version
	requestString += "&release=" + pack.Release
	requestString += "&arch=" + pack.Arch
	req, err := http.NewRequest("GET", requestString, nil)
	if err != nil {
		return r, err
	}
	body, status, err := pulpExec(req)
	if err != nil {
		return r, err
	}
	if status != http.StatusOK {
		return r, fmt.Errorf("HTTP response: %d, expected: %d", status, http.StatusOK)
	}
	err = json.Unmarshal(body, &r)
	if err != nil {
		return r, err
	}
	return r, nil
}

func pulpArtifactInfo(artifact_href string) (PulpCreate, error) {

	var (
		pc = PulpCreate{
			Pulp_href:    "",
			Pulp_created: "",
			File:         "",
			Size:         0,
			Md5:          "",
			Sha1:         "",
			Sha224:       "",
			Sha256:       "",
			Sha384:       "",
			Sha512:       "",
		}
	)

	req, err := http.NewRequest("GET", apiSrv+artifact_href, nil)
	if err != nil {
		return pc, err
	}
	result, status, err := pulpExec(req)
	if err != nil {
		return pc, err
	}
	if status != http.StatusOK {
		return pc, fmt.Errorf("HTTP response: %d, expected: %d", status, http.StatusOK)
	}
	err = json.Unmarshal(result, &pc)
	if err != nil {
		return pc, err
	}
	return pc, nil
}

/*
func pulpRemoteInfo(repo string) (PulpRepositoryResults, error) {

	var r = PulpRepositoryResults{
		Count:    0,
		Next:     "",
		Previous: "",
		Results:  []PulpRepository{},
	}

	req, err := http.NewRequest("GET", apiEnd+"/remotes/rpm/rpm/?name="+repo, nil)
	if err != nil {
		return r, err
	}
	body, status, err := pulpExec(req)
	if err != nil {
		return r, err
	}
	if status == http.StatusNotFound {
		return r, fmt.Errorf("remote does not exist. Are you sure this is a synched repository?")
	}
	if status != http.StatusOK {
		return r, fmt.Errorf("HTTP response: %d", status)
	}
	err = json.Unmarshal(body, &r)
	if err != nil {
		return r, err
	}
	return r, nil
}
*/
