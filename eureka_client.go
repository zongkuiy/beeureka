package beeureka

import (
	"encoding/xml"
	"github.com/astaxie/beego"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

type EurekaClient struct {
	AppMap sync.Map

	communicator EurekaCommunicator

	State string
	RetryRegisterIntervalInSecs int64
	CommunicateWithServerSuccessTimestamp time.Time
}

func (this *EurekaClient) Init(appname string, httpport string, httpsport string, hostname string,
			durationInSecs int, renewalIntervalInSecs int, zones []string){

	log.Println("Init with parameters appname =",appname, "httpport =", httpport, "httpsport =", httpsport, "hostname =",  hostname)

	this.State = EUREKA_CLIENT_STATUS_STARTING
	this.RetryRegisterIntervalInSecs = 2

	this.communicator = EurekaCommunicator{
		AppName: appname,
		HttpPort: httpport,
		HttpsPort: httpsport,
		HostName: hostname,
		DurationInSecs: durationInSecs,

		RenewalIntervalInSecs: renewalIntervalInSecs,
		Zones: zones,
	}

	this.communicator.Init()

	go this.handleExitSignal()

	go this.listenCommunicatorResponse()

	this.start()
}
func (this *EurekaClient) listenCommunicatorResponse() {
	for {
		response := <- this.communicator.responseChan

		if response.Err == nil {
			this.CommunicateWithServerSuccessTimestamp = time.Now()
		}

		switch response.Action {
			case EUREKA_MSG_REGISTER:
				if response.Err != nil {
					this.sendMsg(EUREKA_MSG_REGISTER)
					break
				}
				if response.ResponseCode == 204 {
					this.State = EUREKA_CLIENT_STATUS_UP

					log.Printf("register succeed")

					go this.startHeartbeat()
					go this.startQueryApps()
				}
			case EUREKA_MSG_DEREGISTER:
				this.State = EUREKA_CLIENT_STATUS_DOWN
			case EUREKA_MSG_HEARTBEAT:
				if response.Err != nil {
					this.sendMsg(EUREKA_MSG_HEARTBEAT)
				} else {
					if response.ResponseCode == 404 {
						this.State = EUREKA_CLIENT_STATUS_STARTING
						this.sendMsg(EUREKA_MSG_REGISTER)
					}
				}
			case EUREKA_MSG_FULL_REFRESH:
				if response.Err == nil && response.ResponseCode == 200 {

					allAppsResponse := AllApplicationResponse{}

					err := xml.Unmarshal([]byte(response.Body), &allAppsResponse.Applications)

					if err == nil {
						for _, app := range *&allAppsResponse.Applications.Application {
							var cis [...] ClientInstance
							for idx, inst := range app.Instance{
								log.Printf(inst.InstanceID, inst.Status, inst.IPAddr, inst.Port, inst.SecurePort)
								cis[idx] = ClientInstance{
									inst.InstanceID,
									inst.Status,
									inst.IPAddr,
									strings.EqualFold(inst.Port.Enabled, "true"),
									inst.Port.Text,
									strings.EqualFold(inst.SecurePort.Enabled, "true"),
									inst.SecurePort.Text,
								}
							}
							this.AppMap.Store(app.Name, cis)
						}
					}


				}
			case EUREKA_MSG_DELTA_REFRESH:
		}
	}
}
func (this *EurekaClient) start(){
	this.sendMsg(EUREKA_MSG_REGISTER)
}

func (this *EurekaClient) sendMsg(msg int){
	select {
	case this.communicator.msgChan <- msg:
	default:
	}
}


func (this *EurekaClient) startHeartbeat() {
	heartbeat := time.Tick(10 * time.Second)
	for {
		select {
			case <- heartbeat:
				this.sendMsg(EUREKA_MSG_HEARTBEAT)
				break

		}
	}
}

func (this *EurekaClient) handleExitSignal() {
	sigs := make(chan os.Signal,1)

	// register notifications of the specified signals
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// receive signals and mark the quit flag
	go func() {
		sig := <-sigs
		beego.Info("received signal", sig)
		this.deregisterEureka()
	}()

}

func (this *EurekaClient) deregisterEureka() {
	this.sendMsg(EUREKA_MSG_DEREGISTER)
}


func (this *EurekaClient) startQueryApps() {
	queryFlag := time.Tick(10 * time.Second)
	for {
		select {
		case <- queryFlag:
			this.sendMsg(EUREKA_MSG_FULL_REFRESH)
			break
		}
	}
}

func (this *EurekaClient) GetServiceHttpUrl(appname string) (bool, string){
	ok, inst := this.getActServiceInstance(appname, "http")
	if ok {
		return true, "http://" + inst.Host + ":" + inst.HttpPort
	}
	return false, ""
}

func (this *EurekaClient) GetServiceHttpsUrl(appname string) (bool, string){
	ok, inst := this.getActServiceInstance(appname, "https")
	if ok {
		return true, "https://" + inst.Host + ":" + inst.HttpPort
	}
	return false, ""
}


func (this *EurekaClient) getActServiceInstance(appname string, protocalType string) (bool, ClientInstance){
	cis, ok := this.AppMap.Load(appname)
	if ok {
		for _, inst := range cis.([]ClientInstance){
			if strings.EqualFold(inst.Status, EUREKA_CLIENT_STATUS_UP) {
				if (strings.EqualFold(protocalType, "http") && inst.HttpEnabled) || (strings.EqualFold(protocalType, "https") && inst.HttpsEnabled) {
					return true, inst
				}
			}
		}
	}
	return false, ClientInstance{}
}

