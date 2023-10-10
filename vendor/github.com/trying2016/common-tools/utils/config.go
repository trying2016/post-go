package utils

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// LoadConfigFromFile from file
func LoadConfigFromFile(filepath string, cfg interface{}) error {
	if confContent, err := ioutil.ReadFile(filepath); err != nil {
		return err
	} else if err := yaml.Unmarshal([]byte(confContent), cfg); err != nil {
		return fmt.Errorf("Parser %s error. %v", filepath, err)
	}
	return nil
}
