package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/cgi"
	"os"
	"os/exec"
)


type writerWrapper struct {
	http.ResponseWriter
}

func (wrapper *writerWrapper) Write(data []byte) (int, error) {
	if bytes.Equal(data, []byte("0000")) {
		return 4, nil
	}
	return wrapper.ResponseWriter.Write(data)
}

func createGitHandler(projectPath string) *cgi.Handler {
	return &cgi.Handler{
		Path: "/usr/local/Cellar/git/1.7.11.5/libexec/git-core/git-http-backend",
		Env: []string{
			fmt.Sprintf("GIT_PROJECT_ROOT=%s", projectPath),
			"GIT_HTTP_EXPORT_ALL=true",
		},
	}
}

func (ctx requestContext) projRepoHandler(w http.ResponseWriter, req *http.Request) {
	git := createGitHandler(ctx.projectPath)

	// test if projectPath exists
	if _, err := os.Stat(ctx.projectPath); err != nil {
		if os.IsNotExist(err) {
			// it doesn't, create it
			os.MkdirAll(ctx.projectPath, 0770)
			if git_exec(ctx.projectPath, "init") != nil {
				http.Error(w, "Couldn't init project repo", http.StatusInternalServerError)
				return
			}
			if git_exec(ctx.projectPath, "config", "http.receivepack", "true") != nil {
				http.Error(w, "Couldn't configure http.receivepack on project repo", http.StatusInternalServerError)
				return
			}
			if git_exec(ctx.projectPath, "config", "receive.denyCurrentBranch", "ignore") != nil {
				http.Error(w, "Couldn't configure receive.denyCurrentBranch on project repo", http.StatusInternalServerError)
				return
			}
		} else {
			// other error
			http.Error(w, "Project path was not readable", http.StatusInternalServerError)
			return
		}
	}
	http.StripPrefix(fmt.Sprintf("/proj/%s/repo", ctx.vars["project"]), git).ServeHTTP(w, req)
}

func (ctx requestContext) projRepoReceivePackHandler(w http.ResponseWriter, req *http.Request) {
	git := createGitHandler(ctx.projectPath)

	wrapper := &writerWrapper{w}
	http.StripPrefix(fmt.Sprintf("/proj/%s/repo", ctx.vars["project"]), git).ServeHTTP(wrapper, req)
	printf(w, "herp a derp")
	fmt.Fprintln(w, "0000")
}

func printf(w http.ResponseWriter, format string, args ...interface{}) {
	fmt.Fprintln(w, pktFmt(fmt.Sprintf(format, args...)))
	if ok := w.(http.Flusher); ok != nil {
		ok.Flush()
	}
}

func pktFmt(msg string) (out string) {
	return fmt.Sprintf("%04x\x02%s", len(msg)+6, msg)
}

func git_exec(dir string, args ...string) (err error) {
	cmd := exec.Command("/usr/local/bin/git", args...)
	cmd.Dir = dir
	return cmd.Run()
}
