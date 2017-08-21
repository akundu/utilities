package RTJobRunner

import (
	"fmt"
	"bytes"
	"text/template"
	"os"
	"math/rand"
	"github.com/akundu/utilities/statistics/distribution"
	"regexp"
	"log"
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

var template_expanders_regex *regexp.Regexp = nil
func add_jobs(jhjp *JSONJobContainer, json_jobs *JobHandler) error {
	json_job := jhjp.Job
	t, err := template.New(jhjp.GetName()).Parse(json_job.CommandToExecute)
	if err != nil {
		return err
	}


	//create the results for each of the cases
	//1. get the list of keys
	res := template_expanders_regex.FindAllStringSubmatch(json_job.CommandToExecute, -1)
	if len(res) == 0 { //didnt find anything to expand - simply add the job as it exists
		json_jobs.AddJob(NewRTRequestResultObject(
			&JSONJobProcessor{
				Name:			  jhjp.GetName(),
				CommandToExecute: json_job.CommandToExecute,
				JSONFields:       json_job.JSONFields,
			}))
		return nil
	}
	keys := make([]string, len(res))
	for i, match := range(res) {
		keys[i] = match[1]
	}

	//2. initialize some params
	path := make([]*KV, 0, 10)
	f_results := make([]map[string]string, 0, 10)
	//3. expand the substitutes
	expandSubstitutes(keys, 0, json_job.Substitutes, path, &f_results)
	//4. template expand and start the job
	lcw := bytes.NewBufferString("")
	for _, temp := range f_results {
		//t.Execute(lcw, f_results[i]) //expand the template
		t.Execute(lcw, temp)
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

	//for i,data := range(path) {
	//	fmt.Printf("%d:%v ", i, data)
	//}
	//fmt.Println()

	if index == len(keys) {
		//take the path till now and generate the data out of that
		obj_to_return := make(map[string]string)
		for _, path_obj := range path {
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


	kv_info, ok := substitutes[keys[index]]
	if(ok == false) {
		return
	}
	uniform_distr := distribution.NewuniformGenerator(kv_info.Lower, kv_info.Upper)
	gaussian_distr := distribution.NewgaussianGenerator(kv_info.Lower, kv_info.Upper, kv_info.NumToGenerate)

	for i := 0; i < kv_info.NumToGenerate; i++ {
		val := i
		if(kv_info.Type == "random") {
			val = uniform_distr.GenerateNumber()
		} else if (kv_info.Type == "gaussian") {
			val = gaussian_distr.GenerateNumber()
		}
		path = append(path, &KV{
			key:   keys[index],
			value: val,
		})

		expandSubstitutes(keys, index+1, substitutes, path, result)
		//pop from path
		path = path[:len(path)-1]
	}
	return
}

func init() {
	rand.Seed(int64(os.Getpid()))

	var err error
	if template_expanders_regex, err = regexp.Compile(`{{\.([0-9a-zA-Z]+)}}`); err != nil {
		log.Fatal("couldnt create regex obj for template_expanders")
	}
}
