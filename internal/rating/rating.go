// rating provides the function that calculates leaderboard rating changes.
package rating

import (
	"fmt"
	"math"
	"time"
)

// CalculateRatingChange calculates a rating based on the deviation of the timer and the planned time. 
// Returns the algorithm version and the rating change.
func CalculateRatingChange(start, stop, startTimer, stopTimer time.Time, anchor time.Duration) (string, float64) {
	startDeviation := float64(startTimer.Unix() - start.Unix()) // did the user start correctly
	stopDeviation := float64(stopTimer.Unix() - stop.Unix()) // did the user stop correctly
	durationDeviation := stopDeviation - startDeviation // did the user deviate from the planned event duration

	ratingChange := 0.0
	ratingChange += anchor.Seconds() - math.Abs(startDeviation)
	ratingChange += anchor.Seconds() - math.Abs(stopDeviation)
	ratingChange += 2 * (anchor.Seconds() - math.Abs(durationDeviation))

	return fmt.Sprintf("v0.0.1-%s", anchor.String()), ratingChange
}
