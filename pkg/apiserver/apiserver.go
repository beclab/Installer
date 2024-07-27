package apiserver

import (
	"net/http"
	"path"

	"bytetrade.io/web3os/installer/pkg/api/response"
	apisV1alpha1 "bytetrade.io/web3os/installer/pkg/apis/backend/v1"
	"bytetrade.io/web3os/installer/pkg/constants"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/storage"
	"bytetrade.io/web3os/installer/pkg/core/util"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/go-openapi/spec"
	"github.com/pkg/errors"
	urlruntime "k8s.io/apimachinery/pkg/util/runtime"
)

type APIServer struct {
	Server          *http.Server
	StorageProvider storage.Provider
	container       *restful.Container
}

func New() (*APIServer, error) {
	s := &APIServer{}

	server := &http.Server{
		Addr: constants.ApiServerListenAddress,
	}

	s.Server = server
	return s, nil
}

func (s *APIServer) PrepareRun() error {
	s.container = restful.NewContainer()
	s.container.RecoverHandler(logStackOnRecover)
	s.container.Filter(cors)
	s.container.Filter(func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		defer func() {
			if e := recover(); e != nil {
				response.HandleInternalError(resp, errors.Errorf("server internal error: %v", e))
			}
		}()

		chain.ProcessFilter(req, resp)
	})
	// s.container.Filter(authenticate)
	s.container.Router(restful.CurlyRouter{})

	s.installStaticResources()
	s.installStorage()
	s.installModuleAPI()
	s.installAPIDocs()

	var modulePaths []string
	for _, ws := range s.container.RegisteredWebServices() {
		modulePaths = append(modulePaths, ws.RootPath())
	}
	logger.Infow("registered module", "paths", modulePaths)

	s.Server.Handler = s.container
	return nil
}

func (s *APIServer) Run() error {
	err := s.Server.ListenAndServe()
	if err != nil {
		return errors.Errorf("listen and serve err: %v", err)
	}
	return nil
}

func (s *APIServer) installAPIDocs() {
	config := restfulspec.Config{
		WebServices:                   s.container.RegisteredWebServices(), // you control what services are visible
		APIPath:                       "./apidocs.json",
		PostBuildSwaggerObjectHandler: enrichSwaggerObject}
	s.container.Add(restfulspec.NewOpenAPIService(config))
}

func (s *APIServer) installStaticResources() {
	ws := &restful.WebService{}

	ws.Route(ws.GET("/").To(staticFromPathParam))            // staticFromQueryParam
	ws.Route(ws.GET("/{subpath:*}").To(staticFromPathParam)) // staticFromPathParam

	s.container.Add(ws)
}

func (s *APIServer) installStorage() {
	storageDir := path.Join(constants.WorkDir, "db")
	if ok := util.IsExist(storageDir); !ok {
		util.CreateDir(storageDir)
	}
	s.StorageProvider = storage.NewSQLiteProvider(storageDir)
	if err := s.StorageProvider.StartupCheck(); err != nil {
		logger.Errorf("db connect failed: %v", err)
		panic(err)
	}
}

func (s *APIServer) installModuleAPI() {
	urlruntime.Must(apisV1alpha1.AddContainer(s.container, s.StorageProvider))
}

func enrichSwaggerObject(swo *spec.Swagger) {
	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "Installer API Server Docs",
			Description: "Backend For Installer",
			Contact: &spec.ContactInfo{
				ContactInfoProps: spec.ContactInfoProps{
					Name:  "bytetrade",
					Email: "dev@bytetrade.io",
					URL:   "http://bytetrade.io",
				},
			},
			License: &spec.License{
				LicenseProps: spec.LicenseProps{
					Name: "Apache License 2.0",
					URL:  "http://www.apache.org/licenses/LICENSE-2.0",
				},
			},
			Version: "1.0.0",
		},
	}
	swo.Tags = []spec.Tag{{TagProps: spec.TagProps{
		Name:        "Installer",
		Description: "Terminus Installer"}}}
}
