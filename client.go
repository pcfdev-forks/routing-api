package routing_api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"code.cloudfoundry.org/routing-api/models"
	trace "code.cloudfoundry.org/trace-logger"
	"github.com/tedsuo/rata"
	"github.com/vito/go-sse/sse"
)

const (
	defaultMaxRetries = uint16(0)
)

//go:generate counterfeiter -o fake_routing_api/fake_client.go . Client
type Client interface {
	SetToken(string)
	UpsertRoutes([]models.Route) error
	Routes() ([]models.Route, error)
	DeleteRoutes([]models.Route) error
	RouterGroups() ([]models.RouterGroup, error)
	UpdateRouterGroup(models.RouterGroup) error
	UpsertTcpRouteMappings([]models.TcpRouteMapping) error
	DeleteTcpRouteMappings([]models.TcpRouteMapping) error
	TcpRouteMappings() ([]models.TcpRouteMapping, error)

	SubscribeToEvents() (EventSource, error)
	SubscribeToEventsWithMaxRetries(retries uint16) (EventSource, error)
	SubscribeToTcpEvents() (TcpEventSource, error)
	SubscribeToTcpEventsWithMaxRetries(retries uint16) (TcpEventSource, error)
}

func NewClient(url string, skipTLSVerification bool) Client {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: skipTLSVerification,
	}
	httpClient := &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).Dial,
			TLSClientConfig: tlsConfig,
		},
	}
	streamingClient := &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSClientConfig: tlsConfig,
		},
	}

	return &client{
		httpClient:          httpClient,
		streamingHTTPClient: streamingClient,

		tokenMutex: &sync.RWMutex{},

		reqGen: rata.NewRequestGenerator(url, Routes()),
	}
}

type client struct {
	httpClient          *http.Client
	streamingHTTPClient *http.Client

	tokenMutex *sync.RWMutex
	authToken  string

	reqGen *rata.RequestGenerator
}

func (c *client) SetToken(token string) {
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()
	c.authToken = token
}

func (c *client) UpsertRoutes(routes []models.Route) error {
	return c.doRequest(UpsertRoute, nil, nil, routes, nil)
}

func (c *client) Routes() ([]models.Route, error) {
	var routes []models.Route
	err := c.doRequest(ListRoute, nil, nil, nil, &routes)
	return routes, err
}

func (c *client) UpdateRouterGroup(group models.RouterGroup) error {
	return c.doRequest(UpdateRouterGroup, rata.Params{"guid": group.Guid}, nil, group, nil)
}

func (c *client) RouterGroups() ([]models.RouterGroup, error) {
	var routerGroups []models.RouterGroup
	err := c.doRequest(ListRouterGroups, nil, nil, nil, &routerGroups)
	return routerGroups, err
}

func (c *client) DeleteRoutes(routes []models.Route) error {
	return c.doRequest(DeleteRoute, nil, nil, routes, nil)
}

func (c *client) UpsertTcpRouteMappings(tcpRouteMappings []models.TcpRouteMapping) error {
	return c.doRequest(UpsertTcpRouteMapping, nil, nil, tcpRouteMappings, nil)
}

func (c *client) TcpRouteMappings() ([]models.TcpRouteMapping, error) {
	var tcpRouteMappings []models.TcpRouteMapping
	err := c.doRequest(ListTcpRouteMapping, nil, nil, nil, &tcpRouteMappings)
	return tcpRouteMappings, err
}

func (c *client) DeleteTcpRouteMappings(tcpRouteMappings []models.TcpRouteMapping) error {
	return c.doRequest(DeleteTcpRouteMapping, nil, nil, tcpRouteMappings, nil)
}

func (c *client) SubscribeToEvents() (EventSource, error) {
	eventSource, err := c.doSubscribe(EventStreamRoute, defaultMaxRetries)
	if err != nil {
		return nil, err
	}
	return NewEventSource(eventSource), nil
}

func (c *client) SubscribeToTcpEvents() (TcpEventSource, error) {
	eventSource, err := c.doSubscribe(EventStreamTcpRoute, defaultMaxRetries)
	if err != nil {
		return nil, err
	}
	return NewTcpEventSource(eventSource), nil
}

func (c *client) SubscribeToEventsWithMaxRetries(retries uint16) (EventSource, error) {
	eventSource, err := c.doSubscribe(EventStreamRoute, retries)
	if err != nil {
		return nil, err
	}
	return NewEventSource(eventSource), nil
}

func (c *client) SubscribeToTcpEventsWithMaxRetries(retries uint16) (TcpEventSource, error) {
	eventSource, err := c.doSubscribe(EventStreamTcpRoute, retries)
	if err != nil {
		return nil, err
	}
	return NewTcpEventSource(eventSource), nil
}

func (c *client) doSubscribe(routeName string, retries uint16) (RawEventSource, error) {
	config := sse.Config{
		Client: c.streamingHTTPClient,
		RetryParams: sse.RetryParams{
			MaxRetries:    retries,
			RetryInterval: time.Second,
		},
		RequestCreator: func() *http.Request {
			request, err := c.reqGen.CreateRequest(routeName, nil, nil)
			c.tokenMutex.RLock()
			defer c.tokenMutex.RUnlock()
			request.Header.Add("Authorization", "bearer "+c.authToken)
			if err != nil {
				panic(err) // totally shouldn't happen
			}

			trace.DumpRequest(request)
			return request
		},
	}
	eventSource, err := config.Connect()
	if err != nil {
		bre, ok := err.(sse.BadResponseError)
		if ok && bre.Response.StatusCode == http.StatusUnauthorized {
			return nil, Error{Type: "unauthorized", Message: "unauthorized"}
		}
		return nil, err
	}

	return eventSource, nil
}

func (c *client) createRequest(requestName string, params rata.Params, queryParams url.Values, request interface{}) (*http.Request, error) {
	requestJson, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := c.reqGen.CreateRequest(requestName, params, bytes.NewReader(requestJson))
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = queryParams.Encode()
	req.ContentLength = int64(len(requestJson))
	req.Header.Set("Content-Type", "application/json")
	c.tokenMutex.RLock()
	defer c.tokenMutex.RUnlock()
	req.Header.Add("Authorization", "bearer "+c.authToken)

	return req, nil
}

func (c *client) doRequest(requestName string, params rata.Params, queryParams url.Values, request, response interface{}) error {
	req, err := c.createRequest(requestName, params, queryParams, request)
	if err != nil {
		return err
	}
	return c.do(req, response)
}

func (c *client) do(req *http.Request, response interface{}) error {
	trace.DumpRequest(req)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = res.Body.Close()
	}()

	trace.DumpResponse(res)

	if res.StatusCode == http.StatusUnauthorized {
		return NewError(UnauthorizedError, "unauthorized")
	}

	if res.StatusCode > 299 {
		return transformResponseError(res)
	}

	if response != nil {
		return json.NewDecoder(res.Body).Decode(response)
	}

	return nil
}

func transformResponseError(res *http.Response) error {
	errResponse := Error{}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return NewError(ResponseError, "failed to read response body")
	}

	err = json.Unmarshal(data, &errResponse)
	if err != nil {
		return NewError(ResponseError, string(data))
	}

	if errResponse.Type == "" {
		return NewError(ResponseError, string(data))
	}
	return errResponse
}
