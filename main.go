/* Pulp CLI
 *
 * - version 1.1.4 - 2021/11/09
 *     Fixed bug that prevented all versions of a repo from being show, because
 *     the code did not iterate through all publications.
 * - Version 1.1.3 - 2021/09/21
 *     Fixed bug from version 1.1.1 and 1.1.2 and improved error logging
 *     for failed 'wait' tasks.
 * - Version 1.1.2 - 2021/09/17
 *     Added debugging statement to investigated unfinished tasks bail-out.
 * - Version 1.1.1 - 2021/08/27
 *     Fixed possible memoryleak for open http connections in upload thread.
 * - Version 1.1.0 - 2021/08/19
 *     Added Chunked Uploads to be able to upload large packages
 * - Version 1.0.1 - 2021/08/17
 *     Improved error reporting on failed package uploads.
 * - Version 1.0.0 - 2021/07/31
 *     Initial fully working CLI
 */
package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const (
	VERSION        string        = "1.1.4"
	API_ENDPOINT   string        = "/pulp/api/v3"
	CLIENT_TIMEOUT time.Duration = 300
	CHUNKSIZE      int64         = 8388608 // Pulp3 Nginx limits uploads to 10Mb, so chunk sizes must be lower than that.
	MAX_THREADS    int           = 10
)

// API user, password, server, endpoint, environments and client connection
var (
	apiUser, apiPass, apiSrv, apiEnd string
	apiEnv                           = [...]string{"dev", "uat", "oat", "prd"} // The first element is the default
	apiClt                           *http.Client
)

