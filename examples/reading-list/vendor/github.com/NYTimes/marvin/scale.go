package marvin

import (
	"context"
	"net/http"
	"strings"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	appengine "google.golang.org/api/appengine/v1"
	appengin "google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/module"
)

// ScalingHandler is a utility endpoint for adjusting the 'min_idle_instances' of the
// current service on the fly. While App Engine's automatic scaling is sufficient for
// most uses cases, we've found that some very large bursts of traffic (i.e. the spike
// just after the new NYT crossword is published) cannot keep up without a large number
// of idle instances at the ready. To avoid the high cost of always having large amounts
// of idle instances, this endpoint can be used to preemptively scale your service up
// before a spike and then back down to normal levels post spike.
//
// This handler expects to be registered at `/_ah/push-handlers/scale/{dir:(up|down}`
// as a JSON Endpoint and it expects the environment to be provisioned with
// `IDLE_INSTANCES_UP` and `IDLE_INSTANCES_DOWN` environment variables to set the number
// of min_idle_instances before and after scaling events.
//
// This endpoint can be hit using App Engine cron for regularly recurring spikes but to
// make this endpoint capable of sucurely accepting pushes from PubSub, it has the
// `/_ah/push-handlers/` prefix.
func ScalingHandler(ctx context.Context, _ interface{}) (interface{}, error) {
	// get the scaling settings out of the environment
	var scaling struct {
		Up   int64 `envconfig:"IDLE_INSTANCES_UP"`
		Down int64 `envconfig:"IDLE_INSTANCES_DOWN"`
	}
	envconfig.MustProcess("", &scaling)

	svcName := appengin.ModuleName(ctx)

	// find the default version of the current service
	vrsn, err := module.DefaultVersion(ctx, svcName)
	if err != nil {
		log.Errorf(ctx, "unable to lookup the default version: %s", err)
		return nil, NewJSONStatusResponse("error",
			http.StatusInternalServerError)
	}

	var idleInstances int64
	// get the scaling direction (up/down) and set the appropriate value
	dir := strings.TrimPrefix(ctx.Value(httptransport.ContextKeyRequestPath).(string),
		"/_ah/push-handlers/scale/")
	switch strings.ToLower(dir) {
	case "up":
		idleInstances = scaling.Up
	case "down":
		idleInstances = scaling.Down
	default:
		log.Errorf(ctx, "invalid scaling direction (%s) returning OK to empty queue", dir)
		return "OK", nil
	}

	// init the admin API client/service/call
	clnt, err := appengine.New(oauth2.NewClient(ctx,
		google.AppEngineTokenSource(ctx, appengine.CloudPlatformScope)))
	if err != nil {
		log.Errorf(ctx, "unable to init client: %s", err)
		return nil, NewJSONStatusResponse("error",
			http.StatusInternalServerError)
	}

	// setup the call for the admin API
	svc := appengine.NewAppsServicesVersionsService(clnt)
	call := svc.Patch(appengin.AppID(ctx), svcName, vrsn, &appengine.Version{
		AutomaticScaling: &appengine.AutomaticScaling{
			MinIdleInstances: idleInstances,
		},
	}).UpdateMask("automaticScaling.min_idle_instances")

	// execute the version patch request
	_, err = call.Do()
	if err != nil {
		log.Errorf(ctx, "unable to migrate traffic: %s", err)
		return nil, NewJSONStatusResponse("error",
			http.StatusInternalServerError)
	}

	return "OK", nil
}
