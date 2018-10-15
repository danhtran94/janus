package basic

import (
	"errors"

	"github.com/go-chi/chi"

	"github.com/hellofresh/janus/pkg/plugin"
	"github.com/hellofresh/janus/pkg/proxy"
	"github.com/hellofresh/janus/pkg/router"
)

var (
	repo        Repository
	adminRouter router.Router
)

func init() {
	plugin.RegisterEventHook(plugin.StartupEvent, onStartup)
	plugin.RegisterEventHook(plugin.AdminAPIStartupEvent, onAdminAPIStartup)

	plugin.RegisterPlugin("basic_auth", plugin.Plugin{
		Action: setupBasicAuth,
	})
}

func setupBasicAuth(def *proxy.RouterDefinition, rawConfig plugin.Config) error {
	if repo == nil {
		return errors.New("the repository was not set by onStartup event")
	}

	def.AddMiddleware(NewBasicAuth(repo))
	return nil
}

func onAdminAPIStartup(event interface{}) error {
	e, ok := event.(plugin.OnAdminAPIStartup)
	if !ok {
		return errors.New("could not convert event to admin startup type")
	}

	adminRouter = e.Router
	return nil
}

func onStartup(event interface{}) error {
	var err error

	e, ok := event.(plugin.OnStartup)
	if !ok {
		return errors.New("could not convert event to startup type")
	}

	if e.MongoSession == nil {
		return ErrInvalidMongoDBSession
	}

	if adminRouter == nil {
		return ErrInvalidAdminRouter
	}

	repo, err = NewMongoRepository(e.MongoSession)
	if err != nil {
		return err
	}

	handlers := NewHandler(repo)
	adminRouter.Group("*", "/credentials/basic_auth", func(r chi.Router) {
		r.Get("/", handlers.Index())
		r.Post("/", handlers.Create())
		r.Get("/{username}", handlers.Show())
		r.Put("/{username}", handlers.Update())
		r.Delete("/{username}", handlers.Delete())
	})

	return nil
}
