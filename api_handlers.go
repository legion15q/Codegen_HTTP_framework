package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
)

func (obj *MyApi) HandlerProfile(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	in, api_err := ValidateProfileParams(r)

	if api_err != nil {
		w.WriteHeader(api_err.HTTPStatus)
		w.Write([]byte(`{"error": "` + api_err.Error() + `"}`))
		return
	}
	resp, err := obj.Profile(ctx, *in)
	if err != nil {
		switch t := err.(type) {
		case ApiError:
			w.WriteHeader(t.HTTPStatus)
			w.Write([]byte(`{"error": "` + t.Error() + `"}`))
		case error:
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "` + err.Error() + `"}`))
		}
		return
	}
	json_out, _ := CompileJSONUser(resp)
	w.Write(json_out)
}
func (obj *MyApi) HandlerCreate(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte(`{"error": "bad method"}`))
		return
	}

	if r.Header.Get("X-Auth") != "100500" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": "unauthorized"}`))
		return
	}
	ctx := context.Background()
	in, api_err := ValidateCreateParams(r)

	if api_err != nil {
		w.WriteHeader(api_err.HTTPStatus)
		w.Write([]byte(`{"error": "` + api_err.Error() + `"}`))
		return
	}
	resp, err := obj.Create(ctx, *in)
	if err != nil {
		switch t := err.(type) {
		case ApiError:
			w.WriteHeader(t.HTTPStatus)
			w.Write([]byte(`{"error": "` + t.Error() + `"}`))
		case error:
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "` + err.Error() + `"}`))
		}
		return
	}
	json_out, _ := CompileJSONNewUser(resp)
	w.Write(json_out)
}
func (obj *OtherApi) HandlerCreate(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte(`{"error": "bad method"}`))
		return
	}

	if r.Header.Get("X-Auth") != "100500" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": "unauthorized"}`))
		return
	}
	ctx := context.Background()
	in, api_err := ValidateOtherCreateParams(r)

	if api_err != nil {
		w.WriteHeader(api_err.HTTPStatus)
		w.Write([]byte(`{"error": "` + api_err.Error() + `"}`))
		return
	}
	resp, err := obj.Create(ctx, *in)
	if err != nil {
		switch t := err.(type) {
		case ApiError:
			w.WriteHeader(t.HTTPStatus)
			w.Write([]byte(`{"error": "` + t.Error() + `"}`))
		case error:
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "` + err.Error() + `"}`))
		}
		return
	}
	json_out, _ := CompileJSONOtherUser(resp)
	w.Write(json_out)
}
func ValidateProfileParams(r *http.Request) (*ProfileParams, *ApiError) {
	obj := ProfileParams{}
	var req int
	_ = req
	var err_conv error
	_ = err_conv
	default_val := ""
	if r.FormValue("login") == "" {
		err := ApiError{http.StatusBadRequest, fmt.Errorf("login must be not empty")}
		return nil, &err
	}
	if default_val == "" {
		obj.Login = r.FormValue("login")
	} else {
		obj.Login = default_val
	}
	return &obj, nil
}
func ValidateCreateParams(r *http.Request) (*CreateParams, *ApiError) {
	obj := CreateParams{}
	var req int
	_ = req
	var err_conv error
	_ = err_conv
	default_val := ""
	if r.FormValue("login") == "" {
		err := ApiError{http.StatusBadRequest, fmt.Errorf("login must be not empty")}
		return nil, &err
	}
	if len(r.FormValue("login")) < 10 {
		err := ApiError{http.StatusBadRequest, fmt.Errorf("login len must be >= 10")}
		return nil, &err
	}
	if default_val == "" {
		obj.Login = r.FormValue("login")
	} else {
		obj.Login = default_val
	}
	if default_val == "" {
		obj.Name = r.FormValue("full_name")
	} else {
		obj.Name = default_val
	}
	if r.FormValue("status") == "" {
		default_val = "user"
	}
	k := 0
	if r.FormValue("status") != "user" && default_val != "user" {
		k++
	}
	if r.FormValue("status") != "moderator" && default_val != "moderator" {
		k++
	}
	if r.FormValue("status") != "admin" && default_val != "admin" {
		k++
	}
	if k == 3 {
		err := ApiError{http.StatusBadRequest, fmt.Errorf("status must be one of [user, moderator, admin]")}
		return nil, &err
	}
	if default_val == "" {
		obj.Status = r.FormValue("status")
	} else {
		obj.Status = default_val
	}
	req, err_conv = strconv.Atoi(r.FormValue("age"))
	if err_conv != nil {
		err := ApiError{http.StatusBadRequest, fmt.Errorf("age must be int")}
		return nil, &err
	}
	if req < 0 {
		err := ApiError{http.StatusBadRequest, fmt.Errorf("age must be >= 0")}
		return nil, &err
	}
	req, err_conv = strconv.Atoi(r.FormValue("age"))
	if err_conv != nil {
		err := ApiError{http.StatusBadRequest, fmt.Errorf("age must be int")}
		return nil, &err
	}
	if req > 128 {
		err := ApiError{http.StatusBadRequest, fmt.Errorf("age must be <= 128")}
		return nil, &err
	}
	obj.Age = req
	return &obj, nil
}
func CompileJSONUser(obj *User) ([]byte, error) {
	result := "{" + "\n"
	result = result + "\"error\"" + ":" + "\"" + "\"" + "," + "\n"
	result = result + "\"response\"" + ":" + "{" + "\n"
	result = result + "\"id\"" + ":" + strconv.FormatUint(obj.ID, 10) + "," + "\n"
	result = result + "\"login\"" + ":" + "\"" + obj.Login + "\"" + "," + "\n"
	result = result + "\"full_name\"" + ":" + "\"" + obj.FullName + "\"" + "," + "\n"
	result = result + "\"status\"" + ":" + strconv.FormatInt(int64(obj.Status), 10) + "\n"
	result = result + "}"
	result = result + "}"
	return []byte(result), nil
}
func CompileJSONNewUser(obj *NewUser) ([]byte, error) {
	result := "{" + "\n"
	result = result + "\"error\"" + ":" + "\"" + "\"" + "," + "\n"
	result = result + "\"response\"" + ":" + "{" + "\n"
	result = result + "\"id\"" + ":" + strconv.FormatUint(obj.ID, 10) + "\n"
	result = result + "}"
	result = result + "}"
	return []byte(result), nil
}
func ValidateOtherCreateParams(r *http.Request) (*OtherCreateParams, *ApiError) {
	obj := OtherCreateParams{}
	var req int
	_ = req
	var err_conv error
	_ = err_conv
	default_val := ""
	if r.FormValue("username") == "" {
		err := ApiError{http.StatusBadRequest, fmt.Errorf("username must be not empty")}
		return nil, &err
	}
	if len(r.FormValue("username")) < 3 {
		err := ApiError{http.StatusBadRequest, fmt.Errorf("username len must be >= 3")}
		return nil, &err
	}
	if default_val == "" {
		obj.Username = r.FormValue("username")
	} else {
		obj.Username = default_val
	}
	if default_val == "" {
		obj.Name = r.FormValue("account_name")
	} else {
		obj.Name = default_val
	}
	if r.FormValue("class") == "" {
		default_val = "warrior"
	}
	k := 0
	if r.FormValue("class") != "warrior" && default_val != "warrior" {
		k++
	}
	if r.FormValue("class") != "sorcerer" && default_val != "sorcerer" {
		k++
	}
	if r.FormValue("class") != "rouge" && default_val != "rouge" {
		k++
	}
	if k == 3 {
		err := ApiError{http.StatusBadRequest, fmt.Errorf("class must be one of [warrior, sorcerer, rouge]")}
		return nil, &err
	}
	if default_val == "" {
		obj.Class = r.FormValue("class")
	} else {
		obj.Class = default_val
	}
	req, err_conv = strconv.Atoi(r.FormValue("level"))
	if err_conv != nil {
		err := ApiError{http.StatusBadRequest, fmt.Errorf("level must be int")}
		return nil, &err
	}
	if req < 1 {
		err := ApiError{http.StatusBadRequest, fmt.Errorf("level must be >= 1")}
		return nil, &err
	}
	req, err_conv = strconv.Atoi(r.FormValue("level"))
	if err_conv != nil {
		err := ApiError{http.StatusBadRequest, fmt.Errorf("level must be int")}
		return nil, &err
	}
	if req > 50 {
		err := ApiError{http.StatusBadRequest, fmt.Errorf("level must be <= 50")}
		return nil, &err
	}
	obj.Level = req
	return &obj, nil
}
func CompileJSONOtherUser(obj *OtherUser) ([]byte, error) {
	result := "{" + "\n"
	result = result + "\"error\"" + ":" + "\"" + "\"" + "," + "\n"
	result = result + "\"response\"" + ":" + "{" + "\n"
	result = result + "\"id\"" + ":" + strconv.FormatUint(obj.ID, 10) + "," + "\n"
	result = result + "\"login\"" + ":" + "\"" + obj.Login + "\"" + "," + "\n"
	result = result + "\"full_name\"" + ":" + "\"" + obj.FullName + "\"" + "," + "\n"
	result = result + "\"level\"" + ":" + strconv.FormatInt(int64(obj.Level), 10) + "\n"
	result = result + "}"
	result = result + "}"
	return []byte(result), nil
}
func (obj *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/profile":
		obj.HandlerProfile(w, r)
	case "/user/create":
		obj.HandlerCreate(w, r)

	default:
		// 404
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "unknown method"}`))
		return
	}
}
func (obj *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/create":
		obj.HandlerCreate(w, r)

	default:
		// 404
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "unknown method"}`))
		return
	}
}
