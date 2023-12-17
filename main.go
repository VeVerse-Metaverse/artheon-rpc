package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)
import "dev.hackerman.me/artheon/artheon-rpc/web"

func main() {
	_ = viper.BindEnv("web.host", "WEB_HOST")
	_ = viper.BindEnv("web.port", "WEB_PORT")

	viper.SetDefault("web.host", "0.0.0.0")
	viper.SetDefault("web.port", "8080")

	//todo debug
	//c, err := models.RequestKick(models.VivoxTokenPayload{})
	//if err != nil {
	//     fmt.Printf("%s", err.Error())
	//} else {
	//     fmt.Printf("%s", c)
	//}

	webHost, webPort :=
		viper.GetString("web.host"),
		viper.GetString("web.port")

	webServer := web.NewWebServer(webHost, webPort)

	log.Infof("Starting web server... Host:%s, Port:%s", webHost, webPort)

	webServer.Start()
}
