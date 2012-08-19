package main

import (
	"launchpad.net/goyaml"
)

type Config struct {
	Buildpack string
}

func Parse(data []byte) (c *Config, err error) {
	c = &Config{}
	err = goyaml.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}
	return c, nil
}
