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
type App struct {
	Get    func(route string, toDo func(req Req) UrlResp)
	Post   func(route string, toDo func(req Req) UrlResp)
	Put    func(route string, toDo func(req Req) UrlResp)
	Delete func(route string, toDo func(req Req) UrlResp)
	Patch  func(route string, toDo func(req Req) UrlResp)
	Listen func(port int)
}

type Req struct {
	Method  string
	Route   string
	Params  map[string]string
	Body    string
	Props   map[string]string
	W       http.ResponseWriter
	R       *http.Request
	GetFile func(filename string) (multipart.File, *multipart.FileHeader, error)
	Form    map[string][]string
}

type UrlResp struct {
	body        string
	contentType string
	filename    string
}

var allRoutes [][]string
var routeFunc map[string]func(req Req) UrlResp

func removeIndex(s []string, index int) []string {
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
				routeValues = removeIndex(routeValues, 0)

				ll := strings.Split(jk[0], "/")
				ll = removeIndex(ll, 0)

				if len(routeValues) == len(ll) && routeValues[0] == ll[0] {
					// check to see if it matches better with another route

					matching_map := make(map[string]int)

					for _, route1 := range allRoutes {

						route1Values := strings.Split(route1[0], "/")
						route1Values = removeIndex(route1Values, 0)
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

					// check again
					for idx8, jo := range routeValues {

						if jo != ll[idx8] {
							if strings.Contains(ll[idx8], "{") == false {
								if strings.Contains(ll[idx8], "}") == false {
									continue OUTER
								}
							}
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

func appConstructor(ap App) App {

	routeFunc = make(map[string]func(req Req) UrlResp)

	// handle get request
	ap.Get = func(route string, toDo func(req Req) UrlResp) {

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
		routeSlice = append(routeSlice, http.MethodGet)
		allRoutes = append(allRoutes, routeSlice)

		routeFunc[route] = toDo

		http.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {

			if route != r.URL.Path {
				// check if its a route with a parameter
				parameterRoute, hashRoute, routeMethod, url_vars := checkRouteParams(r.URL.Path)
				if parameterRoute == true {
					if routeMethod == r.Method && routeMethod == http.MethodGet {
						url_params := make(map[string]string)

						// add url param key and values to map
						for k, v := range r.URL.Query() {
							url_params[k] = v[0]
						}

						requestObj := Req{
							Method: r.Method,
							Route:  r.URL.Path,
							Params: url_params,
							Body:   "",
							Props:  url_vars,
							W:      w,
							R:      r,
							Form:   r.Form,
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

						requestObj := Req{
							Method: r.Method,
							Route:  r.URL.Path,
							Params: url_params,
							Body:   bodyString,
							Props:  url_vars,
							W:      w,
							R:      r,
							Form:   r.Form,
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

					requestObj := Req{
						Method: r.Method,
						Route:  r.URL.Path,
						Params: url_params,
						Body:   "",
						W:      w,
						R:      r,
						Form:   r.Form,
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
	ap.Post = func(route string, toDo func(req Req) UrlResp) {

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
		routeSlice1 = append(routeSlice1, http.MethodPost)
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
							requestObj := Req{
								Method: r.Method,
								Route:  r.URL.Path,
								Params: url_params,
								Body:   bodyString,
								Props:  url_vars,
								W:      w,
								R:      r,
								GetFile: func(filename string) (multipart.File, *multipart.FileHeader, error) {

									return r.FormFile(filename)
								},
								Form: r.Form,
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

					requestObj := Req{
						Method: r.Method,
						Route:  r.URL.Path,
						Params: url_params,
						Body:   bodyString,
						W:      w,
						R:      r,
						GetFile: func(filename string) (multipart.File, *multipart.FileHeader, error) {
							file, header, err := r.FormFile(filename)
							return file, header, err
						},
						Form: r.Form,
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

	// handle a Put request
	ap.Put = func(route string, toDo func(req Req) UrlResp) {

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
		routeSlice1 = append(routeSlice1, http.MethodPut)
		allRoutes = append(allRoutes, routeSlice1)

		routeFunc[route] = toDo

		http.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPut {
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
							requestObj := Req{
								Method: r.Method,
								Route:  r.URL.Path,
								Params: url_params,
								Body:   bodyString,
								Props:  url_vars,
								W:      w,
								R:      r,
								GetFile: func(filename string) (multipart.File, *multipart.FileHeader, error) {
									file, header, err := r.FormFile(filename)
									return file, header, err
								},
								Form: r.Form,
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

					requestObj := Req{
						Method: r.Method,
						Route:  r.URL.Path,
						Params: url_params,
						Body:   bodyString,
						W:      w,
						R:      r,
						GetFile: func(filename string) (multipart.File, *multipart.FileHeader, error) {
							file, header, err := r.FormFile(filename)
							return file, header, err
						},
						Form: r.Form,
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

	// handle a Delete Request
	ap.Delete = func(route string, toDo func(req Req) UrlResp) {

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
		routeSlice1 = append(routeSlice1, http.MethodDelete)
		allRoutes = append(allRoutes, routeSlice1)

		routeFunc[route] = toDo

		http.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodDelete {
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
							requestObj := Req{
								Method: r.Method,
								Route:  r.URL.Path,
								Params: url_params,
								Body:   bodyString,
								Props:  url_vars,
								W:      w,
								R:      r,
								GetFile: func(filename string) (multipart.File, *multipart.FileHeader, error) {
									file, header, err := r.FormFile(filename)
									return file, header, err
								},
								Form: r.Form,
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

					requestObj := Req{
						Method: r.Method,
						Route:  r.URL.Path,
						Params: url_params,
						Body:   bodyString,
						W:      w,
						R:      r,
						GetFile: func(filename string) (multipart.File, *multipart.FileHeader, error) {
							file, header, err := r.FormFile(filename)
							return file, header, err
						},
						Form: r.Form,
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

	// handle a Patch Request
	ap.Patch = func(route string, toDo func(req Req) UrlResp) {

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
		routeSlice1 = append(routeSlice1, http.MethodPatch)
		allRoutes = append(allRoutes, routeSlice1)

		routeFunc[route] = toDo

		http.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPatch {
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
							requestObj := Req{
								Method: r.Method,
								Route:  r.URL.Path,
								Params: url_params,
								Body:   bodyString,
								Props:  url_vars,
								W:      w,
								R:      r,
								GetFile: func(filename string) (multipart.File, *multipart.FileHeader, error) {
									file, header, err := r.FormFile(filename)
									return file, header, err
								},
								Form: r.Form,
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

					requestObj := Req{
						Method: r.Method,
						Route:  r.URL.Path,
						Params: url_params,
						Body:   bodyString,
						W:      w,
						R:      r,
						GetFile: func(filename string) (multipart.File, *multipart.FileHeader, error) {
							file, header, err := r.FormFile(filename)
							return file, header, err
						},
						Form: r.Form,
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
	ap.Listen = func(port2 int) {
		http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
		http.ListenAndServe("0.0.0.0:"+strconv.Itoa(port2), nil)

		fmt.Println("Listening on port " + strconv.Itoa(port2))
	}

	return ap
}

// make the app
// app := Server()
func Server() App {
	app := App{}
	app = appConstructor(app)
	return app
}

// render html template
// returns text from html file
func RenderHtml(filepath string, temp_data interface{}) UrlResp {
	if temp_data == nil {
		t, _ := template.ParseFiles(filepath)
		var tpl bytes.Buffer

		type templatedata struct{}
		template_data := templatedata{}

		t.Execute(&tpl, template_data)
		html := tpl.String()

		return_value := UrlResp{
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

		return_value := UrlResp{
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

		return_value := UrlResp{
			body:        string(html),
			filename:    "",
			contentType: "html",
		}
		return return_value
	}

}

func SendStr(bodyu string) UrlResp {
	return_value := UrlResp{
		body:        bodyu,
		filename:    "",
		contentType: "html",
	}
	return return_value
}

// converts a string, map, slice, or array to json and respondes to request
func SendJson(bodyu interface{}) UrlResp {

	// if type is string
	if reflect.TypeOf(bodyu).Kind() == reflect.String {
		real_body := fmt.Sprintf("%v", bodyu)

		return_value := UrlResp{
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
		return_value := UrlResp{
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

		return_value := UrlResp{
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

		return_value := UrlResp{
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

		return_value := UrlResp{
			body:        string(json_data),
			filename:    "",
			contentType: "json",
		}

		return return_value
	}

	return UrlResp{body: "other", filename: "", contentType: "json"}
}

func SendFile(filepath string) UrlResp {
	return_data := UrlResp{
		body:        filepath,
		filename:    "",
		contentType: "file",
	}
	return return_data
}

func DownloadFile(filepath string, filenamee string) UrlResp {
	return_data := UrlResp{
		body:        filepath,
		filename:    filenamee,
		contentType: "download",
	}
	return return_data
}

type Cookie struct {
	Name  string
	Value string

	Expires time.Time
}

func SetCookie(request Req, cookie_data Cookie) {

	http_cookie := &http.Cookie{
		Name:     cookie_data.Name,
		Value:    cookie_data.Value,
		Expires:  cookie_data.Expires,
		HttpOnly: false,
		Path:     "/",
	}

	http.SetCookie(request.W, http_cookie)
}

func GetCookie(request Req, name string) (string, error) {
	c, err := request.R.Cookie(name)

	if err != nil {
		return "", err
	}
	return c.Value, nil
}

func RemoveCookie(request Req, name string) {
	http_cookie := &http.Cookie{
		Name:     name,
		Value:    "",
		Expires:  time.Now().Add(time.Minute - (time.Second)*30),
		HttpOnly: false,
		Path:     "/",
		MaxAge:   -1,
	}
	http.SetCookie(request.W, http_cookie)
}

func main() {
	fmt.Println(http.MethodPatch)
	app := Server()
	// routes
	app.Post("/home/{id}", func(req Req) UrlResp {

		return SendFile("./img.jpg")
	})

	app.Get("/", func(req Req) UrlResp {

		return RenderHtml("./templates/index.html", nil)
	})

	app.Get("/about/{id}/{type}", func(req Req) UrlResp {
		cookie_val, err := GetCookie(req, "cook1")
		if err != nil {
			fmt.Println(err)
		}

		return SendStr("Id: " + req.Props["id"] + "<br>" + "Type: " + req.Props["type"] + "<br> " + cookie_val)

	})

	app.Get("/s", func(req Req) UrlResp {

		my_mape := make(map[string]int)
		my_mape["k1"] = 8

		return DownloadFile("./img.jpg", "myimage11.jpg")
	})

	app.Get("/videos/{id}", func(req Req) UrlResp {
		id := req.Props["id"]
		RemoveCookie(req, "cook1")

		return SendStr("This is the videos page " + id)
	})
	app.Get(`/videos/about`, func(req Req) UrlResp {

		SetCookie(req, Cookie{
			Name:    "cook1",
			Value:   "mycookeieval",
			Expires: time.Now().Add(time.Hour + time.Hour),
		})

		return SendStr("video about page")
	})

	app.Get("/img/{ids}", func(req Req) UrlResp {
		SetCookie(req, Cookie{
			Name:    "cook1",
			Value:   "changed-cookie",
			Expires: time.Now().Add(time.Hour + time.Hour),
		})

		return SendStr("ssss " + req.Props["ids"])
	})

	app.Get("/agg/{id}/{name}", func(req Req) UrlResp {
		return SendStr(req.Props["id"] + `  ` + req.Props["name"])
	})
	app.Get("/agg/videos/{id}", func(req Req) UrlResp {

		return SendStr("image:   " + req.Props["id"])
	})

	app.Get("/agg", func(req Req) UrlResp {

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

		return RenderHtml(`./templates/temp.html`, data2)
	})
	app.Get("/upload", func(req Req) UrlResp {
		return RenderHtml(`./templates/upload.html`, nil)
	})
	app.Post("/file-up", func(req Req) UrlResp {

		file, header, err := req.GetFile("myfile")

		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(header)
		fmt.Println(file)

		return SendStr("good")
	})

	app.Get(`/users/{id}`, func(req Req) UrlResp {
		return SendStr("user: " + req.Props["id"])
	})

	app.Get(`/users/{id}/posts`, func(req Req) UrlResp {
		return SendStr("user: " + req.Props["id"] + " posts")
	})
	app.Get(`/users/{id}/followers`, func(req Req) UrlResp {
		return SendStr("user: " + req.Props["id"] + ` followers`)
	})

	app.Put(`/another/{id}`, func(req Req) UrlResp {
		fmt.Println(req.Props["id"])
		return SendStr(req.Body)
	})

	app.Delete(`/del/{name}`, func(req Req) UrlResp {
		fmt.Println(req.Props["name"])
		return SendStr(req.Body)
	})

	app.Patch(`/patch/{id}`, func(req Req) UrlResp {

		fmt.Println("patch")
		fmt.Println(req.Props["id"])
		return SendStr(req.Body)
	})

	app.Listen(8090)

}
