package chromecast

import (
	"context"
	"encoding/json"
	"github.com/buger/jsonparser"
	"github.com/vishen/go-chromecast/cast"
	castproto "github.com/vishen/go-chromecast/cast/proto"
	//"log"
	"net"
	"sync"
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
	UUID           string     `json:"uuid"`
	Device         string     `json:"device"`
	DeviceName     string     `json:"name"`
	Address        net.IP     `json:"address"`
	Port           int        `json:"port"`
	FirstSeen      *time.Time `json:"first_seen"`
	LastSeen       *time.Time `json:"last_seen"`
	requestIdMutex sync.Mutex
	conn           *cast.Connection
	requestId      int
	recvMsgChan    chan *castproto.CastMessage
	resultChanMap  map[int]chan *castproto.CastMessage
	application    *cast.Application
	volumeReceiver *cast.Volume
}

func (d *Device) send(payload cast.Payload, sourceID, destinationID, namespace string) (int, error) {
	d.requestIdMutex.Lock()
	d.requestId += 1
	requestId := d.requestId
	d.requestIdMutex.Unlock()
	payload.SetRequestId(requestId)
	return requestId, d.conn.Send(d.requestId, payload, sourceID, destinationID, namespace)
}

func (d *Device) sendAndWait(payload cast.Payload, sourceID, destinationID, namespace string) (*castproto.CastMessage, error) {
	requestId, err := d.send(payload, sourceID, destinationID, namespace)
	if err != nil {
		return nil, err
	}

	// Set a timeout to wait for the response
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	resultChan := make(chan *castproto.CastMessage, 1)
	d.requestIdMutex.Lock()
	d.resultChanMap[requestId] = resultChan
	d.requestIdMutex.Unlock()
	defer func() {
		d.requestIdMutex.Lock()
		delete(d.resultChanMap, requestId)
		d.requestIdMutex.Unlock()
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-resultChan:
		return result, nil
	}
}

func (d *Device) sendAndWaitDefaultRecv(payload cast.Payload) (*castproto.CastMessage, error) {
	return d.sendAndWait(payload, defaultSender, defaultRecv, namespaceRecv)
}

func (d *Device) recvMessages() {
	for msg := range d.recvMsgChan {
		requestId, err := jsonparser.GetInt([]byte(*msg.PayloadUtf8), "requestId")
		if err == nil {
			if resultChan, ok := d.resultChanMap[int(requestId)]; ok {
				resultChan <- msg
				continue
			}
		}
	}
}

func (d *Device) getReceiverStatus() (*cast.ReceiverStatusResponse, error) {
	apiMessage, err := d.sendAndWaitDefaultRecv(&cast.GetStatusHeader)
	if err != nil {
		return nil, err
	}
	var response cast.ReceiverStatusResponse
	if err := json.Unmarshal([]byte(*apiMessage.PayloadUtf8), &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (d *Device) Connect() error {
	d.recvMsgChan = make(chan *castproto.CastMessage, 5)
	d.resultChanMap = map[int]chan *castproto.CastMessage{}
	d.conn = cast.NewConnection(d.recvMsgChan)
	go d.recvMessages()
	if err := d.conn.Start(d.Address.String(), d.Port); err != nil {
		return err
	}
	if _, err := d.send(&cast.ConnectHeader, defaultSender, defaultRecv, namespaceConn); err != nil {
		return err
	}
	if err := d.Update(); err != nil {
		return err
	}
	return nil
}

func (d *Device) Launch() error {
	if d.conn == nil {
		if err := d.Connect(); err != nil {
			return err
		}
	}
	if d.application != nil && d.application.AppId == plexAppId {
		return nil
	}
	payload := &cast.LaunchRequest{
		PayloadHeader: cast.LaunchHeader,
		AppId:         plexAppId,
	}
	if _, err := d.sendAndWaitDefaultRecv(payload); err != nil {
		return err
	}
	return nil
}

func (d *Device) Reset() error {
	if d.conn == nil {
		if err := d.Connect(); err != nil {
			return err
		}
	}
	payload := &cast.LaunchRequest{
		PayloadHeader: cast.LaunchHeader,
		AppId:         defaultChromecastAppId,
	}
	if _, err := d.sendAndWaitDefaultRecv(payload); err != nil {
		return err
	}
	return nil
}

func (d *Device) Update() error {
	if d.conn == nil {
		if err := d.Connect(); err != nil {
			return err
		}
	}
	recvStatus, err := d.getReceiverStatus()
	if err != nil {
		return err
	}
	// There may be more than one app, so use the last one
	for _, app := range recvStatus.Status.Applications {
		d.application = &app
	}
	d.volumeReceiver = &recvStatus.Status.Volume
	return nil
}
