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

type JHJSONParserString struct {
	PostJobs      []*JHJSONParserString `json: "postJobs"`
	PreJobs       []*JHJSONParserString `json: "preJobs"`
	Job           *JSONJob                `json: "job"`
	Name          string                `json: "name"`
	NumIterations int                   `json: "numIterations"`
	Attributes    map[string]string     `json: "attributes"`
}

func (this JHJSONParserString) GetPostJobs() []*JHJSONParserString {
	return this.PostJobs
}
func (this JHJSONParserString) GetPreJobs() []*JHJSONParserString {
	return this.PreJobs
}
func (this JHJSONParserString) GetJob() *JSONJob {
	return this.Job
}
func (this JHJSONParserString) GetName() string {
	if len(this.Name) > 0 {
		return this.Name
	}
	if len(this.Job.CommandToExecute) > 0 {
		return this.Job.CommandToExecute
	}
	return ""
}

func CreateJHJSONParserString() *JHJSONParserString {
	return &JHJSONParserString{}
}

func processJobsFromJSON(jhjp *JHJSONParserString, jh *JobHandler) error {
	//create a new job
	var job_tracker_pre sync.WaitGroup
	for index := range jhjp.GetPreJobs() {
		job := jhjp.GetPreJobs()[index]
		job_tracker_pre.Add(1)
		go func() {
			processJobsFromJSON(job, jh)
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
	json_jobs := NewJobHandler(jh.num_run_simultaneously, jh.create_worker_func, jh.print_results)
	for i := 0; i < jhjp.NumIterations; i++ {
		json_jobs.AddJob(NewRTRequestResultObject(jhjp))
	}

	json_jobs.DoneAddingJobs()
	json_jobs.WaitForJobsToComplete()

	for i := range json_jobs.Jobs {
		json_job_result := json_jobs.Jobs[i]
		jh.appendResults(json_job_result)
	}

	if jh.err != nil {
		return jh.err
	}

	//Run the post jobs
	var job_tracker_post sync.WaitGroup
	for index := range jhjp.GetPostJobs() {
		job := jhjp.GetPostJobs()[index]
		job_tracker_post.Add(1)
		go func() {
			processJobsFromJSON(job, jh)
			job_tracker_post.Done()
		}()
	}
	job_tracker_post.Wait()

	return jh.err
}

func ProcessJobsFromJSON(filename string, jh *JobHandler) error {
	file_data, err := ioutil.ReadFile(filename)
	if err != nil {
		logger.Error.Print(err)
		return err
	}

	obj_to_use := CreateJHJSONParserString()
	if err := json.Unmarshal(file_data, obj_to_use); err != nil {
		return err
	}
	if err = processJobsFromJSON(obj_to_use, jh); err != nil {
		return err
	}

	return nil
}
