/* Pulp CLI
 *
 * - Version 1.1.0 - 2021/08/19
 */
package main

import "sync"

type Thread struct {
	mutex sync.Mutex
	count int
	err   error
}

type RepoPublicationList struct {
	Name string `json:"name"`
}

type RepoDetails struct {
	Name         string `json:"name"`
	Distribution string `json:"distribution"`
	Release      string `json:"release"`
	Architecture string `json:"architecture"`
}

type PackageDetails struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Release string `json:"release"`
	Arch    string `json:"arch"`
}

type Configuration struct {
	User string `json:"user"`
	Pass string `json:"pass"`
	Url  string `json:"url"`
}

type Artifact struct {
	Artifact string `json:"artifact"`
	Rel_path string `json:"relative_path"`
}

type UploadStart struct {
	Size int64 `json:"size"`
}

type UploadFinish struct {
	Sha256 string `json:"sha256"`
}

type AddContentUnits struct {
	Add_content_units []string `json:"add_content_units"`
}

type RemoveContentUnits struct {
	Remove_content_units []string `json:"remove_content_units"`
}

type RepoSet struct {
	Repository_version string `json:"repository_version"`
}

type DistroSet struct {
	Base_path     string `json:"base_path"`
	Content_guard string `json:"content_guard"`
	//Pulp_labels   PulpLabels `json:"pulp_labels"`
	Name        string `json:"name"`
	Publication string `json:"publication"`
}

type SyncSet struct {
	Mirror     bool     `json:"mirror"`
	Skip_types []string `json:"skip_types"`
	Optimize   bool     `json:"optimize"`
}

type Task struct {
	Task string `json:"task"`
}

type ProgressReport struct {
	Message string `json:"message"`
	Code    string `json:"code"`
	State   string `json:"state"`
	Total   int    `json:"total"`
	Done    int    `json:"done"`
	Suffix  string `json:"suffix"`
}

type TaskQuery struct {
	Pulp_href                 string           `json:"pulp_href"`
	Pulp_created              string           `json:"pulp_created"`
	State                     string           `json:"state"`
	Name                      string           `json:"name"`
	Logging_cid               string           `json:"logging_cid"`
	Started_at                string           `json:"started_at"`
	Finished_at               string           `json:"finished_at"`
	Error                     PulpError        `json:"error"`
	Worker                    string           `json:"worker"`
	Parent_task               string           `json:"parent_task"`
	Child_tasks               []string         `json:"child_tasks"`
	Task_group                string           `json:"task_group"`
	Progress_reports          []ProgressReport `json:"progress_reports"`
	Created_resources         []string         `json:"created_resources"`
	Reserved_resources_record []string         `json:"reserved_resources_record"`
}

type PulpDistActive struct {
	Distribution      string
	ActivePublication PulpPublish
}

type PulpError struct {
	Traceback   string `json:"traceback"`
	Description string `json:"description"`
}

type PulpLabels struct {
}

type PulpCreate struct {
	Pulp_href    string `json:"pulp_href"`
	Pulp_created string `json:"pulp_created"`
	File         string `json:"file"`
	Size         int    `json:"size"`
	Md5          string `json:"md5"`
	Sha1         string `json:"sha1"`
	Sha224       string `json:"sha224"`
	Sha256       string `json:"sha256"`
	Sha384       string `json:"sha384"`
	Sha512       string `json:"sha512"`
}

type PulpPublish struct {
	Pulp_href              string `json:"pulp_href"`
	Pulp_created           string `json:"pulp_created"`
	Repository_version     string `json:"repository_version"`
	Repository             string `json:"repository"`
	Metadata_checksum_type string `json:"metadata_checksum_type"`
	Package_checksum_type  string `json:"package_checksum_type"`
	Gpgcheck               int    `json:"gpgcheck"`
	Repo_gpgcheck          int    `json:"repo_gpgcheck"`
	Sqlite_metadata        bool   `json:"sqlite_metadata"`
}

