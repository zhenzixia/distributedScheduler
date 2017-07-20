package main

import(
	"distributedScheduler/pkg/electionmanager"
	"github.com/golang/glog"
)

func main() {

	client := electionmanager.NewEtcdRegistry()
	resp, err := client.Create("aaa", "v1", 0)
	if err != nil {
		glog.Infof("--- err1 : %+v ---\n", err)
	}

	resp, err = client.Create("aaa", "v2", 0)
	if err != nil {
		glog.Infof("--- err2 : %+v ---\n", err)
	}


	glog.Infof("+++ %+v +++", resp)
}
