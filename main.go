package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// make structs uppercase so they get exported
type app struct {
	get    func(route string, toDo func(req req) urlResp)
	post   func(route string, toDo func(req req) urlResp)
	listen func(port int)
}

type req struct {
	method  string
	route   string
	params  map[string]string
	body    string
	props   map[string]string
	w       http.ResponseWriter
	r       *http.Request
	getFile func(filename string) (multipart.File, *multipart.FileHeader, error)
}

type urlResp struct {
	body        string
	contentType string
	filename    string
}

var allRoutes [][]string
var routeFunc map[string]func(req req) urlResp

func RemoveIndex(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}

// function that returns if route is paramter, route, route method, and route function
func checkRouteParams(route string) (bool, string, string, map[string]string) {
OUTER:
	for _, jk := range allRoutes {
		if jk[0] == route {
			return false, "", "", make(map[string]string)
		} else {
			if strings.Contains(jk[0], "{") == true && strings.Contains(jk[0], "}") == true {

				routeValues := strings.Split(route, "/")
				routeValues = RemoveIndex(routeValues, 0)

				ll := strings.Split(jk[0], "/")
				ll = RemoveIndex(ll, 0)

				if len(routeValues) == len(ll) && routeValues[0] == ll[0] {
					// check to see if it matches better with another route

					matching_map := make(map[string]int)

					for _, route1 := range allRoutes {

						route1Values := strings.Split(route1[0], "/")
						route1Values = RemoveIndex(route1Values, 0)
						matching_score := 0
						for _, cc := range routeValues {

							for _, uu := range route1Values {

								if uu == cc {
									matching_score = matching_score + 1
								}
							}
						}
						matching_map[route1[0]] = matching_score

					}
					theRouteScore := matching_map[jk[0]]
					for _, yy := range matching_map {

						if yy > theRouteScore {
							continue OUTER
						}
					}
					// create hashmap of url parameters

					params_map := make(map[string]string)
					for idx, keyy := range ll {
						if strings.HasPrefix(keyy, "{") && strings.HasSuffix(keyy, "}") {
							key2 := strings.Replace(keyy, "{", "", 1)
							key2 = strings.Replace(key2, "}", "", 1)
							params_map[key2] = routeValues[idx]
						}
					}
					return true, jk[0], jk[1], params_map

				} else {
					continue
				}

			}
		}

	}
	return false, "", "", make(map[string]string)
}

