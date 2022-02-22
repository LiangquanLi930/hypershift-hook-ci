package config

import (
	"github.com/emicklei/go-restful"
	"hook/internal/controller"
	"hook/internal/util/log"
	"strings"
)

func Register(container *restful.Container) {

	ws := new(restful.WebService)
	githubCon := controller.NewController()
	cors := restful.CrossOriginResourceSharing{
		ExposeHeaders:  []string{"X-My-Header"},
		AllowedHeaders: []string{"Content-Type", "Accept"},
		AllowedMethods: []string{"GET", "POST", "DELETE", "PATCH"},
		CookiesAllowed: false,
		Container:      container}

	//容器过滤器 跨域
	container.Filter(cors.Filter)
	//容器过滤器 跨域 配置OPTIONS 请求
	container.Filter(container.OPTIONSFilter)
	ws.
		Path("/api/v1/hook").
		Consumes(restful.MIME_XML, restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML) // you can specify this per route as well

	// WebService过滤器 日志
	ws.Filter(NCSACommonLogFormatLogger())

	ws.Route(ws.POST("/github").To(githubCon.GithubHook))

	container.Add(ws)
}

func NCSACommonLogFormatLogger() restful.FilterFunction {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		var username = "-"
		if req.Request.URL.User != nil {
			if name := req.Request.URL.User.Username(); name != "" {
				username = name
			}
		}
		chain.ProcessFilter(req, resp)
		log.Info.Println(strings.Split(req.Request.RemoteAddr, ":")[0], " - ",
			username, " \"",
			req.Request.Method, " ",
			req.Request.URL.RequestURI(), " ",
			req.Request.Proto, "\" ",
			resp.StatusCode(), " ",
			resp.ContentLength(),
		)
	}
}
