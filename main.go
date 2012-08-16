package main

import (
	"code.google.com/p/gorilla/mux"
	"fmt"
	"net/http"
	"net/http/cgi"
)

func projHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	git := &cgi.Handler{
		Path: "/usr/local/Cellar/git/1.7.11.2/libexec/git-core/git-http-backend",
		Env: []string{
			fmt.Sprintf("GIT_PROJECT_ROOT=/tmp/derploy/%s", vars["project"]),
			"GIT_HTTP_EXPORT_ALL=true",
		},
	}
	http.StripPrefix(fmt.Sprintf("/proj/%s", vars["project"]), git).ServeHTTP(w, req)

}

func main() {
	r := mux.NewRouter()
	r.PathPrefix("/proj/{project}").HandlerFunc(projHandler)

	http.ListenAndServe(":9090", r)
}