func AppConstructor(ap app) app {

	routeFunc = make(map[string]func(req req) urlResp)

	// handle get request
	ap.get = func(route string, toDo func(req req) urlResp) {

		// make sure route starts with a slash
		if strings.HasPrefix(route, "/") {
		} else {
			panic("Route Must Start With a /")
			return
		}

		// check if route already exists
		for _, ccc := range allRoutes {
			if ccc[0] == route {
				panic("Route Already Exists")
				return
			} else {
				continue
			}
		}

		var routeSlice []string
		routeSlice = append(routeSlice, route)
		routeSlice = append(routeSlice, "GET")
		allRoutes = append(allRoutes, routeSlice)

		routeFunc[route] = toDo

		http.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {

			if route != r.URL.Path {
				// check if its a route with a parameter
				parameterRoute, hashRoute, routeMethod, url_vars := checkRouteParams(r.URL.Path)
				if parameterRoute == true {
					if routeMethod == r.Method && routeMethod == "GET" {
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
							props:  url_vars,
							w:      w,
							r:      r,
						}

						resp := routeFunc[hashRoute](requestObj)

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
					} else if routeMethod == r.Method && routeMethod != "GET" {
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
							props:  url_vars,
							w:      w,
							r:      r,
						}

						resp := routeFunc[hashRoute](requestObj)
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
						fmt.Fprintf(w, "Method wrong")
						return
					}

				} else {
					// if not return 404
					fmt.Fprintf(w, "404 Not Found")
					return
				}

			} else {
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
						w:      w,
						r:      r,
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

			}

		})

	}

	// handle post request
	ap.post = func(route string, toDo func(req req) urlResp) {

		if strings.HasPrefix(route, "/") {
		} else {
			panic("Route Must Start With a /")
			return
		}

		// check if route already exists
		for _, ccc := range allRoutes {
			if ccc[0] == route {
				panic("Route Already Exists")
				return
			} else {
				continue
			}
		}

		var routeSlice1 []string
		routeSlice1 = append(routeSlice1, route)
		routeSlice1 = append(routeSlice1, "POST")
		allRoutes = append(allRoutes, routeSlice1)

		routeFunc[route] = toDo

		http.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				if route != r.URL.Path {
					// check if its a route with a parameter
					parameterRoute, hashRoute, routeMethod, url_vars := checkRouteParams(r.URL.Path)
					if parameterRoute == true {
						if routeMethod == r.Method {

							url_params := make(map[string]string)

							// addd url param key and values to map
							for k, v := range r.URL.Query() {
								url_params[k] = v[0]
							}

							// get http body
							bodyBytes, err := ioutil.ReadAll(r.Body)
							if err != nil {
								log.Fatal(err)
							}
							bodyString := string(bodyBytes)

							// make request object
							requestObj := req{
								method: r.Method,
								route:  r.URL.Path,
								params: url_params,
								body:   bodyString,
								props:  url_vars,
								w:      w,
								r:      r,
								getFile: func(filename string) (multipart.File, *multipart.FileHeader, error) {
									file, header, err := r.FormFile(filename)
									return file, header, err
								},
							}

							resp := routeFunc[hashRoute](requestObj)
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
							// if not return wront method
							fmt.Fprintf(w, "Wrong Method")
							return
						}

					} else {
						// if not return 404
						fmt.Fprintf(w, "404 Not Found")
						return
					}
				} else {
					r.ParseMultipartForm(5 * 1024 * 1024)

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
						w:      w,
						r:      r,
						getFile: func(filename string) (multipart.File, *multipart.FileHeader, error) {
							file, header, err := r.FormFile(filename)
							return file, header, err
						},
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
				}

			} else {
				fmt.Fprintf(w, "Method not allowed")
			}
		})
	}

	// listen on a port
	ap.listen = func(port2 int) {
		http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
		http.ListenAndServe("0.0.0.0:"+strconv.Itoa(port2), nil)

		fmt.Println("Listening on port " + strconv.Itoa(port2))
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
func renderHtml(filepath string, temp_data interface{}) urlResp {
	if temp_data == nil {
		t, _ := template.ParseFiles(filepath)
		var tpl bytes.Buffer

		type templatedata struct{}
		template_data := templatedata{}

		t.Execute(&tpl, template_data)
		html := tpl.String()

		return_value := urlResp{
			body:        string(html),
			filename:    "",
			contentType: "html",
		}
		return return_value
	}

	if reflect.TypeOf(temp_data).Kind() == reflect.Struct {
		t, _ := template.ParseFiles(filepath)

		var tpl bytes.Buffer
		t.Execute(&tpl, temp_data)

		html := tpl.String()

		return_value := urlResp{
			body:        string(html),
			filename:    "",
			contentType: "html",
		}

		return return_value
	} else {
		t, _ := template.ParseFiles(filepath)
		var tpl bytes.Buffer

		type templatedata struct{}
		template_data := templatedata{}

		t.Execute(&tpl, template_data)
		html := tpl.String()

		return_value := urlResp{
			body:        string(html),
			filename:    "",
			contentType: "html",
		}
		return return_value
	}

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

func getCode(r *http.Request, defaultCode int) (int, string) {
	p := strings.Split(r.URL.Path, "/")
	if len(p) == 1 {
		return defaultCode, p[0]
	} else if len(p) > 1 {
		code, err := strconv.Atoi(p[0])
		if err == nil {
			return code, p[1]
		} else {
			return defaultCode, p[1]
		}
	} else {
		return defaultCode, ""
	}
}

type Cookie struct {
	Name  string
	Value string

	Expires time.Time
}

func setCookie(request req, cookie_data Cookie) {

	http_cookie := &http.Cookie{
		Name:     cookie_data.Name,
		Value:    cookie_data.Value,
		Expires:  cookie_data.Expires,
		HttpOnly: false,
		Path:     "/",
	}

	http.SetCookie(request.w, http_cookie)
}

func getCookie(request req, name string) (string, error) {
	c, err := request.r.Cookie(name)

	if err != nil {
		return "", err
	}
	return c.Value, nil
}

func removeCookie(request req, name string) {
	http_cookie := &http.Cookie{
		Name:     name,
		Value:    "",
		Expires:  time.Now().Add(time.Minute - (time.Second)*30),
		HttpOnly: false,
		Path:     "/",
		MaxAge:   -1,
	}
	http.SetCookie(request.w, http_cookie)
}

func main() {
	app := Server()
	// routes
	app.post("/home/{id}", func(req req) urlResp {

		return sendFile("./img.jpg")
	})

	app.get("/", func(req req) urlResp {

		return renderHtml("./templates/index.html", nil)
	})

	app.get("/about/{id}/{type}", func(req req) urlResp {
		cookie_val, err := getCookie(req, "cook1")
		if err != nil {
			fmt.Println(err)
		}

		return sendStr("Id: " + req.props["id"] + "<br>" + "Type: " + req.props["type"] + "<br> " + cookie_val)

	})

	app.get("/s", func(req req) urlResp {

		my_mape := make(map[string]int)
		my_mape["k1"] = 8

		return downloadFile("./img.jpg", "myimage11.jpg")
	})

	app.get("/videos/{id}", func(req req) urlResp {
		id := req.props["id"]
		removeCookie(req, "cook1")

		return sendStr("This is the videos page " + id)
	})
	app.get(`/videos/about`, func(req req) urlResp {

		setCookie(req, Cookie{
			Name:    "cook1",
			Value:   "mycookeieval",
			Expires: time.Now().Add(time.Hour + time.Hour),
		})

		return sendStr("video about page")
	})

	app.get("/img/{ids}", func(req req) urlResp {
		setCookie(req, Cookie{
			Name:    "cook1",
			Value:   "changed-cookie",
			Expires: time.Now().Add(time.Hour + time.Hour),
		})

		return sendStr("ssss " + req.props["ids"])
	})

	app.get("/agg/{id}/{name}", func(req req) urlResp {
		return sendStr(req.props["id"] + `  ` + req.props["name"])
	})
	app.get("/agg/videos/{id}", func(req req) urlResp {

		return sendStr("image:   " + req.props["id"])
	})

	app.get("/agg", func(req req) urlResp {

		// make struct of data to pass to template
		type newsAggPage struct {
			Title string
			News  string
			Posts []string
		}

		data2 := newsAggPage{
			Title: "",
			News:  "Fake news!",
			Posts: []string{"Post 1", "Post 2", "Post3"},
		}

		return renderHtml(`./templates/temp.html`, data2)
	})
	app.get("/upload", func(req req) urlResp {
		return renderHtml(`./templates/upload.html`, nil)
	})
	app.post("/file-up", func(req req) urlResp {

		file, header, err := req.getFile("myfile")

		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(header)
		fmt.Println(file)

		return sendStr("good")
	})

	app.listen(8090)

}
