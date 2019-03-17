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

type worker struct { }
func CreateWorker() RTJobRunner.Worker {
	return &worker{}
}
func (this *worker) PostRun() { }
func (this *worker) PreRun() { }
func (this *worker) Run(id int, jh *RTJobRunner.JobHandler) {
	//for jobInfo := range jobs {
	for j := jh.GetJob(); j != nil ; j = jh.GetJob() {
		job, ok := j.Req.(*RTJobRunner.JSONJobProcessor)
		if ok == false {
			j.Resp = &RTJobRunner.BasicResponseResult{
				Err : utilities.NewBasicError("object cant cast properly"),
				Result : nil,
			}
			jh.DoneJob(j)
			logger.Error.Printf("got error while processing %v\n", job)
			continue
		}
		j.Resp = &RTJobRunner.BasicResponseResult{
			Err : nil,
			Result : job.CommandToExecute,
		}
		jh.DoneJob(j)
	}
}

func main() {
	jh := RTJobRunner.NewJobHandler(1, CreateWorker)
	jh.SetPrintIndividualResults(*print_results_ptr)
	jh.SetPrintStatistics(true)
	if err := RTJobRunner.ProcessJobsFromJSON(*json_file_ptr, jh, *num_simultaneously_to_run_ptr); err != nil {
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
