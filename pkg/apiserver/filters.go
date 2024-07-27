package apiserver

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/utils"

	"github.com/emicklei/go-restful/v3"
)

func logStackOnRecover(panicReason interface{}, w http.ResponseWriter) {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("recover from panic situation: - %v\r\n", panicReason))
	for i := 2; ; i += 1 {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
	}
	logger.Error(buffer.String())
}

func logRequestAndResponse(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	start := time.Now()
	chain.ProcessFilter(req, resp)

	// Always log error response
	logger.Infof("%s - \"%s %s %s\" %d %d %dms",
		utils.RemoteIp(req.Request),
		req.Request.Method,
		req.Request.URL,
		req.Request.Proto,
		resp.StatusCode(),
		resp.ContentLength(),
		time.Since(start)/time.Millisecond,
	)
}

func cors(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	resp.AddHeader("Access-Control-Allow-Origin", "*")

	resp.AddHeader("Content-Type", "application/json, application/x-www-form-urlencoded, text/html;charset=utf-8")
	resp.AddHeader("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	resp.AddHeader("Access-Control-Allow-Headers", "Accept, Content-Type, Accept-Encoding, X-Authorization")

	if req.Request.Method == "OPTIONS" {
		resp.WriteHeader(http.StatusOK)
		resp.Write([]byte("ok"))
		return
	}

	chain.ProcessFilter(req, resp)
}
