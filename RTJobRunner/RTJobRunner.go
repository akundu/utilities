package RTJobRunner

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"time"

	"github.com/akundu/utilities/logger"
	"github.com/satori/go.uuid"
	"github.com/montanaflynn/stats"
	"bytes"
)

type JobHandler struct {
	req_chan chan *JobInfo
	res_chan chan *JobInfo

	ws_job_tracker         sync.WaitGroup
	num_added              int32
	num_run_simultaneously int
	create_worker_func     CreateWorkerFunction

	done_channel 				chan bool
	worker_list  				[]Worker
	id           				string
	err          				error
	print_individual_results 	bool
	print_statistics 			bool

	Jobs 						[]*JobInfo
}

func (this *JobHandler) SetPrintIndividualResults(val bool) {
	this.print_individual_results = val
}

func (this *JobHandler) SetPrintStatistics(val bool) {
	this.print_statistics = val
}

func (this *JobHandler) GetJob() *JobInfo {
	job:= <-this.req_chan
	if job == nil {
		return nil
	}

	job.job_start_time = time.Now()
	return job
}

func (this *JobHandler) DoneJob(job *JobInfo) {
	job.job_end_time = time.Now()
	this.res_chan <- job
}

func NewJobHandler(num_to_setup int, createWorkerFunc CreateWorkerFunction) *JobHandler {
	jh := &JobHandler{
		req_chan:               make(chan *JobInfo, num_to_setup),
		res_chan:               make(chan *JobInfo, num_to_setup),
		num_run_simultaneously: num_to_setup,
		num_added:              0,
		create_worker_func:     createWorkerFunc,
		worker_list:            make([]Worker, num_to_setup),
		done_channel:           make(chan bool, 1),
		id:                     fmt.Sprintf("%s", uuid.NewV4()),
		err:                    nil,
		print_statistics:       true,
	}

	for w := 0; w < num_to_setup; w++ {
		worker := createWorkerFunc()
		jh.worker_list[w] = worker
		worker.PreRun()
		go worker.Run(w, jh)
	}

	jh.ws_job_tracker.Add(1) //goroutine to wait for results
	go jh.waitForResults()

	return jh
}

func (this *JobHandler) AddJob(job *JobInfo) {
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
func (this *JobHandler) waitForResults() {
	timing_results := make(stats.Float64Data, 0)

	var job_name string
	var num_processed int32 = 0
	var job_complete bool

	go func() {
		for ;; {
			time.Sleep(2000*time.Millisecond)
			logger.Trace.Printf("%s %d/%d\n",
						  job_name,
						  num_processed,
						  atomic.LoadInt32(&this.num_added))
			if job_complete == true {
				break
			}
		}
	}()

	done_adding := false
	for done_adding == false || num_processed < atomic.LoadInt32(&this.num_added) {
		select {
		case result := <-this.res_chan:
			num_processed++
			if result == nil {
				continue
			}

			timing := (float64(result.JobTime())/1000000)
			timing_results = append(timing_results, timing)

			job_name = result.Req.GetName()
			if this.print_individual_results == true {
				logger.Trace.Printf("%0.3fms %s %v\n",
					timing,
					result.Req.GetName(),
					result.Resp)
			}
			this.appendResults(result)

		case done_adding = <-this.done_channel:
			continue
		}
	}
	job_complete = true


	//clean up the workers if needed
	for i := range this.worker_list {
		this.worker_list[i].PostRun()
	}

	if this.print_statistics == true && timing_results.Len() > 0{
		buffered_writer := bytes.NewBufferString("")
		buffered_writer.WriteString(fmt.Sprintln("Results: ", job_name))
		if nth_percentile,err := timing_results.Percentile(10.0); err == nil {
			buffered_writer.WriteString(fmt.Sprintf("10%%  :   %10.2f\n", nth_percentile))
		}
		if nth_percentile,err := timing_results.Percentile(25.0); err == nil {
			buffered_writer.WriteString(fmt.Sprintf("25%%  :   %10.2f\n", nth_percentile))
		}
		if nth_percentile,err := timing_results.Percentile(50.0); err == nil {
			buffered_writer.WriteString(fmt.Sprintf("50%%  :   %10.2f\n", nth_percentile))
		}
		if nth_percentile,err := timing_results.Percentile(75.0); err == nil {
			buffered_writer.WriteString(fmt.Sprintf("75%%  :   %10.2f\n", nth_percentile))
		}
		if nth_percentile,err := timing_results.Percentile(90.0); err == nil {
			buffered_writer.WriteString(fmt.Sprintf("90%%  :   %10.2f\n", nth_percentile))
		}
		if nth_percentile,err := timing_results.Percentile(95.0); err == nil {
			buffered_writer.WriteString(fmt.Sprintf("95%%  :   %10.2f\n", nth_percentile))
		}
		if nth_percentile,err := timing_results.Percentile(99.0); err == nil {
			buffered_writer.WriteString(fmt.Sprintf("99%%  :   %10.2f\n", nth_percentile))
		}
		if nth_percentile,err := timing_results.Percentile(100.0); err == nil {
			buffered_writer.WriteString(fmt.Sprintf("100%% :   %10.2f\n", nth_percentile))
		}

		buffered_writer.WriteString(fmt.Sprintln())
		buffered_writer.WriteString(fmt.Sprintf("NumRun : %10d\n", timing_results.Len()))
		val,_ := timing_results.Median()
		buffered_writer.WriteString(fmt.Sprintf("Medan  : %10.2f\n", val))
		val,_ = timing_results.Mean()
		buffered_writer.WriteString(fmt.Sprintf("Mean   : %10.2f\n", val))
		buffered_writer.WriteString(fmt.Sprintf("Req/sec: %10.2f\n", 1000/float64(val)))
		val,_ = timing_results.Max()
		buffered_writer.WriteString(fmt.Sprintf("Max    : %10.2f\n", val))
		val,_ = timing_results.Min()
		buffered_writer.WriteString(fmt.Sprintf("Min    : %10.2f\n", val))
		val,_ = timing_results.StandardDeviation()
		buffered_writer.WriteString(fmt.Sprintf("StdDev : %10.2f\n", val))
		buffered_writer.WriteString(fmt.Sprintln())

		logger.Info.Println(buffered_writer.String())
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

