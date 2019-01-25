package main

import (
	"log"
	"net/http"

	"encoding/json"
	"fmt"
	"os"

	"github.com/gorilla/mux"
	mgo "gopkg.in/mgo.v2"
)

var mongoClient *mgo.Session

func main() {
	log.SetOutput(os.Stdout)
	mongoClient = MongoGetSession(os.Getenv("MONGO_IP"), os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"))
	MongoCreateCollectionIndexes(mongoClient)

	router := mux.NewRouter()
	// Insert
	router.HandleFunc("/hitec/repository/app/store/app-page/google-play/", postAppPageGooglePlay).Methods("POST")
	router.HandleFunc("/hitec/repository/app/store/app-review/google-play/", postAppReviewGooglePlay).Methods("POST")
	router.HandleFunc("/hitec/repository/app/observe/app/google-play/package-name/{package_name}/interval/{interval}", postObserveAppGooglePlay).Methods("POST")
	router.HandleFunc("/hitec/repository/app/non-existing/app-review/google-play/", postNonExistingAppReviewsGooglePlay).Methods("POST")

	// Get
	router.HandleFunc("/hitec/repository/app/observable/google-play", getObsevableGooglePlay).Methods("GET")
	router.HandleFunc("/hitec/repository/app/google-play/package-name/{package_name}/class/{class}", getAppReviewsOfClass).Methods("GET")

	fmt.Println("server now starts")

	log.Fatal(http.ListenAndServe(":9681", router))
}

func postAppPageGooglePlay(w http.ResponseWriter, r *http.Request) {
	// get data from the request
	var appPage AppPageGooglePlay
	err := json.NewDecoder(r.Body).Decode(&appPage)
	if err != nil {
		fmt.Printf("ERROR: %s for request body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// insert data into the db
	m := mongoClient.Copy()
	defer m.Close()
	ok := MongoInsertAppPageGooglePlay(m, appPage)

	// send response
	if ok {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func postAppReviewGooglePlay(w http.ResponseWriter, r *http.Request) {
	// get data from the request
	var appReviews []AppReviewGooglePlay
	err := json.NewDecoder(r.Body).Decode(&appReviews)
	if err != nil {
		fmt.Printf("ERROR: %s for request body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// insert data into the db
	m := mongoClient.Copy()
	defer m.Close()
	for _, review := range appReviews {
		MongoInsertAppReviewGooglePlay(m, review)
	}

	// send response
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
}

func postObserveAppGooglePlay(w http.ResponseWriter, r *http.Request) {
	// get data from the request
	params := mux.Vars(r)
	packageName := params["package_name"]
	interval := params["interval"] // possible intervals: minutely, hourly, daily, monthly

	var observalbe = ObservableGooglePlay{PackageName: packageName, Interval: interval}

	// insert data into the db
	m := mongoClient.Copy()
	defer m.Close()
	ok := MongoInsertObservableGooglePlay(m, observalbe)

	// send response
	if ok {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func postNonExistingAppReviewsGooglePlay(w http.ResponseWriter, r *http.Request) {
	// get data from the request
	var appReviews []AppReviewGooglePlay
	err := json.NewDecoder(r.Body).Decode(&appReviews)
	if err != nil {
		fmt.Printf("ERROR: %s for request body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// insert data into the db
	m := mongoClient.Copy()
	defer m.Close()
	nonExistingAppReviews := MongoGetNonExistingAppReviewGooglePlay(m, appReviews)

	// send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(nonExistingAppReviews)
}

func getObsevableGooglePlay(w http.ResponseWriter, r *http.Request) {
	// get data from the db
	m := mongoClient.Copy()
	defer m.Close()
	observables := MongoGetAllObservableGooglePlay(m)

	// send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(observables)
}

func getAppReviewsOfClass(w http.ResponseWriter, r *http.Request) {
	// get request param
	params := mux.Vars(r)
	packageName := params["package_name"]
	reviewClass := params["class"]

	// query db
	m := mongoClient.Copy()
	bugReports := MongoGetGooglePlayReviewOfClass(m, packageName, reviewClass)

	// if no recent data exist, return false
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bugReports)
}
