package monitor

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/rancher/rio/types/client/rio/v1beta1"
	"github.com/sirupsen/logrus"
)

type UpWatcher struct {
	sync.Mutex
	c             *client.Client
	subCounter    int
	subscriptions map[int]*Subscription
}

func (m *UpWatcher) Subscribe() *Subscription {
	m.Lock()
	defer m.Unlock()

	m.subCounter++
	sub := &Subscription{
		id: m.subCounter,
		C:  make(chan *Event, 1024),
	}
	m.subscriptions[sub.id] = sub

	return sub
}

func (m *UpWatcher) Unsubscribe(sub *Subscription) {
	m.Lock()
	defer m.Unlock()

	close(sub.C)
	delete(m.subscriptions, sub.id)
}

func NewUpWatcher(c *client.Client) *UpWatcher {
	return &UpWatcher{
		c:             c,
		subscriptions: map[int]*Subscription{},
	}
}

func (m *UpWatcher) Start(stackID string) error {
	schema, ok := m.c.Types["subscribe"]
	if !ok {
		return fmt.Errorf("Not authorized to subscribe")
	}

	urlString := schema.Links["collection"]
	u, err := url.Parse(urlString)
	if err != nil {
		return err
	}

	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	}

	q := u.Query()
	q.Add("eventNames", "resource.change")
	q.Add("eventNames", "service.kubernetes.change")

	u.RawQuery = q.Encode()

	conn, resp, err := m.c.Websocket(u.String(), nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != 101 {
		return fmt.Errorf("Bad status code: %d %s", resp.StatusCode, resp.Status)
	}

	logrus.Debugf("Connected to: %s", u.String())

	return m.watch(conn, stackID)
}

func (m *UpWatcher) watch(conn *websocket.Conn, stackID string) error {
	//serviceIds := map[string]struct{}{}
	for {
		v := Event{}
		_, r, err := conn.NextReader()
		if err != nil {
			return err
		}
		if err := json.NewDecoder(r).Decode(&v); err != nil {
			logrus.Errorf("Failed to parse json in message")
			continue
		}

		logrus.Debugf("Event: %s %s %s", v.Name, v.ResourceType, v.ResourceID)
		//if v.ResourceType == "stack" {
		//	stackData := &client.Stack{}
		//	if err := unmarshalling(v.Data["resource"], stackData); err != nil {
		//		logrus.Errorf("failed to unmarshalling err: %v", err)
		//		continue
		//	}
		//	if stackData.ID == stackID {
		//		stackID = stackData.ID
		//		for _, serviceID := range stackData.ServiceIds {
		//			serviceIds[serviceID] = struct{}{}
		//		}
		//		switch stackData.Transitioning {
		//		case "error":
		//			return errors.Errorf("Failed to launch stack %s. Error message: %s", stackID, stackData.TransitioningMessage)
		//		}
		//	}
		//} else if v.ResourceType == "serviceLog" {
		//	serviceLogData := &client.ServiceLog{}
		//	if err := unmarshalling(v.Data["resource"], serviceLogData); err != nil {
		//		logrus.Errorf("failed to unmarshalling err: %v", err)
		//		continue
		//	}
		//	if service, err := m.c.Service.ById(serviceLogData.ServiceId); err == nil {
		//		if service.StackId == stackID {
		//			msg := fmt.Sprintf("%s ServiceLog: %s", serviceLogData.Created, serviceLogData.Description)
		//			switch serviceLogData.Level {
		//			case "info":
		//				logrus.Infof(msg)
		//			case "error":
		//				logrus.Error(err)
		//				return err
		//			}
		//		}
		//	}
		//}
	}
}

func unmarshalling(data interface{}, v interface{}) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return errors.Wrapf(err, "failed to marshall object. Body: %v", data)
	}
	if err := json.Unmarshal(raw, &v); err != nil {
		return errors.Wrapf(err, "failed to unmarshall object. Body: %v", string(raw))
	}
	return nil
}
