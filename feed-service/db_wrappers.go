package main

import "bsky-schwartz/db"

func GetUserFromDB(handle string) (string, map[string]float64, error) {
	return db.GetUserFromDB(handle)
}

func SaveUser(handle, did string) error {
	return db.SaveUser(handle, did)
}

func SaveWeights(userHandle string, weights map[string]float64) error {
	return db.SaveWeights(userHandle, weights)
}

func SaveWeight(userHandle, valueID string, weight float64) error {
	return db.SaveWeight(userHandle, valueID, weight)
}

func GetPostAnalysis(postAtURI string) ([]interface{}, error) {
	analyses, err := db.GetPostAnalysis(postAtURI)
	if err != nil {
		return nil, err
	}
	result := make([]interface{}, len(analyses))
	for i, a := range analyses {
		result[i] = a
	}
	return result, nil
}
