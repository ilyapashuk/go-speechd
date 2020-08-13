package ssip

// This test will try to connect to your speechd and set client name, then disconnect.
// This will test operability of key library functions.
import "testing"
import "net"
func TestSsip(t *testing.T) {
var spaddr = "/run/user/1000/speech-dispatcher/speechd.sock"
t.Logf("connecting to %s", spaddr)
conn,err := net.Dial("unix", spaddr)
if err != nil {
t.Fatalf("can't connect to speech server: %s", err.Error())
}
sconn := NewSsipConn(conn)
defer sconn.Close()
t.Log("setting client name")
t.Log("sending: set self client_name user:tester:tester")
err = sconn.WriteLine("set self client_name user:tester:tester")
if err != nil {
t.Fatal(err)
}
t.Log("reading response")
resp,err := sconn.ReadMessage()
if err != nil {
t.Fatal(err)
}
t.Log(resp)

}
