package controller

import (
	"github.com/emicklei/go-restful/log"
	"time"
	"github.com/coreos/etcd/client"
)

type JobController struct {
	client   client.Client
	interval time.Duration
}

func New(interval time.Duration, client client.Client) *JobController {
	return &JobController{
		client: client,
		interval: interval,
	}
}

func (this *JobController) Run(){
	log.Printf("Running JobController...")
}

func (this *JobController) List() error {

	return nil
}

func (this *JobController) runJob() error {

	return nil
}

func (this *JobController) deleteJob() error {

	return nil
}

func (this *JobController) updateJob() error {

	return nil
}





