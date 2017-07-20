package main

import (
	"distributedScheduler/pkg/utils"
	"distributedScheduler/pkg/controller"
	"time"
	"flag"
	"distributedScheduler/pkg/electionmanager"
	"github.com/golang/glog"
)

func main() {

	id := flag.String("lockid", "slave003", "Host name for slave node.")

	//Initialize etcd client
	client := utils.GetEtcdClient("http://127.0.0.1:2379")

	//Init and start scheduler
	jobController := controller.New(2*time.Second, client)
	//go utils.RunPeriodically(2*time.Second, jobController.Run)

	// Create an etcd client and a election manager.
	electionManager, _ := electionmanager.NewMaster(electionmanager.NewEtcdRegistry(), "schedulerlock", *id, 30)

	//1. Start a go routine to process events.
	go eventHandler(electionManager.EventsChan(), jobController)

	//2. Start the attempt to acquire the lock.
	electionManager.Start()

	select{}
}

func eventHandler(eventsCh <-chan electionmanager.MasterEvent, jobController *controller.JobController ) {
	for {
		glog.Infof("++++++++", len(eventsCh))
		select {
		case e := <-eventsCh:
			glog.Infof(" --- Capture an event: %+v ---", e)
			if e.Type == electionmanager.MasterAdded {
				glog.Infof("Observed a add event, starting jobController....")
				// Acquired the lock.
				jobController.Run()
			} else if e.Type == electionmanager.MasterDeleted {
				glog.Infof("Observed a deletion event, stopping jobController....")
				// Lost the lock.
				jobController.StopChan <- true
			} else {
				jobController.StopChan <- true
				// Lock ownership changed.
			}
		}
	}
}

