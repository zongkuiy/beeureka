package beeureka

import (
	"github.com/astaxie/beego"
	"strings"
)

type BeeurekaClient struct {
}
func (this *BeeurekaClient) Init(){
	zoneList := beego.AppConfig.String("eureka::defaultZone")

	eurekaCient := EurekaClient{}

	eurekaCient.Init(
		beego.AppConfig.String("appname"),
		beego.AppConfig.String("httpport"),
		beego.AppConfig.String("httpsport"),
		beego.AppConfig.String("hostname"),
		beego.AppConfig.DefaultInt("durationInSecs", 20),

		beego.AppConfig.DefaultInt("renewalIntervalInSecs", 10),
		strings.Split(zoneList, ","),
	)
}
