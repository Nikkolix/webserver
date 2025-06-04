package webserver

import (
	"errors"
	"io"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"fmt"
)

type HTTPMethod string

const (
	HTTPMethodGet     HTTPMethod = http.MethodGet
	HTTPMethodHead    HTTPMethod = http.MethodHead
	HTTPMethodPost    HTTPMethod = http.MethodPost
	HTTPMethodPut     HTTPMethod = http.MethodPut
	HTTPMethodPatch   HTTPMethod = http.MethodPatch
	HTTPMethodDelete  HTTPMethod = http.MethodDelete
	HTTPMethodConnect HTTPMethod = http.MethodConnect
	HTTPMethodOptions HTTPMethod = http.MethodOptions
	HTTPMethodTrace   HTTPMethod = http.MethodTrace
)

func init() {
	extToType := map[string]string{
		".aac":    "audio/aac",
		".abw":    "application/x-abiword",
		".arc":    "application/x-freearc",
		".avif":   "image/avif",
		".avi":    "video/x-msvideo",
		".azw":    "application/vnd.amazon.ebook",
		".bin":    "application/octet-stream",
		".bmp":    "image/bmp",
		".bz":     "application/x-bzip",
		".bz2":    "application/x-bzip2",
		".cda":    "application/x-cdf",
		".csh":    "application/x-csh",
		".css":    "text/css",
		".csv":    "text/csv",
		".doc":    "application/msword",
		".docx":   "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".eot":    "application/vnd.ms-fontobject",
		".epub":   "application/epub+zip",
		".gz":     "application/gzip",
		".gif":    "image/gif",
		".htm":    "text/html",
		".html":   "text/html",
		".ico":    "image/vnd.microsoft.icon",
		".ics":    "text/calendar",
		".jar":    "application/java-archive",
		".jpeg":   "image/jpeg",
		".jpg":    "image/jpeg",
		".js":     "text/javascript",
		".json":   "application/json",
		".jsonld": "application/ld+json",
		".mid":    "audio/midi",
		".midi":   "audio/midi",
		".mjs":    "text/javascript",
		".mp3":    "audio/mpeg",
		".mp4":    "video/mp4",
		".mpeg":   "video/mpeg",
		".mpkg":   "application/vnd.apple.installer+xml",
		".odp":    "application/vnd.oasis.opendocument.presentation",
		".ods":    "application/vnd.oasis.opendocument.spreadsheet",
		".odt":    "application/vnd.oasis.opendocument.text",
		".oga":    "audio/ogg",
		".ogv":    "video/ogg",
		".ogx":    "application/ogg",
		".opus":   "audio/opus",
		".otf":    "font/otf",
		".png":    "image/png",
		".pdf":    "application/pdf",
		".php":    "application/x-httpd-php",
		".ppt":    "application/vnd.ms-powerpoint",
		".pptx":   "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".rar":    "application/vnd.rar",
		".rtf":    "application/rtf",
		".sh":     "application/x-sh",
		".svg":    "image/svg+xml",
		".tar":    "application/x-tar",
		".tif":    "image/tiff",
		".tiff":   "image/tiff",
		".ts":     "video/mp2t",
		".ttf":    "font/ttf",
		".txt":    "text/plain",
		".vsd":    "application/vnd.visio",
		".wav":    "audio/wav",
		".weba":   "audio/webm",
		".webm":   "video/webm",
		".webp":   "image/webp",
		".woff":   "font/woff",
		".xhtml":  "application/xhtml+xml",
		".xls":    "application/vnd.ms-excel",
		".xlsx":   "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".xml":    "application/xml",
		".xul":    "application/vnd.mozilla.xul+xml",
		".zip":    "application/zip",
		".3gp":    "video/3gpp",
		".3g2":    "video/3gpp2",
		".7z":     "application/x-7z-compressed",
	}

	for ext, mimeType := range extToType {
		err := mime.AddExtensionType(ext, mimeType)
		if err != nil {
			panic(err)
		}
	}
}

//public

type WebServer struct {
	server *http.Server
	mux    *http.ServeMux

	getMux     *http.ServeMux
	headMux    *http.ServeMux
	postMux    *http.ServeMux
	putMux     *http.ServeMux
	patchMux   *http.ServeMux
	deleteMux  *http.ServeMux
	connectMux *http.ServeMux
	optionsMux *http.ServeMux
	traceMux   *http.ServeMux

	customMux *http.ServeMux

	settings Settings

	fileExtensionFilter []string

	middleware []func(http.ResponseWriter, *http.Request) bool
}

