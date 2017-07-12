package main

import (
	"net/http"
	"log"
	"distributedScheduler/routes"
	"distributedScheduler/utils"
	"distributedScheduler/controller"
	"time"
)

func main() {

	//Initialize etcd client
	client := utils.GetEtcdClient("http://127.0.0.1:2379")

	//Init and start scheduler
	jobController := controller.New(2*time.Second, client)
	go utils.RunPeriodically(2*time.Second, jobController.Run)

	//Init and start web front end
	service := &routes.JobRouter{}
	service.Initialize(client)
	service.Register()
	log.Println("Starting Http Server!")
	log.Fatal(http.ListenAndServe(":8080", nil))

}


