package v1

import (
	"fmt"
	"os"
	"strconv"

	"bytetrade.io/web3os/installer/pkg/api/response"
	"bytetrade.io/web3os/installer/pkg/common"
	corecommon "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/storage"
	"bytetrade.io/web3os/installer/pkg/model"
	"bytetrade.io/web3os/installer/pkg/phase/mock"
	"bytetrade.io/web3os/installer/pkg/pipelines"
	"github.com/emicklei/go-restful/v3"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	// apis.Base
	// appService *app_service.Client
	validate        *validator.Validate
	StorageProvider storage.Provider
}

func New(db storage.Provider) *Handler {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterValidation("kubeTypeValid", model.KubeTypeValid)
	return &Handler{
		validate:        v,
		StorageProvider: db,
	}
}

func (h *Handler) handlerInstall(req *restful.Request, resp *restful.Response) {
	logger.Infof("handler install req: %s", req.Request.Method)

	var reqModel model.InstallModelReq
	err := req.ReadEntity(&reqModel)
	if err != nil {
		response.HandleError(resp, err)
		return
	}

	if err = h.validate.Struct(&reqModel); err != nil {
		if validationErrors := err.(validator.ValidationErrors); validationErrors != nil {
			logger.Errorf("handler install request parameter invalid: %v", validationErrors)
			response.HandleError(resp, fmt.Errorf("handler install request parameter invalid"))
			return
		}
	}

	if reqModel.Config.DomainName == "" {
		reqModel.Config.DomainName = corecommon.DefaultDomainName
	}

	arg := common.Argument{
		KsEnable:         true,
		KsVersion:        common.DefaultKubeSphereVersion,
		Provider:         h.StorageProvider,
		Request:          reqModel,
		InstallPackages:  false,
		SKipPushImages:   false,
		ContainerManager: common.Containerd,
		RegistryMirrors:  GetEnv("REGISTRY_MIRRORS", reqModel.Config.RegistryMirrors),
		Proxy:            GetEnv("PROXY", reqModel.Config.Proxy),
	}

	switch reqModel.Config.KubeType {
	case common.K3s:
		arg.KubernetesVersion = common.DefaultK3sVersion
	case common.K8s:
		arg.KubernetesVersion = common.DefaultK8sVersion
	}

	if err := pipelines.InstallTerminusPipeline(arg); err != nil {
		response.HandleError(resp, err)
		return
	}

	response.SuccessNoData(resp)
}

func (h *Handler) handlerStatus(req *restful.Request, resp *restful.Response) {
	var timespan = req.QueryParameter("time")
	if timespan == "" {
		timespan = "0"
	}

	tspan, err := strconv.ParseInt(timespan, 10, 64)
	if err != nil {
		response.HandleError(resp, err)
		return
	}

	data, err := h.StorageProvider.QueryInstallState(tspan)
	if err != nil {
		response.HandleError(resp, err)
		return
	}

	var res = make(map[string]interface{})
	var msgs = make([]map[string]interface{}, 0)

	if data == nil || len(data) == 0 {
		response.HandleError(resp, fmt.Errorf("get status failed"))
		return
	}

	var last = data[len(data)-1]

	for _, d := range data {
		if d.Time.UnixMilli() == tspan {
			continue
		}
		var r = make(map[string]interface{})
		r["info"] = d.Message
		r["time"] = d.Time.UnixMilli()
		msgs = append(msgs, r)
	}
	// if msgs == nil {
	// 	msgs = make([]map[string]interface{}, 0)
	// }

	res["percent"] = fmt.Sprintf("%.2f%%", float64(float64(last.Percent)/100))
	res["status"] = last.State
	res["msg"] = msgs

	response.Success(resp, res)
}

func (h *Handler) handlerGreetings(req *restful.Request, resp *restful.Response) {
	logger.Infof("handler greetings req: %s", req.Request.Method)

	if err := mock.Greetings(); err != nil {
		logger.Errorf("greetings failed %v", err)
	}

	response.SuccessNoData(resp)
}

func (h *Handler) handlerInstallTerminus(req *restful.Request, resp *restful.Response) {
	// logger.Infof("handler installer req: %s", req.Request.Method)

	// arg := common.Argument{}
	// if err := pipelines.InstallTerminusPipeline(arg); err != nil {
	// 	fmt.Println("---api installer terminus / err---", err)
	// }

	response.SuccessNoData(resp)
}

func GetEnv(key string, arg string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return arg
}
