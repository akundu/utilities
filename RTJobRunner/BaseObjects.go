package RTJobRunner

type Response interface{}
type Request interface{}

type Worker interface {
	PreRun()
	PostRun()
	Run(id int, jobs <-chan Request, results chan<- Response)
}
type CreateWorkerFunction func() Worker

