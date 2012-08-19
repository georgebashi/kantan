package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cgi"
	"os/exec"
)

func CreateGitHandler(repoPath string) *cgi.Handler {
	cmd, err := exec.LookPath("git")
	if err != nil {
		panic("Couldn't find git!")
	}

	return &cgi.Handler{
		Path: cmd,
		Args: []string{"http-backend"},
		Env:  []string{fmt.Sprintf("GIT_PROJECT_ROOT=%s", repoPath), "GIT_HTTP_EXPORT_ALL=true"},
	}
}

type GitIgnoreFlushWriter struct {
	http.ResponseWriter
}

func (gifw *GitIgnoreFlushWriter) Write(data []byte) (int, error) {
	if bytes.Equal(data, []byte("0000")) {
		return 4, nil
	}
	return gifw.ResponseWriter.Write(data)
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

func NewHerokuStyleLogger(gow *GitOutputWriter, major bool) *log.Logger {
	prefix := "       "
	if major {
		prefix = "-----> "
	}
	return log.New(gow, prefix, 0)
}

func PipeToGitWriter(rc io.Reader, gow GitOutputWriter) {
	br := bufio.NewReader(rc)
	for {
		line, isPrefix, err := br.ReadLine()
		if err != nil {
			return
		}

		if !isPrefix {
			line = append(line, "\n"...)
		}

		gow.Write(line)
	}
}
