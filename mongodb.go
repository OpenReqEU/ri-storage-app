package main

import (
	"fmt"
	"log"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	database                       = "app_data"
	collectionAppReviewsGooglePlay = "app_reviews_google_play"
	collectionAppPageGooglePlay    = "app_page_google_play"
	collectionObservableGooglePlay = "observable_google_play"
)

// MongoGetSession returns a session
func MongoGetSession(mongoIP, username, password string) *mgo.Session {
	info := &mgo.DialInfo{
		Addrs:    []string{mongoIP},
		Timeout:  60 * time.Second,
		Database: database,
		Username: username,
		Password: password,
	}

	session, err := mgo.DialWithInfo(info)
	if err != nil {
		log.Fatal(err)
	}

	session.SetMode(mgo.Monotonic, true)

	return session
}

// MongoCreateCollectionIndexes creates the indexes
func MongoCreateCollectionIndexes(mongoClient *mgo.Session) {
	// Index
	appReviewGooglePlayUniqueIndex := mgo.Index{
		Key:        []string{"review_id"},
		Unique:     true,
		Background: true,
		Sparse:     true,
	}

	// Index
	appReviewGooglePlayIndex := mgo.Index{
		Key:        []string{"date_posted"},
		Unique:     false,
		Background: true,
		Sparse:     true,
	}
	appReviewGooglePlayCollection := mongoClient.DB(database).C(collectionAppReviewsGooglePlay)
	err := appReviewGooglePlayCollection.EnsureIndex(appReviewGooglePlayUniqueIndex)
	if err != nil {
		panic(err)
	}
	err = appReviewGooglePlayCollection.EnsureIndex(appReviewGooglePlayIndex)
	if err != nil {
		panic(err)
	}

	// Index
	appPageGooglePlayIndex := mgo.Index{
		Key:        []string{"package_name", "last_update"},
		Unique:     true,
		Background: true,
		Sparse:     true,
	}
	appPageGooglePlayCollection := mongoClient.DB(database).C(collectionAppPageGooglePlay)
	err = appPageGooglePlayCollection.EnsureIndex(appPageGooglePlayIndex)
	if err != nil {
		panic(err)
	}

	// Index
	observableGooglePlayIndex := mgo.Index{
		Key:        []string{"package_name"},
		Unique:     true,
		Background: true,
		Sparse:     true,
	}
	observableGoogleCollection := mongoClient.DB(database).C(collectionObservableGooglePlay)
	err = observableGoogleCollection.EnsureIndex(observableGooglePlayIndex)
	if err != nil {
		panic(err)
	}
}

// MongoInsertAppPageGooglePlay returns ok if the app page was inserted or already existed
func MongoInsertAppPageGooglePlay(mongoClient *mgo.Session, appPage AppPageGooglePlay) bool {
	err := mongoClient.DB(database).C(collectionAppPageGooglePlay).Insert(appPage)
	if err != nil && !mgo.IsDup(err) {
		fmt.Println(err)
		return false
	}

	return true
}

// MongoInsertAppReviewGooglePlay returns ok if the review was inserted or updated
func MongoInsertAppReviewGooglePlay(mongoClient *mgo.Session, review AppReviewGooglePlay) bool {
	col := mongoClient.DB(database).C(collectionAppReviewsGooglePlay)
	change := mgo.Change{
		Update:    review,
		Upsert:    true,
		ReturnNew: true,
	}
	var newReview AppReviewGooglePlay
	_, err := col.Find(bson.M{"review_id": review.ReviewID}).Apply(change, &newReview)
	if err != nil {
		fmt.Println(err)
		return false
	}

	return true
}

// MongoGetNonExistingAppReviewGooglePlay returns a list of app reviews that do not yet exist in the db
func MongoGetNonExistingAppReviewGooglePlay(mongoClient *mgo.Session, reviews []AppReviewGooglePlay) []AppReviewGooglePlay {
	var uniqueAppReviews []AppReviewGooglePlay
	col := mongoClient.DB(database).C(collectionAppReviewsGooglePlay)

	for _, review := range reviews {
		var entry AppReviewGooglePlay
		col.Find(bson.M{"review_id": review.ReviewID}).One(&entry)

		if entry.PackageName == "" {
			uniqueAppReviews = append(uniqueAppReviews, review)
		}
	}

	return uniqueAppReviews
}

// MongoInsertObservableGooglePlay returns ok if the package name was inserted or already existed
func MongoInsertObservableGooglePlay(mongoClient *mgo.Session, observable ObservableGooglePlay) bool {
	err := mongoClient.DB(database).C(collectionObservableGooglePlay).Insert(observable)
	if err != nil && !mgo.IsDup(err) {
		fmt.Println(err)
		return false
	}

	return true
}

// MongoGetAllObservableGooglePlay returns all observable apps
func MongoGetAllObservableGooglePlay(mongoClient *mgo.Session) []ObservableGooglePlay {
	var observables []ObservableGooglePlay
	err := mongoClient.
		DB(database).
		C(collectionObservableGooglePlay).
		Find(nil).
		All(&observables)
	if err != nil {
		fmt.Println("ERR", err)
		panic(err)
	}

	return observables
}

// MongoGeGooglePlayReviewOfClass returns all reviews belonging to the given package name and class
func MongoGetGooglePlayReviewOfClass(mongoClient *mgo.Session, packageName string, reviewClass string) []AppReviewGooglePlay {
	var reviewCLassField string
	if reviewClass == "bug_report" {
		reviewCLassField = "cluster_is_bug_report"
	} else if reviewClass == "feature_request" {
		reviewCLassField = "cluster_is_feature_request"
	}
	var reviews []AppReviewGooglePlay
	err := mongoClient.
		DB(database).
		C(collectionAppReviewsGooglePlay).
		Find(bson.M{"package_name": packageName, reviewCLassField: true}).
		All(&reviews)
	if err != nil {
		fmt.Println("ERR", err)
		panic(err)
	}

	return reviews
}
