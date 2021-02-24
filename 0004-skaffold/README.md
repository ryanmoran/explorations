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

## File Sync

[File Sync](https://skaffold.dev/docs/pipeline-stages/filesync/) is one of the
pipeline stages of the skaffold worflow. (See
[diagram](https://skaffold.dev/docs/pipeline-stages/)). With File Sync,
whenever a change is made to a subset of files, they are copied to the deployed
container, instead of going through a rebuild + redeploy + restart of the
deployed pod. This is done for files that do not need to need to be 'built'
(e.g. static files, js files etc)

There are 3 modes of File Sync: `manual`, `inferred`, `auto`. Inferred mode is
only supported with Dockerfiles. Auto sync mode is enabled by default
for Buildpacks artifacts. To disable auto sync, set `sync.auto = false` in
`skaffold.yaml`. To [use manual
mode](https://skaffold.dev/docs/pipeline-stages/filesync/#manual-sync-mode)
with buildpacks, users can provide sync rules in the `skaffold.yml`.

As of now, `auto` mode File Sync works for Google buildpacks, but does not
work for Paketo buildpacks. When using paketo buildpacks, it rebuilds on every
change unless manual syncing is used.

Skaffold requires buildpacks to do the following for the `auto` sync to work.
* Buildpacks must set under the Bill-of-Materials (BOM) metadata a
  `metadata.devmode.sync` key whose value lists the sync rules based on the
  language/ecosystem of the buildpack. Those sync rules will then be used by
  Skaffold without the user having to configure them manually.

See an example of BOM set by the google nodejs buildpack on the image:

`docker inspect google-built-image | jq -r '.[].Config.Labels["io.buildpacks.build.metadata"]' | jq .bom`

```json
[
  {
    "name": "",
    "metadata": {
      "devmode.sync": [
        {
          "dest": "/workspace",
          "src": "**/*.js"
        },
        {
          "dest": "/workspace",
          "src": "**/*.mjs"
        },
        {
          "dest": "/workspace",
          "src": "**/*.coffee"
        },
        {
          "dest": "/workspace",
          "src": "**/*.litcoffee"
        },
        {
          "dest": "/workspace",
          "src": "**/*.json"
        },
        {
          "dest": "/workspace",
          "src": "public/**"
        }
      ]
    },
    "buildpack": {
      "id": "google.nodejs.npm",
      "version": "0.9.0"
    }
  }
]
```

Skaffold injects `GOOGLE_DEVMODE=1` environment variable (as if it's in a
`project.toml`) and sets `sync.auto = false` when running `skaffold dev`.  The
Buildpacks can use this env var value as signal to change the way the
application is built so that it reloads the changes or rebuilds the app on each
change.

#### How can Paketo Buildpacks support Auto File Sync?

Paketo buildpacks have to do the following functions:

- Check if `GOOGLE_DEVMODE` variable is set.
  - If not set:
    - Usual worflow setting start commands etc.
  - Else if set:
    - Add sync metadata relevant to the language/ecosystem to the BOM under the
      key `metadata.devmode.sync` (i.e. `jq -r '.[].Config.Labels["io.buildpacks.build.metadata"] | fromjson.bom[].metadata["devmode.sync"]'`)
    - Modify the start command to be wrapped in a filewatcher (like watchexec) so
      that start command is restarted on file change.

See [gcp yarn buildpack](https://github.com/GoogleCloudPlatform/buildpacks/blob/10ca4b2e7d2606480238f63df45633bd0d282197/cmd/nodejs/yarn/main.go#L98)

* There's an opportunity to build `<lang-family>-skaffold-filesync` buildpack that does the above functions.

#### Open questions

* What about deletions? Are they synced?
* File Sync requires tar. Does tiny have tar?

## Remote Debugging

Skaffold has support for remote debugging the containers that it deploys. You
can run a `skaffold debug` command on any Skaffold pipeline and that pipeline
will run in "debug mode". As it exists today, Skaffold does all of the work to
enable this. When `skaffold debug` is run, the buildpacks execute as normal and
produce an image. Skaffold then modifies the Pod start command such that it
invokes the right kind of debug tooling. For example, Skaffold will modify a Go
image by rewriting the Pod start command to execute `dlv` and bind it to a
port. Additionally, it includes the `dlv` command via a volume mount that it
includes in the Pod spec.

## Links of Interest
[Quick Start](https://skaffold.dev/docs/quickstart/)
[GitHub Repo](https://github.com/GoogleContainerTools/skaffold)
