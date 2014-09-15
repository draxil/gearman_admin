package gearman_admin

import (
	"testing"
	"strings"
	"reflect"
)

const read_until_stop_data = `
blah
blah
blue.
.
`

func TestReadUntilStop( t *testing.T ){
	rdr := strings.NewReader( read_until_stop_data  )
	list, err := read_until_stop( rdr )

	if err != nil {
		t.Error( err )
	}
	
	if ! reflect.DeepEqual( list ,[]string{ "\n", "blah\n", "blah\n", "blue.\n" }) {
		t.Error( "Unexpected response from read_until_stop : ", list )
	}

}


func TestWorkersFromLines( t *testing.T ){
	good_data := []string{
		"33 127.0.0.1 qqjxnvrxsnqcrovdqxcjtyztiyzjpk : add sub bb:ot\n",
		"33 127.0.0.1 - : green blue\n",
		"33 127.0.0.1 - :\n",
	}
	workers, err := workers_from_lines( good_data )
	if err != nil {
		t.Error( err )
		return
	}
	if len(workers) != 3 {
		t.Error("unexpected number of workers")
	}
	if workers[0].Addr != `127.0.0.1` {
		t.Error("Wrong address on worker 0")
	}
	if workers[1].Addr != `127.0.0.1` {
		t.Error("Wrong address on worker 0")
	}
	if workers[0].ClientId != `qqjxnvrxsnqcrovdqxcjtyztiyzjpk` {
		t.Error("Wrong clientid: " +  workers[0].ClientId)
	}
	if len(workers[0].Functions) != 3 {
		t.Error("Unexpected number of functions: ", len(workers[0].Functions))
	}
	if ! workers[0].HasFunction("add") {
		t.Error("Function 'add' not reported by HasFunction")
	}
	if ! workers[0].HasFunction("bb:ot") {
		t.Error("Function 'bb:ot' not reported by HasFunction")
	}
	if workers[0].HasFunction("green") {
		t.Error("Function 'green' falsely reported by HasFunction")
	}
	if ! workers[1].HasFunction("green") {
		t.Error("Function 'green' not reported by HasFunction on the second worker")
	}

	bad_data := []string{
		"33 127.0.0.1 qqjxnvrxsnqcrovdqxcjtyztiyzjpk : add sub bb:ot",
		"33 127.0.0.1 -",
	}
	
	_, err = workers_from_lines( bad_data )

	if err == nil {
		t.Error("Did not get an error")
	}
	_, ispe := err.(ProtocolError)

	if ! ispe {
		t.Error("Did not get a protocol error")
	}
}

