package RTJobRunner

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	"github.com/akundu/utilities/logger"
)

type SubstituteData struct {
	NumToGenerate int    `json: "numToGenerate,omitempty"`
	Type          string `json: "type,omitempty"` //can be "int|string"
}

type JSONFields struct {
	AdditionalStatements []string          `json:"additionalStatements,omitempty"`
	ExtraFields          map[string]string `json:"extraFields,omitempty"`
}
type JSONCommands struct {
	CommandToExecute string                     `json:"commandToExecute"`
	Substitutes      map[string]*SubstituteData `json:"substitutes,omitempty"`
}
type JSONJob struct {
	JSONCommands
	JSONFields
}

type JSONJobContainer struct {
	PostJobs      []*JSONJobContainer `json: "postJobs"`
	PreJobs       []*JSONJobContainer `json: "preJobs"`
	Job           *JSONJob                `json: "job"`
	Name          string                `json: "name"`
	NumIterations int                   `json: "numIterations"`
	Attributes    map[string]string     `json: "attributes"`
}

func (this JSONJobContainer) GetPostJobs() []*JSONJobContainer {
	return this.PostJobs
}
func (this JSONJobContainer) GetPreJobs() []*JSONJobContainer {
	return this.PreJobs
}
func (this JSONJobContainer) GetJob() *JSONJob {
	return this.Job
}
func (this JSONJobContainer) GetName() string {
	if len(this.Name) > 0 {
		return this.Name
	}
	if len(this.Job.CommandToExecute) > 0 {
		return this.Job.CommandToExecute
	}
	return ""
}

func CreateJHJSONParserString() *JSONJobContainer {
	return &JSONJobContainer{}
}

func processJobsFromJSON(jhjp *JSONJobContainer, jh *JobHandler, num_to_run_simultaneously int) error {
	//create a new job
	var job_tracker_pre sync.WaitGroup
	for index := range jhjp.GetPreJobs() {
		job := jhjp.GetPreJobs()[index]
		job_tracker_pre.Add(1)
		go func() {
			processJobsFromJSON(job, jh, num_to_run_simultaneously)
			job_tracker_pre.Done()
		}()
	}
	job_tracker_pre.Wait()

	//see if there were any errors from the pre jobs - because if there were, we wont proceed
	if jh.err != nil {
		return jh.err
	}



	//Run the current job
	if jhjp.NumIterations == 0 {
		jhjp.NumIterations = 1
	}
	var got_error error = nil
	json_jobs := NewJobHandler(num_to_run_simultaneously, jh.create_worker_func, jh.print_results)
	for i := 0; i < jhjp.NumIterations; i++ {
		if(len(jhjp.Job.Substitutes) == 0) {
			json_jobs.AddJob(NewRTRequestResultObject(&JSONJobProcessor{
				Name: jhjp.GetName(),
				CommandToExecute: jhjp.Job.CommandToExecute,
				JSONFields: jhjp.Job.JSONFields,
			}))
		} else { //we have to expand the substitutes
			if err := add_jobs(jhjp, json_jobs); err != nil {
				got_error = err
			}
		}
	}
	json_jobs.DoneAddingJobs()
	json_jobs.WaitForJobsToComplete()
	for i := range json_jobs.Jobs {
		json_job_result := json_jobs.Jobs[i]
		jh.appendResults(json_job_result)
	}
	if got_error != nil {
		return got_error
	} else if jh.err != nil {
		return jh.err
	}



	//Run the post jobs
	var job_tracker_post sync.WaitGroup
	for index := range jhjp.GetPostJobs() {
		job := jhjp.GetPostJobs()[index]
		job_tracker_post.Add(1)
		go func() {
			processJobsFromJSON(job, jh, num_to_run_simultaneously)
			job_tracker_post.Done()
		}()
	}
	job_tracker_post.Wait()

	return jh.err
}

func ProcessJobsFromJSON(filename string, jh *JobHandler, num_to_run_simultaneously int) error {
	file_data, err := ioutil.ReadFile(filename)
	if err != nil {
		logger.Error.Print(err)
		return err
	}

	obj_to_use := CreateJHJSONParserString()
	if err := json.Unmarshal(file_data, obj_to_use); err != nil {
		return err
	}
	if err = processJobsFromJSON(obj_to_use, jh, num_to_run_simultaneously); err != nil {
		return err
	}

	return nil
}

