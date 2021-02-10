# Tilt with Paketo Buildpacks

## How does Tilt work?

[Tilt](https://tilt.dev/) runs as a daemon on your local machine. It reads a
`Tiltfile` in the root directory.  The `Tiltfile` contains directives that tell
the `tilt` daemon how to build and run your code. There's even a [`pack`
directive](https://github.com/tilt-dev/tilt-extensions/tree/master/pack) that
runs `pack build`.

An example Go app `Tiltfile` might look like this:

```python
# load the 'pack' module, call it `pack` for the remainder of this file.
load('ext://pack', 'pack')

# run `pack build example-go-image --builder gcr.io/paketo-buildpacks/builder:tiny`
pack(
  'example-go-image',
  builder='gcr.io/paketo-buildpacks/builder:tiny'
)

# register a YAML file that contains k8s resources
k8s_yaml('deployment.yml')

# deploy the k8s resource called 'example-go', port-forward the container to localhost:8080
k8s_resource(
  'example-go',
  port_forwards=8080
)
```

The `tilt` daemon creates a dependency graph and then executes the pieces of
that graph to completion. For example, the above `Tiltfile` resolves to
something like the following graph:

```
+--------------------------+
| k8s_resource(example-go) |
+--------------------------+
  |
  |
  v
+--------------------------+
| k8s_yaml(deployment.yml) |
+--------------------------+
  |
  |
  v
+--------------------------+
|  pack(example-go-image)  |
+--------------------------+
  |
  |
  v
+--------------------------+
|      <source code>       |
+--------------------------+
```

And so to get to a deployed `k8s_resource`, Tilt will follow the graph down to
the bottom and then work its way back up, performing each step along the way.

1. Run `pack build` to build an image called `example-go-image`.
1. Update the `deployment.yml` with a new image reference to the built
   `example-go-image` image.
1. Apply the `deployment.yml` manifest to the k8s cluster.

Tilt is constantly monitoring each of the parts of this dependency graph and
trying to keep everything up-to-date. So, when you modify your source code or
`deployment.yml`, you end up with your latest changes deployed to the k8s
cluster. It also monitors the `Tiltfile` itself, and will update the graph if
anything changes.

Tilt also has a nice UI that shows you all of this happening in realtime.

![Tilt UI](/0002-tilt/assets/ui.png)

## What is the inner loop like?

Well, currently, when I make changes to my app, it triggers a full `pack
build`, which depending upon your machine could be pretty slow (on the order of
90 seconds). This is worrying because a comparable `Tiltfile` that uses the
`docker_build` directive to build the image completes very quickly (on the
order of a handful of seconds).

### Why is `pack build` slow?

We tried a few experiments to see if we could speed up `pack build`. We started
by taking a look at what the `pack` directive was running under the covers.
That command ended up looking something like this:

```
pack build image-name:tilt-build-pack-caching \
  --path . \
  --builder gcr.io/paketo-buildpacks/builder:base
```

There are a few ways we can optimize this:

1. We can use the `tiny` builder so that we aren't pulling down such large
   images.

   ```
   pack build image-name:tilt-build-pack-caching \
     --path . \
     --builder gcr.io/paketo-buildpacks/builder:tiny
   ```

   This helped mostly on the initial building of the container, not so much on
   subsequent rebuilds.

1. We can specify which buildpack we want to use so that we aren't wasting time
   detecting. You can do this without pulling down a new buildpack image by
   just referring to the buildpack by its ID and version number:

   ```
   pack build image-name:tilt-build-pack-caching \
     --path . \
     --builder gcr.io/paketo-buildpacks/builder:tiny \
     --buildpack paketo-buildpacks/go@0.3.2
   ```

    There wasn't really any sort of noticable speed up in doing this.

1. We can make sure that the builder is trusted. This will run all of the build
   phases in the same container and save us any performance overhead that
   running an untrusted builder might incur.

   ```
   pack build image-name:tilt-build-pack-caching \
     --path . \
     --builder gcr.io/paketo-buildpacks/builder:tiny \
     --buildpack paketo-buildpacks/go@0.3.2 \
     --trust-builder
   ```

    We found that there was some advantage in doing this (on the order of
    10-20% improvement in build time).

1. We can also set the pull policy to `never` or `if-not-present`, which should
   reduce the overhead of looking for builder and run images at the beginning
   of the build process.

   ```
   pack build image-name:tilt-build-pack-caching \
     --path . \
     --builder gcr.io/paketo-buildpacks/builder:tiny \
     --buildpack paketo-buildpacks/go@0.3.2 \
     --trust-builder \
     --pull-policy never
   ```

    This had a pretty dramatic improvement on the rebuild performance (on the
    order of 20-30%).

Given all of the above changes, we were seeing rebuilds taking about half as long
was they were without any optimizations. These are all changes we could pretty
easily pull request into the original `pack` extension.

### Can it go any faster?

Fundamentally, the majority of the performance difference we are seeing between
the `pack` and `docker_build` directives is not in the "build" phase of the
buildpack API. We are seeing most of the time being spent in "detect",
"analyze", "restore", and "export" (almost a 10:1 ratio). Further performance
gains will likely need to come from the platform side as the buildpacks just
aren't doing anything that is all that slow.

## The "Live Update" paradigm

Tilt declares in its documentation that it is "focused on helping teams that
develop multi-service apps". This framing seems to make a lot of sense. Its
clear that if you have a bunch of apps communicating, Tilt can help keep that
sane. Its also important that when changes in these apps take place, that they
get updated quickly. Its clear that the Tilt team has thought about this
problem a bit and offers up their solution, "Live Update".

Conceptually, a live update is a process that "patches" a running container.
There are a few ways to do this, but maybe the easiest to understand involves a
`local_resource`. You can define a `local_resource` that, for example, builds
your application binary:

```python
local_resource(
  'example-go-compile',
  'GOOS=linux GOARCH=amd64 go build -o my-app ./',
)
```

This `local_resource` is performing a process on your local workstation. It
results in a binary that you can then copy up into your running container. You
can perform the "syncing" part of the live update in a few ways, but one of the
easiest is to modify your `Tiltfile` to use the [`docker_build_with_restart`
directive](https://github.com/tilt-dev/tilt-extensions/tree/master/restart_process).

```python
docker_build_with_restart(
  'example-go-image',
  '.',
  entrypoint=['/my-app'],
  dockerfile='./Dockerfile',
  live_update=[
    sync('./my-app', '/my-app'),
  ],
)
```

This directive will build your image using the `Dockerfile` in the current
directory. Before it runs the build though, it modifies the `Dockerfile` to
include [some extra
bits](https://github.com/tilt-dev/tilt-extensions/blob/20477627fff083f228214a288228c39d1a18b564/restart_process/Tiltfile#L32-L40)
that allow the container to receive the updated binary and restart the process.

With these modifications, changing your application source code will cause your
app to be recompiled on your local machine, and then synced to a running
container where the running process will then be restarted. For a simple change
to a Go application, like the example included in this repo, this results in a
build that can take as little as 1 second.

### How could buildpacks get involved?

Looking at the parts of the "Live Update" feature set, it seems pretty
straight-forward to extrapolate how a buildpack might be built to include a
"Live Update" syncing utility when there is a `Tiltfile` in the working
directory. There are [existing tools](https://github.com/eradman/entr/) that we
could leverage to handle hot-reloading of the processes in the container.

The thing that's strange about what we've done above though is that we've
pushed the guts of the buildpack out onto the developers workstation. They are
now responsible for building their own binary, ensuring they have the right
modules and distribution to do this in the process. That whole idea seems to
run somewhat counter to what buildpacks is trying to achieve. Developers
shouldn't need to know any of those details.

Imagine you work at a company that uses Tilt and you've been tasked with
integrating a new Go service with a set of existing microservices that have
been written in a bunch of other languages (Python, Java, Node.js). In all
likelihood, you aren't a Python, Java, or Node.js developer, and it would
probably be pretty taxing to try to setup their apps if they were using the
`local_resource` directive to "build" their code before syncing it into their
containers.

## Live Updates with `run()` directives

There is another part of the "Live Update" feature that might help with this.
Tilt provides a `run` directive that can be used to execute arbitrary code on a
remote container. This means that we could remove the `local_resource`
directive above, and modify the `docker_build_with_restart` direction to look
something like the following:

```python
docker_build_with_restart(
  'example-go-image',
  '.',
  entrypoint=['/my-app'],
  dockerfile='./Dockerfile',
  live_update=[
    sync('./main.go', '/main.go'),
    run('go build -o /my-app main.go'),
  ],
)
```

This works, but it also assumes that we have a Go distribution available on the
container. It also reimplements a bit of the build process outside of the
buildpacks as the `run` directive specifies the exact command to execute.

### How could buildpacks get involved?

Much of the overhead caused when `pack build` rebuilds an app image could be
avoided if the `run` directive was used to _re-run the build phase of the
lifecycle inside the running container_. Suppose there were a `rebuild` binary
in the running app container that properly re-executes the build phase. This
binary could be invoked via a `run()` directive.

 The advantages here are:
* The build process remains abstracted from users, who don't need to specify
  build commands themselves
* Rebuilds can trigger changes to any/all of the layers for which the
  buildpacks are responsible
    * For example, `rebuild` can invoke both `go mod vendor` and `go build`, as
      appropriate
 * Builds occur in the app container, not on the developer's local, maintaining
   the isolation benefits of containers
 * (Re)builds are reproducible - the rebuild runs the same bits as the fresh
   build

There are, however, a number of complicating factors/prerequisites for this,
some of which are outlined here.

When the lifecycle invokes a [buildpack's `build`
phase](https://github.com/buildpacks/spec/blob/main/buildpack.md#phase-3-build),
it has the following **preconditions**:

> GIVEN:
>
> * The final ordered group of buildpacks determined during the detection
> phase,
> * A directory containing application source code,
> * The Buildpack Plan,
> * Any `<layers>/<layer>.toml` files placed on the filesystem during
> the analysis phase,
> * Any locally cached `<layers>/<layer>` directories, and
> * A shell, if needed,

In the final app image produced by a `pack build` many of these are no longer
available, because
* buildpacks may remove source code from the image during `build`
* the `group.toml` and `plan.toml` are not included in the built image
* locally cached `<layers>/<layer>` are provided by the platform at (re)build
 time
* the `tiny` run image does not have a shell
* the buildpacks that executed during the build phase are not included in the
  built image

If we want to use Tilt's `run()` API to rebuild by re-invoking buildpacks'
`build` binaries, the image produced by the initial build must provide for
these prerequisites. The running container would need, at least:
* the application source code in a known location
* record of what buildpacks were detected (`group.toml`)
* record of what the build plans were (`plan.toml`)
* buildpacks would need to be exported into the app image
* layers would need to be exported into the app image regardless of the
  presence of the `launch` flag
* the app image should be based on top of the build stack image
  * includes required mixins
  * (maybe) has a shell

With all of the above included in the app image, a `rebuild` binary could:
1. Stop the running process
1. Unset launch-time environment variables
1. Use information in `<layers>` to appropriately set up build-time environment
   variables
1. Invoke buildpacks' `build` binaries with the correct arguments
1. Unset build-time environment variables,
1. Set launch-time environment variables
1. Invoke the start command for the app

Some complicating factors/open questions:
* These images would be much bigger than those we currently produce.
* There are non-trivial differences between build and run images.
* We won't have any record of user-provided build-time environment variables?

## Other Potentially Buildpacks-Friendly Tilt User Personas

### Developers who maintain "dev" and "prod" Dockerfiles

As discussed earlier in this doc, using the `live_update` feature of docker
builds with Tilt often involves building components of the app on the dev's
local, so that they can be copied into a running container via a `sync()`
directive. This results in [Dockerfiles that do little more than copy pre-built
binaries](https://github.com/tilt-dev/tilt-example-go/blob/master/3-recommended/deployments/Dockerfile).
Obviously, these aren't production-ready. We wonder whether Tilt pushes devs
toward maintaining separate dev and prod Dockerfiles, where dev Dockerfiles
play nicely with Tilt, and prod ones follow best practices, are highly
portable, pass compliance checks, etc.

Since buildpacks' main goal is to build excellent, production-ready images,
perhaps they can solve devs' pain of maintaining two Dockerfiles. They can
continue to use a Dockerfiles for when they `tilt up`, but use buildpacks to
produce their prod images.

For these devs, exposing more information about the commands run during build
could help them match the `local_resource()` in their Tiltfile to the build
process that'll be used for prod.

### Developers who run "Hybrid" or "All Remote" builds

Not everyone who uses Tilt will build all of the microservices in their project
on their local machine, or deploy them in a local cluster. The [Local vs Remote
Services](https://docs.tilt.dev/local_vs_remote.html) section of the docs
highlights a few other workflows. Buildpacks could help teams with "Hybrid"
setups, who may rely on

> Local pre-built services installed from an existing image or Helm chart

by helping teams generate images for other teams to use, or provide an "it just
works" experience to build other teams' images that might be part of a `tilt
up`, but aren't the service that the dev team is iterating on.

Buildpacks could help teams with "All Remote" setups by addressing some of the
pain points in the ["Remote
Builds"](https://docs.tilt.dev/local_vs_remote.html#remote-builds) section:

> * Configuring the build jobs
> * Communication between the build jobs and your cluster image registry
> * Caching builds effectively
> * Sending only diffs of the build context, instead of re-uploading the same
>   files over and over

These are some of the pains that buildpacks (and kpack) are already trying to address.
