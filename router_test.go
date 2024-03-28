package main

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouter(t *testing.T) {
	tests := []struct {
		name          string
		endpoint      string
		configuration configuration
		token         string
		expectedCode  int
		expectedBody  string
	}{
		{
			"When the healthcheck endpoint is called, it should return 200",
			"/health",
			configuration{},
			"",
			http.StatusOK,
			"OK",
		},
		{
			"When the hook endpoint is called without a token, it should return 401",
			"/hook?script=b9f71a96-0d23-11ee-860e-ff55b106c448",
			configuration{DefaultToken: "test", Scripts: []script{{ID: parseUUIDOrPanic("b9f71a96-0d23-11ee-860e-ff55b106c448")}}},
			"",
			http.StatusUnauthorized,
			"Missing authorization token\n",
		},
		{

			"When the hook endpoint is called with a bad token, it should return 401",
			"/hook?script=b9f71a96-0d23-11ee-860e-ff55b106c448",
			configuration{DefaultToken: "test", Scripts: []script{{ID: parseUUIDOrPanic("b9f71a96-0d23-11ee-860e-ff55b106c448")}}},
			"nonya",
			http.StatusUnauthorized,
			"Invalid authorization token\n",
		},
		{
			"When the hook endpoint is called with a token without a script, it should return 400",
			"/hook",
			configuration{DefaultToken: "test"},
			"test",
			http.StatusBadRequest,
			"Missing script parameter or invalid script parameter\n",
		},
		{
			"When the hook endpoint is called with an non-existent script, it should return 404",
			"/hook?script=b9f71a96-0d23-11ee-860e-ff55b106c448",
			configuration{DefaultToken: "test"},
			"test",
			http.StatusNotFound,
			"Script not found\n",
		},
		{
			"When the hook endpoint is called with an non-uuid script it should return 400",
			"/hook?script=blabla",
			configuration{DefaultToken: "test"},
			"test",
			http.StatusBadRequest,
			"Missing script parameter or invalid script parameter\n",
		},
		{
			"When the hook endpoint is called with a correct script and token, it should return 200",
			"/hook?script=b9f71a96-0d23-11ee-860e-ff55b106c448",
			configuration{DefaultToken: "test", Scripts: []script{{ID: parseUUIDOrPanic("b9f71a96-0d23-11ee-860e-ff55b106c448"), Path: "./scripts/success.sh"}}},
			"test",
			http.StatusOK,
			"ok\n",
		},
		{
			"When the hook endpoint is called with a correct script that fails and token, it should return 500",
			"/hook?script=b9f71a96-0d23-11ee-860e-ff55b106c448",
			configuration{DefaultToken: "test", Scripts: []script{{ID: parseUUIDOrPanic("b9f71a96-0d23-11ee-860e-ff55b106c448"), Path: "./scripts/failure.sh"}}},
			"test",
			http.StatusInternalServerError,
			"ko\n\n\nexit status 1\n",
		},
		{
			"When the hook endpoint is called with a correct script that uses its own token using the default token, it should return 401",
			"/hook?script=b9f71a96-0d23-11ee-860e-ff55b106c448",
			configuration{DefaultToken: "test", Scripts: []script{{ID: parseUUIDOrPanic("b9f71a96-0d23-11ee-860e-ff55b106c448"), Path: "./scripts/success.sh", Token: "nonya"}}},
			"test",
			http.StatusUnauthorized,
			"Invalid authorization token\n",
		},
		{
			"When the hook endpoint is called with a correct script that uses its own token using the this token, it should return 200",
			"/hook?script=b9f71a96-0d23-11ee-860e-ff55b106c448",
			configuration{DefaultToken: "test", Scripts: []script{{ID: parseUUIDOrPanic("b9f71a96-0d23-11ee-860e-ff55b106c448"), Path: "./scripts/success.sh", Token: "nonya"}}},
			"nonya",
			http.StatusOK,
			"ok\n",
		},
		{
			"When the hook endpoint is called with a script that uses inline then it should return 200",
			"/hook?script=47878e38-a700-11ee-bc6d-f3d25921fcde",
			configuration{DefaultToken: "test", Scripts: []script{{ID: parseUUIDOrPanic("47878e38-a700-11ee-bc6d-f3d25921fcde"), Inline: "echo inline", Token: "nonya"}}},
			"nonya",
			http.StatusOK,
			"inline\n",
		},
		{
			"When an environment variable is specified then it should be passed to the script",
			"/hook?script=47878e38-a700-11ee-bc6d-f3d25921fcde",
			configuration{DefaultToken: "test", Scripts: []script{{ID: parseUUIDOrPanic("47878e38-a700-11ee-bc6d-f3d25921fcde"), Inline: "echo $NAME", Token: "nonya", Environment: []environment{{Key: "NAME", Value: "Gandalf"}}}}},
			"nonya",
			http.StatusOK,
			"Gandalf\n",
		},
		{
			"When a global environment variable is specified then it should be passed to the script",
			"/hook?script=47878e38-a700-11ee-bc6d-f3d25921fcde",
			configuration{DefaultToken: "test", Environment: []environment{{Key: "NAME", Value: "Gandalf"}}, Scripts: []script{{ID: parseUUIDOrPanic("47878e38-a700-11ee-bc6d-f3d25921fcde"), Inline: "echo $NAME", Token: "nonya"}}},
			"nonya",
			http.StatusOK,
			"Gandalf\n",
		},
		{
			"When the same global environment variable and script one are is specified then the script should use the script scoped one",
			"/hook?script=47878e38-a700-11ee-bc6d-f3d25921fcde",
			configuration{DefaultToken: "test", Environment: []environment{{Key: "NAME", Value: "Gandalf"}}, Scripts: []script{{ID: parseUUIDOrPanic("47878e38-a700-11ee-bc6d-f3d25921fcde"), Inline: "echo $NAME", Token: "nonya", Environment: []environment{{Key: "NAME", Value: "frodo"}}}}},
			"nonya",
			http.StatusOK,
			"frodo\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := getRouter(test.configuration)
			req, _ := http.NewRequest("GET", test.endpoint, nil)
			req.Header.Set("Authorization", test.token)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			assert.Equal(t, test.expectedCode, rr.Code)
			if test.expectedBody != "" {
				assert.Equal(t, test.expectedBody, rr.Body.String())
			}
		})
	}
}

func parseUUIDOrPanic(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		panic(err)
	}
	return id
}
