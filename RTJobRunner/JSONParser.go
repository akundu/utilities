package RTJobRunner

import (
	"sync"
	"io/ioutil"
	"encoding/json"
	"github.com/akundu/utilities/logger"
)

type ParserObject interface {
	GetPostJobs() []ParserObject
	GetJob() Request
	GetName() string
}
type CreateParserObjectFunc func() ParserObject





type JHJSONParserString struct {
	PostJobs []*JHJSONParserString 	`json: "postJobs"`
	PreJobs []*JHJSONParserString 	`json: "preJobs"`
	Job           string          	`json: "job"`
	Name          string          	`json: "name"`
	NumIterations int          		`json: "numIterations"`
	Attributes    map[string]string `json: "attributes"`
}

func (this JHJSONParserString) GetPostJobs() []*JHJSONParserString {
	return this.PostJobs
}
func (this JHJSONParserString) GetPreJobs() []*JHJSONParserString {
	return this.PreJobs
}
func (this JHJSONParserString) GetJob() Request {
	return this.Job
}
func (this JHJSONParserString) GetName() string {
	return this.Name
}

func CreateJHJSONParserString() *JHJSONParserString {
	return &JHJSONParserString{}
}




//func (this *JobHandler) processJobsFromJSON(jhjp ParserObject) error {
func processJobsFromJSON(jhjp *JHJSONParserString, jh *JobHandler) error {
	//create a new job
	var job_tracker_pre sync.WaitGroup
	for i := range jhjp.GetPreJobs() {
		job_tracker_pre.Add(1)
		job := jhjp.GetPreJobs()[i]
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
	if(jhjp.NumIterations == 0) {
		jhjp.NumIterations = 1
	}
	json_jobs := NewJobHandler(jh.num_run_simultaneously, jh.create_worker_func, jh. print_results)
	for i := 0 ; i < jhjp.NumIterations ; i++ {
		json_jobs.AddJob(NewRTRequestResultObject(jhjp))
	}

	json_jobs.DoneAddingJobs()
	json_jobs.WaitForJobsToComplete()

	for i := range(json_jobs.Jobs) {
		json_job_result := json_jobs.Jobs[i]
		jh.appendResults(json_job_result)
	}

	if jh.err != nil {
		return jh.err
	}


	//Run the post jobs
	var job_tracker_post sync.WaitGroup
	for i := range jhjp.GetPostJobs() {
		job_tracker_post.Add(1)
		job := jhjp.GetPostJobs()[i]
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

	//obj_to_use := parserObjectCreator()
	obj_to_use := CreateJHJSONParserString()
	if err := json.Unmarshal(file_data, obj_to_use); err != nil {
		return err
	}
	if err = processJobsFromJSON(obj_to_use, jh); err != nil {
		return err
	}

	return nil
}

