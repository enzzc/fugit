package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	RE_TAG          = "#[a-zA-Z]([a-zA-Z0-9-])+"
	RE_TASK_LINE    = "^ +?([0-2][0-9]:[0-5][0-9]-[0-2][0-9]:[0-5][0-9]) +?([^\n]+)"
	RE_DAY_LINE     = "^([0-9]{4}-[0-9]{2}-[0-9]{2}) *?$" // https://regexland.com/regex-dates/
	STATE_START     = 0
	STATE_IN_STANZA = 1
)

type Task struct {
	Start       time.Time
	Duration    time.Duration
	Description string
}

type Config struct {
	From      time.Time
	To        time.Time
	FilterTag string
}

func ParseInput(f io.Reader) <-chan *Task {
	scanner := bufio.NewScanner(f)
	out := make(chan *Task, 100)
	go func() {
		err := ParseStanzaStream(out, scanner)
		close(out)
		if err != nil {
			panic(err)
		}
	}()
	return out
}

func FilterRangeStream(tasks <-chan *Task, start, end time.Time) <-chan *Task {
	out := make(chan *Task, 100)
	go func() {
		for t := range tasks {
			if t.Start.After(end) {
				break
			}
			if t.Start.After(start) {
				out <- t
			}
		}
		close(out)
	}()
	return out
}

func CountSpentTimeTag(tasks <-chan *Task, tag string) (time.Duration, error) {
	spent, _ := time.ParseDuration("0")
	reTag := regexp.MustCompile(RE_TAG)
	if !reTag.MatchString(tag) {
		return spent, errors.New("Not a valid tag")
	}
	for t := range tasks {
		if strings.Contains(t.Description, tag) {
			spent += t.Duration
		}
	}
	return spent, nil
}

func ParseStanzaStream(channel chan<- *Task, lines *bufio.Scanner) error {
	re_day := regexp.MustCompile(RE_DAY_LINE)
	re_task := regexp.MustCompile(RE_TASK_LINE)

	line_no := 0
	state := STATE_START
	current_day := "0000-00-00"

	for lines.Scan() {
		line := lines.Text()
		line_no++

		if line == "" {
			continue
		}

		// Match a task
		m := re_task.FindStringSubmatch(line)
		if len(m) > 0 {
			if state == STATE_START { // If we're not in a stanza already
				return errors.New(fmt.Sprintf("Line %d: Cannot parse tasks outside a stanza", line_no))
			}
			interval := m[1]
			desc := m[2]
			task := parseTaskLine(current_day, interval, desc)
			channel <- task
			continue
		}

		// Match a date (new stanza)
		m = re_day.FindStringSubmatch(line)
		if len(m) > 0 {
			state = STATE_IN_STANZA
			day_string := m[1]
			_, err := time.Parse(time.RFC3339, day_string+"T00:00:00Z")
			if err != nil {
				return errors.New(fmt.Sprintf("Line %d: Wrong date format", line_no))
			}
			current_day = day_string
			continue
		}

		// Nothing matched: syntax error
		return errors.New(fmt.Sprintf("Line %d: Syntax error", line_no))

	}
	return nil
}

func ParseStanzas(lines io.Reader) ([]Task, error) {
	tasks := []Task{}
	inStream := ParseInput(lines)
	for t := range inStream {
		tasks = append(tasks, *t)
	}
	return tasks, nil
}

func parseTaskLine(day string, interval string, desc string) *Task {
	splited := strings.Split(interval, "-")
	start, _ := time.Parse(time.RFC3339, day+"T"+splited[0]+":00Z")
	end, _ := time.Parse(time.RFC3339, day+"T"+splited[1]+":00Z")
	dur := end.Sub(start)
	return &Task{start, dur, desc}
}

func parseCommandLine(progname string, args []string) *Config {
	return parseCommandLineWithTime(time.Now().UTC(), progname, args)
}

func parseCommandLineWithTime(refTime time.Time, progname string, args []string) *Config {
	flagSet := flag.NewFlagSet(progname, flag.ContinueOnError)

	tFlag := flagSet.String("t", "", "Filter on tag")
	dFlag := flagSet.Bool("d", false, "This day")
	mFlag := flagSet.Bool("m", false, "This month")
	wFlag := flagSet.Bool("w", false, "This week (starts on Sun)")
	wwFlag := flagSet.Bool("W", false, "This week (starts on Mon)")
	aFlag := flagSet.Bool("a", false, "All time (until now)")
	flagSet.Parse(args)

	// All time (until now)
	if *aFlag {
		start := time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)
		end := time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC)
		return &Config{start, end, *tFlag}
	}

	// From the begining of this day to the end of this day
	if *dFlag {
		start := time.Date(refTime.Year(), refTime.Month(), refTime.Day(), 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 0, 1).Add(-time.Second)
		return &Config{start, end, *tFlag}
	}

	// From the begining of this month to the end of this month
	if *mFlag {
		start := time.Date(refTime.Year(), refTime.Month(), 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 1, 0).Add(-time.Second)
		return &Config{start, end, *tFlag}
	}

	// This week (starts on Sun)
	if *wFlag {
		start := time.Date(refTime.Year(), refTime.Month(), refTime.Day(), 0, 0, 0, 0, time.UTC)
		start = start.AddDate(0, 0, -int(refTime.Weekday()))
		end := time.Date(refTime.Year(), refTime.Month(), refTime.Day(), 0, 0, 0, 0, time.UTC)
		end = end.AddDate(0, 0, 7-int(refTime.Weekday())).Add(-time.Second)
		return &Config{start, end, *tFlag}
	}

	// This week (starts on Mon)
	if *wwFlag {
		start := time.Date(refTime.Year(), refTime.Month(), refTime.Day(), 0, 0, 0, 0, time.UTC)
		start = start.AddDate(0, 0, -int(refTime.Weekday()-1))
		end := time.Date(refTime.Year(), refTime.Month(), refTime.Day(), 0, 0, 0, 0, time.UTC)
		end = end.AddDate(0, 0, 8-int(refTime.Weekday())).Add(-time.Second)
		return &Config{start, end, *tFlag}
	}

	return &Config{refTime, refTime, *tFlag}
}

func main() {
	config := parseCommandLine(os.Args[0], os.Args[1:])

	filename := os.Getenv("FUGIT_FILE")
	f, err := os.Open(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	defer f.Close()

	inStream := ParseInput(f)
	tasks := FilterRangeStream(inStream, config.From, config.To)

	firstDate := time.Now()
	lastDate := time.Now()
	count := 0
	totalSpent, _ := time.ParseDuration("0m")
	for task := range tasks {
		if count == 0 {
			firstDate = task.Start
		}
		lastDate = task.Start.Add(task.Duration)
		totalSpent += task.Duration
		count++
	}

	fmt.Printf("OK: %s\n\nRead %d tasks\n", filename, count)
	if count > 0 {
		fmt.Printf("From %s\nTo   %s\n\n", firstDate, lastDate)
		fmt.Printf("Time spent: %s\n", totalSpent)
	}
}
