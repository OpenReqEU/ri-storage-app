package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

func TestMain(m *testing.M) {
	fmt.Println("--- Start Tests")
	setup()

	// run the test cases defined in this file
	retCode := m.Run()

	tearDown()

	// call with result of m.Run()
	os.Exit(retCode)
}

func setup() {
	fmt.Println("--- --- setup")
	setupRouter()
	setupDB()
	fillDB()
}

func setupRouter() {
	router = mux.NewRouter()
	// Insert
	router.HandleFunc("/hitec/repository/app/store/app-page/google-play/", postAppPageGooglePlay).Methods("POST")
	router.HandleFunc("/hitec/repository/app/store/app-review/google-play/", postAppReviewGooglePlay).Methods("POST")
	router.HandleFunc("/hitec/repository/app/observe/app/google-play/package-name/{package_name}/interval/{interval}", postObserveAppGooglePlay).Methods("POST")
	router.HandleFunc("/hitec/repository/app/non-existing/app-review/google-play/", postNonExistingAppReviewsGooglePlay).Methods("POST")

	// Get
	router.HandleFunc("/hitec/repository/app/observable/google-play", getObsevableGooglePlay).Methods("GET")
	router.HandleFunc("/hitec/repository/app/google-play/package-name/{package_name}/class/{class}", getAppReviewsOfClass).Methods("GET")
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

func buildRequest(method, endpoint string, payload io.Reader, t *testing.T) *http.Request {
	req, err := http.NewRequest(method, endpoint, payload)
	if err != nil {
		t.Errorf("An error occurred. %v", err)
	}

	return req
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	return rr
}

func TestPostAppPageGooglePlay(t *testing.T) {
	fmt.Println("start TestPostAppPageGooglePlay")
	var method = "POST"
	var endpoint = "/hitec/repository/app/store/app-page/google-play/"

	/*
	 * test for faillure
	 */
	payload := new(bytes.Buffer)
	err := json.NewEncoder(payload).Encode([]byte(`[{ "wrong_json_format": true }]`))
	if err != nil {
		t.Errorf("Could not convert example tweet to json byte")
	}

	req := buildRequest(method, endpoint, payload, t)
	rr := executeRequest(req)

	//Confirm the response has the right status code
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusBadRequest, status)
	}

	/*
	 * test for success
	 */
	payload = new(bytes.Buffer)
	err = json.NewEncoder(payload).Encode(
		AppPageGooglePlay{
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
		},
	)
	if err != nil {
		t.Errorf("Could not convert example tweet to json byte")
	}
	req = buildRequest(method, endpoint, payload, t)
	rr = executeRequest(req)

	//Confirm the response has the right status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}
}

func TestPostAppReviewGooglePlay(t *testing.T) {
	fmt.Println("start TestPostAppReviewGooglePlay")
	var method = "POST"
	var endpoint = "/hitec/repository/app/store/app-review/google-play/"

	/*
	 * test for faillure
	 */
	payload := new(bytes.Buffer)
	err := json.NewEncoder(payload).Encode([]byte(`[{ "wrong_json_format": true }]`))
	if err != nil {
		t.Errorf("Could not convert example tweet to json byte")
	}

	req := buildRequest(method, endpoint, payload, t)
	rr := executeRequest(req)

	//Confirm the response has the right status code
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusBadRequest, status)
	}

	/*
	 * test for success
	 */
	payload = new(bytes.Buffer)
	err = json.NewEncoder(payload).Encode(
		[]AppReviewGooglePlay{AppReviewGooglePlay{
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
		}},
	)
	if err != nil {
		t.Errorf("Could not convert example tweet to json byte")
	}
	req = buildRequest(method, endpoint, payload, t)
	rr = executeRequest(req)

	//Confirm the response has the right status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}
}

func TestPostObserveAppGooglePlay(t *testing.T) {
	fmt.Println("start TestPostObserveAppGooglePlay")
	var method = "POST"
	var endpoint = "/hitec/repository/app/observe/app/google-play/package-name/%s/interval/%s"

	/*
	 * test for success
	 */
	endpoint = fmt.Sprintf(endpoint, "test", "1h")
	req := buildRequest(method, endpoint, nil, t)
	rr := executeRequest(req)

	//Confirm the response has the right status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}
}

func TestGetObsevableGooglePlay(t *testing.T) {
	fmt.Println("start TestGetObsevableGooglePlay")
	var method = "GET"
	var endpoint = "/hitec/repository/app/observable/google-play"

	/*
	 * test for success
	 */
	req := buildRequest(method, endpoint, nil, t)
	rr := executeRequest(req)

	//Confirm the response has the right status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}

	var observables []ObservableGooglePlay
	err := json.NewDecoder(rr.Body).Decode(&observables)
	if err != nil {
		t.Errorf("Did not receive a proper formed json")
	}
	if len(observables) != 1 {
		t.Errorf("response length differs. Expected %d .\n Got %d instead", 1, len(observables))
	}
}

func TestGetAppReviewsOfClass(t *testing.T) {
	fmt.Println("start TestGetAppReviewsOfClass")
	var method = "GET"
	var endpoint = "/hitec/repository/app/google-play/package-name/%s/class/%s"

	/*
	 * test for success CHECK 1
	 */
	endpointCheckOne := fmt.Sprintf(endpoint, "eu.openreq", "feature_request")
	req := buildRequest(method, endpointCheckOne, nil, t)
	rr := executeRequest(req)

	//Confirm the response has the right status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}

	var appReviews []AppReviewGooglePlay
	err := json.NewDecoder(rr.Body).Decode(&appReviews)
	if err != nil {
		t.Errorf("Did not receive a proper formed json")
	}
	if len(appReviews) != 1 {
		t.Errorf("response length differs. Expected %d .\n Got %d instead", 1, len(appReviews))
	}
}
