package main

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	//"encoding/csv"
)

// table driven test or test by endpoint?
// >>combination of both for error handling per route

type testCase struct {
	fileNameInput    string
	testName         string
	expectedResponse string
	//err              error
}

// Given a fileName paths and url endpoint, creates the form-data request or errors out
func createFormRequest(fileName, urlEndPoint string) (*http.Request, error) {
	tmpfile, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, tmpfile); err != nil {
		return nil, err
	}

	writer.Close()
	tmpfile.Close()

	req, err := http.NewRequest("POST", urlEndPoint, bytes.NewReader(body.Bytes()))
	if err != nil {
		//t.Fatal(err)
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType()) //"multipart/form-data" ==> huh this doesnt give boundary(new learning :)

	return req, nil
}

func TestFlattenHandler(t *testing.T) {
	testCases := []testCase{
		{
			fileNameInput:    "test1.csv",
			testName:         "Happy path",
			expectedResponse: "1,2,3,4,5,6,7,8,9",
		},
		{
			fileNameInput:    "test3.csv",
			testName:         "strings in data",
			expectedResponse: "1,2,3,4,5,6,7,8,f",
		},
		{
			fileNameInput:    "empty.csv",
			testName:         "Invalid file",
			expectedResponse: "invalid file",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			req, err := createFormRequest(tc.fileNameInput, "/flatten")
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(flattenHandler)

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				if tc.expectedResponse != rr.Body.String() {
					t.Errorf("handler returned wrong status code %d: got %swant %s", status,
						rr.Body.String(), tc.expectedResponse)
				}
			}

			if rr.Body.String() != tc.expectedResponse {
				t.Errorf("handler returned unexpected body: got %v want %v",
					rr.Body.String(), tc.expectedResponse)
			}
		})
	}
}

func TestInvertHandler(t *testing.T) {
	/*
		11,14,17
		12,15,18
		13,16,0
	*/
	testCases := []testCase{
		{
			fileNameInput:    "test1.csv",
			testName:         "Happy path",
			expectedResponse: "1,4,7\n2,5,8\n3,6,9\n", //prettyfy this to above format**
		},
		{
			fileNameInput:    "test2.csv",
			testName:         "unbalanced square", //hypothetical error for missing data but should be caught during parsing
			expectedResponse: "11,14,17\n12,15,18\n13,16,0\n",
		},
		{
			fileNameInput:    "test3.csv",
			testName:         "string in data",
			expectedResponse: "1,4,7\n2,5,8\n3,6,f\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			req, err := createFormRequest(tc.fileNameInput, "/invert")
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(invertHandler)

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, http.StatusOK)
			}

			if rr.Body.String() != tc.expectedResponse {
				t.Errorf("handler returned unexpected body: got %v want %v",
					rr.Body.String(), tc.expectedResponse)
			}
		})
	}
}

func TestMultiplyHandler(t *testing.T) {
	testCases := []testCase{
		{
			fileNameInput:    "test1.csv",
			testName:         "Happy path",
			expectedResponse: "362880",
		},
		{
			fileNameInput:    "test3.csv",
			testName:         "Parsing skip", //in case strings are found...skipped for this case
			expectedResponse: "40320",
		},
		{
			fileNameInput:    "empty.csv",
			testName:         "Invalid file",
			expectedResponse: "invalid file",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			req, err := createFormRequest(tc.fileNameInput, "/multiply")
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(multiplyHandler)

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				if tc.expectedResponse != rr.Body.String() {
					t.Errorf("handler returned wrong status code %d: got %s want %s", status,
						rr.Body.String(), tc.expectedResponse)
				}
			}

			if rr.Body.String() != tc.expectedResponse {
				t.Errorf("handler returned unexpected body: got %v want %v",
					rr.Body.String(), tc.expectedResponse)
			}
		})
	}
}

func TestSumHandler(t *testing.T) {
	testCases := []testCase{
		{
			fileNameInput:    "test1.csv",
			testName:         "Happy path",
			expectedResponse: "45",
		},
		{
			fileNameInput:    "test2.csv",
			testName:         "Happy path2",
			expectedResponse: "116",
		},
		{
			fileNameInput:    "test3.csv",
			testName:         "Parsing skip", //in case strings are found...skipped for this case
			expectedResponse: "36",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			req, err := createFormRequest(tc.fileNameInput, "/sum")
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(sumHandler)

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, http.StatusOK)
			}

			if rr.Body.String() != tc.expectedResponse {
				t.Errorf("handler returned unexpected body: got %v want %v",
					rr.Body.String(), tc.expectedResponse)
			}
		})
	}
}

func TestEchoHandler(t *testing.T) {
	testCases := []testCase{
		{
			fileNameInput:    "test1.csv",
			testName:         "Happy path",
			expectedResponse: "1,2,3\n4,5,6\n7,8,9\n",
		},
		{
			fileNameInput:    "test2.csv",
			testName:         "Happy path2",
			expectedResponse: "11,12,13\n14,15,16\n17,18,0\n",
		},
		{
			fileNameInput:    "test3.csv",
			testName:         "Happy path3",
			expectedResponse: "1,2,3\n4,5,6\n7,8,f\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			req, err := createFormRequest(tc.fileNameInput, "/echo")
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(echoHandler)

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, http.StatusOK)
			}

			if rr.Body.String() != tc.expectedResponse {
				t.Errorf("handler returned unexpected body: got %v want %v",
					rr.Body.String(), tc.expectedResponse)
			}
		})
	}
}
