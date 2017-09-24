package appengine
import "fmt"
import "net/http"
import "encoding/json"
import "io/ioutil"
import "errors"
import "google.golang.org/appengine"
import "google.golang.org/appengine/urlfetch"
import "crypto/tls"

const githubPath = "/projectinfo/v1/github.com/"
const apiGithubPath = "https://api.github.com/repos/"

//https://mholt.github.io/json-to-go/
type gitContriber struct {
	Login             string `json:"login"`
	Contributions     int    `json:"contributions"`
}

type responseJSON struct {
	Repo string              `json:"project"`
	Owner string             `json:"owner"`
	Comitter string          `json:"committer"`
	Commits int              `json:"commits"`
	Languages []string       `json:"language"`
}

//Fetches JSON from path
func getJSON(path string, r *http.Request) ([]byte, error) {
	if len(path) >= len(apiGithubPath) {
		//https://cloud.google.com/appengine/docs/standard/go/issue-requests
		if r != nil {
			ctx := appengine.NewContext(r)
			client := urlfetch.Client(ctx)
			resp, err := client.Get(path)
			if err != nil {
				return nil, err
			}
			data, err := ioutil.ReadAll(resp.Body)
			return data, err
		}
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		resp, err := client.Get(path)
		if err != nil {
			return nil, err
		}
		data, err := ioutil.ReadAll(resp.Body)
		return data, err
	}
	return nil, errors.New("Invalid path, expected " + apiGithubPath)
}

//Fetches JSON from the web and converts it to a map with strings as keys and interface as value
func getAndMapJSON(path string, r *http.Request) (map[string]interface{}, error) {
	data, err := getJSON(path, r)
	if err != nil {
		return nil, err
	}
	m := make(map[string]interface{})
	err = json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func processJSON(githubMap map[string]interface{}, r *http.Request) (*responseJSON, int, error) {
	var response = &responseJSON{}
	//Get name of repo
	if githubMap["full_name"] != nil {
		name, ok := githubMap["full_name"].(string)
		if !ok {
			return nil, http.StatusInternalServerError, errors.New("ERROR, unable to type assert 'full_name' to string")
		}
		response.Repo = "github.com/" + name
	} else {
		return nil, http.StatusBadRequest, errors.New("ERROR, malformed JSON, field 'full_name' not found")
	}
	//Get owner of repo
	if githubMap["owner"] != nil {
		ownerMap, ok := githubMap["owner"].(map[string]interface{})
		if !ok {
			return nil, http.StatusInternalServerError, errors.New("ERROR, unable to type assert 'owner' to map[string]interface{}")
		}
		if ownerMap["login"] != nil {
			owner, ok := ownerMap["login"].(string)
			if !ok {
				return nil, http.StatusInternalServerError, errors.New("ERROR, unable to type assert 'login' to string")
			} 
			response.Owner = owner
		} else {
			return nil, http.StatusBadRequest, errors.New("ERROR, malformed JSON, field 'login' not found")
		}
	} else {
		return nil, http.StatusBadRequest, errors.New("ERROR, malformed JSON, field 'owner' not found")
	}
	//Get the contributor with the most commits
	if githubMap["contributors_url"] != nil {
		contributorURL, ok := githubMap["contributors_url"].(string)
		if !ok {
			return nil, http.StatusInternalServerError, errors.New("ERROR, unable to type assert 'contributors_url' to string")
		}
		githubContribData, err := getJSON(contributorURL, r)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		var contribs = make([]gitContriber, 0)
		err = json.Unmarshal(githubContribData, &contribs)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		if len(contribs) > 0 {
			response.Comitter = contribs[0].Login
			response.Commits = contribs[0].Contributions
		}
	} else {
		return nil, http.StatusBadRequest, errors.New("ERROR, malformed JSON, field 'contributors_url' not found")
	}
	//Get languages
	if githubMap["languages_url"] != nil {
		languageURL, ok := githubMap["languages_url"].(string)
		if !ok {
			return nil, http.StatusInternalServerError ,errors.New("ERROR, unable to type assert 'languages_url' to string")
		}
		langMap, err := getAndMapJSON(languageURL, r)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		response.Languages = make([]string, 0)
		for lang := range langMap {
			response.Languages = append(response.Languages, lang)
		}
	} else {
		return nil, http.StatusBadRequest, errors.New("ERROR, malformed JSON, field 'languages_url' not found")
	}
	return response, http.StatusOK, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Path) >= len(githubPath) {
		githubURI := r.URL.Path[len(githubPath):]
		//Get the main json data
		githubMap, err := getAndMapJSON(apiGithubPath + githubURI, r)
		if err != nil {
			fmt.Fprint(w, err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		//Process the json
		response,status,err := processJSON(githubMap, r)
		if err != nil {
			fmt.Println(err)
			http.Error(w, http.StatusText(status), status)
			return
		}
		//Send data
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

func main() {
	http.HandleFunc(githubPath, handler)
	http.ListenAndServe(":8080", nil)
}