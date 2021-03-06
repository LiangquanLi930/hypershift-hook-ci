package main

import (
	"github.com/emicklei/go-restful"
	"hook/internal/config"
	"hook/internal/util/log"
	"hook/internal/util/yaml"
	"net/http"
)

func init() {
	//log.Config(log.Stdout, log.Stdout, log.Stdout|log.EnableFile, log.Stderr|log.EnableFile,"../../error.log")
	//yaml.Init("/Users/redhat/GolandProjects/hypershift-hook-ci/config.yaml")
	log.Config(log.Stdout, log.Stdout, log.Stdout|log.EnableFile, log.Stderr|log.EnableFile, "./error.log")
	yaml.Init("./config.yaml")
}

func main() {
	log.Info.Println(yaml.GetConfig())
	wsContainer := restful.NewContainer()
	wsContainer.Router(restful.CurlyRouter{})
	//Register
	config.Register(wsContainer)
	log.Info.Println("start listening on localhost: 8080")
	server := &http.Server{Addr: ":" + "8080", Handler: wsContainer}
	log.Error.Fatal(server.ListenAndServe())
}
