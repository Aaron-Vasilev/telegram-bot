package utils

import (
	t "bot/src/utils/types"
	"encoding/json"
	"testing"
	"time"
)

func TestUpdateMembership(tt *testing.T) {
	token2 := t.Token{
		Type:    2,
		Created: time.Date(2023, time.July, 15, 0, 0, 0, 0, time.UTC),
		Valid:   true,
	}

	tests := []struct {
		name       string
		membership t.Membership
		token      t.Token
		expected   t.Membership
	}{
		{
			name: "Type 1 ends expired",
			membership: t.Membership{
				Type:             0,
				Starts:           time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
				Ends:             time.Date(2023, time.February, 1, 0, 0, 0, 0, time.UTC),
				LessonsAvailable: 0,
			},
			token: t.Token{
				Type:    1,
				Created: time.Date(2023, time.July, 13, 0, 0, 0, 0, time.UTC),
				Valid:   true,
			},
			expected: t.Membership{
				Type:             1,
				Starts:           time.Date(2023, time.July, 13, 0, 0, 0, 0, time.UTC),
				Ends:             time.Date(2023, time.August, 9, 0, 0, 0, 0, time.UTC),
				LessonsAvailable: 4,
			},
		},
		{
			name: "Type 1 membership is valid",
			membership: t.Membership{
				Type:             0,
				Starts:           time.Date(2023, time.May, 12, 0, 0, 0, 0, time.UTC),
				Ends:             time.Date(2023, time.July, 22, 0, 0, 0, 0, time.UTC),
				LessonsAvailable: 2,
			},
			token: t.Token{
				Type:    1,
				Created: time.Date(2023, time.July, 13, 0, 0, 0, 0, time.UTC),
				Valid:   true,
			},
			expected: t.Membership{
				Type:             1,
				Starts:           time.Date(2023, time.May, 12, 0, 0, 0, 0, time.UTC),
				Ends:             time.Date(2023, time.August, 19, 0, 0, 0, 0, time.UTC),
				LessonsAvailable: 6,
			},
		},
		{
			name: "Type 1 membership expired, lessons remaining",
			membership: t.Membership{
				Type:             1,
				Starts:           time.Date(2023, time.July, 0, 0, 0, 0, 0, time.UTC),
				Ends:             time.Date(2023, time.July, 29, 0, 0, 0, 0, time.UTC),
				LessonsAvailable: 2,
			},
			token: t.Token{
				Type:    1,
				Created: time.Date(2023, time.August, 1, 0, 0, 0, 0, time.UTC),
				Valid:   true,
			},
			expected: t.Membership{
				Type:             1,
				Starts:           time.Date(2023, time.August, 1, 0, 0, 0, 0, time.UTC),
				Ends:             time.Date(2023, time.August, 28, 0, 0, 0, 0, time.UTC),
				LessonsAvailable: 4,
			},
		},
		{
			name: "Type 1, membership expired, lessons are negative",
			membership: t.Membership{
				Type:             1,
				Starts:           time.Date(2023, time.July, 1, 0, 0, 0, 0, time.UTC),
				Ends:             time.Date(2023, time.July, 29, 0, 0, 0, 0, time.UTC),
				LessonsAvailable: -1,
			},
			token: t.Token{
				Type:    1,
				Created: time.Date(2023, time.August, 1, 0, 0, 0, 0, time.UTC),
				Valid:   true,
			},
			expected: t.Membership{
				Type:             1,
				Starts:           time.Date(2023, time.August, 1, 0, 0, 0, 0, time.UTC),
				Ends:             time.Date(2023, time.August, 28, 0, 0, 0, 0, time.UTC),
				LessonsAvailable: 4,
			},
		},
		{
			name: "Type 1, membership is active, lessons are negative",
			membership: t.Membership{
				Type:             1,
				Starts:           time.Date(2023, time.July, 0, 0, 0, 0, 0, time.UTC),
				Ends:             time.Date(2023, time.September, 7, 0, 0, 0, 0, time.UTC),
				LessonsAvailable: -1,
			},
			token: t.Token{
				Type:    1,
				Created: time.Date(2023, time.September, 5, 0, 0, 0, 0, time.UTC),
				Valid:   true,
			},
			expected: t.Membership{
				Type:             1,
				Starts:           time.Date(2023, time.July, 0, 0, 0, 0, 0, time.UTC),
				Ends:             time.Date(2023, time.October, 5, 0, 0, 0, 0, time.UTC),
				LessonsAvailable: 3,
			},
		},
		{
			name: "Type 2 ends expired",
			membership: t.Membership{
				Type:             0,
				Starts:           time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
				Ends:             time.Date(2023, time.February, 1, 0, 0, 0, 0, time.UTC),
				LessonsAvailable: -1,
			},
			token: token2,
			expected: t.Membership{
				Type:             token2.Type,
				Starts:           token2.Created,
				Ends:             time.Date(2023, time.August, 11, 0, 0, 0, 0, time.UTC),
				LessonsAvailable: 8,
			},
		},
		{
			name: "Type 2 ends is valid",
			membership: t.Membership{
				Type:             2,
				Starts:           time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
				Ends:             time.Date(2023, time.July, 30, 0, 0, 0, 0, time.UTC),
				LessonsAvailable: 5,
			},
			token: token2,
			expected: t.Membership{
				Type:             2,
				Starts:           time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
				Ends:             time.Date(2023, time.August, 27, 0, 0, 0, 0, time.UTC),
				LessonsAvailable: 13,
			},
		},
		{
			name: "Type 8 ends expired",
			membership: t.Membership{
				Type:             2,
				Starts:           time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
				Ends:             time.Date(2023, time.February, 1, 0, 0, 0, 0, time.UTC),
				LessonsAvailable: 0,
			},
			token: t.Token{
				Type:    8,
				Created: time.Date(2023, time.July, 1, 0, 0, 0, 0, time.UTC),
				Valid:   true,
			},
			expected: t.Membership{
				Type:             8,
				Starts:           time.Date(2023, time.July, 1, 0, 0, 0, 0, time.UTC),
				Ends:             time.Date(2023, time.July, 28, 0, 0, 0, 0, time.UTC),
				LessonsAvailable: 0,
			},
		},
		{
			name: "Type 8 ends valid",
			membership: t.Membership{
				Type:             8,
				Starts:           time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
				Ends:             time.Date(2023, time.July, 16, 0, 0, 0, 0, time.UTC),
				LessonsAvailable: 0,
			},
			token: t.Token{
				Type:    8,
				Created: time.Date(2023, time.July, 14, 0, 0, 0, 0, time.UTC),
				Valid:   true,
			},
			expected: t.Membership{
				Type:             8,
				Starts:           time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
				Ends:             time.Date(2023, time.August, 13, 0, 0, 0, 0, time.UTC),
				LessonsAvailable: 0,
			},
		},
	}

	for _, test := range tests {
		tt.Run(test.name, func(tt *testing.T) {
			UpdateMembership(&test.membership, test.token)

			if test.membership != test.expected {
				m, _ := json.MarshalIndent(test.membership, "", "\t")
				e, _ := json.MarshalIndent(test.expected, "", "\t")

				tt.Errorf("Got: %s", m)
				tt.Errorf("Expected %s", e)
			}
		})
	}
}
