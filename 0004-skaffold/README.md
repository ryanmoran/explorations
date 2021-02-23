# Skaffold with Paketo Buildpacks

## How does Skaffold work?

[Skaffold](https://skaffold.dev/) executes as a command-line tool on your local
workstation. Running a `skaffold` command, like `skaffold run` will result in
your application being built and deployed into your Kubernetes cluster. Once
deployed, you can confirm that your application is running in your Kubernetes
cluster. Skaffold refers to this process as a "pipeline", and it expects to
find the definition of that pipeline in a `skaffold.yaml` file. The pipeline
executes by performing a number of stages in sequence. Pipelines have quite a
few stages, but the 2 primary ones are:
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

## File Sync

[File Sync](https://skaffold.dev/docs/pipeline-stages/filesync/) is one of the
pipeline staages of the skaffold worflow. (See
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
```json
docker inspect google-built-image | jq -r '.[].Config.Labels["io.buildpacks.build.metadata"]' | jq .bom
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

* There's an opportunity to build <lang-family>-skaffold-filesync buildpack that does the above functions.

#### Open questions

* What about deletions? Are they synced?
* File Sync requires tar. Does tiny have tar?

## Links of Interest
[Quick Start](https://skaffold.dev/docs/quickstart/)
[GitHub Repo](https://github.com/GoogleContainerTools/skaffold)
