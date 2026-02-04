package company

import (
	"fmt"
	"gin-demo/internal/application/config"
	"gin-demo/internal/shared/base"
)

func RegisterRoutes(version string, app *config.App, handlers []base.HandlerConfig) {
	group := app.Group(fmt.Sprintf("/api/%s/companies", version))
	for _, h := range handlers {
		group.Handle(h.Method, h.Path, h.Handler)
	}
}