type PulpRepository struct {
	Pulp_href           string     `json:"pulp_href"`
	Pulp_created        string     `json:"pulp_created"`
	Versions_href       string     `json:"versions_href"`
	Pulp_labels         PulpLabels `json:"pulp_labels"`
	Latest_version_href string     `json:"latest_version_href"`
	Name                string     `json:"name"`
	Description         string     `json:"description"`
	Remote              string     `json:"remote"`
}

type PulpDistribution struct {
	Pulp_href     string     `json:"pulp_href"`
	Pulp_created  string     `json:"pulp_created"`
	Base_path     string     `json:"base_path"`
	Base_url      string     `json:"base_url"`
	Content_guard string     `json:"content_guard"`
	Pulp_labels   PulpLabels `json:"pulp_labels"`
	Name          string     `json:"name"`
	Repository    string     `json:"repository"`
	Publication   string     `json:"publication"`
}

type PulpContent struct {
	Pulp_href     string `json:"pulp_href"`
	Pulp_created  string `json:"pulp_created"`
	Md5           string `json:"md5"`
	Sha1          string `json:"sha1"`
	Sha224        string `json:"sha224"`
	Sha256        string `json:"sha256"`
	Sha384        string `json:"sha384"`
	Sha512        string `json:"sha512"`
	Artifact      string `json:"artifact"`
	Name          string `json:"name"`
	Epoch         string `json:"epoch"`
	Version       string `json:"version"`
	Release       string `json:"release"`
	Arch          string `json:"arch"`
	PkgId         string `json:"pkgId"`
	Checksum_type string `json:"checksum_type"`
	Summary       string `json:"summary"`
	Description   string `json:"description"`
	Url           string `json:"url"`
	//Changelogs       []string        `json:"changelogs"`
	//Files            [][]string      `json:"files"`
	//Requires         [][]interface{} `json:"requires"`
	//Provides         [][]interface{} `json:"provides"`
	//Conflicts        [][]interface{} `json:"conflicts"`
	//Obsoletes        [][]interface{} `json:"obsoletes"`
	//Suggests         [][]interface{} `json:"suggests"`
	//Enhances         []string        `json:"enhances"`
	//Recommends       []string        `json:"recommends"`
	//Supplements      []string        `json:"supplements"`
	Location_base    string `json:"location_base"`
	Location_href    string `json:"location_href"`
	Rpm_buildhost    string `json:"rpm_buildhost"`
	Rpm_group        string `json:"rpm_group"`
	Rpm_license      string `json:"rpm_license"`
	Rpm_packager     string `json:"rpm_packager"`
	Rpm_sourcerpm    string `json:"rpm_sourcerpm"`
	Rpm_vendor       string `json:"rpm_vendor"`
	Rpm_header_start int    `json:"rpm_header_start"`
	Rpm_header_end   int    `json:"rpm_header_end"`
	Is_modular       bool   `json:"is_modular"`
	Size_archive     int    `json:"size_archive"`
	Size_installed   int    `json:"size_installed"`
	Size_package     int    `json:"size_package"`
	Time_build       int    `json:"time_build"`
	Time_file        int    `json:"time_file"`
}

type PulpUploadResults struct {
	Pulp_href    string `json:"pulp_href"`
	Pulp_created string `json:"pulp_created"`
	Size         int64  `json:"size"`
	Completed    string `json:"completed"`
}

type PulpCreateResults struct {
	Count    int          `json:"count"`
	Next     string       `json:"next"`
	Previous string       `json:"previous"`
	Results  []PulpCreate `json:"results"`
}

type PulpPublishResults struct {
	Count    int           `json:"count"`
	Next     string        `json:"next"`
	Previous string        `json:"previous"`
	Results  []PulpPublish `json:"results"`
}

type PulpRepositoryResults struct {
	Count    int              `json:"count"`
	Next     string           `json:"next"`
	Previous string           `json:"previous"`
	Results  []PulpRepository `json:"results"`
}

type PulpDistributionResults struct {
	Count    int                `json:"count"`
	Next     string             `json:"next"`
	Previous string             `json:"previous"`
	Results  []PulpDistribution `json:"results"`
}

type PulpContentResults struct {
	Count    int           `json:"count"`
	Next     string        `json:"next"`
	Previous string        `json:"previous"`
	Results  []PulpContent `json:"results"`
}
