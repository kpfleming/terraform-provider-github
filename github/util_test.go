package github

import (
	"testing"
	"unicode"
)

func TestAccValidateTeamIDFunc(t *testing.T) {
	// warnings, errors := validateTeamIDFunc(interface{"1234567"})

	cases := []struct {
		TeamID   interface{}
		ErrCount int
	}{
		{

			TeamID:   "1234567",
			ErrCount: 0,
		},
		{
			// an int cannot be cast to a string
			TeamID:   1234567,
			ErrCount: 1,
		},
		{
			TeamID:   "notAnInt",
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validateTeamIDFunc(tc.TeamID, "keyName")
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected %d validation error but got %d", tc.ErrCount, len(errors))
		}
	}
}

func TestAccGithubUtilRole_validation(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "invalid",
			ErrCount: 1,
		},
		{
			Value:    "valid_one",
			ErrCount: 0,
		},
		{
			Value:    "valid_two",
			ErrCount: 0,
		},
	}

	validationFunc := validateValueFunc([]string{"valid_one", "valid_two"})

	for _, tc := range cases {
		_, errors := validationFunc(tc.Value, "test_arg")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected 1 validation error")
		}
	}
}

func TestAccGithubUtilTwoPartID(t *testing.T) {
	partOne, partTwo := "foo", "bar"

	id := buildTwoPartID(partOne, partTwo)

	if id != "foo:bar" {
		t.Fatalf("Expected two part id to be foo:bar, actual: %s", id)
	}

	parsedPartOne, parsedPartTwo, err := parseTwoPartID(id)
	if err != nil {
		t.Fatal(err)
	}

	if parsedPartOne != "foo" {
		t.Fatalf("Expected parsed part one foo, actual: %s", parsedPartOne)
	}

	if parsedPartTwo != "bar" {
		t.Fatalf("Expected parsed part two bar, actual: %s", parsedPartTwo)
	}
}

func flipUsernameCase(username string) string {
	oc := []rune(username)

	for i, ch := range oc {
		if unicode.IsLetter(ch) {

			if unicode.IsUpper(ch) {
				oc[i] = unicode.ToLower(ch)
			} else {
				oc[i] = unicode.ToUpper(ch)
			}
			break
		}

	}
	return string(oc)
}
