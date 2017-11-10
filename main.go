//Support functions for svarmr
package svarmrgo

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	//    "time"
)

type message struct {
	port net.Conn
	raw  string
}

type Message struct {
	Conn      net.Conn
	Selector  string
	Arg       string
	NamedArgs map[string]string
}

type MessageHandler func(net.Conn, Message)

type MessageHandler2 func(Message) []Message

var connList []net.Conn

func Debug(s string) {
	//fmt.Println(s)
}

func debug(s string) {
	//fmt.Println(s)
}

//Run exec.Cmd, capture and return STDOUT
func QuickCommandStdout(cmd *exec.Cmd) string {
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
func QuickCommandStderr(cmd *exec.Cmd) string {

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

func HandleConnection(conn net.Conn, Q chan message) {
	scanner := bufio.NewScanner(conn)
	for {
		debug("Outer scanner loop")
		for scanner.Scan() {
			debug("Inner scanner loop")
			var m message = message{conn, scanner.Text()}
			Q <- m
		}
	}

}

//Read command line options, connect to svarmr server as directed from the command line
//
//Command line must be "program host port" where host and port are the connection details for the svarmr server
func CliConnect() net.Conn {
	if len(os.Args) < 2 {
		log.Println("Use: \"svarmrModule  host:port\" where host: server ip, port: server port")
		log.Println("or \"svarmrModule pipes\" for pipe IO.")
		os.Exit(1)
	}
	if os.Args[1] == "pipes" {
		return nil
	} else {
		serverPort := os.Args[1]
		return ConnectHub(serverPort)
	}
}

//Connect to a svarmr server on host:port
func ConnectHub(connString string) net.Conn {
	conn, err := net.Dial("tcp", connString)
	if err != nil {
		log.Printf("\nCould not connect to hub because: %v\n\n", err)
		os.Exit(1)
	}
	//fmt.Printf("Connected to server\n")
	return conn
}

//Prepares the wire format version of the message, suitable for printing or sending
func WireFormat(m Message) string {
	out, _ := json.Marshal(m)
	return fmt.Sprintf("%s\n", out)
}

//Build a response to a message.  Messages will soon contain unique ID numbers, allowing responses to be matched to messages
func (m *Message) Response(response Message) string {
	return WireFormat(response)
}

//Respond to a message.  Messages will soon contain unique ID numbers, allowing responses to be matched to messages
//
//Always replies on the same port we received the message from
func (m *Message) Respond(response Message) {
	out := response.Response(response)
	if m.Conn == nil {
		fmt.Fprintf(os.Stdout, out)
	} else {
		fmt.Fprintf(m.Conn, out)
	}
}

//Send a message.  Messages will soon contain unique ID numbers, allowing responses to be matched to messages
//
//If conn is nil, then it will use whatever port the message was received on, the same as RespondMessage
func SendMessage(conn net.Conn, m Message) {
	o, _ := json.Marshal(m)
	out := fmt.Sprintf("%s\n", o)
	if conn == nil {
		conn = m.Conn
	}
	if conn == nil {
		fmt.Fprintf(os.Stdout, out)
	} else {
		fmt.Fprintf(conn, out)
	}
}

type SubProx struct {
	In  io.WriteCloser
	Out io.ReadCloser
	Err io.ReadCloser
	Cmd *exec.Cmd
}

//Handle incoming messages.  This will read a message, unpack the JSON, and call the MessageHandler with the unpacked message
//
// MessageHandler must look like:
//
//    func handleMessage (conn net.Conn, m svarmrgo.Message)
func HandleInputs(conn net.Conn, callback MessageHandler) {
	//fmt.Sprintf("%V", conn)
	//time.Sleep(500 * time.Millisecond)
	r := bufio.NewReader(conn)
	for {
		debug("Outer handle inputs loop")
		l, err := r.ReadString('\n')
		if err != nil {
			os.Exit(1)
		}
		if l != "" {
			var text = l
			if len(text) > 10 {
				var m Message
				err := json.Unmarshal([]byte(text), &m)
				if err != nil {
					log.Println("error decoding message!:", err)
				} else {
					m.Conn = conn
					callback(conn, m)
				}
			} else {
				log.Printf("Invalid message: '%V'\n", []byte(text))
			}
		} else {
			log.Printf("Empty message received\n")
		}
	}
}

//Handle incoming messages.  This will read a message, unpack the JSON, and call the MessageHandler with the unpacked message
//
// MessageHandler must look like:
//
//    func handleMessage (m svarmrgo.Message) []svarmrgo.Message
//
// We call handleMessage with a message, and it returns an array of svarmrgo.Message
// We then send all the returned messages
func HandleInputLoop(conn net.Conn, callback MessageHandler2) {
	//fmt.Sprintf("%V", conn)
	//time.Sleep(500 * time.Millisecond)
	var r *bufio.Reader
	if conn != nil {
		r = bufio.NewReader(conn)
		log.Println("Using network socket ")

		for {
			debug("Outer handle inputs loop")
			l, err := r.ReadString('\n')
			if err != nil {
				os.Exit(1)
			}
			if l != "" {
				var text = l
				if len(text) > 10 {
					var m Message
					err := json.Unmarshal([]byte(text), &m)
					if err != nil {
						log.Println("error decoding message!:", err)
					} else {
						m.Conn = conn
						messages := callback(m)
						for _, message := range messages {
							SendMessage(m.Conn, message)
						}
					}
				} else {
					log.Printf("Invalid message: '%V'\n", []byte(text))
				}
			} else {
				log.Printf("Empty message received\n")
			}
		}
	} else {

		log.Println("Using pipes")
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			l := scanner.Text() // Println will add back the final '\n'

			if l != "" {
				var text = l
				if len(text) > 10 {
					var m Message
					err := json.Unmarshal([]byte(text), &m)
					if err != nil {
						log.Println("error decoding message!:", err)
					} else {
						m.Conn = conn
						messages := callback(m)
						for _, message := range messages {
							SendMessage(m.Conn, message)
						}
					}
				} else {
					log.Printf("Invalid message: '%V'\n", []byte(text))
				}
			} else {
				log.Printf("Empty message received\n")
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
		}
	}
}
