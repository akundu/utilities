package RTJobRunner

import (
	"time"
)

type Result interface{}
type Response interface {
	GetError() error
	GetResult() Result
}
type Request interface {
	GetName() string
}



type StringRequest struct {
	Str string
}
func (this StringRequest) GetName() string {
	return this.Str
}
func (this StringRequest) GetStr() string {
	return this.Str
}





type JobInfo struct {
	Resp Response
	Req  Request

	job_start_time time.Time
	job_end_time   time.Time
}
func (this JobInfo) JobTime() time.Duration {
	return this.job_end_time.Sub(this.job_start_time)
}
func NewRTRequestResultObject(req Request) *JobInfo {
	return &JobInfo{
		Req:  req,
		Resp: nil,
	}
}





type BasicResponseResult struct {
	Err    error
	Result Result
}
func (this BasicResponseResult) GetError() error {
	return this.Err
}
func (this BasicResponseResult) GetResult() Result {
	return this.Result
}





type Worker interface {
	PreRun()
	PostRun()
	Run(id int, jh *JobHandler)
}
type CreateWorkerFunction func() Worker
