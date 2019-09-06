# beeureka
eureka client can be used in beego

### Usage

- Install dependency
```
go get github.com/zongkuiy/beeureka
```

- Import the beereka 
```
go get github.com/zongkuiy/beeureka
```

- Add configurations to the Beego app.conf 
```
[eureka]
defaultZone = http://eureka1:11111/eureka/,http://eureka2:11111/eureka/
```


- Init the beeureka client when starting the Beego

```
func main() {
  ...
	go RunEurekaService()
	beego.Run()
}
func RunEurekaService(){
	beeurekaClient := beeureka.BeeurekaClient{}
	beeurekaClient.Init()
}
```


  
