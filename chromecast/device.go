package chromecast

import (
	"github.com/vishen/go-chromecast/cast"
	castproto "github.com/vishen/go-chromecast/cast/proto"
	"net"
	"time"
)

const (
	// 'CC1AD845' seems to be a predefined app; check link
	// https://gist.github.com/jloutsenhizer/8855258
	// https://github.com/thibauts/node-castv2
	defaultChromecastAppId = "CC1AD845"

	defaultSender = "sender-0"
	defaultRecv   = "receiver-0"

	namespaceConn  = "urn:x-cast:com.google.cast.tp.connection"
	namespaceRecv  = "urn:x-cast:com.google.cast.receiver"
	namespaceMedia = "urn:x-cast:com.google.cast.media"

	plexAppId     = "9AC194DC"
	plexNamespace = "urn:x-cast:plex"
)

type Device struct {
	UUID        string     `json:"uuid"`
	Device      string     `json:"device"`
	DeviceName  string     `json:"name"`
	Address     net.IP     `json:"address"`
	Port        int        `json:"port"`
	FirstSeen   *time.Time `json:"first_seen"`
	LastSeen    *time.Time `json:"last_seen"`
	conn        *cast.Connection
	requestId   int
	recvMsgChan chan *castproto.CastMessage
}

func (d *Device) send(payload cast.Payload, sourceID, destinationID, namespace string) (int, error) {
	if d.conn == nil {
		if err := d.Connect(); err != nil {
			return 0, err
		}
	}
	// TODO: make thread safe
	d.requestId += 1
	payload.SetRequestId(d.requestId)
	return d.requestId, d.conn.Send(d.requestId, payload, sourceID, destinationID, namespace)
}

func (d *Device) Connect() error {
	d.recvMsgChan = make(chan *castproto.CastMessage, 5)
	d.conn = cast.NewConnection(d.recvMsgChan)
	if err := d.conn.Start(d.Address.String(), d.Port); err != nil {
		return err
	}
	if _, err := d.send(&cast.ConnectHeader, defaultSender, defaultRecv, namespaceConn); err != nil {
		return err
	}
	return nil
}

func (d *Device) Launch() error {
	payload := &cast.LaunchRequest{
		PayloadHeader: cast.LaunchHeader,
		AppId:         plexAppId,
	}
	if _, err := d.send(payload, defaultSender, defaultRecv, plexNamespace); err != nil {
		return err
	}
	return nil
}

func (d *Device) Reset() error {
	payload := &cast.LaunchRequest{
		PayloadHeader: cast.LaunchHeader,
		AppId:         defaultChromecastAppId,
	}
	if _, err := d.send(payload, defaultSender, defaultRecv, namespaceRecv); err != nil {
		return err
	}
	return nil
}
