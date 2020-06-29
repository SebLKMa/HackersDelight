// This package runs a HTTP server handling specific routes in main entry point.
// This server is not for production - it is only for testing 1 connection to verify concept
package main

import (
	"flag"
	"fmt"
	"log"
	"html/template"
	"net/http"
	"net/url"
)

// Internal package variables
// Set by command line options
var finalRedirectUrl string

//TODO: remove globals, get data directly and use them per request.
// Set by captureQueryStringMultiple()
var urlString string
var sessionId string
var callbackUrl string
var macAddress string
var urlStr string
var requestIP string
var acceptIP string

// Handles /ping request.
// Returns a heartbeat message response.
func pingHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logHandler(w, r)
		fmt.Fprintln(w, "pong")
	})
}

// Handles /hello
func helloHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logHandler(w, r)

		tmpl, err := template.ParseFiles("hello.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		htmlData := struct {
			AppVersion string
			UrlLink    string
		}{
			AppVersion: "8.2.9",
			UrlLink:    "http://192.168.126.130:8282/logincallback?mac_address=00ee.abb2.8820&redirect_url=https%3A%2F%2Fdzone.com%2Frefcardz%2Fgetting-started-with-etherium-private-blockchain&session_id=181cd76f40abb28382fb7bc94ea8c8149f5b36b1510d2037c8475a6a1fa70180&state_flag=1",
		}

		err = tmpl.Execute(w, htmlData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

// Returns html page using go html/template
func executeCaptivePage(w http.ResponseWriter, r *http.Request) {
	logHandler(w, r)

	log.Println("executeCaptivePage")

	// capture the incoming query string data
	captureQueryStringMultiple(w, r)
	// construct the target URL to be embedded into index.html
	targetURL, _ := makeUrlCallback(callbackUrl, sessionId, macAddress, "1", finalRedirectUrl, "10")

	tmpl, err := template.ParseFiles("blankAuto.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	htmlData := struct {
		AppVersion string
		URLLink    string
	}{
		AppVersion: "8.2.9",
		URLLink:    targetURL,
	}

	err = tmpl.Execute(w, htmlData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Extracts keys and values from URL query string.
func captureQueryStringMultiple(w http.ResponseWriter, r *http.Request) {
	log.Println("captureQueryStringMultiple")

	urlString = r.URL.String()
	u, err := url.Parse(urlString)
	if err != nil {
		log.Println("request: " + urlString)
		log.Println("Params are missing from URL query string")
		return
	}
	q := u.Query()
	sessionId = q.Get("session_id")
	callbackUrl = q.Get("callback")
	macAddress = q.Get("mac_address")
	requestIP = getRequestIP(r)
	captiveString := "session_id: " + sessionId + "\n" +
		"callback_url: " + callbackUrl + "\n" +
		"mac_addrress: " + macAddress + "\n" +
		"request_ip: " + requestIP
	log.Println(captiveString)
}

// GetIP gets a requests IP address by reading off the forwarded-for
// header (for proxies) and falls back to use the remote address.
func getRequestIP(r *http.Request) string {
	// first choice
	ipAddress := r.Header.Get("X-Real-Ip")
	log.Printf("X-Real-Ip=%+v\n", ipAddress)

	// second choice
	if ipAddress == "" {
		ipAddress = r.Header.Get("X-Forwarded-For")
		log.Printf("X-Forwarded-For=%+v\n", ipAddress)
	}

	// last choice
	if ipAddress == "" {
		ipAddress = r.RemoteAddr
		log.Printf("RemoteAddr=%+v\n", ipAddress)
	}
	return ipAddress
}

// Construct URL with query string from the supplied callbackUrl as the target host
func makeUrlCallback(callbackUrl string, sessionId string, macAddress string, stateFlag string, redirectUrl string, accessTime string) (string, int) {
	log.Println("makeUrlCallback")

	// assign query strings data
	data := url.Values{}
	data.Set("mac_address", macAddress)
	data.Set("session_id", sessionId)
	data.Set("state_flag", stateFlag)
	data.Set("redirect_url", redirectUrl)
	data.Set("access_time", accessTime)
	dataLen := len(data.Encode())

	// construct the target URL with query string data
	u, err := url.ParseRequestURI(callbackUrl)
	if err != nil {
		log.Println(err)
	}
	u.RawQuery = data.Encode()
	urlStr := fmt.Sprintf("%v", u) // "https://api.com/user/?name=foo&surname=bar"
	log.Println(urlStr)

	return urlStr, dataLen
}

// Construct URL with query strings using requestIP as target host.
// https://stackoverflow.com/questions/19253469/make-a-url-encoded-post-request-using-http-newrequest
func makeCallbackUrl(requestIP string, callbackUrl string, sessionId string, macAddress string, stateFlag string, redirectUrl string, accessTime string) (string, int) {
	log.Println("makeCallbackUrl")
	//resource := "/resoucepath/"
	data := url.Values{}
	data.Set("mac_address", macAddress)
	data.Set("session_id", sessionId)
	data.Set("state_flag", stateFlag)
	data.Set("redirect_url", redirectUrl)
	data.Set("access_time", accessTime)
	dataLen := len(data.Encode())

	u, err := url.ParseRequestURI(callbackUrl)
	log.Printf("callback host from hardcode: err=%+v url=%+v\n", err, u.Host)

	u.Host = requestIP
	// This is to reset the port
	//host, port, _ := net.SplitHostPort(u.Host)
	//host, _, _ := net.SplitHostPort(u.Host) // ignore port
	//newCallbackHost := host + ":" + "80"
	//u.Host = newCallbackHost
	log.Printf("callback host from request: err=%+v url=%+v\n", err, u.Host)

	//u.Path = resource
	//urlStr := u.String() // "https://api.com/user/"
	u.RawQuery = data.Encode()
	urlStr := fmt.Sprintf("%v", u) // "https://api.com/user/?name=foo&surname=bar"
	log.Println(urlStr)

	return urlStr, dataLen
}

func redirect(w http.ResponseWriter, r *http.Request, urlStr string) {
	log.Printf("redirect callback: %s\n", urlStr)
	http.Redirect(w, r, urlStr, 301)
}

// logs the HTTP request.
func logHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s %s\n", r.Method, r.URL, r.Proto)
	for k, v := range r.Header {
		log.Printf("Header[%q] = %q\n", k, v)
	}
	log.Printf("Host = %q\n", r.Host)
	log.Printf("RemoteAddr = %q\n", r.RemoteAddr)
	if err := r.ParseForm(); err != nil {
		log.Print(err)
	}
	for k, v := range r.Form {
		log.Printf("Form[%q] = %q\n", k, v)
	}
}

// Main entry point handling HTTP requests routes:
// <host:port>/ping  - for heartbeat ping
// <host:port>/hello - just render a html page
//
func main() {

	log.Println("HTTP Server")

	portPtr := flag.String("port", "8288", "port number")
	finalUrlPtr := flag.String("finalRedirectUrl", "https://dzone.com/refcardz/getting-started-with-etherium-private-blockchain", "the final redirect URL to be returned to gateway router.")
	flag.Parse()

	mux := http.NewServeMux()
	mux.Handle("/ping", pingHandler())
	mux.Handle("/hello", helloHandler())

	finalRedirectUrl = *finalUrlPtr
	port := *portPtr
	host := "0.0.0.0:" + port
	log.Println("The final redirect URL for gateway router will be: " + finalRedirectUrl)
	log.Println(host + " up and listening")

	err := http.ListenAndServe(host, mux)
	if err != nil {
		log.Fatal("Error creating server. ", err)
	}
}
