# Tilt with Paketo Buildpacks

## Installing Tilt on your workstation

1. Install [Docker for Mac](https://docs.docker.com/docker-for-mac/install/)
1. In the Docker preferences, click [Enable Kubernetes](https://docs.docker.com/docker-for-mac/#kubernetes)
1. Set `kubectl` context to `docker-desktop`: `kubectl config use-context docker-desktop`
1. Install Tilt: `brew install tilt-dev/tap/tilt`
1. Run `tilt up`
