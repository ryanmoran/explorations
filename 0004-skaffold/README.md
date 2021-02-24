# Skaffold with Paketo Buildpacks

## How does Skaffold work?

[Skaffold](https://skaffold.dev/) executes as a command-line tool on your local
workstation. Running a `skaffold` command, like `skaffold run` will result in
your application being built and deployed into your Kubernetes cluster. Once
deployed, you can confirm that your application is running in your Kubernetes
cluster. Skaffold refers to this process as a "pipeline", and it expects to
find the definition of that pipeline in a `skaffold.yaml` file.

The pipeline executes by performing a number of stages in sequence. Pipelines
have quite a few stages, but the 2 primary ones are:
[Build](https://skaffold.dev/docs/pipeline-stages/builders/) and
[Deploy](https://skaffold.dev/docs/pipeline-stages/deployers/). By default a
pipeline will use a `Dockerfile` to build your application container image, and
then deploy that image using a Kubernetes manifest found in `k8s/*.yaml`. An
example of such a pipeline `skaffold.yaml` file might look like the following:

```yaml
apiVersion: skaffold/v2beta12
kind: Config
build:
  artifacts:
  - image: example-basic
```

This pipeline declares that an image called `example-basic` will be built with
the implied `Dockerfile` located in the same directory as this pipeline file.
By default this image will be built with the local Docker daemon, but Skaffold
does support remote build processes. Once the image is built, Skaffold will
find any Kubernetes manifests located in the `k8s` directory. An example of
such a manifest might look like the following:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: example-basic
spec:
  containers:
  - name: example-basic
    image: example-basic
```

In this case, we are deploying the `example-basic` image as a Pod to our
Kubernetes cluster. Once the Pod is deployed, you can confirm with `kubectl get
pod example-basic` that the application is up and running.

Executing `skaffold run` for this application might look like the following:

```
$ skaffold run
Generating tags...
 - example-basic -> example-basic:540176d
Checking cache...
 - example-basic: Not found. Building
Found [minikube] context, using local docker daemon.
Building [example-basic]...
Sending build context to Docker daemon  3.072kB
Step 1/8 : FROM golang:1.15 as builder
 ---> 05499cedca62
Step 2/8 : COPY main.go .
 ---> 12522855dcb9
Step 3/8 : ARG SKAFFOLD_GO_GCFLAGS
 ---> Running in e682fdde575f
 ---> 914bb1a7873d
Step 4/8 : RUN go build -gcflags="${SKAFFOLD_GO_GCFLAGS}" -o /app main.go
 ---> Running in b0660ba92950
 ---> a5cb3e1fbe1c
Step 5/8 : FROM alpine:3
 ---> 28f6e2705743
Step 6/8 : ENV GOTRACEBACK=single
 ---> Running in 3bd38b184251
 ---> 112abd187d6f
Step 7/8 : CMD ["./app"]
 ---> Running in 7017e8e5697f
 ---> e04dba39f650
Step 8/8 : COPY --from=builder /app .
 ---> 5b3233bf8cd9
Successfully built 5b3233bf8cd9
Successfully tagged example-basic:540176d
Tags used in deployment:
 - example-basic -> example-basic:5b3233bf8cd95c4edf74e0455b70c06bc804c526923156078d4417efe866c50c
Starting deploy...
 - pod/example-basic created
Waiting for deployments to stabilize...
Deployments stabilized in 74.095972ms
You can also run [skaffold run --tail] to get the logs
```

You can find the code for this in [`examples/basic/dockerfile`](examples/basic/dockerfile).

## Using buildpacks

Skaffold doesn't just support building using a `Dockerfile`. There is currently
[beta support](https://skaffold.dev/docs/pipeline-stages/builders/buildpacks/)
for building source code into container images using [Cloud Native
Buildpacks](https://buildpacks.io/). Taking the previous example, we can modify
our pipeline to use the buildpack feature:

```yaml
apiVersion: skaffold/v2beta12
kind: Config
build:
  artifacts:
  - image: example-basic
    buildpacks:
      builder: index.docker.io/paketobuildpacks/builder:base
```

Now, when we run `skaffold run`, our application container image will be built
using the CNB lifecycle and Paketo builder.

```
$ skaffold run
Generating tags...
 - example-basic -> example-basic:540176d-dirty
Checking cache...
 - example-basic: Not found. Building
Found [minikube] context, using local docker daemon.
Building [example-basic]...
base: Pulling from paketobuildpacks/builder
d519e2592276: Pull complete
d22d2dfcfa9c: Pull complete
b3afe92c540b: Pull complete
...
Digest: sha256:d9209fcc8d70314b66b3e67b0a64a09f95a69219fd24371c8f1cd78a8044e769
Status: Downloaded newer image for paketobuildpacks/builder:base
base-cnb: Pulling from paketobuildpacks/run
d519e2592276: Already exists
d22d2dfcfa9c: Already exists
b3afe92c540b: Already exists
e6b49eae9e5b: Pull complete
23204e8f2e10: Pull complete
1f233ac18c1d: Pull complete
Digest: sha256:d7ce225d0061cc80333a06577faeed266efe9e74c578470a948517b668b5630c
Status: Downloaded newer image for paketobuildpacks/run:base-cnb
0.10.2: Pulling from buildpacksio/lifecycle
5749e56bea71: Pull complete
ab6640ec168a: Pull complete
Digest: sha256:c3a070ed0eaf8776b66f9f7c285469edccf5299b3283c453dd45699d58d78003
Status: Downloaded newer image for buildpacksio/lifecycle:0.10.2
===> DETECTING
[detector] 2 of 5 buildpacks participating
[detector] paketo-buildpacks/go-dist  0.3.0
[detector] paketo-buildpacks/go-build 0.2.2
===> ANALYZING
[analyzer] Previous image with name "example-basic:latest" not found
===> RESTORING
===> BUILDING
[builder] Paketo Go Distribution Buildpack 0.3.0
[builder]   Resolving Go version
[builder]     Candidate version sources (in priority order):
[builder]       <unknown> -> ""
[builder]
[builder]     Selected Go version (using <unknown>): 1.15.8
[builder]
[builder]   Executing build process
[builder]     Installing Go 1.15.8
[builder]       Completed in 11.132s
[builder]
[builder] Paketo Go Build Buildpack 0.2.2
[builder]   Executing build process
[builder]     Running 'go build -o /layers/paketo-buildpacks_go-build/targets/bin -buildmode pie .'
[builder]       Completed in 6.749s
[builder]
[builder]   Assigning launch processes
[builder]     web: /layers/paketo-buildpacks_go-build/targets/bin/workspace
===> EXPORTING
[exporter] Adding layer 'paketo-buildpacks/go-build:targets'
[exporter] Adding 1/1 app layer(s)
[exporter] Adding layer 'launcher'
[exporter] Adding layer 'config'
[exporter] Adding layer 'process-types'
[exporter] Adding label 'io.buildpacks.lifecycle.metadata'
[exporter] Adding label 'io.buildpacks.build.metadata'
[exporter] Adding label 'io.buildpacks.project.metadata'
[exporter] Setting default process type 'web'
[exporter] *** Images (d744bf1422d3):
[exporter]       example-basic:latest
[exporter] Adding cache layer 'paketo-buildpacks/go-dist:go'
[exporter] Adding cache layer 'paketo-buildpacks/go-build:gocache'
Tags used in deployment:
 - example-basic -> example-basic:d744bf1422d31c703fe8e56a11bdd293580af3f669fe976fc3fa7d75df74d88d
Starting deploy...
 - pod/example-basic created
Waiting for deployments to stabilize...
Deployments stabilized in 126.091811ms
You can also run [skaffold run --tail] to get the logs
```

You can find the code for this in [`examples/basic/paketo`](examples/basic/paketo).

In Skaffold, buildpacks suffer from the same performance issues we outlined in
the [Tilt exploration](/0002-tilt/README.md#what-is-the-inner-loop-like). These
issues are inherent to the buildpack lifecycle and would need to be addressed
upstream.

## `skaffold dev`

Beyond the simple `skaffold run` command, Skaffold supports a ["development
loop" workflow](https://skaffold.dev/docs/workflows/dev/) that can be invoked
using the `skaffold dev` command. When this command is invoked, Skaffold will
rebuild and redeploy your application every time there are changes to the
source code.

In addition to the previously discussed Build and Deploy stages, `skaffold dev`
introduces a new [File
Sync](https://skaffold.dev/docs/pipeline-stages/filesync/) stage that allows
for faster development loop cycles by allowing users to "live update" their
running application container using a mechanism that is similar to what was
[previously outlined](/0002-tilt/README.md#the-live-update-paradigm) in the
Tilt Exploration.

When you first run `skaffold dev`, Skaffold will run the Build and Deploy
stages as normal, resulting in a running app on your Kubernetes cluster. Then,
on subsequent changes to your source code it will sync those files to the
running container and restart the process, skipping the Build and Deploy stages
entirely.

### Using File Sync with buildpacks

The File Sync feature is supported by the [Google
buildpacks](https://github.com/GoogleCloudPlatform/buildpacks). When Skaffold
runs, it provides the buildpack with a `GOOGLE_DEVMODE` environment variable.
The buildpacks modify their behavior based on the presence of this variable.

First, they emit a bill-of-materials that includes some metadata telling
Skaffold what files it should watch for changes and sync into the running
container.

Second, they include some build-time dependencies (like the Go
distribution) as launch layers to that they are available to rebuild go
applications while running.

Finally, they overwrite the start command to run a special script. This script
will run a process that watches the filesystem for changes. When files change,
it will rebuild the app and then restart the app process.

We can see this in operation if we run `skaffold dev` on an example that uses
the Google buildpacks. The following example builds a Go application that logs
"Hello world!" to `stdout` once every second.

```
$ skaffold dev
Listing files to watch...
 - example-file-sync
Generating tags...
 - example-file-sync -> example-file-sync:084abfe-dirty
Checking cache...
 - example-file-sync: Not found. Building
Found [minikube] context, using local docker daemon.
Building [example-file-sync]...
v1: Pulling from buildpacks/builder
Digest: sha256:20197a42da6a3d326e874a90c1a7178e37c5d0645ce8e9cf654c1d8984293787
Status: Image is up to date for gcr.io/buildpacks/builder:v1
v1: Pulling from buildpacks/gcp/run
Digest: sha256:784f4ff2f5ffa20be59668b08e979874f335e9c81704e73c495af8b245d4e9cf
Status: Image is up to date for gcr.io/buildpacks/gcp/run:v1
0.9.3: Pulling from buildpacksio/lifecycle
Digest: sha256:bc253af2edf1577717618cb3a95f0f16bb18fc9e804efbcc1b85f657d931a757
Status: Image is up to date for buildpacksio/lifecycle:0.9.3
===> DETECTING
[detector] 4 of 6 buildpacks participating
[detector] google.go.runtime  0.9.1
[detector] google.go.gopath   0.9.0
[detector] google.go.build    0.9.0
[detector] google.utils.label 0.0.1
===> ANALYZING
[analyzer] Restoring metadata for "google.go.runtime:go" from app image
[analyzer] Restoring metadata for "google.go.build:bin" from app image
[analyzer] Restoring metadata for "google.go.build:devmode_scripts" from app image
[analyzer] Restoring metadata for "google.go.build:watchexec" from app image
===> RESTORING
[restorer] Restoring data for "google.go.runtime:go" from cache
[restorer] Restoring data for "google.go.build:watchexec" from cache
===> BUILDING
[builder] === Go - Runtime (google.go.runtime@0.9.1) ===
[builder] --------------------------------------------------------------------------------
[builder] Running "curl --fail --show-error --silent --location https://golang.org/dl/?mode=json"
[builder] Done "curl --fail --show-error --silent --location https://golang...." (216.214195ms)
[builder] Using latest runtime version: 1.15.8
[builder] === Go - Gopath (google.go.gopath@0.9.0) ===
[builder] --------------------------------------------------------------------------------
[builder] Running "go get -d (GOPATH=/layers/google.go.gopath/gopath GO111MODULE=off)"
[builder] Done "go get -d (GOPATH=/layers/google.go.gopath/gopath GO111MODUL..." (242.114919ms)
[builder] === Go - Build (google.go.build@0.9.0) ===
[builder] --------------------------------------------------------------------------------
[builder] Running "go list -f {{if eq .Name \"main\"}}{{.Dir}}{{end}} ./..."
[builder] /workspace
[builder] Done "go list -f {{if eq .Name \"main\"}}{{.Dir}}{{end}} ./..." (116.977577ms)
[builder] --------------------------------------------------------------------------------
[builder] Running "go build -o /layers/google.go.build/bin/main ./. (GOCACHE=/layers/google.go.build/gocache)"
[builder] Done "go build -o /layers/google.go.build/bin/main ./. (GOCACHE=/l..." (512.913556ms)
[builder] === Utils - Label Image (google.utils.label@0.0.1) ===
===> EXPORTING
[exporter] Reusing layer 'google.go.runtime:go'
[exporter] Reusing layer 'google.go.build:bin'
[exporter] Reusing layer 'google.go.build:devmode_scripts'
[exporter] Adding layer 'google.go.build:gocache'
[exporter] Reusing layer 'google.go.build:watchexec'
[exporter] Adding 1/1 app layer(s)
[exporter] Reusing layer 'launcher'
[exporter] Reusing layer 'config'
[exporter] Reusing layer 'process-types'
[exporter] Adding label 'io.buildpacks.lifecycle.metadata'
[exporter] Adding label 'io.buildpacks.build.metadata'
[exporter] Adding label 'io.buildpacks.project.metadata'
[exporter] Setting default process type 'web'
[exporter] *** Images (5110be3ac725):
[exporter]       example-file-sync:latest
[exporter] Reusing cache layer 'google.go.runtime:go'
[exporter] Reusing cache layer 'google.go.build:watchexec'
Tags used in deployment:
 - example-file-sync -> example-file-sync:5110be3ac7252e4565247002c2a22c39066f4a0aa4353436188db463ae9b7a29
Starting deploy...
 - pod/example-file-sync created
Waiting for deployments to stabilize...
Deployments stabilized in 112.538089ms
Press Ctrl+C to exit
Watching for changes...
[example-file-sync] Hello world!
[example-file-sync] Hello world!
[example-file-sync] Hello world!
```

Then, when we modify the `main.go` to print "Hello moon!", it syncs that file
and runs the build and restart process from inside of the running container:

```
[example-file-sync] Hello world!
[example-file-sync] Hello world!
[example-file-sync] Hello world!
Syncing 1 files for example-file-sync:5110be3ac7252e4565247002c2a22c39066f4a0aa4353436188db463ae9b7a29
Watching for changes...
[example-file-sync] Hello world!
[example-file-sync] Hello moon!
[example-file-sync] Hello moon!
[example-file-sync] Hello moon!
```

You can find the code for this in [`examples/file-sync`](examples/file-sync).

We can take a look at the metadata that the Google buildpacks attached to the
image under the bill of materials by running the following:

```
$ pack inspect-image example-file-sync --bom
{
  "remote": null,
  "local": [
    {
      "name": "",
      "metadata": {
        "devmode.sync": [
          {
            "dest": "/workspace",
            "src": "**/*.go"
          }
        ]
      },
      "buildpacks": {
        "id": "google.go.build",
        "version": "0.9.0"
      }
    }
  ]
}
```

You can see that the buildpack is telling Skaffold to watch for any `*.go`
files and to sync to the `/workspace` in the running container when they
change.

Further inspection of that same image shows that it is using a start command
called `watch_and_run.sh`:

```
$ pack inspect-image example-file-sync
Inspecting image: example-file-sync

REMOTE:
(not present)

LOCAL:

Stack: google

Base Image:
  Reference: e5f7e62bf8a96d60f65f04a9d7caafb532d69c0ea955a276a0b476b0500b75ee
  Top Layer: sha256:64b8c91fa7e24f21fd7edb153d1f820055b4ae1cd542a1554501617bb0547df9

Run Images:
  gcr.io/buildpacks/gcp/run:v1

Buildpacks:
  ID                        VERSION
  google.go.runtime         0.9.1
  google.go.gopath          0.9.0
  google.go.build           0.9.0
  google.utils.label        0.0.1

Processes:
  TYPE                 SHELL        COMMAND        ARGS
  web (default)                     watch_and_run.sh
```

We can take a look at that script to see what is happening there:

```
$ docker run -it --entrypoint launcher example-file-sync bash
cnb@cb0ad6d9b096:/workspace$ find / -name watch_and_run.sh
/layers/google.go.build/devmode_scripts/bin/watch_and_run.sh
cnb@cb0ad6d9b096:/workspace$ cat /layers/google.go.build/devmode_scripts/bin/watch_and_run.sh
#!/bin/sh
watchexec -r -e go /layers/google.go.build/devmode_scripts/bin/build_and_run.sh
```

So, the start command uses a tool called
[`watchexec`](https://github.com/watchexec/watchexec) to watch for files with
the extension `go` and then run another script called `build_and_run.sh` when
they change.

Let's take a look at the `build_and_run.sh` script:

```
$ docker run -it --entrypoint launcher example-file-sync bash
cnb@d03924735435:/workspace$ cat /layers/google.go.build/devmode_scripts/bin/build_and_run.sh
#!/bin/sh
go build -o /layers/google.go.build/bin/main ./. && /layers/google.go.build/bin/main
```

The `build_and_run.sh` script runs the `go build` process creating a binary
called `main` and then runs the `main` executable.

This entire process is similar to what we outlined as a [possible path
forward](https://github.com/ryanmoran/explorations/tree/main/0002-tilt#how-could-buildpacks-get-involved-1)
in the Tilt exploration. Any implementation should push for a set of features
supported upstream by the buildpack spec such that we could leverage those
features in a platform agnostic way. Today, simply recreating what Google has
implemented would only allow us to integrate with Skaffold, limiting the reach
of the feature.

#### Stack Limitations

The File Sync feature operates by `exec`-ing a `tar` command that is given a
tarball with the contents of the files to be copied into the running container.
This means that the existing File Sync feature can only operate on stacks that
include the `tar` package. Currently, `tar` is not included in the Tiny run
image. Instead of adding `tar` to the Tiny stack, it may be better to have the
buildpack lifecycle base the image off of the build image, thus ensuring all of
the same dependencies that were available during the build process were also
available during this File Sync stage.

## `skaffold debug`

Skaffold has [support for remote
debugging](https://skaffold.dev/docs/workflows/debug/) the containers that it
deploys. You can run a `skaffold debug` command on any Skaffold pipeline and
that pipeline will run in "debug mode". As it exists today, Skaffold does all
of the work to enable this.

Let's explore this via an example. We are going to run `skaffold debug` using
the [`examples/debug`](examples/debug) codebase. This codebase uses a simple
Node.js server to respond to requests with "Hello world!"

```
skaffold debug
Listing files to watch...
 - example-debug
Generating tags...
 - example-debug -> example-debug:0918d37-dirty
Checking cache...
 - example-debug: Not found. Building
Found [minikube] context, using local docker daemon.
Building [example-debug]...
base: Pulling from paketobuildpacks/builder
...
Digest: sha256:45a889434ed64017eb18bfbf30c38db1b52566e0d341eb85d6c41fbed84b664f
Status: Downloaded newer image for paketobuildpacks/builder:base
base-cnb: Pulling from paketobuildpacks/run
...
Digest: sha256:e89f3ba15ab6ef4d43d1521c9238b5c74efcf78c1f52470bfec04bc2a025528b
Status: Downloaded newer image for paketobuildpacks/run:base-cnb
0.10.2: Pulling from buildpacksio/lifecycle
Digest: sha256:c3a070ed0eaf8776b66f9f7c285469edccf5299b3283c453dd45699d58d78003
Status: Image is up to date for buildpacksio/lifecycle:0.10.2
===> DETECTING
[detector] 3 of 6 buildpacks participating
[detector] paketo-buildpacks/node-engine 0.1.9
[detector] paketo-buildpacks/npm-install 0.2.6
[detector] paketo-buildpacks/npm-start   0.0.4
===> ANALYZING
[analyzer] Previous image with name "example-debug:latest" not found
===> RESTORING
===> BUILDING
[builder] Paketo Node Engine Buildpack 0.1.9
[builder]   Resolving Node Engine version
[builder]     Candidate version sources (in priority order):
[builder]                 -> ""
[builder]       <unknown> -> "*"
[builder]
[builder]     Selected Node Engine version (using ): 14.15.5
[builder]
[builder]   Executing build process
[builder]     Installing Node Engine 14.15.5
[builder]       Completed in 4.237s
[builder]
[builder]   Configuring build environment
[builder]     NODE_ENV     -> "production"
[builder]     NODE_HOME    -> "/layers/paketo-buildpacks_node-engine/node"
[builder]     NODE_VERBOSE -> "false"
[builder]
[builder]   Configuring launch environment
[builder]     NODE_ENV     -> "production"
[builder]     NODE_HOME    -> "/layers/paketo-buildpacks_node-engine/node"
[builder]     NODE_VERBOSE -> "false"
[builder]
[builder]     Writing profile.d/0_memory_available.sh
[builder]       Calculates available memory based on container limits at launch time.
[builder]       Made available in the MEMORY_AVAILABLE environment variable.
[builder]
[builder] Paketo NPM Install Buildpack 0.2.6
[builder]   Resolving installation process
[builder]     Process inputs:
[builder]       node_modules      -> "Not found"
[builder]       npm-cache         -> "Not found"
[builder]       package-lock.json -> "Not found"
[builder]
[builder]     Selected NPM build process: 'npm install'
[builder]
[builder]   Executing build process
[builder]     Running 'npm install --unsafe-perm --cache /layers/paketo-buildpacks_npm-install/npm-cache'
[builder]       Completed in 5.708s
[builder]
[builder]   Configuring launch environment
[builder]     NPM_CONFIG_LOGLEVEL -> "error"
[builder]
[builder]   Configuring environment shared by build and launch
[builder]     PATH -> "$PATH:/layers/paketo-buildpacks_npm-install/modules/node_modules/.bin"
[builder]
[builder]
[builder] Paketo NPM Start Buildpack 0.0.4
[builder]   Assigning launch processes
[builder]     web: node src/index.js
===> EXPORTING
[exporter] Adding layer 'paketo-buildpacks/node-engine:node'
[exporter] Adding layer 'paketo-buildpacks/npm-install:modules'
[exporter] Adding layer 'paketo-buildpacks/npm-install:npm-cache'
[exporter] Adding 1/1 app layer(s)
[exporter] Adding layer 'launcher'
[exporter] Adding layer 'config'
[exporter] Adding layer 'process-types'
[exporter] Adding label 'io.buildpacks.lifecycle.metadata'
[exporter] Adding label 'io.buildpacks.build.metadata'
[exporter] Adding label 'io.buildpacks.project.metadata'
[exporter] Setting default process type 'web'
[exporter] *** Images (74a07215c933):
[exporter]       example-debug:latest
[exporter] Adding cache layer 'paketo-buildpacks/node-engine:node'
[exporter] Adding cache layer 'paketo-buildpacks/npm-install:modules'
[exporter] Adding cache layer 'paketo-buildpacks/npm-install:npm-cache'
Tags used in deployment:
 - example-debug -> example-debug:74a07215c933f50c8c87f4012dde5da4b4ae257c66fac15ae5564b6618061167
Starting deploy...
 - service/web created
 - deployment.apps/web created
Waiting for deployments to stabilize...
 - deployment/web is ready.
Deployments stabilized in 4.811 seconds
Press Ctrl+C to exit
Not watching for changes...
[install-nodejs-debug-support] Installing runtime debugging support files in /dbg
[install-nodejs-debug-support] Installation complete
[web] Debugger listening on ws://0.0.0.0:9229/6d1e170a-35d7-477e-8738-20842e4d10d1
[web] For help, see: https://nodejs.org/en/docs/inspector
[web] Example app listening on port 3000!
```

We can see here that the output looks like a normal build process using the
Paketo Node.js buildpack. We can see that our application is up and running on
port 3000.

Our example codebase included a Kubernetes manifest that declared that a
LoadBalancer should attach 2 ports to the Deployment under a Service named
`web`. One for http and another for our debug process.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: web
spec:
  ports:
  - port: 3000
    name: http
  - port: 9229
    name: debug
  type: LoadBalancer
  selector:
    app: web
```

Let's find out how that LoadBalancer is bound to our host machine.

```
$ minikube service web
|-----------|------|-------------|---------------------------|
| NAMESPACE | NAME | TARGET PORT |            URL            |
|-----------|------|-------------|---------------------------|
| default   | web  | http/3000   | http://192.168.64.2:30262 |
|           |      | debug/9229  | http://192.168.64.2:30652 |
|-----------|------|-------------|---------------------------|
```

And when we make a request to that url, we can see our "Hello world!" response.

```
curl -vvv http://192.168.64.2:30262/hello
*   Trying 192.168.64.2...
* TCP_NODELAY set
* Connected to 192.168.64.2 (192.168.64.2) port 32266 (#0)
> GET /hello HTTP/1.1
> Host: 192.168.64.2:32266
> User-Agent: curl/7.64.1
> Accept: */*
>
< HTTP/1.1 200 OK
< X-Powered-By: Express
< Content-Type: text/html; charset=utf-8
< Content-Length: 12
< ETag: W/"c-00hq6RNueFa8QiEjhep5cJRHWAI"
< Date: Wed, 24 Feb 2021 18:24:47 GMT
< Connection: keep-alive
< Keep-Alive: timeout=5
<
* Connection #0 to host 192.168.64.2 left intact
Hello world!
* Closing connection 0
```

There was also some output that mentioned debug ports above.

```
[install-nodejs-debug-support] Installing runtime debugging support files in /dbg
[install-nodejs-debug-support] Installation complete
[web] Debugger listening on ws://0.0.0.0:9229/6d1e170a-35d7-477e-8738-20842e4d10d1
[web] For help, see: https://nodejs.org/en/docs/inspector
```

Let's use the Chrome Inspector to connect to the remote debugger. First, we
will open Chrome and navigate to `chrome://inspect`.

![Chrome Inspector](assets/chrome-inspector.png)

We'll click on the "Configure..." button and add `192.168.64.2:30652` as a
target.

![Configure Target](assets/configure-target.png)

Once we have configured the target, we can jump into a
[REPL](https://en.wikipedia.org/wiki/Read%E2%80%93eval%E2%80%93print_loop) that
executes inside of the running container.

![Inspect Target](assets/inspect-target.png)

Let's do something simple like printing "hello" to `stdout`.

![Debug REPL](assets/debug-repl.png)

Going back to our application logs, we can see that "hello" was printed to
`stdout`.

```
[web] Example app listening on port 3000!
[web] Debugger attached.
[web] hello
```

Now that we've confirmed that we can use `skaffold debug` with Paketo
buildpacks, let's dive a bit deeper into what Skaffold is doing to accomplish
this.

### How does this work?

Skaffold is doing a few things to our application to make this work. First, it
is rewriting the command that gets run in the container so that it enables
remote debugging.

Let's see what our image is specifying as the start command.

```
$ pack inspect-image example-debug
Inspecting image: example-debug

REMOTE:
(not present)

LOCAL:

Stack: io.buildpacks.stacks.bionic

Base Image:
  Reference: 883faa600da37641e47618dc1bbbdeaf9900114e43bef808a233109a2d0d6b7d
  Top Layer: sha256:16b4b083253b07f16420f4df9d60fe8e08be237fc9557c343172911de5e08b5c

Run Images:
  index.docker.io/paketobuildpacks/run:base-cnb
  gcr.io/paketo-buildpacks/run:base-cnb

Buildpacks:
  ID                                   VERSION
  paketo-buildpacks/node-engine        0.1.9
  paketo-buildpacks/npm-install        0.2.6
  paketo-buildpacks/npm-start          0.0.4

Processes:
  TYPE                 SHELL        COMMAND        ARGS
  web (default)        bash         node src/index.js
```

We can see that the buildpack has specified that the start command should be
`node src/index.js`. But when we look at the Pod spec, we see something
different.

```
$ kubectl describe pod web-8599f44c96-bxpnc
Name:         web-8599f44c96-bxpnc
Namespace:    default
Priority:     0
Node:         minikube/192.168.64.2
Start Time:   Wed, 24 Feb 2021 11:19:21 -0800
Labels:       app=web
              app.kubernetes.io/managed-by=skaffold
              pod-template-hash=8599f44c96
              skaffold.dev/run-id=4f1945dd-9b46-441c-ae90-a340c65c97cb
Annotations:  debug.cloud.google.com/config: {"web":{"artifact":"example-debug","runtime":"nodejs","workingDir":"/workspace","ports":{"devtools":9229}}}
Status:       Running
IP:           172.17.0.4
IPs:
  IP:           172.17.0.4
Controlled By:  ReplicaSet/web-8599f44c96
Init Containers:
  install-nodejs-debug-support:
    Container ID:   docker://fd5cb33bf22dad91946f2021eb443627d21546651a0e50bd0fd2b185b512b807
    Image:          gcr.io/k8s-skaffold/skaffold-debug-support/nodejs
    Image ID:       docker-pullable://gcr.io/k8s-skaffold/skaffold-debug-support/nodejs@sha256:33c49a754a87851bb6eecdb9b2f995c48100fe6fcd145171070a85eec89f7479
    Port:           <none>
    Host Port:      <none>
    State:          Terminated
      Reason:       Completed
      Exit Code:    0
      Started:      Wed, 24 Feb 2021 11:19:24 -0800
      Finished:     Wed, 24 Feb 2021 11:19:24 -0800
    Ready:          True
    Restart Count:  0
    Environment:    <none>
    Mounts:
      /dbg from debugging-support-files (rw)
      /var/run/secrets/kubernetes.io/serviceaccount from default-token-l2gqs (ro)
Containers:
  web:
    Container ID:  docker://3173a7599f8c7d7a0bfdab3a3724045b64a1729d545bfee89c7cba8c9b6429af
    Image:         example-debug:c34a5f3119bac9c309a4b300fd721c0b7d09b61d54bc66ce4c25f4f3621ac0a6
    Image ID:      docker://sha256:c34a5f3119bac9c309a4b300fd721c0b7d09b61d54bc66ce4c25f4f3621ac0a6
    Ports:         3000/TCP, 9229/TCP
    Host Ports:    0/TCP, 0/TCP
    Command:
      /cnb/lifecycle/launcher
    Args:
      node --inspect=0.0.0.0:9229 src/index.js
    State:          Running
      Started:      Wed, 24 Feb 2021 11:19:25 -0800
    Ready:          True
    Restart Count:  0
    Environment:
      PATH:  /dbg/nodejs/bin:/cnb/process:/cnb/lifecycle:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
    Mounts:
      /dbg from debugging-support-files (rw)
      /var/run/secrets/kubernetes.io/serviceaccount from default-token-l2gqs (ro)
Conditions:
  Type              Status
  Initialized       True
  Ready             True
  ContainersReady   True
  PodScheduled      True
Volumes:
  debugging-support-files:
    Type:       EmptyDir (a temporary directory that shares a pod's lifetime)
    Medium:
    SizeLimit:  <unset>
  default-token-l2gqs:
    Type:        Secret (a volume populated by a Secret)
    SecretName:  default-token-l2gqs
    Optional:    false
QoS Class:       BestEffort
Node-Selectors:  <none>
Tolerations:     node.kubernetes.io/not-ready:NoExecute op=Exists for 300s
                 node.kubernetes.io/unreachable:NoExecute op=Exists for 300s
Events:
  Type    Reason     Age   From               Message
  ----    ------     ----  ----               -------
  Normal  Scheduled  5m9s  default-scheduler  Successfully assigned default/web-8599f44c96-bxpnc to minikube
  Normal  Pulling    5m8s  kubelet            Pulling image "gcr.io/k8s-skaffold/skaffold-debug-support/nodejs"
  Normal  Pulled     5m8s  kubelet            Successfully pulled image "gcr.io/k8s-skaffold/skaffold-debug-support/nodejs" in 614.902485ms
  Normal  Created    5m7s  kubelet            Created container install-nodejs-debug-support
  Normal  Started    5m7s  kubelet            Started container install-nodejs-debug-support
  Normal  Pulled     5m6s  kubelet            Container image "example-debug:c34a5f3119bac9c309a4b300fd721c0b7d09b61d54bc66ce4c25f4f3621ac0a6" already present on machine
  Normal  Created    5m6s  kubelet            Created container web
  Normal  Started    5m6s  kubelet            Started container web
```

There is a lot of output above, but the key thing to notice is that the
command that will be run by the container has changed to `node
--inspect=0.0.0.0:9229 src/index.js`. We can see that in this snippet:

```yaml
Containers:
  web:
    Command:
      /cnb/lifecycle/launcher
    Args:
      node --inspect=0.0.0.0:9229 src/index.js
```

This will tell the `node` process that it should run in debug mode and bind to
`0.0.0.0:9229` to expose a remote debugger. Skaffold is doing something very
tricky here to make this happen. First, it reads the image metadata and
determines that the image was created by a buildpack. This can be done by
looking to see if the image labels include a label called
`io.buildpacks.build.metadata`. Once it has determined that the image was
created by a buildpack, it reads the image metadata to find the start command.
We can see that metadata by running the following:

```
$ docker inspect example-debug | jq -r -S '.[].Config.Labels["io.buildpacks.build.metadata"]' | jq -r -S .
{
  "bom": [
    {
      "buildpack": {
        "id": "paketo-buildpacks/node-engine",
        "version": "0.1.9"
      },
      "metadata": {
        "licenses": [],
        "name": "Node Engine",
        "sha256": "63a8a4d43c325856c2e4673b30dc44e8ea973a5a4f21ed616fa7ee2c31cfe7f4",
        "stacks": [
          "io.buildpacks.stacks.bionic",
          "org.cloudfoundry.stacks.cflinuxfs3"
        ],
        "uri": "https://buildpacks.cloudfoundry.org/dependencies/node/node_14.15.5_linux_x64_cflinuxfs3_63a8a4d4.tgz",
        "version": "14.15.5"
      },
      "name": "node"
    },
    {
      "buildpack": {
        "id": "paketo-buildpacks/npm-install",
        "version": "0.2.6"
      },
      "metadata": {
        "launch": true
      },
      "name": "node_modules"
    }
  ],
  "buildpacks": [
    {
      "homepage": "https://github.com/paketo-buildpacks/node-engine",
      "id": "paketo-buildpacks/node-engine",
      "version": "0.1.9"
    },
    {
      "homepage": "https://github.com/paketo-buildpacks/npm-install",
      "id": "paketo-buildpacks/npm-install",
      "version": "0.2.6"
    },
    {
      "homepage": "https://github.com/paketo-buildpacks/npm-start",
      "id": "paketo-buildpacks/npm-start",
      "version": "0.0.4"
    }
  ],
  "launcher": {
    "source": {
      "git": {
        "commit": "960cf58",
        "repository": "github.com/buildpacks/lifecycle"
      }
    },
    "version": "0.10.2"
  },
  "processes": [
    {
      "args": null,
      "buildpackID": "paketo-buildpacks/npm-start",
      "command": "node src/index.js",
      "direct": false,
      "type": "web"
    }
  ]
}
```

At the very bottom of that JSON document, we can see a section called
`processes` that specifies that our start command is `node src/index.js`.
Skaffold has some heuristics to determine the type of application this is. In
the case of a Node.js app, it will look to see if the command starts with
`node` or `npm`. Once it has been determined to be a Node.js app, Skaffold
transforms that command into one that will enable debugging. It has detection
heuristics and transformations that work for each of its supported languages
(Go, Node.js, Java, Python, and .Net Core). It then takes that transformed
command and updates our Kubernetes manifest to specify the modified command.

This works seamlessly for runtimes that include their own debugging support,
like Node.js. However, for other language runtimes that need an external
debugger, it also performs another step. It mounts another image that contains
these debugging tools at `/dbg` inside the container. This enables a Go
application that might have a start command like `./my-app` to be debugged
using a command like the following:

```
/dbg/go/bin/dlv exec --headless --listen=0.0.0.0:56268 ./my-app
```

In this case, the [`dlv` executable](https://github.com/go-delve/delve) is
mounted into the container by Skaffold at a `/dbg` mountpoint. We can see this
in our Pod spec here:

```
Containers:
  web:
    Mounts:
      /dbg from debugging-support-files (rw)
```

The debugging-support-files images are maintained in a [separate
repo](https://github.com/GoogleContainerTools/container-debug-support) and
include all of the extra debugging tools that are needed for each language
runtime.

Wow, so Skaffold is doing a lot of heavy lifting to make this feature work. The
cool part is that it "just works" for Paketo buildpacks today, but its less
than ideal that it only works for Skaffold.

### How might buildpacks better support `skaffold debug`?

The work that Skaffold is doing to make debugging possible is something that
makes sense to support more broadly using buildpacks. We can imagine a set of
"debug buildpacks" for each language family that modify the start command and
provide debugging tools to enable the same functionality that Skaffold is
implementing today. The work would mostly require establishing a standardized
API such that Skaffold would be able to indicate to the buildpacks that they
need to include debugging support, and then an API to indicate back to Skaffold
where the remote debugger is bound.

Making these needs more concrete, we could imagine that each of these debug
buildpacks would make their modifications to the container when they see an
environment variable like `BP_DEBUG`. And then those buildpacks could also
include extra buildpack labels to indicate how the remote debugger was
configured, including what port it may be bound to.

## Summary

Overall, Skaffold appears to be a relatively mature "development loop" toolkit
with some pretty good buildpacks integrations. There is a huge overlap in its
feature-set with Tilt, and it suffers many of the same limitations as Tilt with
regards to performance and alignment with the "buildpacks philosophy".

It also seems to have spiked out some ideas on how a faster "live update"
functionality might work with buildpacks. It is worth taking a deeper look at
their implementation with the intent to generalize those ideas into a "Develop
API" that could be standardized across the CNB specification.

Additionally, as also outlined in the Tilt exploration, we should invest in
enabling remote debugging support for our existing buildpacks. It is great to
see that Skaffold has a solution that appears to work with our buildpacks
without any need for those features to be built into the buildpacks, but we
should ensure that we can enable remote debugging on platforms beyond Skaffold
alone.

## Links of Interest
[Quick Start](https://skaffold.dev/docs/quickstart/)
[GitHub Repo](https://github.com/GoogleContainerTools/skaffold)
