package v1

import (
	"net/http"

	"bytetrade.io/web3os/installer/pkg/api/response"
	"bytetrade.io/web3os/installer/pkg/apiserver/runtime"
	"bytetrade.io/web3os/installer/pkg/core/storage"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
)

var ModuleVersion = runtime.ModuleVersion{Name: "webserver", Version: "v1"}

var tags = []string{"apiserver"}

func AddContainer(c *restful.Container, db storage.Provider) error {
	ws := runtime.NewWebService(ModuleVersion)
	ws.Consumes(restful.MIME_JSON)
	ws.Produces(restful.MIME_JSON)

	handler := New(db)

	ws.Route(ws.POST("/install").
		To(handler.handlerInstall).
		Doc("").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{}))

	ws.Route(ws.GET("/status").
		To(handler.handlerStatus).
		Doc("").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{}))

	// ws.Route(ws.POST("/download").
	// 	To(handler.handlerDownloadEx).
	// 	Doc("").
	// 	Metadata(restfulspec.KeyOpenAPITags, tags).
	// 	Returns(http.StatusOK, "", response.Response{}))

	ws.Route(ws.POST("/install_terminus").
		To(handler.handlerInstallTerminus).
		Doc("").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{}))

	ws.Route(ws.GET("/greetings").To(handler.handlerGreetings).
		Doc("").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(http.StatusOK, "", response.Response{}))

	c.Add(ws)

	return nil

}
