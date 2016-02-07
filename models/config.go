package models

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
)

type Config struct {
	FilePath     string
	Name         string
	PulpRootNode *Node `yaml:"pulp_root_node"`
}

func NewConfig() *Config {
	return &Config{
		FilePath: "./tree.yaml",
	}
}

func (c Config) GetStageTree() (stage_tree StageTree) {
	filename, _ := filepath.Abs(c.FilePath)
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(yamlFile, &stage_tree)
	if err != nil {
		panic(err)
	}
	return
}
