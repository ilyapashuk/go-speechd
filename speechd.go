// Client library for SpeechDispatcher speech server for go programming language.
// Note: this library is written on pure go and does not use a C speechd library.
package speechd
import "os/exec"
import "net"
import "github.com/ilyapashuk/go-speechd/ssip"
import "os"
import "path"
import "sync"
import "strings"
import "fmt"
import "strconv"

// This constants represents known speechd event codes.
const (
EventBegin int = 701
EventEnd int = 702
EventCancel int = 703
EventPause int = 704
EventResume int = 705
)

// EventHandler type represents speechd event handler.
type EventHandler func(m ssip.SsipMessage) (persist bool)


// Message represents speechd message.
type Message struct {
Id string
cchan chan bool
}
func newMessage(s *SpeechdSession, id string) Message {
c := Message{Id: id}
c.cchan = make(chan bool, 1)
s.RegisterEventHandler(func(m ssip.SsipMessage) bool {
if m.Result[0] == c.Id {
if m.Code == EventCancel {
c.cchan <- false
return false
}
if m.Code == EventEnd {
c.cchan <- true
return false
}
}
return true
})
return c
}

// Wait function waits until message will be spoken or canceled.
// Warning: you should enable event notifications in the session, or you will get endless block.
func (c Message) Wait() (bool) {
return <- c.cchan
}

// SpeechdAddress represents an address of speech-dispatcher socket.
// It's format is identical to internal speechd format.
type SpeechdAddress string
// NetMethod gets a method string for net package, unix or tcp.
func (c SpeechdAddress) NetMethod() string {
cparts := strings.Split(string(c), ":")
m := cparts[0]
switch m {
case "unix_socket":
return "unix"
case "inet_socket":
return "tcp"
default:
panic("invalid speechd address specification: " + c)
}
}
// NetAddr method returns address string for net package, (ip for tcp, path for unix).
func (c SpeechdAddress) NetAddr() string {
cparts := strings.SplitN(string(c), ":", 2)
m := cparts[0]
var a string
if len(cparts) < 2 {
// address is not set, using default
switch m {
case "unix_socket":
a = path.Join(os.Getenv("XDG_RUNTIME_DIR"), "speech-dispatcher/speechd.sock")
case "inet_socket":
a = "127.0.0.1:6560"
default:
panic("invalid speechd address specification: " + c)
}
} else {
a = cparts[1]
}
return a
}

// GetSpeechdAddress function gets a address of SpeechDispatcher that program should use.
// If SPEECHD_ADDRESS environment variable exist, it's content is used.
// If it does not, the default value will be used, that is correct for default configuration in most systems.
// Sea the speech-dispatcher documentation for details.
func GetSpeechdAddress() SpeechdAddress {
var addr string
addr,res := os.LookupEnv("SPEECHD_ADDRESS")
if ! res {
// This environment variable does not exist, so we are taking default value.
addr = "unix_socket:" + path.Join(os.Getenv("XDG_RUNTIME_DIR"), "speech-dispatcher/speechd.sock")
}
return SpeechdAddress(addr)
}

// SpeechdSession represents an open connection to SpeechDispatcher.
// You can have multiple connections with different parameters.
type SpeechdSession struct {
// Conn field contains a SsipConn of this session. You should not use it directly.
Conn *ssip.SsipConn
mch chan ssip.SsipMessage
m sync.Mutex
rerr error
evhandlers []EventHandler
}

// NewSession opens a new speech-dispatcher session with the given configuration.
// Most clients will prefer to use speechd.Open method to start new session with default configuration.
// If autospawn is true, library will start dispatcher if it is not started.
func NewSession(a SpeechdAddress, autospawn bool) (*SpeechdSession, error) {
if autospawn {
cmd := exec.Command("speech-dispatcher", "--spawn")
cmd.Start()
}
conn,err := net.Dial(a.NetMethod(), a.NetAddr())
if err != nil {
return nil, err
}
c := new(SpeechdSession)
c.Conn = ssip.NewSsipConn(conn)
c.mch = make(chan ssip.SsipMessage)
go func() {
for {
m,err := c.Conn.ReadMessage()
if err != nil {
c.rerr = err
return
}
if m.Code >= 700 && m.Code < 800 {
for i,evhandler := range c.evhandlers {
if evhandler == nil {
continue
}
eres := evhandler(m)
if ! eres {
c.evhandlers[i] = nil
}
}
continue
}
c.mch <- m
}
}()
return c,nil
}

// Open function starts new speechd session with default configuration. It is preferable for most clients.
func Open() (*SpeechdSession, error) {
a := GetSpeechdAddress()
return NewSession(a, true)
}

func (c *SpeechdSession) Close() {
c.Conn.Close()
}

// Command function sends raw command to the dispatcher and reads response.
func (c *SpeechdSession) Command(cmd string) (ssip.SsipMessage, error) {
c.m.Lock()
defer c.m.Unlock()
return c.command(cmd)
}

