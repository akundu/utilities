package RTJobRunner

import (
	"fmt"
	"bytes"
	"text/template"
)

type KV struct {
	key   string
	value int
}

/*
func print_path(path []*KV) string {
	var result string
	for i := range(path) {
		result += fmt.Sprintf("%s ", path[i].key)
	}
	return result
}
*/

type JSONJobProcessor struct {
	Name              	string
	CommandToExecute 	string
	JSONFields
}
func (this JSONJobProcessor) GetName() string {
	if len(this.Name) > 0 {
		return this.Name
	}
	return this.CommandToExecute
}

func add_jobs(jhjp *JSONJobContainer, json_jobs *JobHandler) error {
	json_job := jhjp.Job
	t, err := template.New(jhjp.GetName()).Parse(json_job.CommandToExecute)
	if err != nil {
		return err
	}

	//create the results for each of the cases
	//1. get the list of keys
	keys := make([]string, len(json_job.Substitutes))
	var i int = 0
	for k := range json_job.Substitutes {
		keys[i] = k
		i++
	}
	//2. initialize some params
	path := make([]*KV, 0, 10)
	f_results := make([]map[string]string, 0, 10)
	//3. expand the substitutes
	expandSubstitutes(keys, 0, json_job.Substitutes, path, &f_results)
	//4. template expand and start the job
	lcw := bytes.NewBufferString("")
	for i = range f_results {
		t.Execute(lcw, f_results[i]) //expand the template
		json_jobs.AddJob(NewRTRequestResultObject(
			&JSONJobProcessor{
				Name:			  jhjp.GetName(),
				CommandToExecute: lcw.String(),
				JSONFields:       json_job.JSONFields,
			}))

		lcw.Reset()
	}

	return nil
}

func expandSubstitutes(keys []string,
	index int,
	substitutes map[string]*SubstituteData,
	path []*KV, //array of an object of
	result *[]map[string]string) {

	if index == len(keys) {
		//take the path till now and generate the data out of that
		obj_to_return := make(map[string]string)
		for i := range path {
			path_obj := path[i]
			kv_info := substitutes[path_obj.key]

			if kv_info.Type == "string" {
				obj_to_return[path_obj.key] = fmt.Sprintf("%s-%d", path_obj.key, path_obj.value)
			} else {
				obj_to_return[path_obj.key] = fmt.Sprintf("%d", path_obj.value)
			}
		}
		*result = append(*result, obj_to_return)
		return
	}

	kv_info := substitutes[keys[index]]
	for i := 0; i < kv_info.NumToGenerate; i++ {
		path = append(path, &KV{
			key:   keys[index],
			value: i,
		})
		expandSubstitutes(keys, index+1, substitutes, path, result)
		//pop from path
		path = path[:len(path)-1]
	}

	return
}
