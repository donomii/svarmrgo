package svarmrgo
import (
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



func CliConnect() net.Conn {
    if len(os.Args) < 2 {
        panic ("Use:  host port")
    }
    server := os.Args[1]
    port := os.Args[2]
    return ConnectHub(server, port)
}

func ConnectHub(server string, port string) net.Conn {
    conn, err := net.Dial("tcp", fmt.Sprintf("%v:%v", server, port))
    if err != nil {
        panic(fmt.Sprintf("Could not connect to hub because: %v", err))
    }
    fmt.Printf("Connected to server\n")
    return conn
}

func RespondWith(conn net.Conn, response Message) {
	out, _ := json.Marshal(response)
	fmt.Fprintf(conn, fmt.Sprintf("%s\n", out))
}

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
