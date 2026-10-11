package profile

import "time"

type Profile struct {
	Posts     int
	ContentMD string
	UpdatedAt time.Time
}
