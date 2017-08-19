package RTJobRunner

import (
	"bufio"
	"os"
	"strings"
	"sync"
	"fmt"
	"sync/atomic"

	"github.com/akundu/utilities/logger"
    "github.com/satori/go.uuid"
	"time"
)

type JobHandler struct {
	req_chan            chan *JobInfo
	res_chan            chan *JobInfo

	ws_job_tracker      sync.WaitGroup
	num_added           int32
	num_run_simultaneously           int
	create_worker_func  CreateWorkerFunction
	print_results       bool

	done_channel        chan bool
	worker_list         []Worker
	id                  string
	err         		error

	Jobs             	[]*JobInfo
}

func NewJobHandler(num_to_setup int, createWorkerFunc CreateWorkerFunction, print_results bool) *JobHandler {
	jh := &JobHandler{
		req_chan:                make(chan *JobInfo, num_to_setup),
		res_chan:             make(chan *JobInfo, num_to_setup),
		num_run_simultaneously:           num_to_setup,
		num_added:           0,
		create_worker_func:  createWorkerFunc,
		worker_list:         make([]Worker, num_to_setup),
		done_channel:        make(chan bool, 1),
		id:                  fmt.Sprintf("%s", uuid.NewV4()),
		print_results:       print_results,
		err:         nil,
	}

	for w := 0; w < num_to_setup; w++ {
		worker := createWorkerFunc()
		jh.worker_list[w] = worker
		worker.PreRun()
		go worker.Run(w, jh.req_chan, jh.res_chan)
	}

	jh.ws_job_tracker.Add(1) //goroutine to wait for results
	go jh.waitForResults(print_results)

	return jh
}

func (this *JobHandler) AddJob(job *JobInfo) {
	job.job_start_time = time.Now()
	this.req_chan <- job
	atomic.AddInt32(&this.num_added, 1)
}

type JobHandlerLineOutputFilter func(string) Request //line - outputline - if outputline is empty - dont add anything
func (this *JobHandler) GetJobsFromStdin(jhlo JobHandlerLineOutputFilter) {
	//read from stdin
	bio := bufio.NewReader(os.Stdin)
	for {
		line, err := bio.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.Trim(line, "\n \r\n")
		logger.Trace.Println("adding ", line)
		if jhlo == nil {
			this.AddJob(NewRTRequestResultObject(&StringRequest{line}))
		} else {
			filtered_job := jhlo(line)
			if filtered_job != nil {
				this.AddJob(NewRTRequestResultObject(filtered_job))
			}
		}
	}

	//call the handler one last time - in case the filter wants to add anything else
	if jhlo != nil {
		filtered_job := jhlo("")
		if filtered_job != nil {
			this.AddJob(NewRTRequestResultObject(filtered_job))
		}
	}
}

func (this *JobHandler) WaitForJobsToComplete() {
	this.ws_job_tracker.Wait()
}

func (this *JobHandler) appendResults(r *JobInfo) {
	this.Jobs = append(this.Jobs, r)
	if r.Resp.GetError() != nil {
		this.err = r.Resp.GetError()
	}
}
func (this *JobHandler) waitForResults(print_results bool) {
	var num_processed int32 = 0
	done_adding := false
	for done_adding == false || num_processed < atomic.LoadInt32(&this.num_added) {
		select {
		case result := <-this.res_chan:
			num_processed++
			if(result == nil) {
				continue
			}
			result.job_end_time = time.Now()

			if print_results == true {
				logger.Info.Printf("%dus %s %v\n",
									int(result.JobTime()/1000),
									result.Req.GetId(),
									result.Resp)
			}
			this.appendResults(result)
		case done_adding = <-this.done_channel:
			continue
		}
	}
	logger.Trace.Println("done processing results")

	//clean up the workers if needed
	for i := range this.worker_list {
		this.worker_list[i].PostRun()
	}

	this.ws_job_tracker.Done()
}

func (this *JobHandler) DoneAddingJobs() {
	close(this.req_chan)
	if atomic.LoadInt32(&this.num_added) == 0 {
		close(this.res_chan)
	}
	this.done_channel <- true
}