func NewWebServer(settings Settings) *WebServer {
	mux := http.NewServeMux()

	fmt.Println("init server:", settings.BindAddr())

	webServer := &WebServer{
		server: &http.Server{
			Handler: mux,
			Addr:    settings.BindAddr(),
		},
		mux: mux,

		getMux:     http.NewServeMux(),
		headMux:    http.NewServeMux(),
		postMux:    http.NewServeMux(),
		putMux:     http.NewServeMux(),
		patchMux:   http.NewServeMux(),
		deleteMux:  http.NewServeMux(),
		connectMux: http.NewServeMux(),
		optionsMux: http.NewServeMux(),
		traceMux:   http.NewServeMux(),

		customMux: http.NewServeMux(),

		settings: settings,

		fileExtensionFilter: []string{},
	}

	if webServer.settings.Logger == nil {
		webServer.settings.Logger = log.New(os.Stdout, "", log.LstdFlags)
	}

	webServer.mux.HandleFunc("/", webServer.mainHandler)
	webServer.getMux.HandleFunc("/", webServer.fileHandler)

	return webServer
}

func (webServer *WebServer) NewHandleFunc(method HTTPMethod, pattern string, handler func(http.ResponseWriter, *http.Request)) {
	switch method {
	case http.MethodGet:
		webServer.getMux.HandleFunc(pattern, handler)
	case http.MethodHead:
		webServer.headMux.HandleFunc(pattern, handler)
	case http.MethodPost:
		webServer.postMux.HandleFunc(pattern, handler)
	case http.MethodPut:
		webServer.putMux.HandleFunc(pattern, handler)
	case http.MethodPatch:
		webServer.patchMux.HandleFunc(pattern, handler)
	case http.MethodDelete:
		webServer.deleteMux.HandleFunc(pattern, handler)
	case http.MethodConnect:
		webServer.connectMux.HandleFunc(pattern, handler)
	case http.MethodOptions:
		webServer.optionsMux.HandleFunc(pattern, handler)
	case http.MethodTrace:
		webServer.traceMux.HandleFunc(pattern, handler)
	default:
		webServer.customMux.HandleFunc(pattern, handler)
	}
}

func (webServer *WebServer) NewHandlerBody(method HTTPMethod, pattern string, handler func(http.ResponseWriter, *http.Request, []byte)) {
	webServer.NewHandleFunc(method, pattern, func(rw http.ResponseWriter, req *http.Request) {
		bodyData, err := io.ReadAll(req.Body)
		if err != nil {
			panic(err)
		}

		err = req.Body.Close()
		if err != nil {
			panic(err)
		}
		handler(rw, req, bodyData)
	})
}

func (webServer *WebServer) NewHandler(method HTTPMethod, pattern string, handler http.Handler) {
	switch method {
	case http.MethodGet:
		webServer.getMux.Handle(pattern, handler)
	case http.MethodHead:
		webServer.headMux.Handle(pattern, handler)
	case http.MethodPost:
		webServer.postMux.Handle(pattern, handler)
	case http.MethodPut:
		webServer.putMux.Handle(pattern, handler)
	case http.MethodPatch:
		webServer.patchMux.Handle(pattern, handler)
	case http.MethodDelete:
		webServer.deleteMux.Handle(pattern, handler)
	case http.MethodConnect:
		webServer.connectMux.Handle(pattern, handler)
	case http.MethodOptions:
		webServer.optionsMux.Handle(pattern, handler)
	case http.MethodTrace:
		webServer.traceMux.Handle(pattern, handler)
	default:
		webServer.customMux.Handle(pattern, handler)
	}
}

func (webServer *WebServer) NewMiddleware(m func(http.ResponseWriter, *http.Request) bool) {
	webServer.middleware = append(webServer.middleware, m)
}

func (webServer *WebServer) SetRoot(root string) {
	webServer.settings.Root = root
}

func (webServer *WebServer) SetFileExtensionsFilter(fileExtensions ...string) {
	for _, extension := range fileExtensions {
		if !slices.Contains(webServer.fileExtensionFilter, extension) {
			webServer.fileExtensionFilter = append(webServer.fileExtensionFilter, extension)
		}

	}
}

