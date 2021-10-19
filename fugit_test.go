package main

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestParseStanzas(t *testing.T) {

	stanza := `
2021-09-30
    21:00-22:30 Birth
    22:30-23:00 Growing up

2021-10-01
  10:00-11:00 Work on project 1

  11:30-13:00 Coding


`
	sreader := strings.NewReader(stanza)

	task1_dur, _ := time.ParseDuration("1h30m")
	task1_start, _ := time.Parse(time.RFC3339, "2021-09-30T21:00:00Z")

	task2_dur, _ := time.ParseDuration("30m")
	task2_start, _ := time.Parse(time.RFC3339, "2021-09-30T22:30:00Z")

	task3_dur, _ := time.ParseDuration("1h")
	task3_start, _ := time.Parse(time.RFC3339, "2021-10-01T10:00:00Z")

	task4_dur, _ := time.ParseDuration("1h30m")
	task4_start, _ := time.Parse(time.RFC3339, "2021-10-01T11:30:00Z")

	want := []Task{
		Task{task1_start, task1_dur, "Birth"},
		Task{task2_start, task2_dur, "Growing up"},
		Task{task3_start, task3_dur, "Work on project 1"},
		Task{task4_start, task4_dur, "Coding"},
	}

	got, err := ParseStanzas(sreader)
	if err != nil {
		t.Errorf("err %q", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestCountSpentTime(t *testing.T) {
	stanza := `
2021-09-30
    21:00-22:30 Work on #project1 and #project3
    22:30-23:00 Work on #project2

2021-10-01
    10:00-11:00 Work on #project1
    11:30-13:00 Lunch

2021-10-05
    13:37-14:28 Do stuff
    15:30-16:15 Work on #project3
`

	input := strings.NewReader(stanza)
	inStream := ParseInput(input)
	got, err := CountSpentTimeTag(inStream, "#project1")
	want, _ := time.ParseDuration("2h30m")
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}

	input = strings.NewReader(stanza)
	inStream = ParseInput(input)
	got, err = CountSpentTimeTag(inStream, "#project2")
	want, _ = time.ParseDuration("30m")
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}

	input = strings.NewReader(stanza)
	inStream = ParseInput(input)
	got, err = CountSpentTimeTag(inStream, "#project3")
	want, _ = time.ParseDuration("2h15m")
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}

	input = strings.NewReader(stanza)
	inStream = ParseInput(input)
	got, err = CountSpentTimeTag(inStream, "#project4")
	want, _ = time.ParseDuration("0")
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}

}

func NoTestStream(t *testing.T) {
	stanza := `
2021-09-30
    21:00-22:30 Work on #project1
    22:30-23:00 Work on #project2

2021-10-01
    10:00-11:00 Work on #project1
    11:30-13:00 Lunch

2021-10-05
    13:37-14:28 Do stuff
    15:30-16:00 Work on #project3
`
	input := strings.NewReader(stanza)
	ch := ParseInput(input)

	want := 6
	got := 0

	for range ch {
		got++
	}

	if got != want {
		t.Errorf("got %d tasks, want %d", got, want)
	}

}

func TestStreamSelectRange(t *testing.T) {
	stanza := `
2021-09-28
    21:00-22:30 Work on #project1
    22:30-23:00 Work on #project2

2021-09-30
    21:00-22:30 Work on #project1
    22:30-23:00 Work on #project2

2021-10-01
    10:00-11:00 Work on #project1
    11:30-13:00 Lunch

2021-10-05
    13:37-14:28 Do stuff
    15:30-16:00 Work on #project3
    18:00-23:00 Reading a good book

2021-10-06
    13:37-14:28 Do stuff
    15:30-16:00 Work on #project3
`
	input := strings.NewReader(stanza)
	start, _ := time.Parse(time.RFC3339, "2021-09-30T00:00:00Z")
	end, _ := time.Parse(time.RFC3339, "2021-10-05T23:59:59Z")
	inStream := ParseInput(input)
	tasks := FilterRangeStream(inStream, start, end)

	want := 7
	got := 0

	for range tasks {
		got++
	}

	if got != want {
		t.Errorf("got %d tasks, want %d", got, want)
	}

}

func TestParseCommandLineDay(t *testing.T) {
	args := []string{"progname", "-t", "#ok", "-d"}

	thisNow, _ := time.Parse(time.RFC3339, "2021-10-19T18:07:23Z")
	start, _ := time.Parse(time.RFC3339, "2021-10-19T00:00:00Z")
	end, _ := time.Parse(time.RFC3339, "2021-10-19T23:59:59Z")

	want := Config{start, end, "#ok"}
	got := *parseCommandLineWithTime(thisNow, args[0], args[1:])

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseCommandLineMonth(t *testing.T) {
	args := []string{"progname", "-t", "#ok", "-m"}

	thisNow, _ := time.Parse(time.RFC3339, "2021-10-19T13:37:00Z")
	start, _ := time.Parse(time.RFC3339, "2021-10-01T00:00:00Z")
	end, _ := time.Parse(time.RFC3339, "2021-10-31T23:59:59Z")

	want := Config{start, end, "#ok"}
	got := *parseCommandLineWithTime(thisNow, args[0], args[1:])
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseCommandLineWeek(t *testing.T) {
	args := []string{"progname", "-t", "#ok", "-w"}

	thisNow, _ := time.Parse(time.RFC3339, "2021-09-02T13:37:00Z") // Tue
	start, _ := time.Parse(time.RFC3339, "2021-08-29T00:00:00Z")   // Sun
	end, _ := time.Parse(time.RFC3339, "2021-09-04T23:59:59Z")     // Sat

	want := Config{start, end, "#ok"}
	got := *parseCommandLineWithTime(thisNow, args[0], args[1:])
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
