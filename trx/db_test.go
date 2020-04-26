package trx

import (
	"fmt"
	"testing"
)

var url = "./trx.db"

func TestRunDb(t *testing.T) {
	InitDBTest()
}

func InitDBTest() {
	re, err := InitDB(url)
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
