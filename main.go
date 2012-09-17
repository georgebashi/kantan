package main

import (
	"code.google.com/p/gorilla/mux"
	"net/http"
	"os"
)

type globalContext struct {
	kantan_root string
}

type requestContext struct {
	globalContext
	vars    map[string]string
	Project Project
}

type requestHandler struct {
	globalContext
	f func(requestContext, http.ResponseWriter, *http.Request)
}

func (handler *requestHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	rCtx := requestContext{
		globalContext: globalContext{
			kantan_root: handler.globalContext.kantan_root,
		},
		vars:    vars,
		Project: NewProject(handler.globalContext.kantan_root, vars["project"]),
	}
	handler.f(rCtx, w, req)
}

func main() {
	kantan_root, _ := os.Getwd()
	if env_kantan_root := os.Getenv("KANTAN_ROOT"); env_kantan_root != "" {
		kantan_root = env_kantan_root
	}
	r := mux.NewRouter()

	ctx := globalContext{kantan_root}

	r.Path("/projects/{project}/repo/git-receive-pack").Handler(&requestHandler{ctx, requestContext.projRepoReceivePackHandler})
	r.PathPrefix("/projects/{project}/repo").Handler(&requestHandler{ctx, requestContext.projRepoHandler})

	http.ListenAndServe(":9090", r)
}
