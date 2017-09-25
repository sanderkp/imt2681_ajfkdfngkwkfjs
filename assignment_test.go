package appengine

import "testing"
import "encoding/json"
import "net/http"
import "net/http/httptest"
import "fmt"

func Test_getJSON(t *testing.T) {
	bogusPaths := []string{"", "https://com.hubgit.api/repos/", "/projectinfo/v1/github.com/", "sdfhasdkjfhasldkjfhlaskdjfhladskjfhlksdjfhlskjdfh"}
	goodPaths := []string{"https://api.github.com/repos/golang/go", "https://api.github.com/repos/apache/kafka", "https://api.github.com/repos/nothings/stb"}

	//Test bogus paths
	for _, bogus := range bogusPaths {
		data, err := getJSON(bogus, nil)
		if data != nil {
			t.Errorf("Expected data to be nil, got: %p path is:%s error is: %s", data, bogus, err)
		}
	}
	//Test good paths
	for _, good := range goodPaths {
		data, err := getJSON(good, nil)
		if data == nil {
			t.Errorf("Failed to get data from %s, error is: %s", good, err)
		}
	}
}

func Test_getAndMapJSON(t *testing.T) {
	//Test bogus path
	bogusPath := "https://com.hubgit.api/repos/"
	data, err := getAndMapJSON(bogusPath, nil)
	if data != nil {
		t.Errorf("Expected data to be nil, got: %p path is:%s error is: %s", data, bogusPath, err)
	}
	//Test invalid json
	invldJsn := "https://sanderkp.no/invalid_json.json"
	data, err = getAndMapJSON(invldJsn, nil)
	if data != nil {
		t.Errorf("Expected data to be nil, got: %p path is:%s error is: %s", data, invldJsn, err)
	}
	//Test good path
	goodPath := "https://api.github.com/repos/nothings/stb"
	data, err = getAndMapJSON(goodPath, nil)
	if data == nil {
		t.Errorf("Failed to create map for path %s, error is: %s", goodPath, err)
	}
}

func testProcessJSON(t *testing.T, s string) (*responseJSON, int, error) {
	fmt.Println(s)
	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
		t.Errorf("Failed to unmarshal test json: %s error: %s", s, err)
		return nil, 0, err
	}
	resp, status, error := processJSON(m, nil)
	return resp, status, error
}

