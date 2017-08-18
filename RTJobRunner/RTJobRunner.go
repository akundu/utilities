package RTJobRunner

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"fmt"
	"sync/atomic"

	"github.com/akundu/utilities/logger"
    "github.com/satori/go.uuid"
)

type JobHandler struct {
	jobs                chan Request
	results             chan Response
	ws_job_tracker      sync.WaitGroup
	num_added           int32

	done_channel        chan bool

	worker_list         []Worker
	id                  string

	Results             []interface{}
}

func NewJobHandler(num_to_setup int,
	createWorkerFunc CreateWorkerFunction,
	print_results bool,
	) *JobHandler {

	jh := &JobHandler{
		jobs:                make(chan Request, num_to_setup),
		results:             make(chan Response, num_to_setup),
		num_added:           0,
		worker_list:         make([]Worker, num_to_setup),
		done_channel:        make(chan bool, 1),
		id:                  fmt.Sprintf("%s", uuid.NewV4()),
	}

	for w := 0; w < num_to_setup; w++ {
		worker := createWorkerFunc()
		jh.worker_list[w] = worker
		worker.PreRun()
		go worker.Run(w, jh.jobs, jh.results)
	}

	jh.ws_job_tracker.Add(1) //goroutine to wait for results
	go jh.waitForResults(print_results)

	return jh
}

func (this *JobHandler) AddJob(job Request) {
	this.jobs <- job
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
			this.AddJob(line)
		} else {
			filtered_job := jhlo(line)
			if filtered_job != nil {
				this.AddJob(filtered_job)
			}
		}
	}

	//call the handler one last time - in case the filter wants to add anything else
	if jhlo != nil {
		filtered_job := jhlo("")
		if filtered_job != nil {
			this.AddJob(filtered_job)
		}
	}
}

//func (this *JobHandler) processJobsFromJSON(jhjp ParserObject) error {
func (this *JobHandler) processJobsFromJSON(jhjp *JHJSONParserString) error {

	var job_tracker_pre sync.WaitGroup
	for i := range jhjp.GetPreJobs() {
		job_tracker_pre.Add(1)
		job := jhjp.GetPreJobs()[i]
		go func() {
			this.processJobsFromJSON(job)
			job_tracker_pre.Done()
		}()
	}
	job_tracker_pre.Wait()


	if(jhjp.NumIterations == 0) {
		jhjp.NumIterations = 1
	}
	for i := 0 ; i < jhjp.NumIterations ; i++ {
		this.AddJob(jhjp)
	}

	var job_tracker_post sync.WaitGroup
	for i := range jhjp.GetPostJobs() {
		job_tracker_post.Add(1)
		job := jhjp.GetPostJobs()[i]
		go func() {
			this.processJobsFromJSON(job)
			job_tracker_post.Done()
		}()
	}
	job_tracker_post.Wait()

	return nil
}

//func (this *JobHandler) ProcessJobsFromJSON(filename string, parserObjectCreator CreateParserObjectFunc) error {
func (this *JobHandler) ProcessJobsFromJSON(filename string) error {
	file_data, err := ioutil.ReadFile(filename)
	if err != nil {
		logger.Error.Print(err)
		return err
	}

	/*
	if parserObjectCreator == nil {
		return utilities.NewBasicError("parserObjectCreator has to be provided")
	}
	*/

	//obj_to_use := parserObjectCreator()
	obj_to_use := CreateJHJSONParserString()
	if err := json.Unmarshal(file_data, obj_to_use); err != nil {
		return err
	}
	if err = this.processJobsFromJSON(obj_to_use); err != nil {
		return err
	}

	return nil
}

func (this *JobHandler) WaitForJobsToComplete() {
	this.ws_job_tracker.Wait()
}

func (this *JobHandler) waitForResults(print_results bool) {
	var num_processed int32 = 0
	done_adding := false
	for done_adding == false || num_processed < atomic.LoadInt32(&this.num_added) {
		select {
		case result := <-this.results:
			num_processed++
			if result != nil && print_results == true {
				logger.Info.Println(result)
			}
			this.Results = append(this.Results, result)
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
	close(this.jobs)
	if atomic.LoadInt32(&this.num_added) == 0 {
		close(this.results)
	}
	this.done_channel <- true
}
