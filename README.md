
# pulp-admin
# [![Language Badge](https://img.shields.io/badge/Language-Go-blue.svg)](https://go.dev) [![License: GPL v3](https://img.shields.io/github/license/jdavid5815/pulp-admin)](https://www.gnu.org/licenses/gpl-3.0) [![Issues](https://img.shields.io/github/issues/jdavid5815/pulp-admin)](https://github.com/jdavid5815/pulp-admin/issues) [![GoDoc](https://godoc.org/github.com/jdavid5815/pulp-admin?status.svg)](https://godoc.org/github.com/jdavid5815/pulp-admin)

When Redhat introduced Pulp version 3, the familiar 'pulp-admin' tool was no longer included. As sticking to Pulpv2 was not an option, I looked for an alternative. There was another project writing a 'pulp-admin' version in Python, but the first time I tried it, I immediately landed in Python dependency hell. Couldn't get it to work properly, so I decided to write my own version in Go.

Caveat: I only wrote what I needed, so there is only support for RPM repositories and the tool assumes you have 4 environments: 'dev', 'uat', 'oat' and 'prd'. With a bit of work, this can be more generalized if needed (feel free to do so :-)

```
Usage:
	pulp-admin config -u user -p password url
	pulp-admin add    -r repository rpm_package
	pulp-admin del    -r repository rpm_package
	pulp-admin del    -v version repository
	pulp-admin list
	pulp-admin list   -v repository
	pulp-admin list   -d repository
	pulp-admin set    -v version distribution
	pulp-admin clean
	pulp-admin sync repository
	pulp-admin version
```

*config* sets up the necessary permissions to connect to Pulp. Information gets stored in ~/pulp/.admin.conf.

*add* allows you to add an RPM package to a repository.

*del* allows you to remove an RPM package from a repository or to remove a specific version of that package.

*list* show you a list of all repositories. You can also list specific versions or distributions of a repository.

*set* allows you to set a specific version for a distribution.

*clean* cleans up stuff, like orphaned packages.

*sync* forces pulp to perform a synchronize operation with an external upstream repository.

*version* displays the version of this tool.



