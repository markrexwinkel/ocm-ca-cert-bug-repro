# OCM CA Cert error reproduction

This repo contains code that reproduces an error when transfering a component version including its resources.
OCM version: v0.16.0

This error did not occur when using OCM version v0.15.0.

>NOTE: the checked in version references v0.15.0. The code should run without errors. Changing the version to v0.16.0 the error will occur.

## Running the code

Use `docker compose up -d` to start two OCI registries, one running on port 5000 the other runs on port 5001. Both are secured with a self signed certificate and use authentication. Credentials can be found in the code.

Run the Go code.
