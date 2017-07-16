# The 'Reading List' Example

This example implements a clone of NYT's 'saved articles API' that allows users to save, delete and retrieve nytimes.com article URLs to the service.

Instead of utilizing NYT's auth, this example leans on Google OAuth for user identity. When running locally, GAE's dev_appserver.py appears to always inject a user with an ID of "0".

To run this service, you must have 'legacy' App Engine SDK installed at the latest version that supports Go 1.8 then execute the following command in the nested `/api` directory:

`goapp serve`

At that point the application should be served on http://localhost:8080.

A few highlights of this service worth calling out:

* reading-list.yaml
  * an Open API specification that describes the endpoints and how they support JSON _or_ Protobufs.
* gen-proto.sh 
  * a script that relies on github.com/NYTimes/openapi2proto to generate Protobuf IDL from the Open API spec along with the Protobuf stubs.
* api/sevice.go 
  * contains the actual marvin.Service implementation.

To have truly reproducable builds across environments in the GAE Standard environment, this example uses the `dep` command to ensure all dependencies.
