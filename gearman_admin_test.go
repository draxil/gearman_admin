package gearman_admin

import (
	"bufio"
	_ "log"
	"net"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

const read_until_stop_data = `
blah
blah
blue.
.
`

func TestReadUntilStop(t *testing.T) {
	rdr := strings.NewReader(read_until_stop_data)
	list, err := read_until_stop(rdr)

	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(list, []string{"\n", "blah\n", "blah\n", "blue.\n"}) {
		t.Error("Unexpected response from read_until_stop : ", list)
	}

}

func TestProcessLine(t * testing.T){
	in := []string{
		"1 2 3 4",
		"1  2  3  4",
		"1	2	3	4",
		"1234",
	}
	expect := [][]string{
		[]string{"1","2","3","4"},
		[]string{"1","2","3","4"},
		[]string{"1","2","3","4"},
		[]string{"1234"},
	}
	for i := range(expect) {
		res := process_line(in[i])
		for j, v := range(expect[i]) {
			if res[j] != v {
				t.Error("Bad result for row:",i, " col:", j, " expected: ", v, " got ", res[j])
			}
		}
	}
}

func TestFuncStatusesFromLines(t *testing.T) {
	good_data := []string{
		`bb:user_import 0 0 4`,
		`bb:colour_swap 1 2 4`,
		`bb:pdf_splitter 0 0 4`,
		`pews_ping 0 0 1`,
		`bb:check_user_print_file 1 1 4`,
		`pews_event 0 0 1`,
	}

	funcs, err := func_statuses_from_lines(good_data)
	if err != nil {
		t.Error(err)
	}

	if len(funcs) != 6 {
		t.Error("Unexpected number of funcs" + strconv.Itoa(len(funcs)))
	}

	expect := []FunctionStatus{
		FunctionStatus{
			Name : "bb:user_import",
			UnfinishedJobs : 0,
			RunningJobs : 0,
			Workers : 4,
		},
		FunctionStatus{
			Name : "bb:colour_swap",
			UnfinishedJobs : 1,
			RunningJobs : 2,
			Workers : 4,
		},
		FunctionStatus{
			Name : "bb:pdf_splitter",
			UnfinishedJobs : 0,
			RunningJobs : 0,
			Workers : 4,
		},
		FunctionStatus{
			Name : "pews_ping",
			UnfinishedJobs : 0,
			RunningJobs : 0,
			Workers : 1,
		},
		FunctionStatus{
			Name : "bb:check_user_print_file",
			UnfinishedJobs : 1,
			RunningJobs : 1,
			Workers : 4,
		},
		FunctionStatus{
			Name : "pews_event",
			UnfinishedJobs : 0,
			RunningJobs : 0,
			Workers : 1,
		},
	}

	for i, v := range(expect) {

		if funcs[i].Name != v.Name {
			t.Error("Bad func name for ", i);
		}
		if funcs[i].UnfinishedJobs != v.UnfinishedJobs {
			t.Error("Bad unfinished for ", i);
		}
		if funcs[i].RunningJobs != v.RunningJobs {
			t.Error("Bad running  for ", i);
		}
		if funcs[i].Workers != v.Workers {
			t.Error("Bad workers for ", i ," / ", v.Name, " got ", funcs[i].Workers);
		}
	}

	
}

func TestWorkersFromLines(t *testing.T) {
	good_data := []string{
		"33 127.0.0.1 qqjxnvrxsnqcrovdqxcjtyztiyzjpk : add sub bb:ot\n",
		"33 127.0.0.1 - : green blue\n",
		"33 127.0.0.1 - :\n",
	}
	workers, err := workers_from_lines(good_data)
	if err != nil {
		t.Error(err)
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
		t.Error("Wrong clientid: " + workers[0].ClientId)
	}
	if len(workers[0].Functions) != 3 {
		t.Error("Unexpected number of functions: ", len(workers[0].Functions))
	}
	if !workers[0].HasFunction("add") {
		t.Error("Function 'add' not reported by HasFunction")
	}
	if !workers[0].HasFunction("bb:ot") {
		t.Error("Function 'bb:ot' not reported by HasFunction")
	}
	if workers[0].HasFunction("green") {
		t.Error("Function 'green' falsely reported by HasFunction")
	}
	if !workers[1].HasFunction("green") {
		t.Error("Function 'green' not reported by HasFunction on the second worker")
	}

	bad_data := []string{
		"33 127.0.0.1 qqjxnvrxsnqcrovdqxcjtyztiyzjpk : add sub bb:ot",
		"33 127.0.0.1 -",
	}

	_, err = workers_from_lines(bad_data)

	if err == nil {
		t.Error("Did not get an error")
	}
	_, ispe := err.(ProtocolError)

	if !ispe {
		t.Error("Did not get a protocol error")
	}
}

func TestOverTCPErr(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Error(err)
	}
	addr := l.Addr().String()
	w, err := Connect("tcp", addr)
	if err != nil {
		t.Error("Failed to connect " + err.Error())
	}
	if w == nil {
		t.Error("Connecting gave a null connection")
	}
	con, err := l.Accept()
	if err != nil {
		t.Error(err)
	}
	//	wesent := make( chan string )
	go func() {
		_, err := bufio.NewReader(con).ReadString('\n')
		if err != nil {
			panic(err)
		}
		out := bufio.NewWriter(con)
		out.WriteString("33 127.0.0.1 - : green blue\n")
		//out.WriteString(".\n")
		con.Close()
		//	wesent <- in

	}()

	_, err /*:*/ = w.Workers()
	if err == nil {
		t.Error("No EOF error")
	}
	if err != nil && err.Error() != "Error getting worker list: EOF" {
		t.Error("Non EOF error: " + err.Error())
	}
	/*	s := <- wesent
		if s != "workersyo\n" {
			t.Error("We sent: " + s , "  expected \"workers\\n\"")
		}*/

}

