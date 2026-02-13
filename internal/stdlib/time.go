package stdlib

import (
	"time"

	"github.com/ATSOTECK/rage/internal/runtime"
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

// pythonToGoFormat maps Python strftime/strptime format codes to Go time layout strings.
// Used by both convertStrftime and convertStrptimeFormat.
var pythonToGoFormat = map[byte]string{
	'Y': "2006",
	'y': "06",
	'm': "01",
	'd': "02",
	'H': "15",
	'I': "03",
	'M': "04",
	'S': "05",
	'p': "PM",
	'a': "Mon",
	'A': "Monday",
	'b': "Jan",
	'B': "January",
	'c': "Mon Jan 2 15:04:05 2006",
	'x': "01/02/06",
	'X': "15:04:05",
	'Z': "MST",
	'j': "002",
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
			// %w needs special handling: it computes a value, not a Go layout format.
			if code == 'w' {
				result += string('0' + byte(t.Weekday()))
			} else if goLayout, ok := pythonToGoFormat[code]; ok {
				result += t.Format(goLayout)
			} else {
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
			if goLayout, ok := pythonToGoFormat[code]; ok {
				result += goLayout
			} else {
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