// This function contains actual implementation of the command.
func (c *SpeechdSession) command(cmd string) (ssip.SsipMessage, error) {
err := c.Conn.WriteLine(cmd)
if err != nil {
return ssip.SsipMessage{}, err
}
if c.rerr != nil {
return ssip.SsipMessage{},c.rerr
}
msg := <- c.mch
return msg,nil
}

// Speak function sends message for speaking.
func (c *SpeechdSession) Speak(t string) (Message, error) {
c.m.Lock()
defer c.m.Unlock()
res,err := c.command("speak")
if err != nil {
return Message{},err
}
if res.Code < 200 && res.Code > 299 {
return Message{},fmt.Errorf("Server Error: %v %v", res.Code, res.Result)
}
tt := strings.ReplaceAll(t, "\r", "")
lines := strings.Split(tt, "\n")
for _,line := range lines {
err := c.Conn.WriteForSpeak(line)
if err != nil {
return Message{},err
}
}
res,err = c.command(".")
if err != nil {
return Message{}, err
}
if res.Code < 200 || res.Code > 299 {
return Message{}, fmt.Errorf("Server Error: %v %v", res.Code, res.Result)
}
return newMessage(c, res.Result[0]), nil
}

// Set method sets session parameters.
// If special function for this parameters exists, you should use it instead.
func (c *SpeechdSession) Set(n, v string) error {
res,err := c.Command("set self " + n + " " + v)
if err != nil {
return err
}
if res.Code < 200 || res.Code > 299 {
return fmt.Errorf("Server Error: %v %v", res.Code, res.Result)
}
return nil
}

// SetClientName method sets a client name for the speech server. Sea speechd documentation for more details.
func (c *SpeechdSession) SetClientName(user, progname, component string) error {
return c.Set("client_name", user + ":" + progname + ":" + component)
}

// SetPriority function sets priority for all next messages. Sea the ssip protocol specification for more info about priorities.
func (c *SpeechdSession) SetPriority(p string) error {
return c.Set("priority", p)
}

// SetOutputModule method sets output module to use. Use ListOutputModules to get valid values.
func (c *SpeechdSession) SetOutputModule(mn string) error {
return c.Set("Output_Module", mn)
}

// SetLanguage method sets a two letter language code to use. This can change selected voice.
func (c *SpeechdSession) SetLanguage(lc string) error {
return c.Set("language", lc)
}

// SetSpelling method switches a spelling mode.
func (c *SpeechdSession) SetSpelling(v bool) error {
var vv string
if v {
vv = "on"
} else {
vv = "off"
}
return c.Set("spelling", vv)
}

// SetRate command sets speech rate. The argument should be int within the range from -100 to 100
func (c *SpeechdSession) SetRate(v int) error {
if v < -100 || v > 100 {
panic("invalid value")
}
vv := strconv.Itoa(v)
return c.Set("rate", vv)
}

// SetVolume command sets speech volume. The argument should be int within the range from -100 to 100
func (c *SpeechdSession) SetVolume(v int) error {
if v < -100 || v > 100 {
panic("invalid value")
}
vv := strconv.Itoa(v)
return c.Set("volume", vv)
}

// SetPitch command sets speech pitch. The argument should be int within the range from -100 to 100
func (c *SpeechdSession) SetPitch(v int) error {
if v < -100 || v > 100 {
panic("invalid value")
}
vv := strconv.Itoa(v)
return c.Set("pitch", vv)
}

// SetSynthVoice method sets synthesizer voice to use. This can override language setting.
// Use ListSynthVoices method to get valid voice names.
func (c *SpeechdSession) SetSynthVoice(v string) error {
return c.Set("synthesis_voice", v)
}

// ListOutputModules method gets list of available output modules.
func (c *SpeechdSession) ListOutputModules() ([]string, error) {
res,err := c.Command("list output_modules")
if err != nil {
return nil,err
}
if res.Code < 200 || res.Code > 299 {
return nil,fmt.Errorf("ServerError: %v", res.Result)
}
return res.Result[:len(res.Result)-1], nil
}

// ListSynthVoices lists available voices for selected output module.
func (c *SpeechdSession) ListSynthVoices() ([]string, error) {
res,err := c.Command("list synthesis_voices")
if err != nil {
return nil,err
}
if res.Code < 200 || res.Code > 299 {
return nil,fmt.Errorf("ServerError: %v", res.Result)
}
return res.Result[:len(res.Result)-1], nil
}

// SetEventNotifications enables or disables all event notifications. If you want to use events, you should enable this before any speaking commands.
func (c *SpeechdSession) SetEventNotifications(v bool) error {
var vv string
if v {
vv = "on"
} else {
vv = "off"
}
return c.Set("notification all", vv)
}

func (c *SpeechdSession) RegisterEventHandler(f EventHandler) {
c.evhandlers = append(c.evhandlers, f)
}

// Stop method sends a stop command to the dispatcher. View ssip specification for details.
func (c *SpeechdSession) Stop() {
c.Command("stop self")
}

// Cancel method sends a cancel command to the dispatcher.
func (c *SpeechdSession) Cancel() {
c.Command("cancel self")
}
