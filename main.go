package svarmrgo
import (
    "net"
    "bufio"
    "fmt"
    "os"
)

type message struct {
    port net.Conn
    raw string
}

var connList []net.Conn

func Broadcast(Q chan message) {
    for {
            m := <- Q
            for _, c := range connList {
                if ( c != nil && c != m.port) {
                    fmt.Printf("%V\n", c)
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
    if len(os.Args) < 4 {
        panic ("Use: sendMessage host port selector message")
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
