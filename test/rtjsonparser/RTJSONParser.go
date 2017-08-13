package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/akundu/utilities/RTJobRunner"
	"github.com/akundu/utilities/logger"
)

type worker struct {
}

func CreateWorker() RTJobRunner.Worker {
	return &worker{}
}

func (this *worker) PostRun() {
}
func (this *worker) PreRun() {
}
func (this *worker) Run(id int, jobs <-chan RTJobRunner.Request, results chan<- RTJobRunner.Response) {
	for j := range jobs {
		job, ok := j.(string)
		if ok == false {
			results <- nil
			continue
		}
		results <- job
	}
}

func main() {
	jh := RTJobRunner.NewJobHandler(*num_simultaneously_to_run_ptr, CreateWorker, false)
	if err := jh.ProcessJobsFromJSON(*json_file_ptr); err != nil {
		logger.Error.Println(err)
		return
	}
	//mark that there are no more jobs to add
	jh.DoneAddingJobs()
	//wait for the results
	jh.WaitForJobsToComplete()

	//print results
	for i := range jh.Results {
		if r, ok := jh.Results[i].(string); ok {
			fmt.Println(r)
		}
	}
}

var (
	num_simultaneously_to_run_ptr = flag.Int("p", 1, "num times to run action")
	print_results_ptr             = flag.Int("v", 0, "print results")
	json_file_ptr                 = flag.String("f", "", "json file to load jobs from")
)

func init() {
	logger.DefaultLoggerInit()
	logger.Init(ioutil.Discard, os.Stdout, ioutil.Discard, os.Stderr)
	runtime.GOMAXPROCS(runtime.NumCPU() - 1)

	flag.Parse()
}
