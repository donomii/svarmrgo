package svarmrgo
import (
    "net"
    "bufio"
    "fmt"
    "os"
	"encoding/json"
)

type message struct {
    port net.Conn
    raw string
}

type MessageHandler func(net.Conn, Message) 

var connList []net.Conn

func Broadcast(Q chan message) {
    for {
            m := <- Q
            for _, c := range connList {
                if ( c != nil && c != m.port) {
                    //fmt.Printf("%V\n", c)
                    c.Write([]byte(m.raw))
                    c.Write([]byte("\r\n"))
                }
            }
        }
}

func HandleConnection (conn net.Conn, Q chan message) {
    scanner := bufio.NewScanner(conn)

    for {
            for scanner.Scan() {
                var m message = message{ conn, scanner.Text() }
                Q <- m
        }
    }

}

type Message struct {
    Selector string
    Arg string
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
        l,_ := r.ReadString('\n')
        if (l!="") {
                var text = l
                //fmt.Printf("%v\n", text)
                var m Message
                err := json.Unmarshal([]byte(text), &m)
                if err != nil {
                    //fmt.Println("error decoding message!:", err)
                } else {
                    //fmt.Printf("%v", m)
                    callback(conn, m)
                }
            }
        }
    }
