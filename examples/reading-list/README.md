# The 'Reading List' Example

This example implements a clone of NYT's 'saved articles API' that allows users to save, delete and retrieve nytimes.com article URLs.

Instead of utilizing NYT's auth, this example leans on Google OAuth for user identity. When running locally, GAE's dev_appserver.py appears to always inject a user with an ID of "0".

To run this service, you must have the 'legacy' App Engine SDK installed at the latest version that supports Go 1.8, then execute the following command in the nested `/api` directory:

`goapp serve`

At that point the application should be served on http://localhost:8080.

A few highlights of this service worth calling out:

* reading-list.yaml
  * An Open API specification that describes the endpoints and how they support JSON _or_ Protobufs.
* gen-proto.sh 
  * A script that relies on github.com/NYTimes/openapi2proto to generate Protobuf IDL from the Open API spec along with the Protobuf stubs.
* api/service.go 
  * The actual marvin.Service implementation.
* client.go
  * A go-kit client for programmatically accessing the API.
* cmd/cli/main.go
  * A CLI wrapper around the go-kit client.
* Gopkg.toml
  * To have truly reproducible builds across environments in the GAE Standard environment, this example uses the [dep](https://github.com/golang/dep) command to ensure all dependencies.
* api/app.yaml
  * The app config is in a nested directory to enable vendoring.
  * This structure (along with using the legacy SDK) is the only way we've been able to get it to work with the current SDKs available.
* .drone.yaml
  * An example configuration file for [Drone CI](http://readme.drone.io/) using the [NYTimes/drone-gae](https://github.com/nytimes/drone-gae) plugin for managing automated deployments to App Engine.


This live demo is currently running on [https://nyt-reading-list.appspot.com](https://nyt-reading-list.appspot.com/list.json).
