package beeureka

import (
	"encoding/xml"
	"github.com/astaxie/beego"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	EUREKA_MSG_REGISTER = iota
	EUREKA_MSG_HEARTBEAT
	EUREKA_MSG_FULL_REFRESH
	EUREKA_MSG_DELTA_REFRESH
	EUREKA_MSG_DEREGISTER
)
const (
	EUREKA_HTTP_POST = "POST"
	EUREKA_HTTP_GET = "GET"
	EUREKA_HTTP_PUT = "PUT"
	EUREKA_HTTP_DELETE = "DELETE"
)

const (
	EUREKA_CLIENT_STATUS_UP = "UP"
	EUREKA_CLIENT_STATUS_DOWN = "DOWN"
	EUREKA_CLIENT_STATUS_STARTING = "STARTING"
	EUREKA_CLIENT_STATUS_OUT_OF_SERVICE = "OUT_OF_SERVIE"
	EUREKA_CLIENT_STATUS_UNKOWN = "UNKNOWN"
)

const (
	EUREKA_CLIENT_ACTION_TYPE_ADDED = "ADDED"
	EUREKA_CLIENT_ACTION_TYPE_MODIFIED = "MODIFIED"
)

const (
	NOT_AWS_DATACENTER_NAME = "MyOwn"
)

type EurekaCommunicator struct {
	AppName string
	HostName string
	HttpPort string
	HttpsPort string
	Zones []string
	CurrentZoneIndex int
	CurrentZoneFailureCount int

	RenewalIntervalInSecs int
	DurationInSecs int

	Applications *[]Application
	AppMap sync.Map


	msgChan chan int
	responseChan chan EurekaHttpResponse

	lock sync.Mutex
}

type EurekaHttpRequest struct {
	HttpMethod string
	Url string
	Body string
}

type EurekaHttpResponse struct {
	Err error
	Action int
	ResponseCode int
	Body string
}

func (this *EurekaCommunicator) getInstanceId() string {
	return this.HostName + ":" + this.AppName + ":" + this.HttpPort
}
func (this *EurekaCommunicator) Init(){

	log.Printf("EurekaCommunicator Init")

	this.msgChan = make(chan int, 5)
	this.responseChan = make(chan EurekaHttpResponse, 5)

	go this.ListenMessage()
}

func (this *EurekaCommunicator) ListenMessage() {

	for {
		msg := <- this.msgChan

		go this.doProcess(msg)
	}
}

func (this *EurekaCommunicator) doProcess(msg int) {

	req := this.makeReq(msg)
	err, rc, content := this.sendMsg(req)

	if err != nil {
		this.tryFailover()
		this.responseChan <- EurekaHttpResponse{
			Err: err,
			Action: msg,
			Body: content,
			ResponseCode: rc,
		}
	} else {
		this.CurrentZoneFailureCount = 0
		this.responseChan <- EurekaHttpResponse{
			Err: nil,
			Action: msg,
			Body: content,
			ResponseCode: rc,
		}
	}

}

func (this *EurekaCommunicator) heartbeat() {
	req := this.makeHeartbeatReq()
	this.sendMsg(req)
}

func (this *EurekaCommunicator) fullRefresh() {
	req := this.makeFullRefreshReq()
	this.sendMsg(req)
}

func (this *EurekaCommunicator) deltaRefresh() {
	
}

func (this *EurekaCommunicator) deregister() {
	req := this.makeDeregisterReq()
	this.sendMsg(req)
}

func (this *EurekaCommunicator) GetCurrentZone() string {
	return this.Zones[this.CurrentZoneIndex]
}

func (this *EurekaCommunicator) sendMsg(reqHttp EurekaHttpRequest) (error, int, string) {

	log.Printf("sendMsg " + reqHttp.Url)

	payload := strings.NewReader(reqHttp.Body)

	req, _ := http.NewRequest(reqHttp.HttpMethod, reqHttp.Url, payload)

	req.Header.Add("Content-Type", "application/xml")
	req.Header.Add("Accept", "application/xml")

	http.DefaultClient.Timeout = time.Second * 5

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		beego.Error(err)
		return err, 0, ""
	}

	defer res.Body.Close()

	resBody, err := ioutil.ReadAll(res.Body)

	if err == nil {
		resContent := (string(resBody))
		return nil, res.StatusCode, resContent
	} else {
		beego.Error(err)
		return err, 0, ""
	}

}

