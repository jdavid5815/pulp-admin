/* Pulp CLI
 *
 * - Version 1.1.0 - 2021/08/19
 */
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"
)

func calcSHA256(pack string) (string, error) {

	file, err := os.Open(pack)
	if err != nil {
		return "", err
	}
	defer file.Close()
	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func getFileSize(file string) (int64, error) {

	info, err := os.Stat(file)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func getAuthorization() error {

	var config Configuration

	osuser, err := user.Current()
	if err != nil {
		return err
	}
	adminpath := osuser.HomeDir + "/.pulp/admin.conf"
	adminfile, err := os.Open(adminpath)
	if err != nil {
		return err
	} else {
		defer adminfile.Close()
		decoder := json.NewDecoder(adminfile)
		config = Configuration{}
		err = decoder.Decode(&config)
		if err != nil {
			return err
		}
		apiUser = config.User
		apiPass = config.Pass
		apiSrv = config.Url
		apiEnd = apiSrv + API_ENDPOINT
		apiClt = &http.Client{Timeout: time.Second * 10}
		defer apiClt.CloseIdleConnections()
		return nil
	}
}

func setAuthorization() (string, error) {

	var config Configuration

	osuser, err := user.Current()
	if err != nil {
		return "", err
	}
	adminpath := osuser.HomeDir + "/.pulp/admin.conf"
	_, err = os.Stat(filepath.Dir(adminpath))
	if os.IsNotExist(err) {
		err = os.Mkdir(filepath.Dir(adminpath), 0700)
		if err != nil {
			return "", err
		}
	}
	config.User = apiUser
	config.Pass = apiPass
	config.Url = apiSrv
	output, err := json.MarshalIndent(config, " ", " ")
	if err != nil {
		return "", err
	} else {
		err = os.WriteFile(adminpath, output, 0600)
		if err != nil {
			return "", err
		} else {
			return adminpath, nil
		}
	}
}

func version() {
	program := filepath.Base(os.Args[0])
	fmt.Fprintf(flag.CommandLine.Output(), "%s: version %s\n", program, VERSION)
}

func usage() {

	program := filepath.Base(os.Args[0])
	fmt.Fprintf(flag.CommandLine.Output(), "Usage:\n")
	fmt.Fprintf(flag.CommandLine.Output(), "\t%s config -u user -p password url\n", program)
	fmt.Fprintf(flag.CommandLine.Output(), "\t%s add    -r repository rpm_package\n", program)
	fmt.Fprintf(flag.CommandLine.Output(), "\t%s del    -r repository rpm_package\n", program)
	fmt.Fprintf(flag.CommandLine.Output(), "\t%s del    -v version repository\n", program)
	fmt.Fprintf(flag.CommandLine.Output(), "\t%s list\n", program)
	fmt.Fprintf(flag.CommandLine.Output(), "\t%s list   -v repository\n", program)
	fmt.Fprintf(flag.CommandLine.Output(), "\t%s list   -d repository\n", program)
	fmt.Fprintf(flag.CommandLine.Output(), "\t%s set    -v version distribution\n", program)
	fmt.Fprintf(flag.CommandLine.Output(), "\t%s clean\n", program)
	fmt.Fprintf(flag.CommandLine.Output(), "\t%s sync repository\n", program)
	fmt.Fprintf(flag.CommandLine.Output(), "\t%s version\n", program)
}

func deconstructRepository(repo string) (RepoDetails, error) {

	var (
		oi RepoDetails = RepoDetails{
			Name:         "",
			Distribution: "",
			Release:      "",
			Architecture: "",
		}
		os string
	)

	regular := regexp.MustCompile(`-`)
	dashsplit := regular.Split(repo, -1)
	length := len(dashsplit)
	oi.Architecture = dashsplit[length-1]
	if oi.Architecture != "x86_64" {
		return oi, fmt.Errorf("%s is an unsupported hardware platform abbreviation", oi.Architecture)
	}
	for _, char := range dashsplit[length-2] {
		if unicode.IsDigit(char) {
			oi.Release += string(char)
		} else {
			os += string(char)
		}
	}
	switch os {
	case "co":
		oi.Distribution = "centos"
	case "fe":
		oi.Distribution = "fedora"
	case "ol":
		oi.Distribution = "oraclelinux"
	case "rh":
		oi.Distribution = "redhat"
	case "rl":
		oi.Distribution = "rockylinux"
	default:
		return oi, fmt.Errorf("%s is an unknown operating system abbreviation", os)
	}
	oi.Name = strings.Join(dashsplit[:length-2], "-")
	return oi, nil
}

func deconstructPackage(pack string) PackageDetails {

	var packinfo PackageDetails

	packinfo.Release = "1" // The release is often not set and is 1 by default.
	regular := regexp.MustCompile(`\.`)
	dotsplit := regular.Split(pack, -1)
	length := len(dotsplit)
	packinfo.Arch = dotsplit[length-2]
	remainder := strings.Join(dotsplit[:length-2], ".")
	regular = regexp.MustCompile(`-`)
	dashsplit := regular.Split(remainder, -1)
	for i, v := range dashsplit {
		r := []rune(v)
		for j := 0; j < len(r); j++ {
			if unicode.IsDigit(r[j]) {
				if packinfo.Version == "" {
					packinfo.Name = strings.Join(dashsplit[:i], "-")
					packinfo.Version = v
				} else {
					packinfo.Release = v
					goto END
				}
				break
			}
		}
	}
END:
	return packinfo
}
