package gb28181

import (
	. "github.com/Monibuca/engine/v2"
	"github.com/Monibuca/engine/v2/util"
	"github.com/Monibuca/plugin-gb28181/transaction"
	. "github.com/logrusorgru/aurora"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

var Devices sync.Map
var config = struct {
	Serial     string
	Realm      string
	ListenAddr string
	Expires    int
	AutoInvite bool
	MediaPort  uint16
}{"34020000002000000001", "3402000000", "127.0.0.1:5060", 3600, true, 6000}

func init() {
	InstallPlugin(&PluginConfig{
		Name:   "GB28181",
		Config: &config,
		Type:   PLUGIN_PUBLISHER,
		Run:    run,
	})
}

func run() {
	ipAddr, err := net.ResolveUDPAddr("", config.ListenAddr)
	if err != nil {
		log.Fatal(err)
	}
	Print(Green("server gb28181 start at"), BrightBlue(config.ListenAddr))
	config := &transaction.Config{
		SipIP:      ipAddr.IP.String(),
		SipPort:    uint16(ipAddr.Port),
		SipNetwork: "UDP",
		Serial:     config.Serial,
		Realm:      config.Realm,
		AckTimeout: 10,
		MediaIP: ipAddr.IP.String(),
		MediaPort: config.MediaPort,
		RegisterValidity:  config.Expires,
		RegisterInterval:  60,
		HeartbeatInterval: 60,
		HeartbeatRetry:    3,

		AudioEnable:      true,
		WaitKeyFrame:     true,
		MediaPortMin:     58200,
		MediaPortMax:     58300,
		MediaIdleTimeout: 30,
	}
	s := transaction.NewCore(config)
	s.OnInvite = onPublish
	http.HandleFunc("/gb28181/list", func(w http.ResponseWriter, r *http.Request) {
		sse := util.NewSSE(w, r.Context())
		for {
			var list []*transaction.Device
			s.Devices.Range(func(key, value interface{}) bool {
				list = append(list, value.(*transaction.Device))
				return true
			})
			sse.WriteJSON(list)
			select {
			case <-time.After(time.Second * 5):
			case <-sse.Done():
				return
			}
		}
	})
	http.HandleFunc("/gb28181/cmd", func(w http.ResponseWriter, r *http.Request){

	})
	s.Start()

}