func (this *EurekaCommunicator) makeRegisterReq() (EurekaHttpRequest) {
	enabledSecurePort := "true"

	if len(this.HttpsPort) == 0 {
		enabledSecurePort = "false"

		// give a default value to make the eureka server happy
		this.HttpsPort = "443"
	}
	req := instance {
		InstanceID: this.getInstanceId(),
		HostName: this.HostName,
		App: this.AppName,
		IPAddr: this.HostName,
		Status: EUREKA_CLIENT_STATUS_UP,
		Port: Port{
			Enabled: "true",
			Text: this.HttpPort,
		},
		SecurePort: SecurePort{
			Enabled: enabledSecurePort,
			Text: this.HttpsPort,
		},
		CountryID: "1",
		DataCenterInfo:DataCenterInfo{
			Name: NOT_AWS_DATACENTER_NAME,
		},
		LeaseInfo: LeaseInfo{
			RenewalIntervalInSecs: strconv.Itoa(this.RenewalIntervalInSecs),
			DurationInSecs: strconv.Itoa(this.DurationInSecs),
		},
		Metadata:Metadata{
			ManagementPort: "",
		},
		HomePageURL: "",
		StatusPageURL: "",
		HealthCheckURL: "",
		VipAddress: "",
		SecureVipAddress: "",
		IsCoordinatingDiscoveryServer: "false",
		ActionType: EUREKA_CLIENT_ACTION_TYPE_ADDED,

	}

	str, _ := xml.MarshalIndent(&req, " ", " ")

	_req := EurekaHttpRequest{}

	_req.Body = string(str)
	_req.Url = this.GetCurrentZone() + "apps/" + this.AppName
	_req.HttpMethod = EUREKA_HTTP_POST
	return _req
}
func (this *EurekaCommunicator) makeDeregisterReq() EurekaHttpRequest {
	deregisterReq := EurekaHttpRequest{
		HttpMethod: EUREKA_HTTP_DELETE,
		Url: this.GetCurrentZone() + "apps/" + this.AppName + "/" + this.getInstanceId(),
		Body: "",
	}

	return deregisterReq
}

func (this *EurekaCommunicator) makeHeartbeatReq() EurekaHttpRequest {
	heartbeatReq := EurekaHttpRequest{
		HttpMethod: EUREKA_HTTP_PUT,
		Url:        this.GetCurrentZone() + "apps/" + this.AppName + "/" + this.getInstanceId(),
		Body:       "",
	}

	return heartbeatReq
}
func (this *EurekaCommunicator) makeFullRefreshReq() EurekaHttpRequest{

	fullRefreshReq := EurekaHttpRequest{
		HttpMethod: EUREKA_HTTP_GET,
		Url:        this.GetCurrentZone() + "apps",
		Body:       "",
	}

	return fullRefreshReq
}

func (this *EurekaCommunicator) makeDeltaRefreshReq() EurekaHttpRequest {
	return EurekaHttpRequest{}
}


func (this *EurekaCommunicator) makeReq(msg int) EurekaHttpRequest {
	switch msg {
		case EUREKA_MSG_REGISTER:
			return this.makeRegisterReq()
		case EUREKA_MSG_DEREGISTER:
			return this.makeDeregisterReq()
		case EUREKA_MSG_HEARTBEAT:
			return this.makeHeartbeatReq()
		case EUREKA_MSG_FULL_REFRESH:
			return this.makeFullRefreshReq()
		case EUREKA_MSG_DELTA_REFRESH:
			return this.makeDeltaRefreshReq()
	}

	return EurekaHttpRequest{}
}

func (this *EurekaCommunicator) doFailover() {

	this.CurrentZoneIndex = this.CurrentZoneIndex + 1
	if this.CurrentZoneIndex > len(this.Zones) - 1 {
		this.CurrentZoneIndex = 0
	}

}

// three failover attempts will trigger the actually failover
func (this *EurekaCommunicator) tryFailover() {
	this.lock.Lock()
	this.CurrentZoneFailureCount ++

	if this.CurrentZoneFailureCount > 3 {
		this.doFailover()
		this.CurrentZoneFailureCount = 0;
	}
	this.lock.Unlock()
}

