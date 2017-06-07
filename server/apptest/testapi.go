package apptest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"google.golang.org/appengine/aetest"

	"github.com/gin-gonic/gin"
)

type commonExpectedErrors struct {
	Errors []*ClientError
}

type ClientError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type TestCommand struct {
	Name         string
	Href         string
	Src          interface{}
	Dst          interface{}
	Code         int
	ErrorMessage string
	Method       string
	Token        string
	Validate     func(t *testing.T, cmd *TestCommand)
}

func (t *TestCommand) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(t.Name)
	// buffer.WriteString(strings.Replace(t.Href, "/", " ", -1))
	// if t.Code == 200 {
	// 	buffer.WriteString(" successful")
	// } else {
	// 	if t.Error != nil {
	// 		buffer.WriteString(" errors: ")
	// 		buffer.WriteString(t.Error.Message)
	// 	}
	// }
	return buffer.String()
}

func CommonApiRunnerAll(
	r *gin.Engine,
	t *testing.T,
	inst aetest.Instance,
	commands []*TestCommand,
) {
	for _, test := range commands {
		if CommonApiRunner(r, t, inst, test) {
			t.Fatal()
		}
	}
}

func CommonApiRunner(
	r *gin.Engine,
	t *testing.T,
	inst aetest.Instance,
	cmd *TestCommand,
) bool {
	expectedError := cmd.ErrorMessage
	var reqReader io.Reader
	var reqBody []byte
	if cmd.Src != nil {
		reqBody, _ = json.Marshal(&cmd.Src)
	}
	reqReader = strings.NewReader(string(reqBody))
	method := "POST"
	if cmd.Method != "" {
		method = cmd.Method
	}
	req, err := inst.NewRequest(method, cmd.Href, reqReader)
	if cmd.Token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", cmd.Token))
	}
	if err != nil {
		t.Fatalf("Failed to create req: %v", err)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	printResp := false
	if w.Code != cmd.Code {
		printResp = true
		t.Errorf("Response code should be %d, was: %d", cmd.Code, w.Code)
	}
	bodyAsString := w.Body.String()
	if len(bodyAsString) > 0 {
		var formResp commonExpectedErrors
		err = json.Unmarshal([]byte(bodyAsString), &formResp)
		if err != nil {
			t.Errorf("Failed to parse response json %s\n\n%s", err, bodyAsString)
		}
		if expectedError != "" {
			if len(formResp.Errors) == 0 {
				t.Errorf("Expected error not found %+v", expectedError)
				printResp = true
			} else if formResp.Errors[0].Message != expectedError {
				t.Errorf("Unexpected error %+v", formResp.Errors[0])
				printResp = true
			}
		} else if len(formResp.Errors) > 0 {
			t.Errorf("Unexpected error %+v", formResp.Errors[0])
			printResp = true
		}
		if cmd.Dst != nil {
			err = json.Unmarshal([]byte(bodyAsString), &cmd.Dst)
			if err != nil {
				t.Errorf("Failed to parse response json to dst %s\n\n%s", err, bodyAsString)
			}
		}
	}
	if !printResp && cmd.Validate != nil {
		cmd.Validate(t, cmd)
	}

	if printResp || t.Failed() {
		t.Logf("name = %+v\n", cmd.Name)
		t.Logf("path = %+v\n", cmd.Href)
		t.Log("Fail: " + cmd.String() + " ")
		t.Logf("reqBody = %+v\n", string(reqBody))
		t.Logf("response = %+v\n", bodyAsString)
	} else {
		t.Log("OK " + cmd.String())

	}
	return printResp
}
