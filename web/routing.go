package web

func (s webServer) initRouting() {

	s.initWebsocketRoutes()

	s.initStaticRoutes()

}
