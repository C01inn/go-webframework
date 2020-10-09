package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strconv"
)

type app struct {
	get    func(route string, toDo func(req req) urlResp)
	post   func(route string, toDo func(req req) urlResp)
	listen func(port int)
}

type req struct {
	method string
	route  string
	params map[string]string
	body   string
}

type urlResp struct {
	body        string
	contentType string
	filename    string
}

type json_body struct {
	Test string
}

func AppConstructor(ap app) app {

	// handle get request
	ap.get = func(route string, toDo func(req req) urlResp) {
		http.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {

				url_params := make(map[string]string)

				// add url param key and values to map
				for k, v := range r.URL.Query() {
					url_params[k] = v[0]
				}

				requestObj := req{
					method: r.Method,
					route:  r.URL.Path,
					params: url_params,
					body:   "",
				}

				resp := toDo(requestObj)
				// check if user is returning html
				if resp.contentType == "html" {
					w.Header().Set("Content-Type", "text/html; charset=utf-8")

					fmt.Fprint(w, resp.body)
				} else if resp.contentType == "json" {
					w.Header().Set("Content-Type", "application/json")
					fmt.Fprint(w, resp.body)
				} else if resp.contentType == "file" {
					http.ServeFile(w, r, resp.body)
				} else if resp.contentType == "download" {
					w.Header().Set(`Content-Disposition`, fmt.Sprintf(`attachment; filename="%s"`, resp.filename))
					http.ServeFile(w, r, resp.body)
				}

			} else {
				fmt.Fprintf(w, "Method not allowed")
			}

		})

	}

	// handle post request
	ap.post = func(route string, toDo func(req req) urlResp) {
		http.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {

				// get http body
				bodyBytes, err := ioutil.ReadAll(r.Body)
				if err != nil {
					log.Fatal(err)
				}
				bodyString := string(bodyBytes)

				url_params := make(map[string]string)

				// addd url param key and values to map
				for k, v := range r.URL.Query() {
					url_params[k] = v[0]
				}

				requestObj := req{
					method: r.Method,
					route:  r.URL.Path,
					params: url_params,
					body:   bodyString,
				}

				resp := toDo(requestObj)
				// check if user is returning html
				if resp.contentType == "html" {
					w.Header().Set("Content-Type", "text/html; charset=utf-8")

					fmt.Fprint(w, resp.body)
				} else if resp.contentType == "json" {
					w.Header().Set("Content-Type", "application/json")
					fmt.Fprint(w, resp.body)
				} else if resp.contentType == "file" {
					http.ServeFile(w, r, resp.body)
				} else if resp.contentType == "download" {
					w.Header().Set(`Content-Disposition`, fmt.Sprintf(`attachment; filename="%s"`, resp.filename))
					http.ServeFile(w, r, resp.body)
				}
			} else {
				fmt.Fprintf(w, "Method not allowed")
			}
		})
	}

	// listen on a port
	ap.listen = func(port int) {
		http.ListenAndServe(":"+strconv.Itoa(port), nil)
		fmt.Println(fmt.Sprintf("Listening on port %s", strconv.Itoa(port)))
	}

	return ap
}

// make the app
// app := Server()
func Server() app {
	app := app{}
	app = AppConstructor(app)
	return app
}

// render html template
// returns text from html file
func renderHtml(filepath string) urlResp {
	html, err := ioutil.ReadFile(filepath)

	if err != nil {
		log.Fatal(err)
	}

	return_value := urlResp{
		body:        string(html),
		filename:    "",
		contentType: "html",
	}

	return return_value
}

func sendStr(bodyu string) urlResp {
	return_value := urlResp{
		body:        bodyu,
		filename:    "",
		contentType: "html",
	}
	return return_value
}

// converts a string, map, slice, or array to json and respondes to request
func sendJson(bodyu interface{}) urlResp {

	// if type is string
	if reflect.TypeOf(bodyu).Kind() == reflect.String {
		real_body := fmt.Sprintf("%v", bodyu)

		return_value := urlResp{
			body:        real_body,
			filename:    "",
			contentType: "json",
		}

		return return_value

	} else if reflect.TypeOf(bodyu).Kind() == reflect.Map {
		real_body, err := json.Marshal(bodyu)

		if err != nil {
			fmt.Println(err)
		}
		return_value := urlResp{
			body:        string(real_body),
			filename:    "",
			contentType: "json",
		}

		return return_value
	} else if reflect.TypeOf(bodyu).Kind() == reflect.Slice {
		real_body, err := json.Marshal(bodyu)

		if err != nil {
			fmt.Println(err)
		}

		return_value := urlResp{
			body:        string(real_body),
			filename:    "",
			contentType: "json",
		}
		return return_value
	} else if reflect.TypeOf(bodyu).Kind() == reflect.Array {
		real_body, err := json.Marshal(bodyu)

		if err != nil {
			fmt.Println(err)
		}

		return_value := urlResp{
			body:        string(real_body),
			filename:    "",
			contentType: "json",
		}

		return return_value
	} else if reflect.TypeOf(bodyu).Kind() == reflect.Struct {
		json_data, err := json.Marshal(bodyu)

		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(json_data))
		return_value := urlResp{
			body:        string(json_data),
			filename:    "",
			contentType: "json",
		}

		return return_value
	}

	return urlResp{body: "other", filename: "", contentType: "json"}
}

func sendFile(filepath string) urlResp {
	return_data := urlResp{
		body:        filepath,
		filename:    "",
		contentType: "file",
	}
	return return_data
}

func downloadFile(filepath string, filenamee string) urlResp {
	return_data := urlResp{
		body:        filepath,
		filename:    filenamee,
		contentType: "download",
	}
	return return_data
}

type TestStruct struct {
	Name    string
	Country string
	Days    map[string]int
}

func main() {
	app := Server()
	// routes
	app.post("/home", func(req req) urlResp {
		fmt.Println(req.body)

		return sendFile("./img.jpg")
	})

	app.get("/", func(req req) urlResp {

		fmt.Println(req.params)

		return renderHtml("./index.html")
	})

	app.get("/about", func(req req) urlResp {

		/* mymap := make([]map[string][]int)
		mymap["key1"] = []int{1, 2, 3}
		fmt.Println(reflect.TypeOf(mymap).Kind())

		return sendJson(mymap)*/
		return sendStr("ss")

	})

	app.get("/s", func(req req) urlResp {

		my_mape := make(map[string]int)
		my_mape["k1"] = 8

		fmt.Println(req.params)

		return downloadFile("./img.jpg", "myimage11.jpg")
	})

	app.listen(8090)
}
