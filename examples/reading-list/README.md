# The 'Reading List' Example

This example implements a clone of NYT's 'saved articles API' that allows users to save, delete and retrieve nytimes.com article URLs.

Instead of utilizing NYT's auth, this example leans on Google OAuth for user identity. When running locally, GAE's `dev_appserver.py` appears to always inject a user with an ID of "0".

To run this service, you must be using [Google Cloud SDK](https://cloud.google.com/appengine/docs/standard/go/download) >= `162.0.0` or the "original" App Engine Go SDK >= `1.9.56`. Then execute the following command:

```sh
# If using Cloud SDK:
dev_appserver.py server/app.yaml
# If using the original SDK:
goapp serve server/app.yaml
```

At that point the application should be served on http://localhost:8080.

A few highlights of this service worth calling out:

* [service.yaml](service.yaml)
  * An Open API specification that describes the endpoints and how they support JSON _or_ Protobufs.
* [gen-proto.sh](gen-proto.sh)
  * A script that relies on github.com/NYTimes/openapi2proto to generate Protobuf IDL from the Open API spec along with the Protobuf stubs via protoc.
* [service.go](service.go)
  * The actual [marvin.MixedService](http://godoc.org/github.com/NYTimes/marvin#MixedService) implementation.
* [client.go](client.go)
  * A go-kit client for programmatically accessing the API.
* [cmd/cli/main.go](cmd/cli/main.go)
  * A CLI wrapper around the go-kit client.
* [Gopkg.toml](Gopkg.toml)
  * To have truly reproducible builds across environments in the GAE Standard environment, this example uses the [dep](https://github.com/golang/dep) command to ensure all dependencies.
* [server/app.yaml](server/app.yaml)
  * The app config is in a nested directory to enable vendoring.
  * This structure (along with using the legacy SDK) is the only way we've been able to get it to work with the current SDKs available.
* [.drone.yaml](.drone.yaml)
  * An example configuration file for [Drone CI](http://readme.drone.io/) using the [NYTimes/drone-gae](https://github.com/nytimes/drone-gae) plugin for managing automated deployments to App Engine.


This live demo is currently running on [https://nyt-reading-list.appspot.com](https://nyt-reading-list.appspot.com/list.json).
