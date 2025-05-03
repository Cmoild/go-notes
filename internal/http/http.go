package http

import (
	"bufio"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Request struct {
	Method  string
	Path    string
	Query   string
	Version string
	Body    []byte
	Headers []string
}

type Response struct {
	Status  string
	Headers map[string]string
}

type ResponseWriter struct {
	net.Conn
	Response Response
}

type Route struct {
	Path    string
	Handler func(ResponseWriter, *Request)
}

func (w ResponseWriter) Write(data []byte) (int, error) {
	contentLength := len(data)
	if contentLength > 0 {
		w.SetHeader("Content-Length", strconv.Itoa(contentLength))
	}
	httpHeader := []byte("HTTP/1.1 " + w.Response.Status + "\r\n")
	for key, value := range w.Response.Headers {
		httpHeader = append(httpHeader, []byte(key+": "+value+"\r\n")...)
	}
	response := append(httpHeader, []byte("\r\n")...)
	response = append(response, data...)
	response = append(response, []byte("\r\n")...)
	return w.Conn.Write(response)
}

func (w *ResponseWriter) SetHeader(key, value string) {
	if w.Response.Headers == nil {
		w.Response.Headers = make(map[string]string)
	}
	w.Response.Headers[key] = value
}

func (w *ResponseWriter) SetStatus(status string) {
	w.Response.Status = status
}

var router []Route

func ListenAndServe(port string) {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	log.Printf("Сервер запущен на порту %s", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Ошибка при подключении: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	var message []string
	contentLength := 0
	var content []byte
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		if len(msg) > 15 && msg[:15] == "Content-Length:" {
			contentLength, err = strconv.Atoi(strings.TrimSpace(msg[15 : len(msg)-2]))
			if err != nil {
				log.Fatal(err)
			}
		}
		if msg != "\r\n" {
			message = append(message, msg)
		} else {
			if contentLength > 0 {
				content = make([]byte, contentLength)
				_, err := reader.Read(content)
				if err != nil {
					log.Fatal(err)
				}
			}
			break
		}
	}

	re := regexp.MustCompile(`(GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS)\s+([^\s?]+)(?:\?([^\s]+))?\s+(HTTP\/\d\.\d)`)
	matches := re.FindStringSubmatch(message[0])
	if len(matches) != 5 {
		return
	}
	method := matches[1]
	path := matches[2]
	query := matches[3]
	version := matches[4]
	if method == "" {
		return
	}
	if path == "" {
		return
	}
	if version == "" {
		return
	}

	var request Request = Request{
		Method:  method,
		Path:    path,
		Query:   query,
		Version: version,
		Body:    content,
		Headers: message,
	}

	log.Println(request.Method, request.Path)

	ok := false
	var handler func(ResponseWriter, *Request)
	for _, route := range router {
		re := regexp.MustCompile(route.Path)
		if re.MatchString(request.Path) {
			handler = route.Handler
			ok = true
			break
		}
	}

	w := ResponseWriter{
		Conn: conn,
	}

	if !ok {
		handler = SendNotFound404
	}

	handler(w, &request)
}

func HandleFunc(address string, handler func(w ResponseWriter, r *Request)) {
	router = append(router, Route{Path: address, Handler: handler})
}

func getFilePathsRecursively(root string, relativeRoot string, paths *[]string) {
	dir, err := os.ReadDir(root)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range dir {
		if file.IsDir() {
			getFilePathsRecursively(root+"/"+file.Name(), relativeRoot+"/"+file.Name(), paths)
		} else {
			*paths = append(*paths, relativeRoot+"/"+file.Name())
		}
	}
}

func SendInternalServerError500(w ResponseWriter, r *Request) {
	w.SetStatus("500 Internal Server Error")
	w.SetHeader("Content-Type", "text/plain")
	w.Write([]byte("Internal Server Error"))
}

func SendBadRequest400(w ResponseWriter, r *Request) {
	w.SetStatus("400 Bad Request")
	w.SetHeader("Content-Type", "text/plain")
	w.Write([]byte("Bad request"))
}

func SendForbidden403(w ResponseWriter, r *Request) {
	w.SetStatus("403 Forbidden")
	w.SetHeader("Content-Type", "text/plain")
	w.Write([]byte("Forbidden"))
}

func SendNotFound404(w ResponseWriter, r *Request) {
	w.SetStatus("404 Not Found")
	w.SetHeader("Content-Type", "text/plain")
	w.Write([]byte("404 Not Found"))
}

func SendConflict409(w ResponseWriter, r *Request) {
	w.SetStatus("409 Conflict")
	w.SetHeader("Content-Type", "text/plain")
	w.Write([]byte("Conflict with the data"))
}

func SendOK200(w ResponseWriter, r *Request) {
	w.SetStatus("200 OK")
	w.SetHeader("Content-Type", "text/plain")
	w.Write([]byte("Success"))
}

func SendCreated201(w ResponseWriter, r *Request) {
	w.SetStatus("201 Created")
	w.SetHeader("Content-Type", "text/plain")
	w.Write([]byte("Successfully created"))
}

var fileServerRoot string = ""

func fileServerHandler(w ResponseWriter, r *Request) {
	filePath := fileServerRoot + r.Path
	file, err := os.ReadFile(filePath)
	if err != nil {
		SendNotFound404(w, r)
		return
	}
	re := `\.[\w]*$`
	fileExt := regexp.MustCompile(re).FindString(filePath)
	w.SetStatus("200 OK")
	switch fileExt {
	case ".html":
		w.SetHeader("Content-Type", "text/html; charset=utf-8")
	case ".css":
		w.SetHeader("Content-Type", "text/css")
	case ".js":
		w.SetHeader("Content-Type", "text/javascript; charset=utf-8")
	case ".png":
		w.SetHeader("Content-Type", "image/png")
	case ".jpg":
		w.SetHeader("Content-Type", "image/jpeg")
	case ".jpeg":
		w.SetHeader("Content-Type", "image/jpeg")
	case ".gif":
		w.SetHeader("Content-Type", "image/gif")
	case ".svg":
		w.SetHeader("Content-Type", "image/svg+xml")
	case ".ico":
		w.SetHeader("Content-Type", "image/x-icon")
	case ".json":
		w.SetHeader("Content-Type", "application/json")
	case ".txt":
		w.SetHeader("Content-Type", "text/plain")
	case ".pdf":
		w.SetHeader("Content-Type", "application/pdf")
	default:
		w.SetHeader("Content-Type", "text/plain")
	}
	_, err = w.Write(file)
	if err != nil {
		log.Println(err)
	}
}

func FileServer(root string) {
	var paths []string
	var relativeRoot string = ""
	fileServerRoot = root
	getFilePathsRecursively(root, relativeRoot, &paths)
	for _, path := range paths {
		HandleFunc(path, fileServerHandler)
	}
}
