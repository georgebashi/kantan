package main

import (
	"bytes"
	"code.google.com/p/gorilla/mux"
	"fmt"
	"net/http"
	"net/http/cgi"
	"os"
	"os/exec"
)

type writerWrapper struct {
	w http.ResponseWriter
}

func (wrapper *writerWrapper) Header() http.Header {
	return wrapper.w.Header()
}

func (wrapper *writerWrapper) Write(data []byte) (int, error) {
	if bytes.Equal(data, []byte("0000")) {
		fmt.Println("IGNORE")
		return 4, nil
	}
	return wrapper.w.Write(data)
}

func (wrapper *writerWrapper) WriteHeader(status int) {
	fmt.Println("WRITEHEADER")
	wrapper.w.WriteHeader(status)
}

func projRepoHandler(w http.ResponseWriter, req *http.Request, derp_root string) {
	vars := mux.Vars(req)
	fmt.Println(req.RequestURI)
	projectPath := fmt.Sprintf("%s/projects/%s", derp_root, vars["project"])
	git := &cgi.Handler{
		Path: "/usr/local/Cellar/git/1.7.11.5/libexec/git-core/git-http-backend",
		Env: []string{
			fmt.Sprintf("GIT_PROJECT_ROOT=%s", projectPath),
			"GIT_HTTP_EXPORT_ALL=true",
		},
	}
	// test if projectPath exists
	if _, err := os.Stat(projectPath); err != nil {
		if os.IsNotExist(err) {
			// it doesn't, create it
			os.MkdirAll(projectPath, 0770)
			if git_exec(projectPath, "init") != nil {
				http.Error(w, "Couldn't init project repo", http.StatusInternalServerError)
				return
			}
			if git_exec(projectPath, "config", "http.receivepack", "true") != nil {
				http.Error(w, "Couldn't configure http.receivepack on project repo", http.StatusInternalServerError)
				return
			}
			if git_exec(projectPath, "config", "receive.denyCurrentBranch", "ignore") != nil {
				http.Error(w, "Couldn't configure receive.denyCurrentBranch on project repo", http.StatusInternalServerError)
				return
			}
		} else {
			// other error
			http.Error(w, "Project path was not readable", http.StatusInternalServerError)
			return
		}
	}
	http.StripPrefix(fmt.Sprintf("/proj/%s/repo", vars["project"]), git).ServeHTTP(w, req)
}

func projRepoReceivePackHandler(w http.ResponseWriter, req *http.Request, derp_root string) {
	vars := mux.Vars(req)
	fmt.Println("RECEIVE PACK")
	projectPath := fmt.Sprintf("%s/projects/%s", derp_root, vars["project"])
	git := &cgi.Handler{
		Path: "/usr/local/Cellar/git/1.7.11.5/libexec/git-core/git-http-backend",
		Env: []string{
			fmt.Sprintf("GIT_PROJECT_ROOT=%s", projectPath),
			"GIT_HTTP_EXPORT_ALL=true",
		},
	}
	wrapper := &writerWrapper{w: w}
	http.StripPrefix(fmt.Sprintf("/proj/%s/repo", vars["project"]), git).ServeHTTP(wrapper, req)
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

func main() {
	derp_root, _ := os.Getwd()
	if env_derp_root := os.Getenv("DERP_ROOT"); env_derp_root != "" {
		derp_root = env_derp_root
	}
	r := mux.NewRouter()

	r.Path("/proj/{project}/repo/git-receive-pack").HandlerFunc(func(w http.ResponseWriter, req *http.Request) { projRepoReceivePackHandler(w, req, derp_root) })
	r.PathPrefix("/proj/{project}/repo").HandlerFunc(func(w http.ResponseWriter, req *http.Request) { projRepoHandler(w, req, derp_root) })

	http.ListenAndServe(":9090", r)
}
