package main

import (
	"fmt"
)

type Project struct {
	Name  string
	Path  string
	Repo  string
	Cache string
}

func NewProject(root string, name string) Project {
	path := fmt.Sprintf("%s/projects/%s", root, name)
	return Project{
		Name:  name,
		Path:  path,
		Repo:  fmt.Sprintf("%s/repo", path),
		Cache: fmt.Sprintf("%s/cache", path),
	}
}
