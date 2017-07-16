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

// ScalingHandler expects to be registered at `/_ah/push-handlers/scale/{dir:(up|down}`
// as a JSON Endpoint and it expects the environment to be provisioned with
// `IDLE_INSTANCES_UP` and `IDLE_INSTANCES_DOWN` environment variables to set the number
// of min_idle_instances before and after scaling events.
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
