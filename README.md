# Smokeshift
[![Build Status](https://travis-ci.org/cyberbliss/smokeshift.svg?branch=master)](https://travis-ci.org/cyberbliss/smokeshift)

`smokeshift` is a command-line utility for smoke testing an Openshift install. It has been forked from Apprenda's Kismatic tool: [kuberang](https://github.com/apprenda/kuberang)

It scales out a pair of services, checks their connectivity, and then scales back in again.

For the latest build check out the [release](https://github.com/opencredo/smokeshift/releases) page.

![smokeshift_demo](https://cloud.githubusercontent.com/assets/5401528/23824487/9c2024ca-066f-11e7-978c-b9e370ac0b0d.gif)

### Features
Smokeshift will tell you if the machine and account from which you run it:
* Has oc installed correctly
* Has available nodes
* Has working pod & service networks
* Has working pod <-> pod DNS
* Has working master(s)
* Has the ability to access pods and services from the node you run it on.

**NOTE:** `smokeshift` will exit with a code of 0 even if some of the above are not possible. It's up to the user to parse the output.

Adding -o json will return a json blob (that can be parsed) instead of a pretty string report.

It is recommended to run `smokeshift` from either outside the Openshift cluster or a Master within the cluster.

### Pre-requisites
* A working oc (or all you'll get is a message complaining about oc)
* Access to a Docker registry with busybox and nginx images

### Usage

```
$ smokeshift --help
smokeshift is intended to perform smoke tests against an Openshift cluster. It expects the oc cli
to be available on the path and a user, with cluster-admin access, to already have authenticated. The actual smoke test
creates a Project (smoketest), deploys an Nginx Pod on to each Node, provisions a busybox and uses that to ensure access
using DNS and IP based connections to the Nginx Pods. Unless the 'skip-cleanup' flag is set all Pods, Services and the
smokeshift Project are deleted on completion

Usage:
  smokeshift [flags]

Flags:
      --registry-url string   Override the default Docker Hub URL to use a local offline registry for required Docker images.
      --skip-cleanup          Don't clean up. Leave all deployed artifacts running on the cluster.

```

# Developer notes
### Pre-requisites
- Go 1.8 installed

### Build using make
We use `make` to clean, build, and produce our distribution package. Take a look at the Makefile for more details.

In order to build the Go binaries (creates binaries in ./bin)
```
make build
```

In order to clean:
```
make clean
```

To run tests:
```
make test
```