func Test_processJSON(t *testing.T) {
	//Test with json that does not have 'full_name' key
	resp, status, error := testProcessJSON(t, "{\"abc\": 123}")
	if resp != nil {
		t.Errorf("Expected resp from json with no 'full_name' to yield no response, got: %p, error is: %s, status is: %d", resp, error, status)
	}
	if status != http.StatusBadRequest {
		t.Errorf("Expected status from json with no 'full_name' to yield no StatusBadRequest, got: %p, error is: %s, status is: %d", resp, error, status)
	}
	//Test with json that does have 'full_name' key but is not a string
	resp, status, error = testProcessJSON(t, "{\"full_name\":123}")
	if resp != nil {
		t.Errorf("Expected resp from json with 'full_name' of type not string to yield no response, got: %p, error is: %s, status is: %d", resp, error, status)
	}
	if status != http.StatusInternalServerError {
		t.Errorf("Expected status from json with 'full_name' of type not string to yield StatusInternalServerError, got: %p, error is: %s, status is: %d", resp, error, status)
	}
	//Test with json that does not have 'owner' key
	resp, status, error = testProcessJSON(t, "{\"full_name\":\"yo\"}")
	if resp != nil {
		t.Errorf("Expected resp from json with no 'owner' to yield no response, got: %p, error is: %s, status is: %d", resp, error, status)
	}
	if status != http.StatusBadRequest {
		t.Errorf("Expected status from json with no 'owner' to yield no StatusBadRequest, got: %p, error is: %s, status is: %d", resp, error, status)
	}
	//Test with json that does have 'owner' key but is not a map[string]interface{}
	resp, status, error = testProcessJSON(t, "{\"full_name\":\"yo\",\"owner\":123}")
	if resp != nil {
		t.Errorf("Expected resp from json with 'owner' of type not map[string]interface{} to yield no response, got: %p, error is: %s, status is: %d", resp, error, status)
	}
	if status != http.StatusInternalServerError {
		t.Errorf("Expected status from json with 'owner' of type not map[string]interface{} to yield StatusInternalServerError, got: %p, error is: %s, status is: %d", resp, error, status)
	}
	//Test with json that does have 'owner' but not login key
	resp, status, error = testProcessJSON(t, "{\"full_name\":\"yo\",\"owner\":{\"test\":123}}")
	if resp != nil {
		t.Errorf("Expected resp from json with 'owner' but not 'login'to yield no response, got: %p, error is: %s, status is: %d", resp, error, status)
	}
	if status != http.StatusBadRequest {
		t.Errorf("Expected resp from json with 'owner' but not 'login'to yield StatusBadRequest, got: %p, error is: %s, status is: %d", resp, error, status)
	}
	//Test with json that does have 'owner' and 'login' but which login key is not string
	resp, status, error = testProcessJSON(t, "{\"full_name\":\"yo\",\"owner\":{\"login\":123}}")
	if resp != nil {
		t.Errorf("Expected resp from json with 'owner' and 'login' where login is not string to yield no response, got: %p, error is: %s, status is: %d", resp, error, status)
	}
	if status != http.StatusInternalServerError {
		t.Errorf("Expected resp from json with 'owner' and 'login' where login is not string to yield StatusInternalServerError, got: %p, error is: %s, status is: %d", resp, error, status)
	}
	//Test with json that does not have 'contributors_url'
	resp, status, error = testProcessJSON(t, "{\"full_name\":\"yo\",\"owner\":{\"login\":\"bob\"}}")
	if resp != nil {
		t.Errorf("Expected resp from json with no 'contributors_url' to yield no response, got: %p, error is: %s, status is: %d", resp, error, status)
	}
	if status != http.StatusBadRequest {
		t.Errorf("Expected resp from json with no 'contributors_url' to yield StatusBadRequest, got: %p, error is: %s, status is: %d", resp, error, status)
	}
	//Test with json that does have 'contributors_url' but is not string
	resp, status, error = testProcessJSON(t, "{\"full_name\":\"yo\",\"owner\":{\"login\":\"bob\"}, \"contributors_url\":123}")
	if resp != nil {
		t.Errorf("Expected resp from json with 'contributors_url' which is not string to yield no response, got: %p, error is: %s, status is: %d", resp, error, status)
	}
	if status != http.StatusInternalServerError {
		t.Errorf("Expected resp from json with 'contributors_url' which is not string to yield StatusInternalServerError, got: %p, error is: %s, status is: %d", resp, error, status)
	}
	//Test with json that does have 'contributors_url' but which points to bogus
	resp, status, error = testProcessJSON(t, "{\"full_name\":\"yo\",\"owner\":{\"login\":\"bob\"}, \"contributors_url\":\"bab\"}")
	if resp != nil {
		t.Errorf("Expected resp from json with 'contributors_url' which points to bogus to yield no response, got: %p, error is: %s, status is: %d", resp, error, status)
	}
	if status != http.StatusInternalServerError {
		t.Errorf("Expected resp from json with 'contributors_url' which points to bogus to yield StatusInternalServerError, got: %p, error is: %s, status is: %d", resp, error, status)
	}
	//Test with json that does have 'contributors_url' but which points to a json with bogus
	resp, status, error = testProcessJSON(t, "{\"full_name\":\"yo\",\"owner\":{\"login\":\"bob\"}, \"contributors_url\":\"https://sanderkp.no/invalid_json.json\"}")
	if resp != nil {
		t.Errorf("Expected resp from json with 'contributors_url' which points to a json with bogus to yield no response, got: %p, error is: %s, status is: %d", resp, error, status)
	}
	if status != http.StatusInternalServerError {
		t.Errorf("Expected resp from json with 'contributors_url' which points to a json with bogus to yield StatusInternalServerError, got: %p, error is: %s, status is: %d", resp, error, status)
	}
	//Test with json that does not have 'languages_url'
	resp, status, error = testProcessJSON(t, "{\"full_name\":\"yo\",\"owner\":{\"login\":\"bob\"}, \"contributors_url\":\"https://sanderkp.no/golang_go_contributors.json\"}")
	if resp != nil {
		t.Errorf("Expected resp from json with no 'languages_url' to yield no response, got: %p, error is: %s, status is: %d", resp, error, status)
	}
	if status != http.StatusBadRequest {
		t.Errorf("Expected resp from json with no 'languages_url' to yield StatusBadRequest, got: %p, error is: %s, status is: %d", resp, error, status)
	}
	//Test with json that does have 'languages_url' but is not string
	resp, status, error = testProcessJSON(t, "{\"full_name\":\"yo\",\"owner\":{\"login\":\"bob\"}, \"contributors_url\":\"https://sanderkp.no/golang_go_contributors.json\", \"languages_url\":123}")
	if resp != nil {
		t.Errorf("Expected resp from json with 'languages_url' not of type string to yield no response, got: %p, error is: %s, status is: %d", resp, error, status)
	}
	if status != http.StatusInternalServerError {
		t.Errorf("Expected resp from json with 'languages_url' not of type string to yield StatusInternalServerError, got: %p, error is: %s, status is: %d", resp, error, status)
	}
	//Test with json that does have 'languages_url' but does not point to anything valid
	resp, status, error = testProcessJSON(t, "{\"full_name\":\"yo\",\"owner\":{\"login\":\"bob\"}, \"contributors_url\":\"https://sanderkp.no/golang_go_contributors.json\", \"languages_url\":\"bab\"}")
	if resp != nil {
		t.Errorf("Expected resp from json with 'languages_url' which does not point anything valid to yield no response, got: %p, error is: %s, status is: %d", resp, error, status)
	}
	if status != http.StatusInternalServerError {
		t.Errorf("Expected resp from json with 'languages_url' which does not point anything valid to yield StatusInternalServerError, got: %p, error is: %s, status is: %d", resp, error, status)
	}
	//Test with valid json
	resp, status, error = testProcessJSON(t, "{\"full_name\":\"yo\",\"owner\":{\"login\":\"bob\"}, \"contributors_url\":\"https://sanderkp.no/golang_go_contributors.json\", \"languages_url\":\"https://sanderkp.no/golang_go_languages.json\"}")
	if resp == nil {
		t.Errorf("Expected valid json to work, resp is: %p, error is: %s, status is: %d", resp, error, status)
	}
}

func Test_handler(t *testing.T) {
	//"STOLEN" from here: https://github.com/marni/imt2681_studentdb/blob/master/api_student_test.go#L12
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()
	//Bogus paths
	tests := []string{ts.URL + "", ts.URL + "////", ts.URL + "/git"}
	for _, test := range tests {
		resp, err := http.Get(test)
		if err != nil {
			t.Errorf("Error, could not make the GET request, err: %s", err)
			continue
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status code StatusBadRequest, got: %d, url is: %s", resp.StatusCode, test)
		}
	}
}
