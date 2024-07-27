package runtime

import (
	"fmt"

	"github.com/emicklei/go-restful/v3"
)

const (
	APIRootPath = "/api"
)

type ModuleVersion struct {
	Name    string
	Version string
}

func NewWebService(mv ModuleVersion) *restful.WebService {
	webservice := restful.WebService{}

	webservice.Path(fmt.Sprintf("%s/%s/%s", APIRootPath, mv.Name, mv.Version)).
		Produces(restful.MIME_JSON)

	return &webservice
}
