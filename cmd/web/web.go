package main

import (
	"net/http"
	"log"
	"distributedScheduler/pkg/routes"
	"distributedScheduler/pkg/utils"
)

func main() {

	//Initialize etcd client
	client := utils.GetEtcdClient("http://127.0.0.1:2379")

	//Init and start web front end
	service := &routes.JobRouter{}
	service.Initialize(client)
	service.Register()
	log.Println("Starting Http Server!")
	log.Fatal(http.ListenAndServe(":8080", nil))

}


