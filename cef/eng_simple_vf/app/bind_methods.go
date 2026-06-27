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

func (m *DemoBind) Test1(data string, data2 int, browserId int) {
	fmt.Println("DemoBind.Test1", data, data2, "browserId:", browserId)
	m.Field3++
}

func (m DemoBind) Test2(datas string, datai int, dataf32 float32, datab bool) string {
	fmt.Println("DemoBind.Test2", datas, datai, dataf32, datab)
	return "Test2"
}

func (m *DemoBind) TestResult() *DemoBind {
	fmt.Println("DemoBind.TestResult", m)
	return &DemoBind{Field1: "Field1", Field2: "Field2" + time.Now().String(), Field3: m.Field3}
}

func (m *DemoBind) test() {
	fmt.Println("demo.test")
}
