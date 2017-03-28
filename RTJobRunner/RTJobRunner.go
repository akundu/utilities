package RTJobRunner

import (
	"bufio"
	"os"
	"strings"
	"sync"

	"github.com/akundu/utilities/logger"
)

type Response interface{}
type Request interface{}
type Worker func(int, <-chan Request, chan<- Response)

type JobHandler struct {
	jobs           chan Request
	results        chan Response
	ws_job_tracker sync.WaitGroup
	num_added      int
	done_adding    bool
	worker         Worker
}

func NewJobHandler(num_to_setup int, worker Worker, print_results bool) *JobHandler {
	jh := &JobHandler{
		jobs:        make(chan Request, num_to_setup),
		results:     make(chan Response, num_to_setup),
		num_added:   0,
		done_adding: false,
	}

	for w := 0; w < num_to_setup; w++ {
		go worker(w, jh.jobs, jh.results)
	}

	jh.ws_job_tracker.Add(1)
	go jh.waitForResults(print_results)

	return jh
}

func (this *JobHandler) AddJob(job Request) {
	this.jobs <- job
	this.num_added++
}

func (this *JobHandler) GetJobsFromStdin() {
	//read from stdin
	bio := bufio.NewReader(os.Stdin)
	for {
		line, err := bio.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.Trim(line, "\n \r\n")
		logger.Trace.Println("adding ", line)
		this.AddJob(line)
	}
}

func (this *JobHandler) WaitForJobsToComplete() {
	this.ws_job_tracker.Wait()
}

func (this *JobHandler) waitForResults(print_results bool) {
	num_processed := 0
	for this.done_adding == false || num_processed < this.num_added {
		result := <-this.results
		num_processed++
		if result != nil && print_results == true{
			logger.Info.Println(result)
		}
	}
	this.ws_job_tracker.Done()
	logger.Trace.Println("done processing results")
}

func (this *JobHandler) DoneAddingJobs() {
	close(this.jobs)
	this.done_adding = true
}
