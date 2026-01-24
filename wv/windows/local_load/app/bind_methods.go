package app

import (
	"fmt"
	"time"
)

type DemoBind struct {
	Field1 string
	Field2 string
	Field3 int
}

func (m *DemoBind) Test1(data string, data2 int) {
	fmt.Println("DemoBind.Test1", data, data2)
	m.Field3++
}

func (m DemoBind) Test2(datas string, datai int, dataf32 float32, datab bool) string {
	fmt.Println("DemoBind.Test2", datas, datai, dataf32, datab)
	return "Test2"
}

func (m *DemoBind) TestResult() *DemoBind {
	fmt.Println("DemoBind.TestResult")
	return &DemoBind{"Field1", "Field2" + time.Now().String(), m.Field3}
}
