package domain

//go:generate mockery --name=Clock --output=../mocks --outpkg=mocks --filename=clock_mock.go --structname=ClockMock --with-expecter

import "time"

// Clock abstracts retrieval of the current time so it can be faked in tests.
type Clock interface {
	Now() time.Time
}
