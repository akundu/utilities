package main

import (
	"flag"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/akundu/utilities/RTJobRunner"
	"github.com/akundu/utilities/logger"
	"github.com/akundu/utilities"
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
func (this *worker) Run(id int, jobs <-chan *RTJobRunner.JobInfo, results chan<- *RTJobRunner.JobInfo) {
	for jobInfo := range jobs {
		job, ok := jobInfo.Req.(*RTJobRunner.JHJSONParserString)
		if ok == false {
			jobInfo.Resp = &RTJobRunner.BasicResponseResult{
				Err : utilities.NewBasicError("object cant cast properly"),
				Result : nil,
			}
			results <- jobInfo
			logger.Error.Printf("got error while processing %v\n", job)
			continue
		}
		jobInfo.Resp = &RTJobRunner.BasicResponseResult{
			Err : nil,
			Result : job.GetJob(),
		}
		results <- jobInfo
	}
}

func main() {
	jh := RTJobRunner.NewJobHandler(*num_simultaneously_to_run_ptr, CreateWorker, *print_results_ptr)
	if err := RTJobRunner.ProcessJobsFromJSON(*json_file_ptr, jh); err != nil {
		logger.Error.Println(err)
		return
	}
	//mark that there are no more jobs to add
	jh.DoneAddingJobs()
	//wait for the results
	jh.WaitForJobsToComplete()

	/*
	//print results
	for i := range jh.Jobs {
		fmt.Println(jh.Jobs[i].Resp)
	}
	*/
}

var (
	num_simultaneously_to_run_ptr = flag.Int("p", 1, "num times to run action")
	print_results_ptr             = flag.Bool("v", false, "print results")
	json_file_ptr                 = flag.String("f", "", "json file to load jobs from")
)

func init() {
	logger.DefaultLoggerInit()
	logger.Init(ioutil.Discard, os.Stdout, ioutil.Discard, os.Stderr)
	runtime.GOMAXPROCS(runtime.NumCPU() - 1)

	flag.Parse()
}
