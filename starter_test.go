package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"

	"gopkg.in/mgo.v2/dbtest"
)

var router *mux.Router
var mockDBServer dbtest.DBServer

var invalidArrayPayload = []byte(`[{ "wrong_json_format": true }]`)
var invalidObjectPayload = []byte(`{ "wrong_json_format": true }`)

func TestMain(m *testing.M) {
	fmt.Println("--- Start Tests")
	setup()

	// run the test cases defined in this file
	retCode := m.Run()

	defer tearDown()

	// call with result of m.Run()
	os.Exit(retCode)
}

func setup() {
	fmt.Println("--- --- setup")
	router = makeRouter()
	setupDB()
	fillDB()
}

func setupDB() {
	tempDir, _ := ioutil.TempDir("", "testing")
	mockDBServer.SetPath(tempDir)

	mongoClient = mockDBServer.Session()
	MongoCreateCollectionIndexes(mongoClient)
}

func fillDB() {
	/*
	 * Insert fake data
	 */
	fmt.Println("Insert fake data")
	err := mongoClient.DB(database).C(collectionAppReviewsGooglePlay).Insert(
		AppReviewGooglePlay{
			ReviewID:       "1234567",
			PackageName:    "eu.openreq",
			Author:         "OpenReqUser",
			Date:           20190131,
			Rating:         5,
			Title:          "Tool usage",
			Body:           "I used the tool for over a year now and like it! I hope for more analytics features coming in the future.",
			PermaLink:      "www.openreq.eu",
			FeatureRequest: true,
			BugReport:      false,
		},
	)
	if err != nil {
		panic(err)
	}

	/*
	 * Insert fake observables
	 */
	err = mongoClient.DB(database).C(collectionObservableGooglePlay).Insert(
		ObservableGooglePlay{
			PackageName: "eu.openreq",
			Interval:    "2h",
		})
	if err != nil {
		panic(err)
	}
}

func tearDown() {
	fmt.Println("--- --- tear down")
	mongoClient.Close()
	mockDBServer.Stop() // Stop shuts down the temporary server and removes data on disk.
}

type endpoint struct {
	method string
	url    string
}

func (e endpoint) withVars(vs ...interface{}) endpoint {
	e.url = fmt.Sprintf(e.url, vs...)
	return e
}

func (e endpoint) executeRequest(payload interface{}) (error, *httptest.ResponseRecorder) {
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(payload)
	if err != nil {
		return err, nil
	}

	req, err := http.NewRequest(e.method, e.url, body)
	if err != nil {
		return err, nil
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	return nil, rr
}

func (e endpoint) mustExecuteRequest(payload interface{}) *httptest.ResponseRecorder {
	err, rr := e.executeRequest(payload)
	if err != nil {
		panic(errors.Wrap(err, `Could not execute request`))
	}

	return rr
}

func isSuccess(code int) bool {
	return code >= 200 && code < 300
}

func assertSuccess(t *testing.T, rr *httptest.ResponseRecorder) {
	if !isSuccess(rr.Code) {
		t.Errorf("Status code differs. Expected success.\n Got status %d (%s) instead", rr.Code, http.StatusText(rr.Code))
	}
}
func assertFailure(t *testing.T, rr *httptest.ResponseRecorder) {
	if isSuccess(rr.Code) {
		t.Errorf("Status code differs. Expected failure.\n Got status %d (%s) instead", rr.Code, http.StatusText(rr.Code))
	}
}

func assertJsonDecodes(t *testing.T, rr *httptest.ResponseRecorder, v interface{}) {
	err := json.Unmarshal(rr.Body.Bytes(), v)
	if err != nil {
		t.Error(errors.Wrap(err, "Expected valid json array"))
	}
}

func TestPostAppPageGooglePlay(t *testing.T) {
	ep := endpoint{"POST", "/hitec/repository/app/store/app-page/google-play/"}

	// Test for failure
	assertFailure(t, ep.mustExecuteRequest(invalidObjectPayload))

	// Test for success
	appPage := AppPageGooglePlay{
		Name:          "OpenReq",
		PackageName:   "eu.openreq",
		DateCrawled:   20190131,
		Category:      "Tools",
		USK:           "all ages",
		Price:         "free",
		PriceValue:    0.0,
		PriceCurrency: "EUR",
		Description:   "contains, among others, analytics features.",
		WhatsNew:      []string{"re-designed dashboard"},
		Rating:        5,
		StarsCount:    10000,
		CountPerRating: StarCountPerRating{
			Five:  10000,
			Four:  0,
			Three: 0,
			Two:   0,
			One:   0,
		},
		EstimatedDownloadNumber: 10000,
		DeveloperName:           "OpenReq Consortium",
		TopDeveloper:            true,
		ContainsAds:             false,
		InAppPurchases:          false,
		LastUpdate:              20190131,
		Os:                      "ANDROID",
		RequiresOsVersion:       "9.0",
		CurrentSoftwareVersion:  "beta",
		SimilarApps:             []string{"supersede"},
	}

	assertSuccess(t, ep.mustExecuteRequest(appPage))
}

func TestPostAppReviewGooglePlay(t *testing.T) {
	ep := endpoint{"POST", "/hitec/repository/app/store/app-review/google-play/"}

	// Test for failure
	assertFailure(t, ep.mustExecuteRequest(invalidObjectPayload))

	// Test for success
	reviews := []AppReviewGooglePlay{{
		ReviewID:       "1234567",
		PackageName:    "eu.openreq",
		Author:         "OpenReqUser",
		Date:           20190131,
		Rating:         5,
		Title:          "Tool usage",
		Body:           "I used the tool for over a year now and like it! I hope for more analytics features coming in the future.",
		PermaLink:      "www.openreq.eu",
		FeatureRequest: true,
		BugReport:      false,
	}}
	assertSuccess(t, ep.mustExecuteRequest(reviews))
}

func TestPostObserveAppGooglePlay(t *testing.T) {
	ep := endpoint{"POST", "/hitec/repository/app/observe/app/google-play/package-name/%s/interval/%s"}

	// Test for success
	assertSuccess(t, ep.withVars("test", "h1").mustExecuteRequest(nil))
}

func TestGetObsevableGooglePlay(t *testing.T) {
	ep := endpoint{"GET", "/hitec/repository/app/observable/google-play"}

	// Test for success
	response := ep.mustExecuteRequest(nil)
	assertSuccess(t, response)
	var observables []ObservableGooglePlay
	assertJsonDecodes(t, response, &observables)
	assert.Len(t, observables, 2)
}

func TestGetAppReviewsOfClass(t *testing.T) {
	ep := endpoint{"GET", "/hitec/repository/app/google-play/package-name/%s/class/%s"}

	// Test for success
	response := ep.withVars("eu.openreq", "feature_request").mustExecuteRequest(nil)
	assertSuccess(t, response)
	var reviews []AppReviewGooglePlay
	assertJsonDecodes(t, response, &reviews)
	assert.Len(t, reviews, 1)
}
