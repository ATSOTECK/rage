package stdlib

import (
	"time"

	"github.com/ATSOTECK/oink/internal/runtime"
)

// InitTimeModule registers the time module
func InitTimeModule() {
	// Get timezone info
	_, offset := time.Now().Zone()
	tzName, _ := time.Now().Zone()

	runtime.NewModuleBuilder("time").
		Doc("Time access and conversions.").
		// Constants
		Const("timezone", runtime.NewInt(int64(-offset))).
		Const("altzone", runtime.NewInt(int64(-offset))).
		Const("daylight", runtime.NewInt(0)).
		Const("tzname", runtime.NewTuple([]runtime.Value{
			runtime.NewString(tzName),
			runtime.NewString(tzName),
		})).
		// Core functions
		Func("time", timeTime).
		Func("time_ns", timeTimeNs).
		Func("sleep", timeSleep).
		Func("localtime", timeLocaltime).
		Func("gmtime", timeGmtime).
		Func("mktime", timeMktime).
		Func("strftime", timeStrftime).
		Func("strptime", timeStrptime).
		Func("ctime", timeCtime).
		Func("asctime", timeAsctime).
		// Performance counters
		Func("perf_counter", timePerfCounter).
		Func("perf_counter_ns", timePerfCounterNs).
		Func("monotonic", timeMonotonic).
		Func("monotonic_ns", timeMonotonicNs).
		Func("process_time", timeProcessTime).
		Func("process_time_ns", timeProcessTimeNs).
		Register()
}

var startTime = time.Now()

// time.time() -> float seconds since epoch
func timeTime(vm *runtime.VM) int {
	now := time.Now()
	secs := float64(now.Unix()) + float64(now.Nanosecond())/1e9
	vm.Push(runtime.NewFloat(secs))
	return 1
}

// time.time_ns() -> int nanoseconds since epoch
func timeTimeNs(vm *runtime.VM) int {
	vm.Push(runtime.NewInt(time.Now().UnixNano()))
	return 1
}

// time.sleep(seconds)
func timeSleep(vm *runtime.VM) int {
	secs := vm.CheckFloat(1)
	duration := time.Duration(secs * float64(time.Second))
	time.Sleep(duration)
	return 0
}

// PyStructTime represents a time struct (like Python's time.struct_time)
type PyStructTime struct {
	TmYear  int // year
	TmMon   int // month (1-12)
	TmMday  int // day of month (1-31)
	TmHour  int // hour (0-23)
	TmMin   int // minute (0-59)
	TmSec   int // second (0-61)
	TmWday  int // weekday (0-6, Monday=0)
	TmYday  int // day of year (1-366)
	TmIsdst int // DST flag (-1, 0, 1)
}

// toStructTime converts time.Time to a tuple representing struct_time
func toStructTime(t time.Time) *runtime.PyTuple {
	wday := int(t.Weekday())
	if wday == 0 {
		wday = 6 // Sunday
	} else {
		wday-- // Monday = 0
	}

	yday := t.YearDay()

	// DST check
	_, offset := t.Zone()
	_, stdOffset := time.Date(t.Year(), time.January, 1, 0, 0, 0, 0, t.Location()).Zone()
	isDst := 0
	if offset != stdOffset {
		isDst = 1
	}

	return runtime.NewTuple([]runtime.Value{
		runtime.NewInt(int64(t.Year())),
		runtime.NewInt(int64(t.Month())),
		runtime.NewInt(int64(t.Day())),
		runtime.NewInt(int64(t.Hour())),
		runtime.NewInt(int64(t.Minute())),
		runtime.NewInt(int64(t.Second())),
		runtime.NewInt(int64(wday)),
		runtime.NewInt(int64(yday)),
		runtime.NewInt(int64(isDst)),
	})
}

// time.localtime([secs]) -> struct_time in local time
func timeLocaltime(vm *runtime.VM) int {
	var t time.Time
	if vm.GetTop() >= 1 && !runtime.IsNone(vm.Get(1)) {
		secs := vm.CheckFloat(1)
		sec := int64(secs)
		nsec := int64((secs - float64(sec)) * 1e9)
		t = time.Unix(sec, nsec)
	} else {
		t = time.Now()
	}
	vm.Push(toStructTime(t))
	return 1
}

