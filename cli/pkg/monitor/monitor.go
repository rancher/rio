package monitor

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/patrickmn/go-cache"
	"github.com/rancher/rio/types/client/rio/v1beta1"
	"github.com/sirupsen/logrus"
)

type Event struct {
	Name         string                 `json:"name"`
	ResourceType string                 `json:"resourceType"`
	ResourceID   string                 `json:"resourceId"`
	Data         map[string]interface{} `json:"data"`
}

type Monitor struct {
	sync.Mutex
	c             *client.Client
	cache         *cache.Cache
	subCounter    int
	subscriptions map[int]*Subscription
}

func (m *Monitor) Subscribe() *Subscription {
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

func (m *Monitor) Unsubscribe(sub *Subscription) {
	m.Lock()
	defer m.Unlock()

	close(sub.C)
	delete(m.subscriptions, sub.id)
}

type Subscription struct {
	id int
	C  chan *Event
}

func New(c *client.Client) *Monitor {
	return &Monitor{
		c:             c,
		cache:         cache.New(5*time.Minute, 30*time.Second),
		subscriptions: map[int]*Subscription{},
	}
}

func (m *Monitor) Start() error {
	schema, ok := m.c.Types["subscribe"]
	if !ok {
		return fmt.Errorf("not authorized to subscribe")
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
		return fmt.Errorf("bad status code: %d %s", resp.StatusCode, resp.Status)
	}

	logrus.Debugf("Connected to: %s", u.String())

	return m.watch(conn)
}

func (m *Monitor) Get(resourceType, resourceID string, obj interface{}) (bool, error) {
	val, ok := m.cache.Get(key(resourceType, resourceID))
	if !ok {
		return ok, nil
	}

	if val == nil {
		return true, nil
	}

	content, err := json.Marshal(val)
	if err != nil {
		return ok, err
	}

	return true, json.Unmarshal(content, obj)
}

func key(a, b string) string {
	return fmt.Sprintf("%s:%s", a, b)
}

func (m *Monitor) put(resourceType, resourceID string, event *Event) {
	if resourceType == "" && resourceID == "" {
		return
	}

	m.cache.Replace(key(resourceType, resourceID), event.Data["resource"], cache.DefaultExpiration)

	m.Lock()
	defer m.Unlock()

	for _, sub := range m.subscriptions {
		sub.C <- event
	}
}

func (m *Monitor) watch(conn *websocket.Conn) error {
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

		logrus.Debugf("Event: %s %s %s %v", v.Name, v.ResourceType, v.ResourceID, v.Data)
		m.put(v.ResourceType, v.ResourceID, &v)
	}
}
