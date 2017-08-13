package RTJobRunner

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/akundu/utilities/logger"
)

type Response interface{}
type Request interface{}
type Worker interface {
	PreRun()
	PostRun()
	Run(id int, jobs <-chan Request, results chan<- Response)
}
type CreateWorkerFunction func() Worker

type JobHandler struct {
	jobs           chan Request
	results        chan Response
	ws_job_tracker sync.WaitGroup
	num_added      int
	done_adding    bool
	worker_list    []Worker
	Results        []interface{}
}

func NewJobHandler(num_to_setup int, createWorkerFunc CreateWorkerFunction, print_results bool) *JobHandler {
	jh := &JobHandler{
		jobs:        make(chan Request, num_to_setup),
		results:     make(chan Response, num_to_setup),
		num_added:   0,
		worker_list: make([]Worker, num_to_setup),
		done_adding: false,
	}

	for w := 0; w < num_to_setup; w++ {
		worker := createWorkerFunc()
		jh.worker_list[w] = worker
		worker.PreRun()
		go worker.Run(w, jh.jobs, jh.results)
	}

	jh.ws_job_tracker.Add(1)
	go jh.waitForResults(print_results)

	return jh
}

func (this *JobHandler) AddJob(job Request) {
	this.jobs <- job
	this.num_added++
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

type JHJsonParser struct {
	DependentJobs []*JHJsonParser `json: "dependentJobs"`
	Job           string          `json: "job"`
	Name          string          `json: "name"`
	RunType       string          `json: "runType"`
}

func (this *JobHandler) processJobsFromJSON(jhjp *JHJsonParser) error {
	var job_tracker sync.WaitGroup
	for i := range jhjp.DependentJobs {
		job_tracker.Add(1)
		job := jhjp.DependentJobs[i]
		go func() {
			this.processJobsFromJSON(job)
			job_tracker.Done()
		}()
	}
	job_tracker.Wait()

	this.AddJob(jhjp.Job)
	return nil
}

func (this *JobHandler) ProcessJobsFromJSON(filename string) error {
	file_data, err := ioutil.ReadFile(filename)
	if err != nil {
		logger.Error.Print(err)
		return err
	}

	var jhjp JHJsonParser
	if err := json.Unmarshal(file_data, &jhjp); err != nil {
		return err
	}

	if err = this.processJobsFromJSON(&jhjp); err != nil {
		return err
	}
	return nil
}

func (this *JobHandler) WaitForJobsToComplete() {
	this.ws_job_tracker.Wait()
}

func (this *JobHandler) waitForResults(print_results bool) {
	num_processed := 0
	for this.done_adding == false || num_processed < this.num_added {
		result := <-this.results
		num_processed++
		if result != nil && print_results == true {
			logger.Info.Println(result)
		}
		this.Results = append(this.Results, result)
	}
	this.ws_job_tracker.Done()
	logger.Trace.Println("done processing results")

	//clean up the workers if needed
	for i := range this.worker_list {
		this.worker_list[i].PostRun()
	}
}

func (this *JobHandler) DoneAddingJobs() {
	close(this.jobs)
	if this.num_added == 0 {
		close(this.results)
	}
	this.done_adding = true
}
