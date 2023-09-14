# trigger

[![Build Status](https://github.com/gerrittrigger/trigger/workflows/ci/badge.svg?branch=main&event=push)](https://github.com/gerrittrigger/trigger/actions?query=workflow%3Aci)
[![codecov](https://codecov.io/gh/gerrittrigger/trigger/branch/main/graph/badge.svg?token=YCXTOSU3WR)](https://codecov.io/gh/gerrittrigger/trigger)
[![Go Report Card](https://goreportcard.com/badge/github.com/gerrittrigger/trigger)](https://goreportcard.com/report/github.com/gerrittrigger/trigger)
[![License](https://img.shields.io/github/license/gerrittrigger/trigger.svg)](https://github.com/gerrittrigger/trigger/blob/main/LICENSE)
[![Tag](https://img.shields.io/github/tag/gerrittrigger/trigger.svg)](https://github.com/gerrittrigger/trigger/tags)



## Introduction

*trigger* is the Gerrit trigger written in Go.



## Prerequisites

- Go >= 1.18.0



## Run

```bash
version=latest make build
./bin/trigger --config-file="$PWD"/config/config.yml
```



## Docker

```bash
version=latest make docker
docker run -v "$PWD"/config:/tmp ghcr.io/gerrittrigger/trigger:latest --config-file=/tmp/config.yml
```



## Usage

```
usage: trigger --config-file=CONFIG-FILE [<flags>]

gerrit trigger

Flags:
  --help                     Show context-sensitive help (also try --help-long
                             and --help-man).
  --version                  Show application version.
  --config-file=CONFIG-FILE  Config file (.yml)
  --log-level="INFO"         Log level (DEBUG|INFO|WARN|ERROR)
```



## Settings

*trigger* parameters can be set in the directory [config](https://github.com/gerrittrigger/trigger/blob/main/config).

An example of configuration in [config.yml](https://github.com/gerrittrigger/trigger/blob/main/config/config.yml):

```yaml
apiVersion: v1
kind: trigger
metadata:
  name: trigger
spec:
  connect:
    frontendUrl: http://localhost:8080
    hostname: localhost
    name: gerrit
    http:
      password: pass
      username: user
    ssh:
      keyfile: /path/to/.ssh/id_rsa
      keyfilePassword: pass
      port: 29418
      username: user
  playback:
    eventsApi: http://localhost:8081/events
  trigger:
    events:
      - commentAdded:
          verdictCategory: "Code-Review"
          value: "+1"
        commentAddedContainsRegularExpression:
          value: "merged.*"
        commitMessage: "message.*"
        name: "patchset-created"
        patchsetCreated:
          excludeDrafts: false
          excludeTrivialRebase: false
          excludeNoCodeChange: false
          excludePrivateChanges: false
          excludeWIPChanges: false
        uploaderName: "name"
    projects:
      - branches:
          - pattern: main
            type: plain
        filePaths:
          - pattern: name
            type: plain
        forbiddenFilePaths:
          - pattern: "**/name"
            type: path
        repo:
          pattern: ".*"
          type: regexp
        topics:
          - pattern: name
            type: plain
  watchdog:
    periodSeconds: 20
    timeoutSeconds: 20
```

- spec.connect.frontendUrl: Gerrit URL
- spec.connect.hostname: Gerrit address
- spec.trigger.events.name: See **Events**
- spec.watchdog.periodSeconds: Period in seconds (0: turn off)
- spec.watchdog.timeoutSeconds: Timeout in seconds (0: turn off)



## Events

```
change-abandoned
change-merged
change-restored
comment-added
draft-published
hashtags-changed
merge-failed
patchset-created
patchset-notified
private-state-changed
project-created
ref-replicated
ref-replicated-done
ref-updated
rerun-check
reviewer-added
topic-changed
vote-deleted
wip-state-changed
```



## Parameters

```
GERRIT_BRANCH
GERRIT_CHANGE_COMMIT_MESSAGE
GERRIT_CHANGE_ID
GERRIT_CHANGE_NUMBER
GERRIT_CHANGE_OWNER
GERRIT_CHANGE_OWNER_EMAIL
GERRIT_CHANGE_OWNER_NAME
GERRIT_CHANGE_PRIVATE_STATE
GERRIT_CHANGE_SUBJECT
GERRIT_CHANGE_URL
GERRIT_CHANGE_WIP_STATE
GERRIT_EVENT_TYPE
GERRIT_HOST
GERRIT_NAME
GERRIT_PATCHSET_NUMBER
GERRIT_PATCHSET_REVISION
GERRIT_PATCHSET_UPLOADER
GERRIT_PATCHSET_UPLOADER_EMAIL
GERRIT_PATCHSET_UPLOADER_NAME
GERRIT_PORT
GERRIT_PROJECT
GERRIT_REFSPEC
GERRIT_SCHEME
GERRIT_TOPIC
```



## License

Project License can be found [here](LICENSE).



## Reference

- [gerrit-events](https://github.com/sonyxperiadev/gerrit-events)
- [gerrit-events-log](https://gerrit.googlesource.com/plugins/events-log/)
- [gerrit-ssh](https://github.com/craftsland/gerrit-ssh)
- [gerrit-ssh](https://gist.github.com/craftslab/2a89da7b65fd32aaf6c598145625e643)
- [gerrit-stream-events](https://gerrit-review.googlesource.com/Documentation/cmd-stream-events.html)
- [gerrit-trigger-playback](https://github.com/jenkinsci/gerrit-trigger-plugin/blob/master/src/main/java/com/sonyericsson/hudson/plugins/gerrit/trigger/playback/GerritMissedEventsPlaybackManager.java)
- [gerrit-trigger-plugin](https://github.com/jenkinsci/gerrit-trigger-plugin)
- [gerrit-trigger-watchdog](https://github.com/sonyxperiadev/gerrit-events/blob/master/src/main/java/com/sonymobile/tools/gerrit/gerritevents/watchdog/StreamWatchdog.java)
- [go-queue](https://github.com/alexsergivan/blog-examples/blob/master/queue)
- [go-ssh](https://golang.hotexamples.com/site/file?hash=0x622d73200b734b5b68931b92861d30d6f4ef184f0872a45c49cedf26a29fa965&fullName=main.go&project=aybabtme/multisshtail)
