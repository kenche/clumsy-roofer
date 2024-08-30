package main_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/kenche/clumsy-roofer/app"
	"github.com/stretchr/testify/assert"
)

type Risk struct {
	ID          string `json:"id"`
	State       string `json:"state"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func TestServer(t *testing.T) {

	t.Run("No items on start", func(t *testing.T) {
		handler := app.NewServer(false)
		svr := httptest.NewServer(handler)
		defer svr.Close()

		resp, err := http.Get(svr.URL + "/v1/risks")
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		var riskList []Risk
		err = json.Unmarshal(body, &riskList)
		assert.NoError(t, err)
		assert.Empty(t, riskList)

	})

	t.Run("GET Non existent item returns 404", func(t *testing.T) {
		handler := app.NewServer(false)
		svr := httptest.NewServer(handler)
		defer svr.Close()

		randomID, _ := uuid.NewUUID()
		resp, err := http.Get(svr.URL + "/v1/risks/" + randomID.String())
		assert.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode)
	})

	t.Run("GET with invalid ID returns 400", func(t *testing.T) {
		handler := app.NewServer(false)
		svr := httptest.NewServer(handler)
		defer svr.Close()

		resp, err := http.Get(svr.URL + "/v1/risks/c0ffee")
		assert.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})

	t.Run("Post one item", func(t *testing.T) {
		handler := app.NewServer(false)
		svr := httptest.NewServer(handler)
		defer svr.Close()

		newRisk := Risk{
			State:       "open",
			Title:       "foo",
			Description: "hello 世界",
		}
		b, err := json.Marshal(newRisk)
		assert.NoError(t, err)
		resp, err := http.Post(svr.URL+"/v1/risks", "application/json", bytes.NewReader(b))
		assert.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode)
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		var createdRisk Risk
		err = json.Unmarshal(body, &createdRisk)
		assert.NoError(t, err)
		assert.NotEmpty(t, createdRisk.ID)
		assert.Equal(t, newRisk.State, createdRisk.State)
		assert.Equal(t, newRisk.Title, createdRisk.Title)
		assert.Equal(t, newRisk.Description, createdRisk.Description)
	})

	t.Run("Missing state returns 400", func(t *testing.T) {
		handler := app.NewServer(false)
		svr := httptest.NewServer(handler)
		defer svr.Close()

		newRisk := Risk{
			Title:       "foo",
			Description: "hello world",
		}
		b, err := json.Marshal(newRisk)
		assert.NoError(t, err)
		resp, err := http.Post(svr.URL+"/v1/risks", "application/json", bytes.NewReader(b))
		assert.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})

	t.Run("Invalid state returns 400", func(t *testing.T) {
		handler := app.NewServer(false)
		svr := httptest.NewServer(handler)
		defer svr.Close()

		newRisk := Risk{
			State:       "half_open",
			Title:       "foo",
			Description: "hello world",
		}
		b, err := json.Marshal(newRisk)
		assert.NoError(t, err)
		resp, err := http.Post(svr.URL+"/v1/risks", "application/json", bytes.NewReader(b))
		assert.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
	})

	t.Run("GET one item", func(t *testing.T) {
		handler := app.NewServer(false)
		svr := httptest.NewServer(handler)
		defer svr.Close()

		newRisk := Risk{
			State:       "investigating",
			Title:       "bar",
			Description: "hello 世界",
		}
		b, err := json.Marshal(newRisk)
		assert.NoError(t, err)
		resp, err := http.Post(svr.URL+"/v1/risks", "application/json", bytes.NewReader(b))
		assert.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode)
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		var createdRisk Risk
		err = json.Unmarshal(body, &createdRisk)
		assert.NoError(t, err)
		assert.NotEmpty(t, createdRisk.ID)

		var gotRisk Risk
		resp, err = http.Get(svr.URL + "/v1/risks/" + createdRisk.ID)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		defer resp.Body.Close()
		body, err = io.ReadAll(resp.Body)
		assert.NoError(t, err)

		err = json.Unmarshal(body, &gotRisk)
		assert.NoError(t, err)
		assert.Equal(t, newRisk.State, gotRisk.State)
		assert.Equal(t, newRisk.Title, gotRisk.Title)
		assert.Equal(t, newRisk.Description, gotRisk.Description)

	})

	t.Run("Concurrently Post items then List", func(t *testing.T) {
		handler := app.NewServer(false)
		svr := httptest.NewServer(handler)
		defer svr.Close()

		possibleStates := []string{
			"open", "closed", "accepted", "investigating",
		}

		var wg sync.WaitGroup
		risksPerState := 1000
		for i := 0; i < len(possibleStates); i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				for j := 0; j < risksPerState; j++ {
					newRisk := Risk{
						State:       possibleStates[i],
						Title:       "foo",
						Description: "ハローワールド",
					}
					b, err := json.Marshal(newRisk)
					assert.NoError(t, err)
					resp, err := http.Post(svr.URL+"/v1/risks", "application/json", bytes.NewReader(b))
					assert.NoError(t, err)
					assert.Equal(t, 201, resp.StatusCode)
					defer resp.Body.Close()
					_, err = io.ReadAll(resp.Body)
					assert.NoError(t, err)
				}

			}(i)
		}

		wg.Wait()

		resp, err := http.Get(svr.URL + "/v1/risks")
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		var riskList []Risk
		err = json.Unmarshal(body, &riskList)
		assert.NoError(t, err)
		assert.Equal(t, len(possibleStates)*risksPerState, len(riskList))

		createdRisk := riskList[len(riskList)-1]
		assert.NotEmpty(t, createdRisk.ID)
		assert.NotEmpty(t, createdRisk.State)
		assert.Equal(t, createdRisk.Title, "foo")
		assert.Equal(t, createdRisk.Description, "ハローワールド")
	})

}
