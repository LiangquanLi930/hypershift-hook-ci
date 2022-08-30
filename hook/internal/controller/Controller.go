package controller

import (
	"errors"
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/quan930/ControlTower/builder/pkg/git"
	"hook/internal/pojo"
	"hook/internal/util/file"
	"hook/internal/util/log"
	"hook/internal/util/slack"
	"hook/internal/util/yaml"
	"io"
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
					//docker login
					err = execute(exec.Command("sh", "-c", fmt.Sprintf(`docker login -u="%s" -p="%s" quay.io`, yaml.GetConfig().Quay.User, yaml.GetConfig().Quay.Password)))
					if err != nil {
						log.Warning.Println(err)
						response.WriteEntity(pojo.NewResponse(500, "docker login error", nil).Body)
						slack.SendSlack("build image error! " + pushPayload.Repository.URL + " new push, branch:" + branch + " " + yaml.GetConfig().Hook.ImageRepo + ":" + shortCommitId)
						return
					}
					// build image
					err = execute(exec.Command("sh", "-c", fmt.Sprintf(`docker buildx build --file temp/Dockerfile --no-cache --platform linux/amd64,linux/arm64,linux/s390x,linux/ppc64le -t %s --push .`, yaml.GetConfig().Hook.ImageRepo+":"+shortCommitId)))
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
					err = execute(exec.Command("sh", "-c", fmt.Sprintf(`docker buildx build --file temp/Dockerfile --platform linux/amd64,linux/arm64,linux/s390x,linux/ppc64le -t %s --push .`, yaml.GetConfig().Hook.ImageRepo+":latest")))
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

func asyncLog(reader io.ReadCloser) error {
	cache := "" //缓存不足一行的日志信息
	buf := make([]byte, 1024)
	for {
		num, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if num > 0 {
			b := buf[:num]
			s := strings.Split(string(b), "\n")
			line := strings.Join(s[:len(s)-1], "\n") //取出整行的日志
			fmt.Printf("%s%s", cache, line)
			cache = s[len(s)-1]
		}
	}
	return nil
}

func execute(cmd *exec.Cmd) error {
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		log.Warning.Printf("Error starting command: %s......\n", err.Error())
		return err
	}

	go asyncLog(stdout)
	go asyncLog(stderr)

	if err := cmd.Wait(); err != nil {
		log.Warning.Printf("Error waiting for command execution: %s......", err.Error())
		return err
	}

	return nil

}