// time.gmtime([secs]) -> struct_time in UTC
func timeGmtime(vm *runtime.VM) int {
	var t time.Time
	if vm.GetTop() >= 1 && !runtime.IsNone(vm.Get(1)) {
		secs := vm.CheckFloat(1)
		sec := int64(secs)
		nsec := int64((secs - float64(sec)) * 1e9)
		t = time.Unix(sec, nsec).UTC()
	} else {
		t = time.Now().UTC()
	}
	vm.Push(toStructTime(t))
	return 1
}

// time.mktime(t) -> float seconds since epoch
func timeMktime(vm *runtime.VM) int {
	tuple := vm.Get(1)
	pyTuple, ok := tuple.(*runtime.PyTuple)
	if !ok || len(pyTuple.Items) < 9 {
		vm.RaiseError("mktime argument must be a 9-tuple")
		return 0
	}

	year := int(runtime.ToGoValue(pyTuple.Items[0]).(int64))
	month := time.Month(runtime.ToGoValue(pyTuple.Items[1]).(int64))
	day := int(runtime.ToGoValue(pyTuple.Items[2]).(int64))
	hour := int(runtime.ToGoValue(pyTuple.Items[3]).(int64))
	min := int(runtime.ToGoValue(pyTuple.Items[4]).(int64))
	sec := int(runtime.ToGoValue(pyTuple.Items[5]).(int64))

	t := time.Date(year, month, day, hour, min, sec, 0, time.Local)
	vm.Push(runtime.NewFloat(float64(t.Unix())))
	return 1
}

// Python strftime format to Go format mapping
var strftimeFormats = map[byte]string{
	'a': "Mon",
	'A': "Monday",
	'b': "Jan",
	'B': "January",
	'c': "Mon Jan 2 15:04:05 2006",
	'd': "02",
	'H': "15",
	'I': "03",
	'j': "002", // Day of year (needs special handling)
	'm': "01",
	'M': "04",
	'p': "PM",
	'S': "05",
	'U': "", // Week number (not directly supported)
	'w': "", // Weekday as number (needs special handling)
	'W': "", // Week number (not directly supported)
	'x': "01/02/06",
	'X': "15:04:05",
	'y': "06",
	'Y': "2006",
	'Z': "MST",
	'%': "%",
}

// time.strftime(format[, t]) -> string
func timeStrftime(vm *runtime.VM) int {
	format := vm.CheckString(1)

	var t time.Time
	if vm.GetTop() >= 2 && !runtime.IsNone(vm.Get(2)) {
		tuple := vm.Get(2).(*runtime.PyTuple)
		year := int(runtime.ToGoValue(tuple.Items[0]).(int64))
		month := time.Month(runtime.ToGoValue(tuple.Items[1]).(int64))
		day := int(runtime.ToGoValue(tuple.Items[2]).(int64))
		hour := int(runtime.ToGoValue(tuple.Items[3]).(int64))
		min := int(runtime.ToGoValue(tuple.Items[4]).(int64))
		sec := int(runtime.ToGoValue(tuple.Items[5]).(int64))
		t = time.Date(year, month, day, hour, min, sec, 0, time.Local)
	} else {
		t = time.Now()
	}

	result := convertStrftime(format, t)
	vm.Push(runtime.NewString(result))
	return 1
}

func convertStrftime(format string, t time.Time) string {
	result := ""
	i := 0
	for i < len(format) {
		if format[i] == '%' && i+1 < len(format) {
			code := format[i+1]
			switch code {
			case 'Y':
				result += t.Format("2006")
			case 'y':
				result += t.Format("06")
			case 'm':
				result += t.Format("01")
			case 'd':
				result += t.Format("02")
			case 'H':
				result += t.Format("15")
			case 'I':
				result += t.Format("03")
			case 'M':
				result += t.Format("04")
			case 'S':
				result += t.Format("05")
			case 'p':
				result += t.Format("PM")
			case 'a':
				result += t.Format("Mon")
			case 'A':
				result += t.Format("Monday")
			case 'b':
				result += t.Format("Jan")
			case 'B':
				result += t.Format("January")
			case 'c':
				result += t.Format("Mon Jan 2 15:04:05 2006")
			case 'x':
				result += t.Format("01/02/06")
			case 'X':
				result += t.Format("15:04:05")
			case 'Z':
				result += t.Format("MST")
			case 'j':
				result += t.Format("002")
			case 'w':
				result += string('0' + byte(t.Weekday()))
			case '%':
				result += "%"
			default:
				result += string(format[i : i+2])
			}
			i += 2
		} else {
			result += string(format[i])
			i++
		}
	}
	return result
}

