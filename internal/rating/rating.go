// rating provides the function that calculates leaderboard rating changes.
package rating

import (
	"fmt"
	"math"
	"time"
)

// CalculateRatingChange calculates a rating based on the deviation of the timer and the planned time.
// The anchor defines the point where a deviation switches from positive to negative rating (anchor = 120s => 140s deviation == -20s).
// Returns the algorithm version and the rating change.
func CalculateRatingChange(start, stop, startTimer, stopTimer time.Time, streak int64, anchor time.Duration) (string, float64) {
	startDeviation := float64(startTimer.Unix() - start.Unix()) // did the user start correctly
	stopDeviation := float64(stopTimer.Unix() - stop.Unix())    // did the user stop correctly
	durationDeviation := stopDeviation - startDeviation         // did the user deviate from the planned event duration

	ratingChange := 0.0
	ratingChange += anchor.Seconds() - math.Abs(startDeviation)
	ratingChange += anchor.Seconds() - math.Abs(stopDeviation)
	ratingChange += 2 * (anchor.Seconds() - math.Abs(durationDeviation))

	// cap change at 3x anchor to avoid unrecoverable rating loss
	// if someone e.g. forgets to stop the event before sleep.
	if ratingChange > anchor.Seconds()*3 {
		ratingChange = anchor.Seconds() * 3
	}

	// streak is pushing the rating very strongly but this is intended to set a focus on streaks (discipline)
	if ratingChange > 0 {
		ratingChange *= float64(streak / 10)
	}

	return fmt.Sprintf("v0.0.1-%s", anchor.String()), ratingChange
}
