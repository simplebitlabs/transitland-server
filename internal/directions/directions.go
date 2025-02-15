package directions

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/interline-io/log"
	"github.com/interline-io/transitland-server/model"
)

type Handler interface {
	Request(context.Context, model.DirectionRequest) (*model.Directions, error)
}

type handlerFunc func() Handler

var handlersLock sync.Mutex
var handlers = map[string]handlerFunc{}

func RegisterRouter(name string, f handlerFunc) error {
	handlersLock.Lock()
	defer handlersLock.Unlock()
	if _, ok := handlers[name]; ok {
		return fmt.Errorf("handler '%s' already registered", name)
	}
	log.Tracef("Registering routing handler: %s", name)
	handlers[name] = f
	return nil
}

func getHandler(name string) (handlerFunc, bool) {
	handlersLock.Lock()
	defer handlersLock.Unlock()
	a, ok := handlers[name]
	return a, ok
}

func HandleRequest(ctx context.Context, pref string, req model.DirectionRequest) (*model.Directions, error) {
	var handler Handler
	handler = &lineRouter{}
	// Default to walking
	if !req.Mode.IsValid() {
		req.Mode = model.StepModeWalk
	}
	// Always use line
	if req.Mode == model.StepModeLine {
		pref = "line"
	}
	// Realtime auto requires aws
	if req.Mode == model.StepModeAuto && req.DepartAt == nil {
		pref = "aws"
	}
	// Default
	if pref == "" {
		pref = os.Getenv("TL_DEFAULT_ROUTER")
		// will default to line if invalid or empty
	}
	if hf, ok := getHandler(pref); ok {
		handler = hf()
	}
	h, err := handler.Request(ctx, req)
	a := log.For(ctx).Trace()
	if err != nil {
		a = log.For(ctx).Error().Err(err)
	}
	a = a.Str("mode", req.Mode.String()).
		Str("handler", pref).
		Float64("from_lat", req.From.Lat).
		Float64("from_lon", req.From.Lon).
		Float64("to_lat", req.To.Lat).
		Float64("to_lon", req.To.Lon)
	if h.Duration != nil {
		a = a.Float64("duration", h.Duration.Duration).Str("duration_units", h.Duration.Units.String())
	}
	if h.Distance != nil {
		a = a.Float64("distance", h.Distance.Distance).Str("distance_units", h.Distance.Units.String())
	}
	a.Msg("directions request")
	return h, err
}

func validateDirectionRequest(req model.DirectionRequest) error {
	if req.From == nil || req.To == nil {
		return errors.New("from and to waypoints required")
	}
	return nil
}

func wpiWaypoint(w *model.WaypointInput) *model.Waypoint {
	if w == nil {
		return nil
	}
	return &model.Waypoint{
		Lon:  w.Lon,
		Lat:  w.Lat,
		Name: w.Name,
	}
}
