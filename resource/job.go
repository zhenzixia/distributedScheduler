package resource

const (
	RUNNING JobStatus = "running"
	FINISHED JobStatus = "finished"
	WAITING JobStatus = "waiting"
)

type JobStatus string

type Jobs []Job

type Job struct {
	Id	string		`json:"id"`
	Name 	string  	`json:"name"`
	//Completed bool 		`json:"completed"`
	Start	string	`json:"startat"`
	Status	JobStatus	`json:"status"`
}
