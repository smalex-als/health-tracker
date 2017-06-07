package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type H map[string]interface{}

func (h H) Get(key string) string {
	return h[key].(string)
}

const SERVER = "https://health-tracker-1366.appspot.com"

func main() {
	var auth *H
	var err error
	auth, err = doReq("POST", "", "/v1/users/signin/", H{"username": "smalex1234", "password": "golang"})
	if err != nil {
		panic(err)
	}
	auth, err = doReq("POST", "", "/v1/users/signin/", H{"username": "smalex", "password": "golang"})
	if err != nil {
		panic(err)
	}
	tokenAdmin := auth.Get("token")
	auth, err = doReq("POST", "", "/v1/users/signin/", H{"username": "kion2412", "password": "123456"})
	if err != nil {
		panic(err)
	}
	tokenUser := auth.Get("token")
	meals, err := doReq("GET", tokenUser, "/v1/meals/?mode=items", nil)
	if err != nil {
		panic(err)
	}
	prettyJson(meals)
	users, err := doReq("GET", tokenUser, "/v1/users/?mode=items", nil)
	if err != nil {
		panic(err)
	}
	prettyJson(users)
	newmeal, err := doReq("POST", tokenUser, "/v1/meals/", H{
		"amount":      600,
		"date":        "2017-06-06T00:00:00Z",
		"description": "Sushi",
		"time":        "1021",
		"typeId":      "3",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("============")
	prettyJson(newmeal)
	deleteresp, err := doReq("DELETE", tokenUser, "/v1/meals/7000001", nil)
	if err != nil {
		panic(err)
	}
	prettyJson(deleteresp)

	meals, err = doReq("GET", tokenAdmin, "/v1/meals/?mode=items", nil)
	if err != nil {
		panic(err)
	}
	prettyJson(meals)
	users, err = doReq("GET", tokenAdmin, "/v1/users/?mode=items", nil)
	if err != nil {
		panic(err)
	}
	prettyJson(users)
}

func doReq(method, token, path string, obj interface{}) (*H, error) {
	var reqReader io.Reader
	var reqBody []byte

	if obj != nil {
		reqBody, _ = json.Marshal(obj)
	}
	reqReader = strings.NewReader(string(reqBody))
	req, err := http.NewRequest(method, SERVER+path, reqReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-type", "application/json")
	if token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	}
	if resp, err := http.DefaultClient.Do(req); err != nil {
		return nil, err
	} else {
		fmt.Println()
		fmt.Printf("path = %+v\n", path)
		fmt.Printf("resp = %+v\n", resp.Status)
		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		mm := H{}
		err = json.Unmarshal([]byte(body), &mm)
		if err != nil {
			return nil, err
		}
		return &mm, nil
	}

	return nil, nil
}

func prettyJson(h *H) {
	b, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Print(string(b))
}