func (webServer *WebServer) Run() error {
	if webServer.settings.UseHttps {
		if webServer.settings.UseHttpRedirect {
			m := http.NewServeMux()
			m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				urlPath := "https://" + webServer.settings.Domain + ":" + webServer.settings.HttpsPort + r.URL.Path
				http.Redirect(w, r, urlPath, http.StatusMovedPermanently)
				webServer.settings.Logger.Println("Redirect: http to https 301 to " + urlPath)
			})
			s := http.Server{
				Addr:    ":" + "80",
				Handler: m,
			}
			go func() {
				err := s.ListenAndServe()
				if err != nil {
					panic(err)
				}
			}()
		}
		webServer.settings.Logger.Println("WebServer running on " + webServer.settings.Url())
		return webServer.server.ListenAndServeTLS(webServer.settings.CertFile, webServer.settings.KeyFile)
	} else {
		webServer.settings.Logger.Println("WebServer running on " + webServer.settings.Url())
		return webServer.server.ListenAndServe()
	}
}

//private

func (webServer *WebServer) fallbackRedirect(rw http.ResponseWriter, req *http.Request) {
	urlPath := "http://" + webServer.settings.Domain + ":" + webServer.settings.HttpPort + webServer.settings.FallbackRedirect
	if webServer.settings.UseHttps {
		urlPath = "https://" + webServer.settings.Domain + ":" + webServer.settings.HttpsPort + webServer.settings.FallbackRedirect

	}
	http.Redirect(rw, req, urlPath, http.StatusTemporaryRedirect)
	webServer.settings.Logger.Println("Fallback Redirect to " + urlPath)
}

func (webServer *WebServer) fileHandler(rw http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	parts := strings.Split(path, ".")
	fileExtension := parts[len(parts)-1]

	if slices.Contains(webServer.fileExtensionFilter, fileExtension) {
		rw.WriteHeader(http.StatusForbidden)
		webServer.settings.Logger.Println("File Handler: 403: " + fileExtension + " (" + path + ")")
		return
	}

	joinPath, err := url.JoinPath(webServer.settings.Root, path)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		webServer.settings.Logger.Println("File Handler: 500: can't join root " + webServer.settings.Root + " and path " + path)
		return
	}

	file, err := os.ReadFile(joinPath)
	if err != nil {
		var pathError *fs.PathError
		if errors.As(err, &pathError) {
			webServer.settings.Logger.Println("File Handler: 404: " + pathError.Error())
			if fileExtension == "html" || fileExtension == "" || len(parts) == 1 {
				webServer.fallbackRedirect(rw, req)
			} else {
				rw.WriteHeader(http.StatusNotFound)
				write, err := rw.Write([]byte{})
				if err != nil {
					webServer.settings.Logger.Println("File Handler: Write Error: " + err.Error() + " (" + strconv.Itoa(write) + " bytes send)")
				}
			}
			return
		} else {
			rw.WriteHeader(http.StatusInternalServerError)
			webServer.settings.Logger.Println("File Handler: 500: " + err.Error())
			return
		}
	}

	rw.Header().Set("Content-Type", mime.TypeByExtension("."+fileExtension))
	rw.WriteHeader(http.StatusOK)
	bytes, err := rw.Write(file)
	if err != nil {
		webServer.settings.Logger.Println("File Handler: Write Error: " + err.Error() + " (" + strconv.Itoa(bytes) + "/" + strconv.Itoa(len(file)) + ")")
	} else {
		webServer.settings.Logger.Println("File Handler: 200: " + path)
	}
}

func (webServer *WebServer) mainHandler(rw http.ResponseWriter, req *http.Request) {
	webServer.settings.Logger.Println(req.Method, req.URL, req.ContentLength)

	for _, m := range webServer.middleware {
		if !m(rw, req) {
			return
		}
	}

	switch strings.ToUpper(req.Method) {
	case http.MethodGet:
		webServer.getMux.ServeHTTP(rw, req)
	case http.MethodHead:
		webServer.headMux.ServeHTTP(rw, req)
	case http.MethodPost:
		webServer.postMux.ServeHTTP(rw, req)
	case http.MethodPut:
		webServer.putMux.ServeHTTP(rw, req)
	case http.MethodPatch:
		webServer.patchMux.ServeHTTP(rw, req)
	case http.MethodDelete:
		webServer.deleteMux.ServeHTTP(rw, req)
	case http.MethodConnect:
		webServer.connectMux.ServeHTTP(rw, req)
	case http.MethodOptions:
		webServer.optionsMux.ServeHTTP(rw, req)
	case http.MethodTrace:
		webServer.traceMux.ServeHTTP(rw, req)
	default:
		webServer.customMux.ServeHTTP(rw, req)
	}
}
