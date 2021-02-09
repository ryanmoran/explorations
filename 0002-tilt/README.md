# Tilt with Paketo Buildpacks

## Installing Tilt on your workstation

1. Install [Docker for Mac](https://docs.docker.com/docker-for-mac/install/)
1. In the Docker preferences, click [Enable Kubernetes](https://docs.docker.com/docker-for-mac/#kubernetes)
1. Set `kubectl` context to `docker-desktop`: `kubectl config use-context docker-desktop`
1. Install Tilt: `brew install tilt-dev/tap/tilt`

## How does Tilt work?

Tilt runs as a daemon on your local machine. It reads a `Tiltfile` in the root
directory.  The `Tiltfile` contains directives that tell the `tilt` daemon how
to build and run your code. There's even a [`pack`
directive](https://github.com/tilt-dev/tilt-extensions/tree/master/pack) that
runs `pack build`.

An example Go app `Tiltfile` might look like this:

```python
# load the 'pack' module, call it `pack` for the remained of this file.
load('ext://pack', 'pack')

# run `pack build example-go-image --builder gcr.io/paketo-buildpacks/builder:tiny`
pack(
  'example-go-image',
  builder='gcr.io/paketo-buildpacks/builder:tiny'
)

# register a YAML file that contains k8s resource
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

### Where is `pack build` slow?

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

1. We can specify which buildpack we want to use so that we are spending
   wasteful time detecting. You can do this without pulling down a new
   buildapck image by just referring to the buildpack by its ID and version
   number:

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
     -- pull-policy never
   ```

    This had a pretty dramatic improvement on the rebuild performance (on the
    order of 20-30%).

Given all of the above changes, we we seeing rebuilds taking about half as long
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
get updated quickly. It clear that the Tilt team has thought about this problem
a bit and offers up their solution, "Live Update".

Conceptually, a live update is a process that "patches" a running container.
There are a few ways to do this, but maybe the easiest to understand involves a
`local_resource`. You can define a `local_resource` that, for example, builds
your application binary:

```python
local_resource(
  'example-go-compile',
  'go build -o my-app ./',
)
```

This `local_resource` is performing a process on your local develop
workstation. It results in a binary that you can then copy up into your running
container. You can perform the "syncing" part of the live update in a few ways,
but one of the easiest is to modify your `Tiltfile` to use the
[`docker_build_with_restart`
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
