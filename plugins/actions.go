package main

import "fmt"

type PathAction struct {
	Path string `json:"path"`
	Action string `json:"action"`
}

var pathActions = []PathAction{
	{
		"POST:/api/datasets","create:dataset",
	},
	{
		"POST:/api/models","create:model",
	},
	{
		"POST:/api/jobs","create:job",
	},
	{
		"POST:/api/jobs/stop/{job_id}", "stop:job",
	},
	{
		"POST:/api/jobs/close/{job_id}", "close:job",
	},
	{
		"POST:/api/jobs/{job_type}", "run:job",
	},
	{
		"POST:/api/jobs/upload/test/{job_id}", "upload:test:job",
	},
}


func GetRequestAction(path string) (
	action string,
    err error,
) {
	for _, v := range(pathActions) {
		if v.Path == path {
            return v.Action, nil
        }
		err = fmt.Errorf("no action found for path %s", path)
		return "", err
	}
	return "", nil  // no action found for the given path
}