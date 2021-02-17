# Using the `ca-certificates` buildpack with Node.js

## Generating certificates

1. Run the following to generate a set of certificates and keys:
   ```
   go run ssl/main.go
   ```
1. Confirm that you can see 3 PEM-encoded files in the `ssl` directory:
   * `ca.pem`: The CA certificate that will be used to verify the presented certificate.
   * `cert.pem`: The certificate that is presented during the TLS handshake.
   * `key.pem`: The key used to encrypt traffic after the handshake completes.

## Building and running the Node.js app image

1. Copy the CA certificate into the binding directory:
   ```
   cp ssl/ca.pem server/binding/ca.pem
   ```
1. Copy the certificate and key into the server directory:
   ```
   cp ssl/cert.pem server/cert.pem
   cp ssl/key.pem server/key.pem
   ```
1. Run the following `pack build` command:
   ```
   pack build node-ca-certificates \
     --path ./server \
     --buildpack paketo-buildpacks/ca-certificates \
     --buildpack paketo-buildpacks/nodejs
   ```
1. Run the following `docker` command to start the container:
   ```
   docker run \
     -it \
     -p 8080:8080 \
     --env PORT=8080 \
     --env SERVICE_BINDING_ROOT=/bindings \
     --env NODE_OPTIONS="--use-openssl-ca" \
     --volume "$(pwd}/server/binding:/bindings/ca-certificates" \
     node-ca-certificates
     ```
     Note that the command includes `NODE_OPTIONS=--use-openssl-ca`. This
     command-line argument to the `node` process tells it to delegate to
     OpenSSL when verifying certificates. This is important because otherwise
     Node will use its bundled CA certificate set and not see that we have
     added a new CA to the root store.

## Making requests to the server

1. Copy the CA certificate, certificate, and key into the server directory:
   ```
   cp ssl/ca.pem client/ca.pem
   cp ssl/cert.pem client/cert.pem
   cp ssl/key.pem client/key.pem
   ```
1. Run the following to make an authenticated request to the server:
   ```
   go run client/main.go
   ```
   You should see that the response reports success.
1. Now try making that same request using a simple `curl` command:
   ```
   curl -vvv --cacert client/ca.pem https://localhost:8080
   ```
   You should see that the response reports that the request is unauthenticated.
1. Now try making that same request using a `curl` command that presents the cert:
   ```
   curl -vvv \
     --cert client/cert.pem \
     --key client/key.pem \
     --cacert client/ca.pem \
     https://localhost:8080
   ```
   You should see that the response is the same as the Go client response.
