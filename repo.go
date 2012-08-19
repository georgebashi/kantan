package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"time"
)

func git_exec(dir string, args ...string) (out []byte, err error) {
	path, _ := exec.LookPath("git")
	cmd := exec.Command(path, args...)
	cmd.Dir = dir
	return cmd.Output()
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (ctx requestContext) projRepoHandler(w http.ResponseWriter, req *http.Request) {
	git := CreateGitHandler(ctx.Project.Repo)

	// test if repo exists
	exists, err := exists(ctx.Project.Repo)
	if err != nil {
		http.Error(w, "Couldn't read project repo dir", http.StatusInternalServerError)
		return
	}

	if !exists {
		err := os.MkdirAll(ctx.Project.Repo, 0770)
		if err != nil {
			http.Error(w, "Couldn't create project repo dir", http.StatusInternalServerError)
			return
		}

		if _, err = git_exec(ctx.Project.Repo, "init", "--bare"); err != nil {
			http.Error(w, "Couldn't init project repo", http.StatusInternalServerError)
			return
		}
		if _, err = git_exec(ctx.Project.Repo, "config", "http.receivepack", "true"); err != nil {
			http.Error(w, "Couldn't configure http.receivepack on project repo", http.StatusInternalServerError)
			return
		}
	}

	http.StripPrefix(fmt.Sprintf("/projects/%s/repo", ctx.vars["project"]), git).ServeHTTP(w, req)
}

func gitURL(path string) (*url.URL, error) {
	rx := regexp.MustCompile("([\\-\\.\\w]+)@([\\-\\.\\w]+):([\\-\\.\\w]+)")
	if parts := rx.FindStringSubmatch(path); parts != nil {
		path = fmt.Sprintf("ssh://%s@%s/%s", parts[1], parts[2], parts[3])
	}
	return url.Parse(path)
}

func urlToDir(url *url.URL) (string, error) {
	rx := regexp.MustCompile(".*/([\\-\\.\\w]+?)(.git)?$")
	if parts := rx.FindStringSubmatch(url.RequestURI()); parts != nil {
		return parts[1], nil
	}
	return "", errors.New("Couldn't parse repository name from URL")
}

func (ctx requestContext) projRepoReceivePackHandler(w http.ResponseWriter, req *http.Request) {
	git := CreateGitHandler(ctx.Project.Repo)

	wrapper := &GitIgnoreFlushWriter{w}
	http.StripPrefix(fmt.Sprintf("/projects/%s/repo", ctx.vars["project"]), git).ServeHTTP(wrapper, req)
	defer fmt.Fprintln(w, "0000")

	gow := &GitOutputWriter{w}
	major := NewHerokuStyleLogger(gow, true)
	minor := NewHerokuStyleLogger(gow, false)
	major.Println("Derploy receiving push")

	yml, err := git_exec(ctx.Project.Repo, "cat-file", "blob", "master:.derploy.yml")
	if err != nil {
		major.Println("Couldn't read derploy config")
		minor.Println("Create .derploy.yml in the repository, containing \"buildpack: git@uri:for/buildpack\"")
		return
	}

	c, err := Parse(yml)
	if err != nil {
		major.Println("Couldn't parse deploy config")
		return
	}

	bpUrl, err := gitURL(c.Buildpack)
	if err != nil {
		major.Println("Couldn't parse buildpack URL")
		return
	}

	bpDir, err := urlToDir(bpUrl)
	if err != nil {
		major.Println(err)
		return
	}

	buildpackPath := fmt.Sprintf("%s/buildpacks/%s", ctx.globalContext.derp_root, bpDir)

	exists, err := exists(buildpackPath)
	if err != nil {
		major.Printf("Couldn't read buildpack dir (%s)", buildpackPath)
		return
	}

	if !exists {
		major.Printf("Fetching buildpack \"%s\" from %s", bpDir, c.Buildpack)
		err = os.MkdirAll(buildpackPath, 0770)
		if _, err = git_exec(buildpackPath, "clone", c.Buildpack, buildpackPath); err != nil {
			major.Printf("Couldn't clone buildpack %s into %s", c.Buildpack, buildpackPath)
			return
		}
		minor.Println("Cleaning buildpack cache")
		if os.RemoveAll(ctx.Project.Cache) != nil {
			minor.Printf("Couldn't remove old cache, leaving in place")
		}
	} else {
		minor.Printf("Using existing buildpack \"%s\"", bpDir)
		git_exec(buildpackPath, "clean", "-fd")
	}

	if os.MkdirAll(ctx.Project.Cache, 0770) != nil {
		major.Println("Couldn't create cache dir")
		return
	}

	releaseId := time.Now().Format("20060102150405")
	releaseDir := fmt.Sprintf("%s/releases/%s", ctx.Project.Path, releaseId)
	minor.Printf("Preparing new release %s", releaseId)
	if os.MkdirAll(releaseDir, 0770) != nil {
		major.Printf("Couldn't create release dir %s", releaseDir)
		return
	}

	if _, err = git_exec(releaseDir, "clone", "--local", "--no-hardlinks", "--depth", "1", "--recursive", ctx.Project.Repo, "."); err != nil {
		major.Printf("Couldn't export HEAD to release dir")
		return
	}

	compileCmd := exec.Command(fmt.Sprintf("%s/bin/compile", buildpackPath), releaseDir, ctx.Project.Cache)
	compileCmd.Dir = ctx.Project.Cache
	outReader, err := compileCmd.StdoutPipe()
	if err != nil {
		major.Println("Couldn't get stdout of compile script")
		return
	}

	errReader, err := compileCmd.StderrPipe()
	if err != nil {
		major.Println("Couldn't get stderr of compile script")
		return
	}

	go PipeToGitWriter(outReader, *gow)
	go PipeToGitWriter(errReader, *gow)
	compileCmd.Run()

	major.Println("Done")

}
