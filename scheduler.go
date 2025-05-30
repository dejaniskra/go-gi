package gogi

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

type job struct {
	name     string
	interval time.Duration
	cron     *cronSchedule
	fn       func()
}

var (
	jobsMu sync.Mutex
	jobs   []job
)

func validateDuration(value interface{}) (time.Duration, bool) {
	duration, ok := value.(time.Duration)
	return duration, ok
}

func AddJobInterval(name string, interval time.Duration, fn func()) error {
	_, ok := validateDuration(interval)
	if !ok {
		return errors.New("invalid duration type, must be time.Duration")
	}
	jobsMu.Lock()
	defer jobsMu.Unlock()
	jobs = append(jobs, job{name: name, interval: interval, fn: fn})
	return nil
}

func AddJobCron(name, cronExpr string, fn func()) error {
	sched, err := parseCron(cronExpr)
	if err != nil {
		return err
	}
	jobsMu.Lock()
	defer jobsMu.Unlock()
	jobs = append(jobs, job{name: name, cron: sched, fn: fn})
	return nil
}

func startScheduler() {
	if len(jobs) == 0 {
		log.Println("No jobs registered, skipping scheduler start")
		return
	}
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for now := range ticker.C {
			jobsMu.Lock()
			for _, j := range jobs {
				if j.interval > 0 {
					if now.Unix()%int64(j.interval.Seconds()) == 0 {
						go safeRun(j.name, j.fn)
					}
				} else if j.cron != nil && j.cron.matches(now) {
					go safeRun(j.name, j.fn)
				}
			}
			jobsMu.Unlock()
		}
	}()
}

func safeRun(name string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("⛔ Job '%s' panicked: %v", name, r)
		}
	}()
	fn()
}

type cronSchedule struct {
	minutes     map[int]bool
	hours       map[int]bool
	daysOfMonth map[int]bool
	months      map[int]bool
	daysOfWeek  map[int]bool
}

func parseCron(expr string) (*cronSchedule, error) {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return nil, ErrInvalidCron
	}

	mins, err := parseCronField(fields[0], 0, 59)
	if err != nil {
		return nil, err
	}
	hrs, err := parseCronField(fields[1], 0, 23)
	if err != nil {
		return nil, err
	}
	doms, err := parseCronField(fields[2], 1, 31)
	if err != nil {
		return nil, err
	}
	mons, err := parseCronField(fields[3], 1, 12)
	if err != nil {
		return nil, err
	}
	dows, err := parseCronField(fields[4], 0, 6) // 0 = Sunday

	if err != nil {
		return nil, err
	}

	return &cronSchedule{
		minutes:     mins,
		hours:       hrs,
		daysOfMonth: doms,
		months:      mons,
		daysOfWeek:  dows,
	}, nil
}

var ErrInvalidCron = errors.New("invalid cron expression")

func parseCronField(field string, min, max int) (map[int]bool, error) {
	result := make(map[int]bool)
	if field == "*" {
		for i := min; i <= max; i++ {
			result[i] = true
		}
		return result, nil
	}

	parts := strings.Split(field, ",")
	for _, part := range parts {
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, ErrInvalidCron
			}
			start, err1 := strconv.Atoi(rangeParts[0])
			end, err2 := strconv.Atoi(rangeParts[1])
			if err1 != nil || err2 != nil || start < min || end > max || start > end {
				return nil, ErrInvalidCron
			}
			for i := start; i <= end; i++ {
				result[i] = true
			}
		} else {
			val, err := strconv.Atoi(part)
			if err != nil || val < min || val > max {
				return nil, ErrInvalidCron
			}
			result[val] = true
		}
	}
	return result, nil
}

func (c *cronSchedule) matches(t time.Time) bool {
	return c.minutes[t.Minute()] &&
		c.hours[t.Hour()] &&
		c.daysOfMonth[t.Day()] &&
		c.months[int(t.Month())] &&
		c.daysOfWeek[int(t.Weekday())]
}
