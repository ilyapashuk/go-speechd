package speechd
import "testing"
import "time"

func TestOpen(t *testing.T) {
t.Log("opening connection with default settings")
c,err := Open()
if err != nil {
t.Fatal(err)
}
defer c.Close()
t.Run("ClientName", func(t *testing.T) {
t.Log("setting client name")
err := c.SetClientName("user", "tester", "tester")
if err != nil {
t.Fatal(err)
}
})
t.Run("PlainSpeak", func(t *testing.T) {
_,err := c.Speak("plain1")
if err != nil {
t.Fatal(err)
}
time.Sleep(3 * time.Second)
_,err = c.Speak("plain2")
if err != nil {
t.Fatal(err)
}
})
time.Sleep(3 * time.Second)
t.Run("rate", func(t *testing.T) {
t.Log("low rate")
err := c.SetRate(-100)
if err != nil {
t.Fatal(err)
}
_,err = c.Speak("low rate")
if err != nil {
t.Fatal(err)
}
time.Sleep(3 * time.Second)
t.Log("high rate")
err = c.SetRate(100)
if err != nil {
t.Fatal(err)
}
_,err = c.Speak("high rate")
if err != nil {
t.Fatal(err)
}
time.Sleep(3 * time.Second)
t.Log("normal rate")
err = c.SetRate(0)
if err != nil {
t.Fatal(err)
}
_,err = c.Speak("normal")
if err != nil {
t.Fatal(err)
}
})
t.Run("pitch", func(t *testing.T) {
t.Log("low pitch")
err := c.SetPitch(-100)
if err != nil {
t.Fatal(err)
}
_,err = c.Speak("low pitch")
if err != nil {
t.Fatal(err)
}
time.Sleep(3 * time.Second)
t.Log("high pitch")
err = c.SetPitch(100)
if err != nil {
t.Fatal(err)
}
_,err = c.Speak("high pitch")
if err != nil {
t.Fatal(err)
}
time.Sleep(3 * time.Second)
t.Log("normal pitch")
err = c.SetPitch(0)
if err != nil {
t.Fatal(err)
}
_,err = c.Speak("normal")
if err != nil {
t.Fatal(err)
}
})
t.Run("SpeakWithWait", func(t *testing.T) {
t.Log("enabling event notifications")
err := c.SetEventNotifications(true)
if err != nil {
t.Fatal(err)
}
t.Log("speaking first word")
tn := time.Now()
msg,err := c.Speak("this is an eventful speaking")
if err != nil {
t.Fatal(err)
}
msg.Wait()
tt := time.Since(tn)
t.Logf("ok speaking eventful, time taken %v", tt / time.Millisecond)
})
t.Run("ListOutputModules", func(t *testing.T) {
t.Log("requesting outmod list")
vl,err := c.ListOutputModules()
if err != nil {
t.Fatal(err)
}
for _,v := range vl {
t.Run("SetOutputModule", func(t *testing.T) {
t.Logf("setting outmod to %s", v)
err := c.SetOutputModule(v)
if err != nil {
t.Fatal(err)
}
c.Speak("module test")
time.Sleep(3 * time.Second)
})
}
})
}
