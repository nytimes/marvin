[![GoDoc](https://godoc.org/github.com/NYTimes/marvin?status.svg)](https://godoc.org/github.com/NYTimes/marvin)

# Marvin is a go-kit server for Google App Engine
### Marvin + GAE -> _let's get it oooonnnn!_
![Marvin!](https://cdn-images-1.medium.com/max/1600/1*btef9arpHHTzuknPFBHdUw.jpeg)

Marvin provides common tools and structure for services being built on Google App Engine by leaning heavily on the [go-kit/kit/transport/http package](http://godoc.org/github.com/go-kit/kit/transport/http). The [service interface here](http://godoc.org/github.com/NYTimes/marvin#Service) is very similar to the service interface in [NYT's gizmo/server/kit package](https://godoc.org/github.com/NYTimes/gizmo/server/kit#Service) so teams can build very similar looking software but use vasty different styles of infrastructure.

Marvin has been built to work with Go 1.8, which is available on App Engine Standard.

<sub>The Marvin logo is based on the Go mascot designed by Renée French and copyrighted under the Creative Commons Attribution 3.0 license.</sub>
