package controller

import (
	"errors"
	"github.com/docker/docker/client"
	"github.com/emicklei/go-restful"
	"github.com/quan930/ControlTower/builder/pkg/docker"
	"github.com/quan930/ControlTower/builder/pkg/git"
	"hook/internal/pojo"
	"hook/internal/util/file"
	"hook/internal/util/log"
	"hook/internal/util/slack"
	"hook/internal/util/yaml"
	"k8s.io/klog/v2"
	"os"
	"os/exec"
	"strings"
)

type Controller struct {
}

func NewController() *Controller {
	log.Info.Println("k8sClientService init")
	return &Controller{}
}

func (c Controller) GithubHook(request *restful.Request, response *restful.Response) {
	pushPayload := new(pojo.PushPayload)
	err := request.ReadEntity(&pushPayload)
	if err != nil {
		log.Info.Println(err)
		response.WriteEntity(pojo.NewResponse(500, "read entity error", nil).Body)
	} else {
		shortCommitId := pushPayload.After[0:7]
		branch := pushPayload.Ref[strings.LastIndex(pushPayload.Ref, "/")+1:]
		log.Info.Println(pushPayload.Repository.URL)
		log.Info.Println(shortCommitId)
		log.Info.Println(branch)
		if pushPayload.Repository.URL == yaml.GetConfig().Hook.GithubRepo {
			for _, s := range yaml.GetConfig().Hook.Branches {
				if s == branch {
					log.Info.Println("need to push")
					// git clone
					err = gitCloneRepo("https://github.com/LiangquanLi930/hypershift-hook-ci", "buildimage")
					if err != nil {
						log.Warning.Println(err)
						response.WriteEntity(pojo.NewResponse(500, "git repo error", nil).Body)
						slack.SendSlack("error! git clone: " + err.Error())
						return
					}
					// build image
					err = buildAndPushImage(yaml.GetConfig().Hook.ImageRepo + ":" + shortCommitId)
					if err != nil {
						log.Warning.Println(err)
						response.WriteEntity(pojo.NewResponse(500, "build image error", nil).Body)
						slack.SendSlack("build image error! " + pushPayload.Repository.URL + " new push, branch:" + branch + " " + yaml.GetConfig().Hook.ImageRepo + ":" + shortCommitId)
						return
					}
					//验证
					err = verifyImage(yaml.GetConfig().Hook.ImageRepo + ":" + shortCommitId)
					if err != nil {
						log.Warning.Println(err)
						response.WriteEntity(pojo.NewResponse(500, "get dockerClient error", nil).Body)
						slack.SendSlack("verifyImage error! " + pushPayload.Repository.URL + " new push, branch:" + branch + " " + yaml.GetConfig().Hook.ImageRepo + ":" + shortCommitId)
						return
					}
					//build image
					err = buildAndPushImage(yaml.GetConfig().Hook.ImageRepo + ":latest")
					if err != nil {
						log.Warning.Println(err)
						response.WriteEntity(pojo.NewResponse(500, "build image error", nil).Body)
						slack.SendSlack("build image error! " + pushPayload.Repository.URL + " new push, branch:" + branch + " " + yaml.GetConfig().Hook.ImageRepo + ":" + shortCommitId)
						return
					} else {
						response.WriteEntity(pojo.NewResponse(200, "successful", nil).Body)
						slack.SendSlack("successful! " + pushPayload.Repository.URL + " new push, branch:" + branch + " " + yaml.GetConfig().Hook.ImageRepo + ":" + shortCommitId)
						return
					}
				}
			}
		}
		response.WriteEntity(pojo.NewResponse(200, "successful,not need push", nil).Body)
	}
}

func verifyImage(image string) error {
	cmd := exec.Command("sh", "-c", "rm -rf tmp && mkdir tmp && oc image extract "+image+" --path /hypershift:tmp")
	out, err := cmd.CombinedOutput()
	trimmed := strings.TrimSpace(string(out))
	klog.Info(trimmed)
	if err != nil {
		klog.Warning(trimmed)
	}
	if file.Exists("/run/tmp/hypershift") {
		return nil
	} else {
		return errors.New("file 'hypershift' not exist")
	}
}

func gitCloneRepo(url string, branch string) error {
	path := "./temp"
	err := os.RemoveAll(path)
	if err != nil {
		log.Warning.Println(err)
	}
	//git clone
	repo, err := git.Clone(url, "./temp")
	if err != nil {
		log.Warning.Println(err)
		return err
	}
	//git checkout
	err = git.Checkout(repo, branch)
	if err != nil {
		log.Warning.Println(err)
		return err
	}
	return nil
}

// yaml.GetConfig().Hook.ImageRepo+":"+shortCommitId)
func buildAndPushImage(imageName string) error {
	//get dockerClient
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Warning.Println(err)
		return err
	}
	//docker build
	err = docker.BuildImage(cli, "Dockerfile", "./temp", imageName)
	if err != nil {
		log.Warning.Println(err)
		return err
	}
	//docker push
	docker.PushImage(cli, yaml.GetConfig().Quay.User, yaml.GetConfig().Quay.Password, imageName)
	if err != nil {
		log.Warning.Println(err)
		return err
	}
	return nil

}
