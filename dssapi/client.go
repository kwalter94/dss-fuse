package dssapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type Client struct {
	baseUrl    string
	apiKey     string
	httpClient http.Client
}

func NewDssClient(url string, apiKey string) (*Client, error) {
	if url == "" {
		return nil, fmt.Errorf("url can not be be blank")
	}

	client := Client{baseUrl: url, apiKey: apiKey}
	err := client.TestConnection()
	if err != nil {
		return nil, err
	}

	return &client, nil
}

func (dss *Client) TestConnection() error {
	// TODO: Test the api key too
	url := fmt.Sprintf("%s/dip/api/get-configuration", dss.baseUrl)
	log.Printf("Pinging DSS at %s", url)

	response, err := dss.httpClient.Get(url)
	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}

		log.Printf("DSS is not available: %v", body)
		return fmt.Errorf("DSS is not available")
	}

	return nil
}

// Lists all projects accessible to apiKey
func (dss *Client) GetProjects() ([]Project, error) {
	log.Print("Fetching projects...")

	request, err := dss.makeApiRequest("GET", "projects/", nil)
	if err != nil {
		return nil, err
	}

	response, err := dss.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var projects []Project
	err = readResponseJSON(response, &projects)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

func (dss *Client) GetRecipes(project Project) ([]Recipe, error) {
	log.Printf("Fetch recipes from project %s", project.Name)

	path := fmt.Sprintf("projects/%s/recipes/", project.ProjectKey)
	request, err := dss.makeApiRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	response, err := dss.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var recipes []Recipe
	err = readResponseJSON(response, &recipes)
	if err != nil {
		return nil, err
	}

	return recipes, nil
}

type recipePayload struct {
	Payload string `json:"payload"`
}

func (dss *Client) GetRecipePayload(recipe Recipe) (string, error) {
	log.Printf("Fetching payload for recipe: %s.%s", recipe.ProjectKey, recipe.Name)

	path := fmt.Sprintf("projects/%s/recipes/%s/", recipe.ProjectKey, recipe.Name)
	request, err := dss.makeApiRequest("GET", path, nil)
	if err != nil {
		return "", err
	}

	response, err := dss.httpClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	var payload recipePayload
	err = readResponseJSON(response, &payload)
	if err != nil {
		return "", err
	}

	return payload.Payload, nil
}

func (dss *Client) SaveRecipePayload(recipe Recipe, payload string) error {
	log.Printf("Saving recipe %s.%s", recipe.ProjectKey, recipe.Name)

	path := fmt.Sprintf("projects/%s/recipes/%s/", recipe.ProjectKey, recipe.Name)

	requestBody, err := json.Marshal(recipePayload{Payload: payload})
	if err != nil {
		return err
	}

	request, err := dss.makeApiRequest("PUT", path, requestBody)
	if err != nil {
		return err
	}

	response, err := dss.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	return nil
}

func (dss *Client) makeApiRequest(method string, path string, body []byte) (*http.Request, error) {
	url := fmt.Sprintf("%s/public/api/%s", dss.baseUrl, path)
	log.Printf("Preparing request: %s %s", method, url)

	var bodyReader io.Reader = nil

	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	request, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	request.SetBasicAuth(dss.apiKey, "")
	request.Header.Set("Content-type", "application/json")

	return request, nil
}

func readResponseJSON(response *http.Response, target any) error {
	contentType := response.Header.Get("Content-type")
	if !strings.Contains(contentType, "application/json") {
		return fmt.Errorf("expected a JSON response but got %s instead", contentType)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, target)
	if err != nil {
		return err
	}

	return nil
}
