package config

import (
	"encoding/json"
	"os"

	"github.com/akundu/utilities/logger"
)

type Config interface{}

func LoadConfiguration(config_file string, config_obj Config) error {
	logger.Trace.Println("reading from file ", config_file)
	config_file_fd, err := os.Open(config_file)
	if err != nil || config_file_fd == nil {
		logger.Error.Println("got error", err, " while opening file", config_file)
		return err
	}
	defer func() {
		_ = config_file_fd.Close()
	}()

	jsonParser := json.NewDecoder(config_file_fd)
	//var config Config
	//if err := jsonParser.Decode(&config); err != nil {
	if err := jsonParser.Decode(config_obj); err != nil {
		return err
	}
	logger.Trace.Println("config = ", config_obj)
	return nil
}