// time.strptime(string, format) -> struct_time
func timeStrptime(vm *runtime.VM) int {
	str := vm.CheckString(1)
	format := vm.CheckString(2)

	// Convert Python format to Go format
	goFormat := convertStrptimeFormat(format)

	t, err := time.Parse(goFormat, str)
	if err != nil {
		vm.RaiseError("time data '%s' does not match format '%s'", str, format)
		return 0
	}

	vm.Push(toStructTime(t))
	return 1
}

func convertStrptimeFormat(format string) string {
	result := ""
	i := 0
	for i < len(format) {
		if format[i] == '%' && i+1 < len(format) {
			code := format[i+1]
			switch code {
			case 'Y':
				result += "2006"
			case 'y':
				result += "06"
			case 'm':
				result += "01"
			case 'd':
				result += "02"
			case 'H':
				result += "15"
			case 'I':
				result += "03"
			case 'M':
				result += "04"
			case 'S':
				result += "05"
			case 'p':
				result += "PM"
			case 'a':
				result += "Mon"
			case 'A':
				result += "Monday"
			case 'b':
				result += "Jan"
			case 'B':
				result += "January"
			case 'Z':
				result += "MST"
			case '%':
				result += "%"
			default:
				result += string(format[i : i+2])
			}
			i += 2
		} else {
			result += string(format[i])
			i++
		}
	}
	return result
}

// time.ctime([secs]) -> string
func timeCtime(vm *runtime.VM) int {
	var t time.Time
	if vm.GetTop() >= 1 && !runtime.IsNone(vm.Get(1)) {
		secs := vm.CheckFloat(1)
		t = time.Unix(int64(secs), 0)
	} else {
		t = time.Now()
	}
	vm.Push(runtime.NewString(t.Format("Mon Jan  2 15:04:05 2006")))
	return 1
}

// time.asctime([t]) -> string
func timeAsctime(vm *runtime.VM) int {
	var t time.Time
	if vm.GetTop() >= 1 && !runtime.IsNone(vm.Get(1)) {
		tuple := vm.Get(1).(*runtime.PyTuple)
		year := int(runtime.ToGoValue(tuple.Items[0]).(int64))
		month := time.Month(runtime.ToGoValue(tuple.Items[1]).(int64))
		day := int(runtime.ToGoValue(tuple.Items[2]).(int64))
		hour := int(runtime.ToGoValue(tuple.Items[3]).(int64))
		min := int(runtime.ToGoValue(tuple.Items[4]).(int64))
		sec := int(runtime.ToGoValue(tuple.Items[5]).(int64))
		t = time.Date(year, month, day, hour, min, sec, 0, time.Local)
	} else {
		t = time.Now()
	}
	vm.Push(runtime.NewString(t.Format("Mon Jan  2 15:04:05 2006")))
	return 1
}

// time.perf_counter() -> float seconds
func timePerfCounter(vm *runtime.VM) int {
	elapsed := time.Since(startTime)
	vm.Push(runtime.NewFloat(elapsed.Seconds()))
	return 1
}

// time.perf_counter_ns() -> int nanoseconds
func timePerfCounterNs(vm *runtime.VM) int {
	elapsed := time.Since(startTime)
	vm.Push(runtime.NewInt(elapsed.Nanoseconds()))
	return 1
}

// time.monotonic() -> float seconds
func timeMonotonic(vm *runtime.VM) int {
	elapsed := time.Since(startTime)
	vm.Push(runtime.NewFloat(elapsed.Seconds()))
	return 1
}

// time.monotonic_ns() -> int nanoseconds
func timeMonotonicNs(vm *runtime.VM) int {
	elapsed := time.Since(startTime)
	vm.Push(runtime.NewInt(elapsed.Nanoseconds()))
	return 1
}

// time.process_time() -> float seconds of CPU time
func timeProcessTime(vm *runtime.VM) int {
	// Go doesn't have direct CPU time access, use wall clock as approximation
	elapsed := time.Since(startTime)
	vm.Push(runtime.NewFloat(elapsed.Seconds()))
	return 1
}

// time.process_time_ns() -> int nanoseconds of CPU time
func timeProcessTimeNs(vm *runtime.VM) int {
	elapsed := time.Since(startTime)
	vm.Push(runtime.NewInt(elapsed.Nanoseconds()))
	return 1
}
