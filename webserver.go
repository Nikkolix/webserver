package webserver

import (
	"errors"
	"golang.org/x/exp/slices"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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

//helper

func getMimeType(fileExtension string) string {
	switch strings.ToLower(fileExtension) {
	case "aac":
		return "audio/aac"
	case "abw":
		return "application/x-abiword"
	case "arc":
		return "application/x-freearc"
	case "avif":
		return "image/avif"
	case "avi":
		return "video/x-msvideo"
	case "azw":
		return "application/vnd.amazon.ebook"
	case "bin":
		return "application/octet-stream"
	case "bmp":
		return "image/bmp"
	case "bz":
		return "application/x-bzip"
	case "bz2":
		return "application/x-bzip2"
	case "cda":
		return "application/x-cdf"
	case "csh":
		return "application/x-csh"
	case "css":
		return "text/css"
	case "csv":
		return "text/csv"
	case "doc":
		return "application/msword"
	case "docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case "eot":
		return "application/vnd.ms-fontobject"
	case "epub":
		return "application/epub+zip"
	case "gz":
		return "application/gzip"
	case "gif":
		return "image/gif"
	case "htm", "html":
		return "text/html"
	case "ico":
		return "image/vnd.microsoft.icon"
	case "ics":
		return "text/calendar"
	case "jar":
		return "application/java-archive"
	case "jpeg", "jpg":
		return "image/jpeg"
	case "js":
		return "text/javascript"
	case "json":
		return "application/json"
	case "jsonld":
		return "application/ld+json"
	case "mid", "midi":
		return "audio/midi" //audio/x-midi
	case "mjs":
		return "text/javascript"
	case "mp3":
		return "audio/mpeg"
	case "mp4":
		return "video/mp4"
	case "mpeg":
		return "video/mpeg"
	case "mpkg":
		return "application/vnd.apple.installer+xml"
	case "odp":
		return "application/vnd.oasis.opendocument.presentation"
	case "ods":
		return "application/vnd.oasis.opendocument.spreadsheet"
	case "odt":
		return "application/vnd.oasis.opendocument.text"
	case "oga":
		return "audio/ogg"
	case "ogv":
		return "video/ogg"
	case "ogx":
		return "application/ogg"
	case "opus":
		return "audio/opus"
	case "otf":
		return "font/otf"
	case "png":
		return "image/png"
	case "pdf":
		return "application/pdf"
	case "php":
		return "application/x-httpd-php"
	case "ppt":
		return "application/vnd.ms-powerpoint"
	case "pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case "rar":
		return "application/vnd.rar"
	case "rtf":
		return "application/rtf"
	case "sh":
		return "application/x-sh"
	case "svg":
		return "image/svg+xml"
	case "tar":
		return "application/x-tar"
	case "tif", "tiff":
		return "image/tiff"
	case "ts":
		return "video/mp2t"
	case "ttf":
		return "font/ttf"
	case "txt":
		return "text/plain"
	case "vsd":
		return "application/vnd.visio"
	case "wav":
		return "audio/wav"
	case "weba":
		return " audio/webm"
	case "webm":
		return "video/webm"
	case "webp":
		return "image/webp"
	case "woff":
		return "font/woff"
	case "xhtml":
		return "application/xhtml+xml"
	case "xls":
		return "application/vnd.ms-excel"
	case "xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "xml":
		return "application/xml" // text/xml (old)
	case "xul":
		return "application/vnd.mozilla.xul+xml"
	case "zip":
		return "application/zip"
	case "3gp":
		return "video/3gpp" // audio/3gpp (only audio)
	case "3g2":
		return "video/3gpp2" // audio/3gpp2 (only audio)
	case "7z":
		return "application/x-7z-compressed"
	default:
		return "application/octet-stream"
	}
}

func urlJoin(parts ...string) string {
	out := "./"
	for _, part := range parts {
		segments := strings.Split(part, "/")
		for _, segment := range segments {
			if segment != "" {
				out += segment + "/"
			}
		}
	}
	return out[:len(out)-1]
}

//helper end

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

	webServer := &WebServer{
		server: &http.Server{
			Handler: mux,
			Addr:    settings.Addr(),
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

// NewMiddleware return value is for deciding to run next middleware/handler
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
				url := "https://" + webServer.settings.Hostname + ":" + webServer.settings.HttpsPort + r.URL.Path
				http.Redirect(w, r, url, http.StatusMovedPermanently)
				webServer.settings.Logger.Println("Redirect: http to https 301 to " + url)
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
	url := "https://" + webServer.settings.Hostname + ":" + webServer.settings.HttpsPort + webServer.settings.FallbackRedirect
	http.Redirect(rw, req, url, http.StatusTemporaryRedirect)
	webServer.settings.Logger.Println("Fallback Redirect to " + url)
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

	file, err := os.ReadFile(urlJoin(webServer.settings.Root, path))
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

	rw.Header().Set("Content-Type", getMimeType(fileExtension))
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
