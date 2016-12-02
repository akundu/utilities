package containers

import "errors"

type Tuple struct {
	Elements []interface{}
}

func NewTuple(inputs ...interface{}) *Tuple {
	t := Tuple{inputs}
	//t := Tuple{make([]interface{}, 0, len(inputs))}
	//for _, num := range nums {
	//t.Elements = append(t.Elements, num)
	//}
	return &t
}
func (t *Tuple) Add(input interface{}) {
	t.Elements = append(t.Elements, input)
}

func (t *Tuple) At(index int) (interface{}, error) {
	if t != nil && index < len(t.Elements) {
		return t.Elements[index], nil
	}
	return nil, errors.New("invalid index")
}

func (t *Tuple) Len() int {
	if t != nil {
		return len(t.Elements)
	}
	return 0
}
