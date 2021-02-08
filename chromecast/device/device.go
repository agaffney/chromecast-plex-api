package device

import (
	"context"
	"encoding/json"
	"github.com/vishen/go-chromecast/cast"
	castproto "github.com/vishen/go-chromecast/cast/proto"
	"log"
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
)

type Device struct {
	UUID           string     `json:"uuid"`
	Device         string     `json:"device"`
	Name           string     `json:"name"`
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

func (d *Device) GetApplication() *cast.Application {
	return d.application
}

func (d *Device) Send(payload cast.Payload, sourceID, destinationID, namespace string) (int, error) {
	d.requestIdMutex.Lock()
	d.requestId += 1
	requestId := d.requestId
	d.requestIdMutex.Unlock()
	payload.SetRequestId(requestId)
	log.Printf("Send(): payload = %#v, sourceID = %s, destinationID = %s, namespace = %s", payload, sourceID, destinationID, namespace)
	return requestId, d.conn.Send(d.requestId, payload, sourceID, destinationID, namespace)
}

func (d *Device) SendAndWait(payload cast.Payload, sourceID, destinationID, namespace string) (*castproto.CastMessage, error) {
	requestId, err := d.Send(payload, sourceID, destinationID, namespace)
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

func (d *Device) SendAndWaitDefaultRecv(payload cast.Payload) (*castproto.CastMessage, error) {
	return d.SendAndWait(payload, defaultSender, defaultRecv, namespaceRecv)
}

func (d *Device) recvMessages() {
	for msg := range d.recvMsgChan {
		payloadHeader := cast.PayloadHeader{}
		if err := json.Unmarshal([]byte(*msg.PayloadUtf8), &payloadHeader); err != nil {
			log.Printf("failed to unmarshal payload header: %s", err)
			continue
		}
		log.Printf("recvMessages(): namespace = %s, sourceId = %s, destinationId = %s", *msg.Namespace, *msg.SourceId, *msg.DestinationId)
		log.Printf("recvMessages(): PayloadType = %s, PayloadUtf8 = %s", *msg.PayloadType, *msg.PayloadUtf8)
		if resultChan, ok := d.resultChanMap[int(payloadHeader.RequestId)]; ok {
			resultChan <- msg
			continue
		} else {
			switch payloadHeader.Type {
			case "RECEIVER_STATUS":
				var response cast.ReceiverStatusResponse
				if err := json.Unmarshal([]byte(*msg.PayloadUtf8), &response); err != nil {
					log.Printf("failed to unmarshal receiver status payload: %s", err)
					continue
				}
				if err := d.updateReceiverStatus(&response); err != nil {
					log.Printf("failed to update receiver status: %s", err)
					continue
				}
			case "MEDIA_STATUS":
				var response cast.MediaStatusResponse
				if err := json.Unmarshal([]byte(*msg.PayloadUtf8), &response); err != nil {
					log.Printf("failed to unmarshal media status payload: %s", err)
					continue
				}
				if err := d.updateMediaStatus(&response); err != nil {
					log.Printf("failed to update media status: %s", err)
					continue
				}
			}
		}
	}
}

func (d *Device) getReceiverStatus() (*cast.ReceiverStatusResponse, error) {
	apiMessage, err := d.SendAndWaitDefaultRecv(&cast.GetStatusHeader)
	if err != nil {
		return nil, err
	}
	var response cast.ReceiverStatusResponse
	if err := json.Unmarshal([]byte(*apiMessage.PayloadUtf8), &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (d *Device) updateReceiverStatus(status *cast.ReceiverStatusResponse) error {
	// There may be more than one app, so use the last one
	for _, app := range status.Status.Applications {
		d.application = &app
	}
	d.volumeReceiver = &status.Status.Volume
	return nil
}

func (d *Device) updateMediaStatus(status *cast.MediaStatusResponse) error {
	log.Printf("media status = %#v", status)
	return nil
}

func (d *Device) Connect() error {
	d.recvMsgChan = make(chan *castproto.CastMessage, 5)
	d.resultChanMap = map[int]chan *castproto.CastMessage{}
	d.conn = cast.NewConnection(d.recvMsgChan)
	go d.recvMessages()
	if err := d.conn.Start(d.Address.String(), d.Port); err != nil {
		return err
	}
	if _, err := d.Send(&cast.ConnectHeader, defaultSender, defaultRecv, namespaceConn); err != nil {
		return err
	}
	if err := d.Update(); err != nil {
		return err
	}
	return nil
}

func (d *Device) Launch(appId string) error {
	if d.conn == nil {
		if err := d.Connect(); err != nil {
			return err
		}
	}
	// Launch specified app if it's not already running
	if d.application == nil || d.application.AppId != appId {
		payload := &cast.LaunchRequest{
			PayloadHeader: cast.LaunchHeader,
			AppId:         appId,
		}
		if _, err := d.SendAndWaitDefaultRecv(payload); err != nil {
			return err
		}
	}
	// Update receiver status
	if err := d.Update(); err != nil {
		return err
	}
	// Connect channel to app
	if _, err := d.Send(&cast.ConnectHeader, defaultSender, d.GetApplication().TransportId, namespaceConn); err != nil {
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
	if _, err := d.SendAndWaitDefaultRecv(payload); err != nil {
		return err
	}
	if err := d.Update(); err != nil {
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
	return d.updateReceiverStatus(recvStatus)
}

func (d *Device) SetVolumeLevel(level float32) error {
	if level > 1 {
		level = 1
	} else if level < 0 {
		level = 0
	}
	volumePayload := &cast.SetVolume{
		PayloadHeader: cast.VolumeHeader,
		Volume: cast.Volume{
			Level: level,
			Muted: false,
		},
	}
	if _, err := d.Send(volumePayload, defaultSender, defaultRecv, namespaceRecv); err != nil {
		return err
	}
	return nil
}

func (d *Device) SetVolumeMuted(muted bool) error {
	volumePayload := &cast.SetVolume{
		PayloadHeader: cast.VolumeHeader,
		Volume: cast.Volume{
			Muted: muted,
		},
	}
	if _, err := d.Send(volumePayload, defaultSender, defaultRecv, namespaceRecv); err != nil {
		return err
	}
	return nil
}

func (d *Device) GetVolume() *cast.Volume {
	return d.volumeReceiver
}