func TestOverTCPWorkers(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Error(err)
	}
	addr := l.Addr().String()
	w, err := Connect("tcp", addr)
	if err != nil {
		t.Error("Failed to connect " + err.Error())
	}
	if w == nil {
		t.Error("Connecting gave a null connection")
	}
	con, err := l.Accept()
	if err != nil {
		t.Error(err)
	}
	wesent := make(chan string)
	go func() {
		in, err := bufio.NewReader(con).ReadString('\n')
		if err != nil {
			panic(err)
		}
		out := bufio.NewWriter(con)
		out.WriteString("33 127.0.0.1 - : green blue\n")
		out.WriteString(".\n")
		out.Flush()
		con.Close()
		wesent <- in

	}()

	workers, err := w.Workers()
	if err != nil {
		t.Error("Unexpected error: " + err.Error())
	}

	s := <-wesent
	if s != "workers\n" {
		t.Error("We sent: "+s, "  expected \"workers\\n\"")
	}

	if len(workers) != 1 {
		t.Error("Expected one worker")
	}
	worker := workers[0]
	if worker.Fd != "33" {
		t.Error("worker.Fd")
	}
	if worker.Addr != "127.0.0.1" {
		t.Error("worker.Addr")
	}

}

func TestOverTCPStatus(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Error(err)
	}
	addr := l.Addr().String()
	w, err := Connect("tcp", addr)
	if err != nil {
		t.Error("Failed to connect " + err.Error())
	}
	if w == nil {
		t.Error("Connecting gave a null connection")
	}
	con, err := l.Accept()
	if err != nil {
		t.Error(err)
	}
	wesent := make(chan string)
	go func() {
		in, err := bufio.NewReader(con).ReadString('\n')
		if err != nil {
			panic(err)
		}
		out := bufio.NewWriter(con)
		out.WriteString("boo 1 2 3\n")
		out.WriteString(".\n")
		out.Flush()
		con.Close()
		wesent <- in

	}()

	funcs, err := w.Status()
	if err != nil {
		t.Error("Unexpected error: " + err.Error())
	}

	s := <-wesent
	if s != "status\n" {
		t.Error("We sent: "+s, "  expected \"status\\n\"")
	}

	if len(funcs) != 1 {
		t.Error("Expected one func")
	}
	funct := funcs[0]
	if funct.Name != "boo" {
		t.Error("funct.Name")
	}
	if funct.UnfinishedJobs != 1 {
		t.Error("funct.UnfinishedJobs")
	}
	if funct.RunningJobs != 2 {
		t.Error("funct.RunningJobs")
	}
	if funct.Workers != 3 {
		t.Error("funct.Workers")
	}



}
