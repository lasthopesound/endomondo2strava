package synchronizer

import (
	"fmt"

	"github.com/michalq/endo2strava/pkg/strava-client"
)

// StravaUploader uploads workouts into strava
type StravaUploader struct {
	stravaClient       *strava.Client
	workoutsRepository Workouts
}

// NewStravaUploader creates instance of StravaUploader
func NewStravaUploader(stravaClient *strava.Client, workoutsRepository Workouts) *StravaUploader {
	return &StravaUploader{stravaClient, workoutsRepository}
}

// UploadAll uploads all provided workouts to strava
func (s *StravaUploader) UploadAll() ([]Workout, error) {
	workouts, err := s.workoutsRepository.FindAll()
	if err != nil {
		return nil, err
	}
	var toImport []Workout
	for _, workout := range workouts {
		if workout.UploadStarted == 0 {
			toImport = append(toImport, workout)
		}
	}

	uploaded, err := s.uploadMany(workouts)
	if len(uploaded) == 0 && err != nil {
		return nil, err
	}
	for _, workout := range uploaded {
		if err := s.workoutsRepository.Update(&workout); err != nil {
			fmt.Println("Err", err)
		}
	}
	return uploaded, err
}

func (s *StravaUploader) uploadMany(workouts []Workout) ([]Workout, error) {
	var uploaded []Workout
	for _, workout := range workouts {
		uploadResponse, err := s.stravaClient.ImportWorkout(strava.UploadParameters{
			ExternalID:  workout.EndomondoID,
			Name:        fmt.Sprintf("Endomondo %s", workout.EndomondoID),
			Description: fmt.Sprintf("Workout imported from endomondo"),
			File:        workout.Path,
			Commute:     "0",
			DataType:    workout.Ext,
			Trainer:     "0",
		})
		if err != nil {
			return nil, err
		}
		workout.StravaID = uploadResponse.ID
		workout.UploadStarted = 1
		uploaded = append(uploaded, workout)
		return uploaded, nil
	}
	return uploaded, nil
}
