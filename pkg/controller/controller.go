package controller

import (
	"github.com/emicklei/go-restful/log"
	"time"
	"github.com/coreos/etcd/client"
	"distributedScheduler/pkg/utils"
	"github.com/golang/glog"
)

type JobController struct {
	client   client.Client
	Interval time.Duration
	StopChan chan bool
}

func New(interval time.Duration, client client.Client) *JobController {
	stopChan := make(chan bool, 1)
	return &JobController{
		client: client,
		Interval: interval,
		StopChan: stopChan,
	}
}

func (this *JobController) Run(){
	go utils.RunPeriodically(2*time.Second, this.run)
	select {
		case <- this.StopChan:
			glog.Infof("Capture a stop signal...")
			return
	}
}

func (this *JobController) run(){
	log.Printf("Running JobController...")
}

func (this *JobController) list() error {

	return nil
}

func (this *JobController) execute() error {

	return nil
}

func (this *JobController) delete() error {

	return nil
}

func (this *JobController) update() error {

	return nil
}





