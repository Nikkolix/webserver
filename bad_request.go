package webserver

import "net/http"

func (webServer *WebServer) BadRequest(rw http.ResponseWriter, msg string) {
	rw.WriteHeader(http.StatusBadRequest)
	_, err := rw.Write([]byte(msg))
	if err != nil {
		webServer.settings.Logger.Fatalln(err)
	}
}
