package profile

import "time"

type Profile struct {
	Posts     int
	Projects  int
	ContentMD string
	UpdatedAt time.Time
}
