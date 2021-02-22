## Links
[Skaffold](https://skaffold.dev/)
[Quick Start](https://skaffold.dev/docs/quickstart/)
[GitHub Repo](https://github.com/GoogleContainerTools/skaffold)

## Installing Skaffold

1. Download and install the `skaffold` CLI:
   ```
   brew install skaffold
   ```
1. Download and install `minikube`:
   ```
   brew install minikube
   ```

## Running the example

1. Run `minikube start`
1. Change directory into the `example-dockerfile` subdirectory.
1. Run `skaffold dev`

## Running the buildpacks example

1. Run `minikube start` if you haven't already.
1. Change directory into the `example-buildpacks` subdirectory.
1. Run `skaffold dev`
