package beeureka

/*
	data structures to communicate with eureka server
 */

type Port struct {
	Enabled string `xml:"enabled,attr"`
	Text    string `xml:",innerxml"`
}
type SecurePort struct {
	Enabled string `xml:"enabled,attr"`
	Text    string `xml:",innerxml"`
}
type DataCenterInfo struct {
	Name string `xml:"name"`
}
type LeaseInfo struct {
	RenewalIntervalInSecs string `xml:"renewalIntervalInSecs"`
	DurationInSecs        string `xml:"durationInSecs"`
}
type Metadata struct {
	//ManagementPort string `xml:"management.port"`
	ManagementPort string `xml:"-"`
}
type instance struct {
	InstanceID                    string         `xml:"instanceId"`
	HostName                      string         `xml:"hostName"`
	App                           string         `xml:"app"`
	IPAddr                        string         `xml:"ipAddr"`
	Status                        string         `xml:"status"`
	Overriddenstatus              string         `xml:"overriddenstatus"`
	Port                          Port           `xml:"port"`
	SecurePort                    SecurePort     `xml:"securePort"`
	CountryID                     string         `xml:"countryId"`
	DataCenterInfo                DataCenterInfo `xml:"dataCenterInfo"`
	LeaseInfo                     LeaseInfo      `xml:"leaseInfo"`
	Metadata                      Metadata       `xml:"metadata"`
	HomePageURL                   string         `xml:"-"`
	StatusPageURL                 string         `xml:"-"`
	HealthCheckURL                string         `xml:"-"`
	VipAddress                    string         `xml:"-"`
	SecureVipAddress              string         `xml:"-"`
	IsCoordinatingDiscoveryServer string         `xml:"isCoordinatingDiscoveryServer"`
	ActionType                    string         `xml:"actionType"`
}

type AllApplicationResponse struct {
	Applications Applications `xml:"applications"`
}
type Application struct {
	Name     string        `xml:"name"`
	Instance []instance `xml:"instance"`
}
type Applications struct {
	VersionsDelta string      `xml:"versions__delta"`
	AppsHashcode  string      `xml:"apps__hashcode"`
	Application   []Application `xml:"application"`
}