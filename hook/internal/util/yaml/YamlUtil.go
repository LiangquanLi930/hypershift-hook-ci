package yaml

import (
	"gopkg.in/yaml.v2"
	"hook/internal/util/err"
	"io/ioutil"
)

var conf config

func GetConfig() config {
	return conf
}

type config struct {
	Hook struct{
		GithubRepo string `yaml:"github_repo"`
		ImageRepo string `yaml:"image_repo"`
		Branches []string `yaml:"branches"`
	} `yaml:"hook"`
	Quay struct{
		User string `yaml:"user"`
		Password string `yaml:"password"`
	} `yaml:"quay"`
}

func Init(filePath string) {
	var erro error
	yamlFile, erro := ioutil.ReadFile(filePath)
	err.GetErr("无法读取客户端文件", erro)
	erro = yaml.Unmarshal(yamlFile, &conf)
	err.GetErr("解析yaml错误", erro)
}
