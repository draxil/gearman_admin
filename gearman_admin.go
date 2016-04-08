package gearman_admin

import ( 
	"net" 
	"io"
	"bufio" 
	"strings"
)

/*
  Connection to a gearman server
*/
type Connection struct{
	net.Conn
}

/*
  Connect to a gearman server

  ie: 
 
  gma, err := gearman_admin.Connect("tcp", "10.0.50.3:4730")

*/
func Connect( network, address string)( connection *Connection, err error ){
	connection = &Connection{}
	connection.Conn, err = net.Dial( network, address )

	if err != nil {
		return 
	}

	return
}

/*
  Query the gearman server as to what workers are available
*/
func (c *Connection) Workers()( workers []Worker, err error ){
	c.Write( []byte("workers\n") )
	lines, err := read_until_stop( c )
	if err != nil {
		return
	}
	
	workers, err = workers_from_lines( lines )

	return
}

func workers_from_lines( lines []string ) ( workers []Worker, err error){
	workers = make( []Worker, 0, len(lines) )

	for _, line := range( lines ) {
		trimmed_line := strings.TrimRight( line, "\n")
		parts := strings.Split( trimmed_line, " " )
		if len(parts) < 4 {
			err = ProtocolError("Incomplete worker entry")
			return
		}

		if parts[3] != `:` {
			err = ProtocolError("Malformed worker entry '" + parts[3] + "'")
			return
		}

		var worker Worker
		
		worker.Fd        = parts[0]
		worker.Addr      = parts[1]
		worker.ClientId  = parts[2]
		if( len(parts) > 4 ){
			worker.Functions = parts[4:len(parts)]
		}
		
		workers = append( workers, worker )
	}

	return
}

/*
   Decoded description of a gearman worker
*/
type Worker struct{
	Fd               string
	Addr             string 
	ClientId         string
	Functions        []string
}

/*
    Check a worker for a particular function
*/
func (w *Worker) HasFunction( funcname string ) (bool){
	for _, v := range w.Functions {
		if v == funcname {
			return true
		}
	}
	return false
}


func read_until_stop( r io.Reader )( lines []string, err error ) {
	rdr := bufio.NewReader( r )
	stop := false
	lines = make( []string, 0, 0 )
	for ! stop {
		line := ""
		line, err = rdr.ReadString('\n')

		if err != nil {
			return
		}
		if line == ".\n" {
			return
		}
		
		lines = append( lines, line )
	}

	return
}

/*
   Some kind of protocol error
*/
type ProtocolError string
func ( p ProtocolError) Error()(string){
	return "Protocol error: " + string(p)
}
