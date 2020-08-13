// This package provides a client implementation of ssip (speech synthesis interface protocol) as mentioned on https://freebsoft.org/doc/speechd/ssip.html for go programming language.
// Note: most clients will prefer to use more high level speechd library.
package ssip

import (
"net"
"errors"
"bufio"
"strings"
"strconv"
)

// SsipMessage type represents server's message, that consists of 3 digits code and collection of strings defining message results.
type SsipMessage struct {
// Code represents 3 digit code returned by the server, that can be used to determine action results.
// Note: Code always equal to the code part of the last response string.
Code int

// Result field contains a string slice with server returned data. Usually it contains only useless human readable descriptions of performed actions like in ftp protocol, but sometimes it contains interesting machine readable output.
Result []string
}

// SsipConn type represents a client ssip session.
type SsipConn struct {
// Conn field contains a underlying connection to the ssip socket.
// Warning: This field should be used only for retrieving information about the connection, like type or address, but you should not change anything.
Conn net.Conn
scnr *bufio.Scanner
}

func NewSsipConn(t net.Conn) *SsipConn {
c := new(SsipConn)
c.Conn = t
c.scnr = bufio.NewScanner(t)
return c
}

// ReadMessage method will read the next ssip message.
// Warning: keep in mind that you can receive a event notification that is not related to your request. You can identify them by 7xx code. However, they are disabled by default.
func (c *SsipConn) ReadMessage() (SsipMessage, error) {
var msg SsipMessage
for {
if ! c.scnr.Scan() {
break
}
t := c.scnr.Text()
scode := string([]rune(t)[:3])
code,err := strconv.Atoi(scode)
if err != nil {
return SsipMessage{},errors.New("invalid result code: " + scode)
}
msg.Code = code
res := string([]rune(t)[4:])
msg.Result = append(msg.Result, res)
delim := []rune(t)[3]
if delim == ' ' {
break
}
}
return msg, c.scnr.Err()
}

// WriteLine method will send the line of text with appropriat windows style line ending to the connection.
// Note: Use WriteForSpeak method to write lines for the speak command.
func (c *SsipConn) WriteLine(p string) (error) {
_,err := c.Conn.Write([]byte(p + "\r\n"))
return err
}

// WriteForSpeak method writes line for speak command with all needed escaping, as defined in protocol specification.
func (c *SsipConn) WriteForSpeak(p string) (error) {
// As defined in the protocol, if input string starts with a dot, we will prepend an other dot to it.
var pp string
if strings.HasPrefix(p, ".") {
pp = "." + p
} else {
pp = p
}
return c.WriteLine(pp)
}

func (c *SsipConn) Close() {
c.WriteLine("quit")
c.Conn.Close()
}

