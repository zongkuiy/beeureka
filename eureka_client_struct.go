package beeureka

/*
	data structure for communicate with eureka server
 */

type ClientApp struct {
	Name string
}

type ClientInstance struct {
	Id string
	Status string
	Host string
	HttpEnabled bool
	HttpPort string
	HttpsEnabled bool
	HttpsPort string
}


