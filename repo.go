package main

import (
	"bytes"
	"errors"
	"fmt"
	"launchpad.net/goyaml"
	"log"
	"net/http"
	"net/http/cgi"
	"net/url"
	"os"
	"os/exec"
	"regexp"
)

func createGitHandler(repoPath string) *cgi.Handler {
	return &cgi.Handler{
		Path: "/usr/local/Cellar/git/1.7.11.5/libexec/git-core/git-http-backend",
		Env: []string{
			fmt.Sprintf("GIT_PROJECT_ROOT=%s", repoPath),
			"GIT_HTTP_EXPORT_ALL=true",
		},
	}
}

func git_exec(dir string, args ...string) (out []byte, err error) {
	cmd := exec.Command("/usr/local/bin/git", args...)
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
	git := createGitHandler(ctx.repoPath)

	// test if repo exists
	exists, err := exists(ctx.repoPath)
	if err != nil {
		http.Error(w, "Couldn't read project repo dir", http.StatusInternalServerError)
		return
	}

	if !exists {
		err := os.MkdirAll(ctx.repoPath, 0770)
		if err != nil {
			http.Error(w, "Couldn't create project repo dir", http.StatusInternalServerError)
			return
		}

		if _, err = git_exec(ctx.repoPath, "init", "--bare"); err != nil {
			http.Error(w, "Couldn't init project repo", http.StatusInternalServerError)
			return
		}
		if _, err = git_exec(ctx.repoPath, "config", "http.receivepack", "true"); err != nil {
			http.Error(w, "Couldn't configure http.receivepack on project repo", http.StatusInternalServerError)
			return
		}
	}

	http.StripPrefix(fmt.Sprintf("/projects/%s/repo", ctx.vars["project"]), git).ServeHTTP(w, req)
}

type writerWrapper struct {
	http.ResponseWriter
}

func (wrapper *writerWrapper) Write(data []byte) (int, error) {
	if bytes.Equal(data, []byte("0000")) {
		return 4, nil
	}
	return wrapper.ResponseWriter.Write(data)
}

type GitOutputWriter struct {
	w http.ResponseWriter
}

func (gow *GitOutputWriter) Write(p []byte) (n int, err error) {
	fmt.Fprintf(gow.w, "%04x\x02%s", len(p)+5, p)
	if ok := gow.w.(http.Flusher); ok != nil {
		ok.Flush()
	}
	return len(p), nil
}

func newHerokuStyleLogger(gow *GitOutputWriter, major bool) *log.Logger {
	prefix := "       "
	if major {
		prefix = "-----> "
	}
	return log.New(gow, prefix, 0)
}

type config struct {
	Buildpack string
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
	git := createGitHandler(ctx.repoPath)

	wrapper := &writerWrapper{w}
	http.StripPrefix(fmt.Sprintf("/projects/%s/repo", ctx.vars["project"]), git).ServeHTTP(wrapper, req)
	defer fmt.Fprintln(w, "0000")

	gow := &GitOutputWriter{w}
	major := newHerokuStyleLogger(gow, true)
	minor := newHerokuStyleLogger(gow, false)
	major.Println("Derploy receiving push")

	// git cat-file blob master:.derploy.yml
	yml, err := git_exec(ctx.repoPath, "cat-file", "blob", "master:.derploy.yml")
	if err != nil {
		major.Println("Couldn't read derploy config")
		minor.Println("Create .derploy.yml in the repository, containing \"buildpack: git@uri:for/buildpack\"")
		return
	}

	c := config{}
	err = goyaml.Unmarshal(yml, &c)
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
	} else {
		major.Printf("Using existing buildpack \"%s\"", bpDir)
		git_exec(buildpackPath, "clean", "-fd")
	}

	cacheDir := fmt.Sprintf("%s/cache/%s/%s", ctx.globalContext.derp_root, bpDir, ctx.vars["project"])
	if os.MkdirAll(cacheDir, 0770) != nil {
		major.Println("Couldn't create cache dir")
		return
	}

}

