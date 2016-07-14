//Support functions for svarmr
package svarmrgo
import (
        "os/exec"
    "strings"
    "bytes"
    "net"
    "bufio"
    "fmt"
    "os"
    "encoding/json"
//    "time"
)

type message struct {
    port net.Conn
    raw string
}

type Message struct {
    Conn  net.Conn
    Selector string
    Arg string
    NamedArgs map[string]string
}


type MessageHandler func(net.Conn, Message)

var connList []net.Conn

func Debug(s string) {
    //fmt.Println(s)
}


func debug(s string) {
    //fmt.Println(s)
}

//Run exec.Cmd, capture and return STDOUT
func QuickCommandStdout (cmd *exec.Cmd) string{
    fmt.Println()
    in := strings.NewReader("")
    cmd.Stdin = in 
    var out bytes.Buffer
    cmd.Stdout = &out
    var err bytes.Buffer
    cmd.Stderr = &err
    //res := cmd.Run()
    cmd.Run()
    //fmt.Printf("Command result: %v\n", res)
    ret := fmt.Sprintf("%s", out)
    //fmt.Println(ret)
    return ret
}

//Run exec.Cmd, capture and return STDERR
func QuickCommandStderr (cmd *exec.Cmd) string{
    fmt.Println()
    in := strings.NewReader("")
    cmd.Stdin = in 
    var out bytes.Buffer
    cmd.Stdout = &out
    var err bytes.Buffer
    cmd.Stderr = &err
    cmd.Run()
    //fmt.Printf("Command result: %v\n", res)
    ret := fmt.Sprintf("%s", err)
    //fmt.Println(ret)
    return ret
}

func HandleConnection (conn net.Conn, Q chan message) {
    scanner := bufio.NewScanner(conn)
    for {
            debug("Outer scanner loop")
            for scanner.Scan() {
                debug("Inner scanner loop")
                var m message = message{ conn, scanner.Text() }
                Q <- m
        }
    }

}



//Read command line options, connect to svarmr server as directed from the command line
//
//Command line must be "program host port" where host and port are the connection details for the svarmr server
func CliConnect() net.Conn {
    if len(os.Args) < 2 {
        fmt.Println ("Use: svarmrModule  host port")
        fmt.Println ("host: server ip, port: server port")
        os.Exit(1)
    }
    server := os.Args[1]
    port := os.Args[2]
    return ConnectHub(server, port)
}

//Connect to a svarmr server on host:port
func ConnectHub(server string, port string) net.Conn {
    conn, err := net.Dial("tcp", fmt.Sprintf("%v:%v", server, port))
    if err != nil {
        fmt.Printf("\nCould not connect to hub because: %v\n\n", err)
        os.Exit(1)
    }
    //fmt.Printf("Connected to server\n")
    return conn
}

//Respond to a message.  Messages will soon contain unique ID numbers, allowing responses to be matched to messages
func (m *Message) Respond(response Message) {
	out, _ := json.Marshal(response)
	fmt.Fprintf(m.Conn, fmt.Sprintf("%s\n", out))
}

//Send a message.  Messages will soon contain unique ID numbers, allowing responses to be matched to messages
func SendMessage(conn net.Conn, response Message) {
	out, _ := json.Marshal(response)
	fmt.Fprintf(conn, fmt.Sprintf("%s\n", out))
}

//Handle incoming messages.  This will read a message, unpack the JSON, and call the MessageHandler with the unpacked message
//
// MessageHandler must look like:
//
//    func handleMessage (conn net.Conn, m svarmrgo.Message)
func HandleInputs (conn net.Conn, callback MessageHandler) {
    //fmt.Sprintf("%V", conn)
    //time.Sleep(500 * time.Millisecond)
    r := bufio.NewReader(conn)
    for {
        debug("Outer handle inputs loop")
        l, err := r.ReadString('\n')
		if err != nil {
			os.Exit(1)
		}
        if (l!="") {
                var text = l
                if len(text)>10 {
                    var m Message
                    err := json.Unmarshal([]byte(text), &m)
                    if err != nil {
                        fmt.Println("error decoding message!:", err)
                    } else {
			    m.Conn = conn
                            callback(conn, m)
                    }
                } else {
                    fmt.Printf("Invalid message: '%V'\n", []byte(text))
                }
            } else {
                fmt.Printf("Empty message received\n")
            }
        }
    }
