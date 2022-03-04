package trx

import (
	"fmt"
	"testing"
)

var url = "tron.db"

func TestRunDb(t *testing.T) {
	InitDBTest()
}

func InitDBTest() {
	re, err := NewDB(url)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = re.Sync()

	if err != nil {
		fmt.Println(err)
		return
	}
	//dbengine = re
}
