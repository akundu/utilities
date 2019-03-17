package jobs

import (
	"sync"
	"time"

	"github.com/akundu/utilities/containers"
)

type GoJob struct {
	tasks []GoTask
}

func (this *GoJob) Run() ([]*containers.Tuple, time.Duration) {
	resultsSaved := make([]*containers.Tuple, len(this.tasks))
	var wg sync.WaitGroup

	startTimeOverall := time.Now()
	for taskIndex, task := range this.tasks {
		localTask := task
		localTaskIndex := taskIndex
		wg.Add(1)
		go func() {
			defer wg.Done()
			var delta time.Duration = 0
			startTime := time.Now()
			result := localTask.Run()
			if result.Status == nil {
				delta = time.Since(startTime)
			}
			resultsSaved[localTaskIndex] = containers.NewTuple(result, delta, localTask)
		}()
	}
	wg.Wait()
	return resultsSaved, time.Since(startTimeOverall)
}
func (this *GoJob) AddTask(gt GoTask) {
	this.tasks = append(this.tasks, gt)
}

func newGoJob() *GoJob {
	job := GoJob{make([]GoTask, 0, 5)}
	return &job
}
