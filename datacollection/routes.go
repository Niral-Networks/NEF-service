package datacollection

import (
	"strings"

	"github.com/Niral-Networks/NEF-service/commom"
	"github.com/Niral-Networks/NEF-service/logger"
	"github.com/free5gc/logger_util"
	"github.com/gin-gonic/gin"
)

type Routes []commom.Route

// NewRouter returns a new router.
func NewRouter() *gin.Engine {
	router := logger_util.NewGinWithLogrus(logger.GinLog)
	AddService(router)
	return router
}
func AddService(engine *gin.Engine) *gin.RouterGroup {
	group := engine.Group("")

	for _, route := range routes {
		switch route.Method {
		case "GET":
			group.GET(route.Pattern, route.HandlerFunc)
		case "POST":
			group.POST(route.Pattern, route.HandlerFunc)
		case "PUT":
			group.PUT(route.Pattern, route.HandlerFunc)
		case "DELETE":
			group.DELETE(route.Pattern, route.HandlerFunc)
		case "PATCH":
			group.PATCH(route.Pattern, route.HandlerFunc)
		}
	}

	return group
}

var routes = Routes{

	{
		"AMFRegistrationAccept",
		strings.ToUpper("Post"),
		"/datacollection/amf-contexts/registration-accept123",
		HTTPAmfRegistrationAccept,
	},
	{
		"AMFRegistrationAccept",
		strings.ToUpper("Post"),
		"/datacollection/amf-contexts/registration-accept",
		SendHTTPAmfRegistrationResponseAMF,
	},
	{
		"AMFRegistrationAccept",
		strings.ToUpper("Post"),
		"/nnef-eventssubscription/v1/subscriptions",
		SendHTTPAmfRegistrationRequest,
	},
	{
		"AMFEventUnsubscribe",
		strings.ToUpper("Post"),
		"/nnef-eventssubscription/v1/unsubscribe",
		SendAmfEventUnsubscribeRequest,
	},
}