func main() {

	var (
		err    error
		status int
	)

	configCmd := flag.NewFlagSet("config", flag.ExitOnError)
	configUsr := configCmd.String("u", "", "User performing administration on Pulp.")
	configPss := configCmd.String("p", "", "Password of user performing administration.")

	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	addRep := addCmd.String("r", "", "The repository to work upon.")

	delCmd := flag.NewFlagSet("del", flag.ExitOnError)
	delRep := delCmd.String("r", "", "The repository to work upon.")
	delVer := delCmd.String("v", "", "The publication version to remove.")

	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	listDis := listCmd.Bool("d", false, "Display the active publication for each distribution of the given repository.")
	listPub := listCmd.Bool("v", false, "Display all publications (versions) for the given repository.")

	setCmd := flag.NewFlagSet("set", flag.ExitOnError)
	setVer := setCmd.Int("v", 0, "Set the version of the publication you want to use for the given distribution.")

	cleanCmd := flag.NewFlagSet("clean", flag.ExitOnError)

	syncCmd := flag.NewFlagSet("sync", flag.ExitOnError)

	versionCmd := flag.NewFlagSet("version", flag.ExitOnError)

	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "config":
		configCmd.Parse(os.Args[2:])
		if len(*configUsr) == 0 || len(*configPss) == 0 {
			fmt.Fprintf(flag.CommandLine.Output(), "ERROR: both the -u and -p options are required for the 'config' subcommand!\n")
			os.Exit(1)
		}
		if len(configCmd.Args()) != 1 {
			fmt.Fprintf(flag.CommandLine.Output(), "ERROR: the 'config' subcommand requires exactly one url as argument!\n")
			usage()
			os.Exit(1)
		}
		temp := configCmd.Args()
		pulpUrl, err := url.ParseRequestURI(temp[0])
		if err != nil {
			fmt.Fprintf(flag.CommandLine.Output(), "ERROR: %s\n", err.Error())
			os.Exit(1)
		}
		apiUser = *configUsr
		apiPass = *configPss
		apiSrv = pulpUrl.Scheme + "://" + pulpUrl.Hostname() + ":" + pulpUrl.Port()
		apiEnd = apiSrv + API_ENDPOINT
		apiClt = &http.Client{Timeout: time.Second * CLIENT_TIMEOUT}
		defer apiClt.CloseIdleConnections()
		status, err = pulpStatus()
		if err != nil {
			fmt.Fprintf(flag.CommandLine.Output(), "ERROR: %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Fprintf(flag.CommandLine.Output(), "Pulp API Status Check: %d\n", status)
		if status == http.StatusOK {
			adminpath, err := setAuthorization()
			if err != nil {
				fmt.Printf("ERROR %s\n", err.Error())
				os.Exit(1)
			} else {
				fmt.Printf("Credentials saved to '%s'\n", adminpath)
			}
		} else {
			os.Exit(1)
		}
	case "add":
		addCmd.Parse(os.Args[2:])
		if len(*addRep) == 0 {
			fmt.Fprintf(flag.CommandLine.Output(), "ERROR: the -r option is required for the 'add' subcommand!\n")
			os.Exit(1)
		}
		if len(addCmd.Args()) != 1 {
			fmt.Fprintf(flag.CommandLine.Output(), "ERROR: the 'add' subcommand requires exactly one rpm package as argument!\n")
			usage()
			os.Exit(1)
		}
		temp := addCmd.Args()
		pack := strings.TrimSpace(temp[0])
		if filepath.Ext(pack) != ".rpm" {
			fmt.Fprintf(flag.CommandLine.Output(), "ERROR: the 'add' subcommand requires exactly one rpm package as argument!\n")
			os.Exit(1)
		}
		err = getAuthorization()
		if err != nil {
			fmt.Fprintf(flag.CommandLine.Output(), "ERROR: %s\n", err.Error())
			os.Exit(1)
		}
		status, err = pulpStatus()
		if err != nil {
			fmt.Fprintf(flag.CommandLine.Output(), "ERROR: %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Fprintf(flag.CommandLine.Output(), "Pulp API Status Check: %d\n", status)
		if status != http.StatusOK {
			fmt.Fprintf(flag.CommandLine.Output(), "ERROR: try running 'config' again!\n")
			os.Exit(1)
		}
		err = pulpVerifyRepo(*addRep)
		if err != nil {
			fmt.Printf("ERROR %s\n", err.Error())
			os.Exit(1)
		}
		err = pulpAddPackage(*addRep, pack)
		if err != nil {
			fmt.Printf("ERROR %s\n", err.Error())
			os.Exit(1)
		}
		pub, err := pulpPublishPackage(*addRep)
		if err != nil {
			fmt.Printf("ERROR %s\n", err.Error())
			os.Exit(1)
		}
		err = pulpDistributePackage(*addRep, pub)
		if err != nil {
			fmt.Printf("ERROR %s\n", err.Error())
			os.Exit(1)
		}
	case "del":
		delCmd.Parse(os.Args[2:])
		if len(*delRep) == 0 && len(*delVer) == 0 {
			fmt.Fprintf(flag.CommandLine.Output(), "ERROR: either the -r option or the -v option is required for the 'del' subcommand!\n")
			usage()
			os.Exit(1)
		}
		if len(*delRep) != 0 && len(*delVer) != 0 {
			fmt.Fprintf(flag.CommandLine.Output(), "ERROR: cannot use option -r and -v at the same time!\n")
			usage()
			os.Exit(1)
		}
		if len(delCmd.Args()) != 1 {
			fmt.Fprintf(flag.CommandLine.Output(), "ERROR: the 'del' subcommand requires exactly one argument!\n")
			usage()
			os.Exit(1)
		}
		err = getAuthorization()
		if err != nil {
			fmt.Fprintf(flag.CommandLine.Output(), "ERROR: %s\n", err.Error())
			os.Exit(1)
		}
		status, err = pulpStatus()
		if err != nil {
			fmt.Fprintf(flag.CommandLine.Output(), "ERROR: %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Fprintf(flag.CommandLine.Output(), "Pulp API Status Check: %d\n", status)
		if status != http.StatusOK {
			fmt.Fprintf(flag.CommandLine.Output(), "ERROR: try running 'config' again!\n")
			usage()
			os.Exit(1)
		}
		temp := delCmd.Args()
		argu := strings.TrimSpace(temp[0])
		if len(*delRep) != 0 {
			if filepath.Ext(argu) != ".rpm" {
				fmt.Fprintf(flag.CommandLine.Output(), "ERROR: the -r option requires a rpm package as argument!\n")
				usage()
				os.Exit(1)
			}
			err = pulpVerifyRepo(*delRep)
			if err != nil {
				fmt.Printf("ERROR %s\n", err.Error())
				os.Exit(1)
			}
			err = pulpDelPackage(*delRep, argu)
			if err != nil {
				fmt.Printf("ERROR %s\n", err.Error())
				os.Exit(1)
			}
			pub, err := pulpPublishPackage(*delRep)
			if err != nil {
				fmt.Printf("ERROR %s\n", err.Error())
				os.Exit(1)
			}
			err = pulpDistributePackage(*delRep, pub)
			if err != nil {
				fmt.Printf("ERROR %s\n", err.Error())
				os.Exit(1)
			}
		}
		if len(*delVer) != 0 {
			err = pulpVerifyRepo(argu)
			if err != nil {
				fmt.Printf("ERROR %s\n", err.Error())
				usage()
				os.Exit(1)
			}
			pubList, err := pulpPublicationList(argu)
			if err != nil {
				fmt.Printf("ERROR %s\n", err.Error())
				os.Exit(1)
			}
			var version string
			for _, pub := range pubList {
				version = path.Base(pub.Repository_version)
				if *delVer == version {
					disList, err := pulpDistributionList(argu)
					if err != nil {
						fmt.Printf("ERROR %s\n", err.Error())
						os.Exit(1)
					}
					for _, dis := range disList {
						version = path.Base(dis.ActivePublication.Repository_version)
						if *delVer == version {
							fmt.Printf("ERROR publication version %s is currently still being used by distribution %s!\n", *delVer, dis.Distribution)
							return
						}
					}
					err = pulpDelPublication(pub)
					if err != nil {
						fmt.Printf("ERROR %s\n", err.Error())
						os.Exit(1)
					}
					fmt.Printf("Publication version %s from repository %s was successfully deleted.\n", *delVer, argu)
					return
				}
			}
			fmt.Printf("ERROR publication version %s from repository %s could not be found!\n", *delVer, argu)
		}
	case "list":
		listCmd.Parse(os.Args[2:])
		if !*listPub && !*listDis {
			if len(listCmd.Args()) > 0 {
				fmt.Fprintf(flag.CommandLine.Output(), "ERROR: 'list' subcommand without options, does not require an argument!\n")
				usage()
				os.Exit(1)
			}
		}
		if *listPub || *listDis {
			if len(listCmd.Args()) != 1 {
				fmt.Fprintf(flag.CommandLine.Output(), "ERROR: 'list' subcommand with -d or -v options requires a repository as argument!\n")
				usage()
				os.Exit(1)
			}
		}
		err = getAuthorization()
		if err != nil {
			fmt.Fprintf(flag.CommandLine.Output(), "ERROR: %s\n", err.Error())
			os.Exit(1)
		}
		status, err = pulpStatus()
		if err != nil {
			fmt.Printf("ERROR %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Fprintf(flag.CommandLine.Output(), "Pulp API Status Check: %d\n", status)
		if status != http.StatusOK {
			fmt.Fprintf(flag.CommandLine.Output(), "ERROR: try running 'config' again!\n")
			os.Exit(1)
		}
		if *listPub {
			temp := listCmd.Args()
			repo := strings.TrimSpace(temp[0])
			err = pulpVerifyRepo(repo)
			if err != nil {
				fmt.Printf("ERROR %s\n", err.Error())
				os.Exit(1)
			}
			res, err := pulpPublicationList(repo)
			if err != nil {
				fmt.Printf("ERROR %s\n", err.Error())
				os.Exit(1)
			}
			for _, pub := range res {
				version := path.Base(pub.Repository_version)
				fmt.Printf("%s\t%s\t%s\n", repo, version, pub.Pulp_created)
			}
		} else if *listDis {
			temp := listCmd.Args()
			repo := strings.TrimSpace(temp[0])
			err = pulpVerifyRepo(repo)
			if err != nil {
				fmt.Printf("ERROR %s\n", err.Error())
				os.Exit(1)
			}
			res, err := pulpDistributionList(repo)
			if err != nil {
				fmt.Printf("ERROR %s\n", err.Error())
				os.Exit(1)
			}
			for _, r := range res {
				version := path.Base(r.ActivePublication.Repository_version)
				fmt.Printf("%s\t%s\t%s\n", r.Distribution, version, r.ActivePublication.Pulp_created)
			}
		} else {
			res, err := pulpRepositoryAll()
			if err != nil {
				fmt.Printf("ERROR %s\n", err.Error())
				os.Exit(1)
			}
			fmt.Println()
			for i := 0; i < res.Count; i++ {
				fmt.Printf("%s\n", res.Results[i].Name)
			}
		}
	case "set":
		setCmd.Parse(os.Args[2:])
		if *setVer == 0 {
			fmt.Fprintf(flag.CommandLine.Output(), "ERROR: the -v option is required for the 'set' subcommand!\n")
			os.Exit(1)
		}
		if len(setCmd.Args()) != 1 {
			fmt.Fprintf(flag.CommandLine.Output(), "ERROR: the 'set' subcommand requires exactly one distribution as argument!\n")
			usage()
			os.Exit(1)
		}
		err = getAuthorization()
		if err != nil {
			fmt.Fprintf(flag.CommandLine.Output(), "ERROR: %s\n", err.Error())
			os.Exit(1)
		}
		status, err = pulpStatus()
		if err != nil {
			fmt.Fprintf(flag.CommandLine.Output(), "ERROR: %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Fprintf(flag.CommandLine.Output(), "Pulp API Status Check: %d\n", status)
		if status != http.StatusOK {
			fmt.Fprintf(flag.CommandLine.Output(), "ERROR: try running 'config' again!\n")
			os.Exit(1)
		}
		temp := setCmd.Args()
		dist := strings.TrimSpace(temp[0])
		dashsplit := strings.Split(dist, "-")
		repo := strings.Join(dashsplit[:len(dashsplit)-1], "-")
		env := dashsplit[len(dashsplit)-1]
		err = pulpVerifyRepo(repo)
		if err != nil {
			fmt.Printf("ERROR %s\n", err.Error())
			os.Exit(1)
		}
		err = pulpSetPubVersion(repo, dist, env, *setVer)
		if err != nil {
			fmt.Printf("ERROR %s\n", err.Error())
			os.Exit(1)
		}
	case "clean":
		cleanCmd.Parse(os.Args[2:])
		if len(os.Args) > 2 {
			fmt.Fprintf(flag.CommandLine.Output(), "ERROR: the 'clean' subcommand requires no options or additional arguments!\n")
			usage()
			os.Exit(1)
		}
		err = getAuthorization()
		if err != nil {
			fmt.Fprintf(flag.CommandLine.Output(), "ERROR: %s\n", err.Error())
			os.Exit(1)
		}
		status, err = pulpStatus()
		if err != nil {
			fmt.Fprintf(flag.CommandLine.Output(), "ERROR: %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Fprintf(flag.CommandLine.Output(), "Pulp API Status Check: %d\n", status)
		if status != http.StatusOK {
			fmt.Fprintf(flag.CommandLine.Output(), "ERROR: try running 'config' again!\n")
			os.Exit(1)
		}
		progress, err := pulpOrphanClean()
		if err != nil {
			fmt.Fprintf(flag.CommandLine.Output(), "ERROR: %s\n", err.Error())
			os.Exit(1)
		}
		for _, p := range progress {
			fmt.Printf("%s: total %d done %d\n", p.Message, p.Total, p.Done)
		}
	case "sync":
		syncCmd.Parse(os.Args[2:])
		if len(syncCmd.Args()) != 1 {
			fmt.Fprintf(flag.CommandLine.Output(), "ERROR: the 'sync' subcommand requires exactly one repository as argument!\n")
			usage()
			os.Exit(1)
		}
		temp := syncCmd.Args()
		repo := strings.TrimSpace(temp[0])
		err = getAuthorization()
		if err != nil {
			fmt.Fprintf(flag.CommandLine.Output(), "ERROR: %s\n", err.Error())
			os.Exit(1)
		}
		status, err = pulpStatus()
		if err != nil {
			fmt.Fprintf(flag.CommandLine.Output(), "ERROR: %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Fprintf(flag.CommandLine.Output(), "Pulp API Status Check: %d\n", status)
		if status != http.StatusOK {
			fmt.Fprintf(flag.CommandLine.Output(), "ERROR: try running 'config' again!\n")
			os.Exit(1)
		}
		err = pulpSyncRepo(repo)
		if err != nil {
			fmt.Printf("ERROR %s\n", err.Error())
			os.Exit(1)
		}
		pub, err := pulpPublishPackage(repo)
		if err != nil {
			testerr := fmt.Sprintf("publication for repository %s already exists", repo)
			if err.Error() == testerr {
				fmt.Printf("Repository was already in sync.\n")
				return
			} else {
				fmt.Printf("ERROR %s\n", err.Error())
			}
			os.Exit(1)
		}
		err = pulpDistributePackage(repo, pub)
		if err != nil {
			fmt.Printf("ERROR %s\n", err.Error())
			os.Exit(1)
		}
	case "version":
		versionCmd.Parse(os.Args[2:])
		if len(os.Args) > 2 {
			fmt.Fprintf(flag.CommandLine.Output(), "ERROR: the 'version' subcommand requires no options or additional arguments!\n")
			usage()
			os.Exit(1)
		}
		version()
	default:
		usage()
	}

}
