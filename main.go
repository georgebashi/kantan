package main

import (
	"code.google.com/p/gorilla/mux"
	"fmt"
	"net/http"
	"os"
)

type globalContext struct {
	derp_root string
}

type requestContext struct {
	globalContext
	vars map[string]string
	projectPath string
}

type requestHandler struct {
	globalContext
	f func(requestContext, http.ResponseWriter, *http.Request)
}

func (handler *requestHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	rCtx := requestContext{
		globalContext: globalContext{
			derp_root: handler.globalContext.derp_root,
		},
		vars: vars,
		projectPath: fmt.Sprintf("%s/projects/%s", handler.globalContext.derp_root, vars["project"]),
	}
	handler.f(rCtx, w, req)
}

func main() {
	derp_root, _ := os.Getwd()
	if env_derp_root := os.Getenv("DERP_ROOT"); env_derp_root != "" {
		derp_root = env_derp_root
	}
	r := mux.NewRouter()

	ctx := globalContext{derp_root}

	r.Path("/proj/{project}/repo/git-receive-pack").Handler(&requestHandler{ctx, requestContext.projRepoReceivePackHandler})
	r.PathPrefix("/proj/{project}/repo").Handler(&requestHandler{ctx, requestContext.projRepoHandler})

	http.ListenAndServe(":9090", r)
}
