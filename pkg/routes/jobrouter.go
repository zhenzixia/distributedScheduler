package routes

import (
	"github.com/emicklei/go-restful"
	"log"
	"github.com/coreos/etcd/client"
	"net/http"
	"simpleWebTest/pkg/utils"
	"distributedScheduler/pkg/resource"
	"context"
	"k8s.io/kubernetes/staging/src/k8s.io/apimachinery/pkg/util/json"
)

type JobRouter struct {
	client client.Client
}

func init() {
	log.Println("Initializing....")
}

func (s *JobRouter) Initialize(client client.Client) {
	s.client = client
}

func (s *JobRouter) Register() {
	service := new(restful.WebService)
	service.Path("/")
	service.Consumes(restful.MIME_JSON)
	service.Produces(restful.MIME_JSON)

	service.Route(service.GET("job").To(s.GetAll))

	service.Route(service.GET("job/{id}").To(s.GetOne))

	service.Route(service.POST("").To(s.UpdateOne))
	//Param(service.QueryParameter("count", "Description...").DataType("int")) //this is to get url para like id=123

	service.Route(service.PUT("").To(s.CreateOne))

	service.Route(service.DELETE("/{id}").To(s.RemoveOne))

	restful.Add(service)
}

func (s *JobRouter) CreateOne(request *restful.Request, response *restful.Response) {
	log.Println("Trying to create one job...")

	job := resource.Job{
		Id : utils.GenerateUUID(),
	}

	err := request.ReadEntity(&job)

	log.Printf(">> Posting one record! Record: %+v ", job)

	if err == nil {
		//marshal job
		value, _ := json.Marshal(job)
		kapi := client.NewKeysAPI(s.client)
		resp, err := kapi.Set(context.Background(), "/job/"+job.Id, string(value), nil)
		if err != nil {
			log.Fatal(err)
		} else {
			// print common key info
			log.Printf("Set is done. Metadata is %q\n", resp)
		}
		response.WriteEntity(job)
	} else {
		response.WriteError(http.StatusInternalServerError,err)
	}
}

func (s *JobRouter) UpdateOne(request *restful.Request, response *restful.Response) {
	log.Println("Trying to update one job...")
	//count, _ := strconv.Atoi(request.QueryParameter("count"))
	job := resource.Job{
		Id : utils.GenerateUUID(),
	}

	err := request.ReadEntity(&job)

	log.Printf(">> Posting one record! Record: %+v ", job)

	if err == nil {
		//marshal job
		value, _ := json.Marshal(job)

		kapi := client.NewKeysAPI(s.client)
		resp, err := kapi.Set(context.Background(), "/job/"+job.Id, string(value), nil)
		if err != nil {
			log.Fatal(err)
		} else {
			// print common key info
			log.Printf("Set is done. Metadata is %q\n", resp)
		}
		response.WriteEntity(job)
	} else {
		response.WriteError(http.StatusInternalServerError,err)
	}
}


func (s *JobRouter) GetAll(request *restful.Request, response *restful.Response) {
	log.Println("Trying to list all the jobs...")
}

func (s *JobRouter) GetOne(request *restful.Request, response *restful.Response) {
	log.Println("Trying to get one job...")
	//Get job id from url
	id := request.PathParameter("id")
	log.Printf("Trying to get job: %v ....", id)

	kapi := client.NewKeysAPI(s.client)
	resp, err := kapi.Get(context.Background(), "job/"+id, nil)
	if err != nil {
		log.Printf("Error getting job: %s", err.Error())
		//log.Fatal(err)
	} else {
		// print common key info
		log.Printf("Get is done. Metadata is %q\n", resp)
		// print value
		log.Printf("%q key has %q value\n", resp.Node.Key, resp.Node.Value)
	}
}



func (s *JobRouter) RemoveOne(request *restful.Request, response *restful.Response) {
	log.Println("Trying to remove one job...")
}
