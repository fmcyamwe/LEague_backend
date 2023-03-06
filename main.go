package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// Run with
//		go run .
// Send request with:
//		curl -F 'file=@/path/matrix.csv' "localhost:8080/echo"

// save all records in memory--might be bad for large stuff
func readFile(r *http.Request) ([][]string, error) {
	file, _, err := r.FormFile("file")
	if err != nil {
		//w.Write([]byte(fmt.Sprintf("error %s", err.Error())))
		return [][]string{}, err
	}
	defer file.Close()
	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return [][]string{}, err
	}
	return records, nil
}

func sendResponse(w http.ResponseWriter, response string, statusCode int) {
	w.WriteHeader(statusCode)
	w.Write([]byte(response))
}

func echoHandler(w http.ResponseWriter, r *http.Request) {

	records, err := readFile(r)
	if err != nil {
		sendResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var response string
	for _, row := range records {
		response = fmt.Sprintf("%s%s\n", response, strings.Join(row, ","))
	}
	//fmt.Fprint(w, response)
	sendResponse(w, response, http.StatusOK)
}

func flattenHandler(w http.ResponseWriter, r *http.Request) {

	records, err := readFile(r)
	if err != nil {
		sendResponse(w, err.Error(), http.StatusOK)
		return
	}
	if len(records) < 1 { //empty file error
		sendResponse(w, "invalid file", http.StatusInternalServerError)
		return
	}

	mangas := []string{}
	for _, row := range records {
		mangas = append(mangas, row...)
	}
	response := strings.Join(mangas, ",")

	sendResponse(w, response, http.StatusOK)
}

func doInvert(records [][]string) ([][]string, error) {
	numRows := len(records)
	numColumns := len(records[0])

	if numRows != numColumns { //i case it's not an m*m matrix and there is missing data
		return records, fmt.Errorf("unbalanced square")
	}

	m := make([][]string, numColumns)
	for i := range records {
		m[i] = make([]string, numRows)
	}

	//populate new matrix
	for i := 0; i < numRows; i++ {
		for j := 0; j < numColumns; j++ {
			m[j][i] = records[i][j]
		}
	}
	return m, nil
}

func invertHandler(w http.ResponseWriter, r *http.Request) {

	records, err := readFile(r)
	if err != nil {
		sendResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	inverted, errp := doInvert(records)
	if errp != nil {
		sendResponse(w, errp.Error(), http.StatusInternalServerError)
		return
	}

	var response string
	for _, row := range inverted {
		response = fmt.Sprintf("%s%s\n", response, strings.Join(row, ","))
	}

	sendResponse(w, response, http.StatusOK)
}

func doSumRow(r []string) int {
	result := 0
	for _, num := range r {
		n, err := strconv.Atoi(num)
		if err != nil {
			fmt.Printf("error: %v ....skipping\n", err)
			continue
		}
		result = result + n
	}
	return result
}

func sumHandler(w http.ResponseWriter, r *http.Request) {

	records, err := readFile(r)
	if err != nil {
		sendResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response string
	sum := 0
	for _, row := range records {
		sum = sum + doSumRow(row)
	}
	response = fmt.Sprintf("%d", sum)

	sendResponse(w, response, http.StatusOK)
}

func doMultiplyRow(r []string) int {
	result := 1
	for _, num := range r {
		n, err := strconv.Atoi(num)
		if err != nil {
			fmt.Printf("error: %v ....skipping\n", err)
			continue
		}
		result = result * n
	}
	return result

}

func multiplyHandler(w http.ResponseWriter, r *http.Request) {
	records, err := readFile(r)
	if err != nil {
		sendResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(records) < 1 { //empty file error
		sendResponse(w, "invalid file", http.StatusInternalServerError)
		return
	}

	product := 1
	for _, row := range records {
		product = product * doMultiplyRow(row)
	}

	response := fmt.Sprintf("%d", product)

	sendResponse(w, response, http.StatusOK)
}

func main() {
	http.HandleFunc("/echo", echoHandler)
	http.HandleFunc("/flatten", flattenHandler)
	http.HandleFunc("/invert", invertHandler)
	http.HandleFunc("/sum", sumHandler)
	http.HandleFunc("/multiply", multiplyHandler)

	//add decorator for handling panics errors thrown during processing--todo
	http.ListenAndServe(":8080", nil)
}
