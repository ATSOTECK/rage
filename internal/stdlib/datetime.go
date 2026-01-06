package stdlib

import (
	"fmt"
	"time"

	"github.com/ATSOTECK/rage/internal/runtime"
)

// InitDatetimeModule registers the datetime module
func InitDatetimeModule() {
	// Register datetime type metatable
	datetimeMT := &runtime.TypeMetatable{
		Name: "datetime",
		Methods: map[string]runtime.GoFunction{
			"date":        datetimeDate,
			"time":        datetimeTime,
			"timetz":      datetimeTimetz,
			"replace":     datetimeReplace,
			"timetuple":   datetimeTimetuple,
			"timestamp":   datetimeTimestamp,
			"weekday":     datetimeWeekday,
			"isoweekday":  datetimeIsoweekday,
			"isoformat":   datetimeIsoformat,
			"strftime":    datetimeStrftime,
			"ctime":       datetimeCtime,
			"toordinal":   datetimeToordinal,
			"year":        datetimeYear,
			"month":       datetimeMonth,
			"day":         datetimeDay,
			"hour":        datetimeHour,
			"minute":      datetimeMinute,
			"second":      datetimeSecond,
			"microsecond": datetimeMicrosecond,
		},
	}
	runtime.RegisterTypeMetatable("datetime", datetimeMT)

	// Register date type metatable
	dateMT := &runtime.TypeMetatable{
		Name: "date",
		Methods: map[string]runtime.GoFunction{
			"replace":    dateReplace,
			"timetuple":  dateTimetuple,
			"weekday":    dateWeekday,
			"isoweekday": dateIsoweekday,
			"isoformat":  dateIsoformat,
			"strftime":   dateStrftime,
			"ctime":      dateCtime,
			"toordinal":  dateToordinal,
			"year":       dateYear,
			"month":      dateMonth,
			"day":        dateDay,
		},
	}
	runtime.RegisterTypeMetatable("date", dateMT)

	// Register time type metatable
	timeMT := &runtime.TypeMetatable{
		Name: "time",
		Methods: map[string]runtime.GoFunction{
			"replace":     timeReplace,
			"isoformat":   timeIsoformat,
			"strftime":    timeStrftimeMethod,
			"hour":        timeHour,
			"minute":      timeMinute,
			"second":      timeSecond,
			"microsecond": timeMicrosecond,
		},
	}
	runtime.RegisterTypeMetatable("time", timeMT)

	// Register timedelta type metatable
	timedeltaMT := &runtime.TypeMetatable{
		Name: "timedelta",
		Methods: map[string]runtime.GoFunction{
			"total_seconds": timedeltaTotalSeconds,
			"days":          timedeltaDays,
			"seconds":       timedeltaSeconds,
			"microseconds":  timedeltaMicroseconds,
		},
	}
	runtime.RegisterTypeMetatable("timedelta", timedeltaMT)

	runtime.NewModuleBuilder("datetime").
		Doc("Basic date and time types.").
		// Constants
		Const("MINYEAR", runtime.NewInt(1)).
		Const("MAXYEAR", runtime.NewInt(9999)).
		// Type constructors
		Func("datetime", datetimeNew).
		Func("date", dateNew).
		Func("time", timeNew).
		Func("timedelta", timedeltaNew).
		// Module-level functions (aliased to type methods)
		Func("now", datetimeNow).
		Func("today", dateToday).
		Func("utcnow", datetimeUtcnow).
		Func("fromtimestamp", datetimeFromtimestamp).
		Func("fromisoformat", datetimeFromisoformat).
		Func("combine", datetimeCombine).
		Register()
}

// =====================================
// PyDatetime - datetime class
// =====================================

type PyDatetime struct {
	Year        int
	Month       int
	Day         int
	Hour        int
	Minute      int
	Second      int
	Microsecond int
}

func (d *PyDatetime) Type() string { return "datetime" }
func (d *PyDatetime) String() string {
	if d.Microsecond > 0 {
		return fmt.Sprintf("datetime.datetime(%d, %d, %d, %d, %d, %d, %d)",
			d.Year, d.Month, d.Day, d.Hour, d.Minute, d.Second, d.Microsecond)
	}
	if d.Second > 0 || d.Minute > 0 || d.Hour > 0 {
		return fmt.Sprintf("datetime.datetime(%d, %d, %d, %d, %d, %d)",
			d.Year, d.Month, d.Day, d.Hour, d.Minute, d.Second)
	}
	return fmt.Sprintf("datetime.datetime(%d, %d, %d)", d.Year, d.Month, d.Day)
}

func (d *PyDatetime) toTime() time.Time {
	return time.Date(d.Year, time.Month(d.Month), d.Day, d.Hour, d.Minute, d.Second, d.Microsecond*1000, time.Local)
}

func datetimeFromTime(t time.Time) *PyDatetime {
	return &PyDatetime{
		Year:        t.Year(),
		Month:       int(t.Month()),
		Day:         t.Day(),
		Hour:        t.Hour(),
		Minute:      t.Minute(),
		Second:      t.Second(),
		Microsecond: t.Nanosecond() / 1000,
	}
}

func wrapDatetime(dt *PyDatetime) *runtime.PyUserData {
	ud := runtime.NewUserData(dt)
	ud.Metatable = runtime.NewDict()
	ud.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("datetime")
	return ud
}

// datetime(year, month, day, hour=0, minute=0, second=0, microsecond=0)
func datetimeNew(vm *runtime.VM) int {
	year := int(vm.CheckInt(1))
	month := int(vm.CheckInt(2))
	day := int(vm.CheckInt(3))

	hour := 0
	minute := 0
	second := 0
	microsecond := 0

	if vm.GetTop() >= 4 && !runtime.IsNone(vm.Get(4)) {
		hour = int(vm.CheckInt(4))
	}
	if vm.GetTop() >= 5 && !runtime.IsNone(vm.Get(5)) {
		minute = int(vm.CheckInt(5))
	}
	if vm.GetTop() >= 6 && !runtime.IsNone(vm.Get(6)) {
		second = int(vm.CheckInt(6))
	}
	if vm.GetTop() >= 7 && !runtime.IsNone(vm.Get(7)) {
		microsecond = int(vm.CheckInt(7))
	}

	// Validate
	if year < 1 || year > 9999 {
		vm.RaiseError("year %d is out of range", year)
		return 0
	}
	if month < 1 || month > 12 {
		vm.RaiseError("month must be in 1..12")
		return 0
	}
	if day < 1 || day > 31 {
		vm.RaiseError("day is out of range for month")
		return 0
	}
	if hour < 0 || hour > 23 {
		vm.RaiseError("hour must be in 0..23")
		return 0
	}
	if minute < 0 || minute > 59 {
		vm.RaiseError("minute must be in 0..59")
		return 0
	}
	if second < 0 || second > 59 {
		vm.RaiseError("second must be in 0..59")
		return 0
	}
	if microsecond < 0 || microsecond > 999999 {
		vm.RaiseError("microsecond must be in 0..999999")
		return 0
	}

	dt := &PyDatetime{
		Year:        year,
		Month:       month,
		Day:         day,
		Hour:        hour,
		Minute:      minute,
		Second:      second,
		Microsecond: microsecond,
	}

	vm.Push(wrapDatetime(dt))
	return 1
}

// datetime.now()
func datetimeNow(vm *runtime.VM) int {
	dt := datetimeFromTime(time.Now())
	vm.Push(wrapDatetime(dt))
	return 1
}

// datetime.utcnow()
func datetimeUtcnow(vm *runtime.VM) int {
	dt := datetimeFromTime(time.Now().UTC())
	vm.Push(wrapDatetime(dt))
	return 1
}

// datetime.fromtimestamp(timestamp)
func datetimeFromtimestamp(vm *runtime.VM) int {
	ts := vm.CheckFloat(1)
	sec := int64(ts)
	nsec := int64((ts - float64(sec)) * 1e9)
	t := time.Unix(sec, nsec)
	dt := datetimeFromTime(t)
	vm.Push(wrapDatetime(dt))
	return 1
}

// datetime.fromisoformat(date_string)
func datetimeFromisoformat(vm *runtime.VM) int {
	s := vm.CheckString(1)

	// Try various ISO formats
	formats := []string{
		"2006-01-02T15:04:05.999999",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05.999999",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	var t time.Time
	var err error
	for _, format := range formats {
		t, err = time.Parse(format, s)
		if err == nil {
			break
		}
	}

	if err != nil {
		vm.RaiseError("Invalid isoformat string: '%s'", s)
		return 0
	}

	dt := datetimeFromTime(t)
	vm.Push(wrapDatetime(dt))
	return 1
}

// datetime.combine(date, time)
func datetimeCombine(vm *runtime.VM) int {
	dateUd := vm.ToUserData(1)
	timeUd := vm.ToUserData(2)

	if dateUd == nil || timeUd == nil {
		vm.RaiseError("combine requires date and time objects")
		return 0
	}

	date, ok := dateUd.Value.(*PyDate)
	if !ok {
		vm.RaiseError("first argument must be a date object")
		return 0
	}

	timeObj, ok := timeUd.Value.(*PyTime)
	if !ok {
		vm.RaiseError("second argument must be a time object")
		return 0
	}

	dt := &PyDatetime{
		Year:        date.Year,
		Month:       date.Month,
		Day:         date.Day,
		Hour:        timeObj.Hour,
		Minute:      timeObj.Minute,
		Second:      timeObj.Second,
		Microsecond: timeObj.Microsecond,
	}

	vm.Push(wrapDatetime(dt))
	return 1
}

// datetime.date() -> date object
func datetimeDate(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected datetime object")
		return 0
	}
	dt, ok := ud.Value.(*PyDatetime)
	if !ok {
		vm.RaiseError("expected datetime object")
		return 0
	}

	d := &PyDate{Year: dt.Year, Month: dt.Month, Day: dt.Day}
	vm.Push(wrapDate(d))
	return 1
}

// datetime.time() -> time object
func datetimeTime(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected datetime object")
		return 0
	}
	dt, ok := ud.Value.(*PyDatetime)
	if !ok {
		vm.RaiseError("expected datetime object")
		return 0
	}

	t := &PyTime{
		Hour:        dt.Hour,
		Minute:      dt.Minute,
		Second:      dt.Second,
		Microsecond: dt.Microsecond,
	}
	vm.Push(wrapTime(t))
	return 1
}

// datetime.timetz() -> time object (same as time() without tzinfo)
func datetimeTimetz(vm *runtime.VM) int {
	return datetimeTime(vm)
}

// datetime.replace(year=None, month=None, ...)
func datetimeReplace(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected datetime object")
		return 0
	}
	dt, ok := ud.Value.(*PyDatetime)
	if !ok {
		vm.RaiseError("expected datetime object")
		return 0
	}

	// Copy current values
	newDt := &PyDatetime{
		Year:        dt.Year,
		Month:       dt.Month,
		Day:         dt.Day,
		Hour:        dt.Hour,
		Minute:      dt.Minute,
		Second:      dt.Second,
		Microsecond: dt.Microsecond,
	}

	// Override with provided values
	if vm.GetTop() >= 2 && !runtime.IsNone(vm.Get(2)) {
		newDt.Year = int(vm.CheckInt(2))
	}
	if vm.GetTop() >= 3 && !runtime.IsNone(vm.Get(3)) {
		newDt.Month = int(vm.CheckInt(3))
	}
	if vm.GetTop() >= 4 && !runtime.IsNone(vm.Get(4)) {
		newDt.Day = int(vm.CheckInt(4))
	}
	if vm.GetTop() >= 5 && !runtime.IsNone(vm.Get(5)) {
		newDt.Hour = int(vm.CheckInt(5))
	}
	if vm.GetTop() >= 6 && !runtime.IsNone(vm.Get(6)) {
		newDt.Minute = int(vm.CheckInt(6))
	}
	if vm.GetTop() >= 7 && !runtime.IsNone(vm.Get(7)) {
		newDt.Second = int(vm.CheckInt(7))
	}
	if vm.GetTop() >= 8 && !runtime.IsNone(vm.Get(8)) {
		newDt.Microsecond = int(vm.CheckInt(8))
	}

	vm.Push(wrapDatetime(newDt))
	return 1
}

// datetime.timetuple() -> struct_time
func datetimeTimetuple(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected datetime object")
		return 0
	}
	dt, ok := ud.Value.(*PyDatetime)
	if !ok {
		vm.RaiseError("expected datetime object")
		return 0
	}

	t := dt.toTime()
	wday := int(t.Weekday())
	if wday == 0 {
		wday = 6
	} else {
		wday--
	}

	vm.Push(runtime.NewTuple([]runtime.Value{
		runtime.NewInt(int64(dt.Year)),
		runtime.NewInt(int64(dt.Month)),
		runtime.NewInt(int64(dt.Day)),
		runtime.NewInt(int64(dt.Hour)),
		runtime.NewInt(int64(dt.Minute)),
		runtime.NewInt(int64(dt.Second)),
		runtime.NewInt(int64(wday)),
		runtime.NewInt(int64(t.YearDay())),
		runtime.NewInt(-1), // DST unknown
	}))
	return 1
}

// datetime.timestamp() -> float seconds since epoch
func datetimeTimestamp(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected datetime object")
		return 0
	}
	dt, ok := ud.Value.(*PyDatetime)
	if !ok {
		vm.RaiseError("expected datetime object")
		return 0
	}

	t := dt.toTime()
	ts := float64(t.Unix()) + float64(t.Nanosecond())/1e9
	vm.Push(runtime.NewFloat(ts))
	return 1
}

// datetime.weekday() -> 0-6 (Monday=0)
func datetimeWeekday(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected datetime object")
		return 0
	}
	dt, ok := ud.Value.(*PyDatetime)
	if !ok {
		vm.RaiseError("expected datetime object")
		return 0
	}

	t := dt.toTime()
	wday := int(t.Weekday())
	if wday == 0 {
		wday = 6
	} else {
		wday--
	}
	vm.Push(runtime.NewInt(int64(wday)))
	return 1
}

// datetime.isoweekday() -> 1-7 (Monday=1)
func datetimeIsoweekday(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected datetime object")
		return 0
	}
	dt, ok := ud.Value.(*PyDatetime)
	if !ok {
		vm.RaiseError("expected datetime object")
		return 0
	}

	t := dt.toTime()
	wday := int(t.Weekday())
	if wday == 0 {
		wday = 7
	}
	vm.Push(runtime.NewInt(int64(wday)))
	return 1
}

// datetime.isoformat(sep='T') -> string
func datetimeIsoformat(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected datetime object")
		return 0
	}
	dt, ok := ud.Value.(*PyDatetime)
	if !ok {
		vm.RaiseError("expected datetime object")
		return 0
	}

	sep := "T"
	if vm.GetTop() >= 2 && !runtime.IsNone(vm.Get(2)) {
		sep = vm.CheckString(2)
	}

	var result string
	if dt.Microsecond > 0 {
		result = fmt.Sprintf("%04d-%02d-%02d%s%02d:%02d:%02d.%06d",
			dt.Year, dt.Month, dt.Day, sep, dt.Hour, dt.Minute, dt.Second, dt.Microsecond)
	} else {
		result = fmt.Sprintf("%04d-%02d-%02d%s%02d:%02d:%02d",
			dt.Year, dt.Month, dt.Day, sep, dt.Hour, dt.Minute, dt.Second)
	}

	vm.Push(runtime.NewString(result))
	return 1
}

// datetime.strftime(format) -> string
func datetimeStrftime(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected datetime object")
		return 0
	}
	dt, ok := ud.Value.(*PyDatetime)
	if !ok {
		vm.RaiseError("expected datetime object")
		return 0
	}

	format := vm.CheckString(2)
	result := convertStrftimeDatetime(format, dt.toTime())
	vm.Push(runtime.NewString(result))
	return 1
}

// datetime.ctime() -> string
func datetimeCtime(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected datetime object")
		return 0
	}
	dt, ok := ud.Value.(*PyDatetime)
	if !ok {
		vm.RaiseError("expected datetime object")
		return 0
	}

	t := dt.toTime()
	vm.Push(runtime.NewString(t.Format("Mon Jan  2 15:04:05 2006")))
	return 1
}

// datetime.toordinal() -> int
func datetimeToordinal(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected datetime object")
		return 0
	}
	dt, ok := ud.Value.(*PyDatetime)
	if !ok {
		vm.RaiseError("expected datetime object")
		return 0
	}

	ordinal := toOrdinal(dt.Year, dt.Month, dt.Day)
	vm.Push(runtime.NewInt(int64(ordinal)))
	return 1
}

// Property accessors for datetime
func datetimeYear(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	dt := ud.Value.(*PyDatetime)
	vm.Push(runtime.NewInt(int64(dt.Year)))
	return 1
}

func datetimeMonth(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	dt := ud.Value.(*PyDatetime)
	vm.Push(runtime.NewInt(int64(dt.Month)))
	return 1
}

func datetimeDay(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	dt := ud.Value.(*PyDatetime)
	vm.Push(runtime.NewInt(int64(dt.Day)))
	return 1
}

func datetimeHour(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	dt := ud.Value.(*PyDatetime)
	vm.Push(runtime.NewInt(int64(dt.Hour)))
	return 1
}

func datetimeMinute(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	dt := ud.Value.(*PyDatetime)
	vm.Push(runtime.NewInt(int64(dt.Minute)))
	return 1
}

func datetimeSecond(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	dt := ud.Value.(*PyDatetime)
	vm.Push(runtime.NewInt(int64(dt.Second)))
	return 1
}

func datetimeMicrosecond(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	dt := ud.Value.(*PyDatetime)
	vm.Push(runtime.NewInt(int64(dt.Microsecond)))
	return 1
}

// =====================================
// PyDate - date class
// =====================================

type PyDate struct {
	Year  int
	Month int
	Day   int
}

func (d *PyDate) Type() string { return "date" }
func (d *PyDate) String() string {
	return fmt.Sprintf("datetime.date(%d, %d, %d)", d.Year, d.Month, d.Day)
}

func (d *PyDate) toTime() time.Time {
	return time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, time.Local)
}

func wrapDate(d *PyDate) *runtime.PyUserData {
	ud := runtime.NewUserData(d)
	ud.Metatable = runtime.NewDict()
	ud.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("date")
	return ud
}

// date(year, month, day)
func dateNew(vm *runtime.VM) int {
	year := int(vm.CheckInt(1))
	month := int(vm.CheckInt(2))
	day := int(vm.CheckInt(3))

	if year < 1 || year > 9999 {
		vm.RaiseError("year %d is out of range", year)
		return 0
	}
	if month < 1 || month > 12 {
		vm.RaiseError("month must be in 1..12")
		return 0
	}
	if day < 1 || day > 31 {
		vm.RaiseError("day is out of range for month")
		return 0
	}

	d := &PyDate{Year: year, Month: month, Day: day}
	vm.Push(wrapDate(d))
	return 1
}

// date.today()
func dateToday(vm *runtime.VM) int {
	now := time.Now()
	d := &PyDate{Year: now.Year(), Month: int(now.Month()), Day: now.Day()}
	vm.Push(wrapDate(d))
	return 1
}

// date.replace(year=None, month=None, day=None)
func dateReplace(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected date object")
		return 0
	}
	d, ok := ud.Value.(*PyDate)
	if !ok {
		vm.RaiseError("expected date object")
		return 0
	}

	newD := &PyDate{Year: d.Year, Month: d.Month, Day: d.Day}

	if vm.GetTop() >= 2 && !runtime.IsNone(vm.Get(2)) {
		newD.Year = int(vm.CheckInt(2))
	}
	if vm.GetTop() >= 3 && !runtime.IsNone(vm.Get(3)) {
		newD.Month = int(vm.CheckInt(3))
	}
	if vm.GetTop() >= 4 && !runtime.IsNone(vm.Get(4)) {
		newD.Day = int(vm.CheckInt(4))
	}

	vm.Push(wrapDate(newD))
	return 1
}

// date.timetuple() -> struct_time
func dateTimetuple(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected date object")
		return 0
	}
	d, ok := ud.Value.(*PyDate)
	if !ok {
		vm.RaiseError("expected date object")
		return 0
	}

	t := d.toTime()
	wday := int(t.Weekday())
	if wday == 0 {
		wday = 6
	} else {
		wday--
	}

	vm.Push(runtime.NewTuple([]runtime.Value{
		runtime.NewInt(int64(d.Year)),
		runtime.NewInt(int64(d.Month)),
		runtime.NewInt(int64(d.Day)),
		runtime.NewInt(0),
		runtime.NewInt(0),
		runtime.NewInt(0),
		runtime.NewInt(int64(wday)),
		runtime.NewInt(int64(t.YearDay())),
		runtime.NewInt(-1),
	}))
	return 1
}

// date.weekday() -> 0-6 (Monday=0)
func dateWeekday(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected date object")
		return 0
	}
	d, ok := ud.Value.(*PyDate)
	if !ok {
		vm.RaiseError("expected date object")
		return 0
	}

	t := d.toTime()
	wday := int(t.Weekday())
	if wday == 0 {
		wday = 6
	} else {
		wday--
	}
	vm.Push(runtime.NewInt(int64(wday)))
	return 1
}

// date.isoweekday() -> 1-7 (Monday=1)
func dateIsoweekday(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected date object")
		return 0
	}
	d, ok := ud.Value.(*PyDate)
	if !ok {
		vm.RaiseError("expected date object")
		return 0
	}

	t := d.toTime()
	wday := int(t.Weekday())
	if wday == 0 {
		wday = 7
	}
	vm.Push(runtime.NewInt(int64(wday)))
	return 1
}

// date.isoformat() -> string 'YYYY-MM-DD'
func dateIsoformat(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected date object")
		return 0
	}
	d, ok := ud.Value.(*PyDate)
	if !ok {
		vm.RaiseError("expected date object")
		return 0
	}

	result := fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
	vm.Push(runtime.NewString(result))
	return 1
}

// date.strftime(format) -> string
func dateStrftime(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected date object")
		return 0
	}
	d, ok := ud.Value.(*PyDate)
	if !ok {
		vm.RaiseError("expected date object")
		return 0
	}

	format := vm.CheckString(2)
	result := convertStrftimeDatetime(format, d.toTime())
	vm.Push(runtime.NewString(result))
	return 1
}

// date.ctime() -> string
func dateCtime(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected date object")
		return 0
	}
	d, ok := ud.Value.(*PyDate)
	if !ok {
		vm.RaiseError("expected date object")
		return 0
	}

	t := d.toTime()
	vm.Push(runtime.NewString(t.Format("Mon Jan  2 15:04:05 2006")))
	return 1
}

// date.toordinal() -> int
func dateToordinal(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected date object")
		return 0
	}
	d, ok := ud.Value.(*PyDate)
	if !ok {
		vm.RaiseError("expected date object")
		return 0
	}

	ordinal := toOrdinal(d.Year, d.Month, d.Day)
	vm.Push(runtime.NewInt(int64(ordinal)))
	return 1
}

// Property accessors for date
func dateYear(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	d := ud.Value.(*PyDate)
	vm.Push(runtime.NewInt(int64(d.Year)))
	return 1
}

func dateMonth(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	d := ud.Value.(*PyDate)
	vm.Push(runtime.NewInt(int64(d.Month)))
	return 1
}

func dateDay(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	d := ud.Value.(*PyDate)
	vm.Push(runtime.NewInt(int64(d.Day)))
	return 1
}

// =====================================
// PyTime - time class
// =====================================

type PyTime struct {
	Hour        int
	Minute      int
	Second      int
	Microsecond int
}

func (t *PyTime) Type() string { return "time" }
func (t *PyTime) String() string {
	if t.Microsecond > 0 {
		return fmt.Sprintf("datetime.time(%d, %d, %d, %d)", t.Hour, t.Minute, t.Second, t.Microsecond)
	}
	if t.Second > 0 {
		return fmt.Sprintf("datetime.time(%d, %d, %d)", t.Hour, t.Minute, t.Second)
	}
	return fmt.Sprintf("datetime.time(%d, %d)", t.Hour, t.Minute)
}

func wrapTime(t *PyTime) *runtime.PyUserData {
	ud := runtime.NewUserData(t)
	ud.Metatable = runtime.NewDict()
	ud.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("time")
	return ud
}

// time(hour=0, minute=0, second=0, microsecond=0)
func timeNew(vm *runtime.VM) int {
	hour := 0
	minute := 0
	second := 0
	microsecond := 0

	if vm.GetTop() >= 1 && !runtime.IsNone(vm.Get(1)) {
		hour = int(vm.CheckInt(1))
	}
	if vm.GetTop() >= 2 && !runtime.IsNone(vm.Get(2)) {
		minute = int(vm.CheckInt(2))
	}
	if vm.GetTop() >= 3 && !runtime.IsNone(vm.Get(3)) {
		second = int(vm.CheckInt(3))
	}
	if vm.GetTop() >= 4 && !runtime.IsNone(vm.Get(4)) {
		microsecond = int(vm.CheckInt(4))
	}

	if hour < 0 || hour > 23 {
		vm.RaiseError("hour must be in 0..23")
		return 0
	}
	if minute < 0 || minute > 59 {
		vm.RaiseError("minute must be in 0..59")
		return 0
	}
	if second < 0 || second > 59 {
		vm.RaiseError("second must be in 0..59")
		return 0
	}
	if microsecond < 0 || microsecond > 999999 {
		vm.RaiseError("microsecond must be in 0..999999")
		return 0
	}

	t := &PyTime{
		Hour:        hour,
		Minute:      minute,
		Second:      second,
		Microsecond: microsecond,
	}
	vm.Push(wrapTime(t))
	return 1
}

// time.replace(hour=None, minute=None, second=None, microsecond=None)
func timeReplace(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected time object")
		return 0
	}
	t, ok := ud.Value.(*PyTime)
	if !ok {
		vm.RaiseError("expected time object")
		return 0
	}

	newT := &PyTime{
		Hour:        t.Hour,
		Minute:      t.Minute,
		Second:      t.Second,
		Microsecond: t.Microsecond,
	}

	if vm.GetTop() >= 2 && !runtime.IsNone(vm.Get(2)) {
		newT.Hour = int(vm.CheckInt(2))
	}
	if vm.GetTop() >= 3 && !runtime.IsNone(vm.Get(3)) {
		newT.Minute = int(vm.CheckInt(3))
	}
	if vm.GetTop() >= 4 && !runtime.IsNone(vm.Get(4)) {
		newT.Second = int(vm.CheckInt(4))
	}
	if vm.GetTop() >= 5 && !runtime.IsNone(vm.Get(5)) {
		newT.Microsecond = int(vm.CheckInt(5))
	}

	vm.Push(wrapTime(newT))
	return 1
}

// time.isoformat() -> string 'HH:MM:SS.ffffff'
func timeIsoformat(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected time object")
		return 0
	}
	t, ok := ud.Value.(*PyTime)
	if !ok {
		vm.RaiseError("expected time object")
		return 0
	}

	var result string
	if t.Microsecond > 0 {
		result = fmt.Sprintf("%02d:%02d:%02d.%06d", t.Hour, t.Minute, t.Second, t.Microsecond)
	} else {
		result = fmt.Sprintf("%02d:%02d:%02d", t.Hour, t.Minute, t.Second)
	}

	vm.Push(runtime.NewString(result))
	return 1
}

// time.strftime(format) -> string
func timeStrftimeMethod(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected time object")
		return 0
	}
	t, ok := ud.Value.(*PyTime)
	if !ok {
		vm.RaiseError("expected time object")
		return 0
	}

	format := vm.CheckString(2)
	// Create a time.Time with the time values
	goTime := time.Date(1900, 1, 1, t.Hour, t.Minute, t.Second, t.Microsecond*1000, time.Local)
	result := convertStrftimeDatetime(format, goTime)
	vm.Push(runtime.NewString(result))
	return 1
}

// Property accessors for time
func timeHour(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	t := ud.Value.(*PyTime)
	vm.Push(runtime.NewInt(int64(t.Hour)))
	return 1
}

func timeMinute(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	t := ud.Value.(*PyTime)
	vm.Push(runtime.NewInt(int64(t.Minute)))
	return 1
}

func timeSecond(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	t := ud.Value.(*PyTime)
	vm.Push(runtime.NewInt(int64(t.Second)))
	return 1
}

func timeMicrosecond(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	t := ud.Value.(*PyTime)
	vm.Push(runtime.NewInt(int64(t.Microsecond)))
	return 1
}

// =====================================
// PyTimedelta - timedelta class
// =====================================

type PyTimedelta struct {
	Days         int64
	Seconds      int64
	Microseconds int64
}

func (td *PyTimedelta) Type() string { return "timedelta" }
func (td *PyTimedelta) String() string {
	if td.Microseconds > 0 {
		return fmt.Sprintf("datetime.timedelta(days=%d, seconds=%d, microseconds=%d)",
			td.Days, td.Seconds, td.Microseconds)
	}
	if td.Seconds > 0 {
		return fmt.Sprintf("datetime.timedelta(days=%d, seconds=%d)", td.Days, td.Seconds)
	}
	return fmt.Sprintf("datetime.timedelta(days=%d)", td.Days)
}

func (td *PyTimedelta) TotalSeconds() float64 {
	return float64(td.Days)*86400 + float64(td.Seconds) + float64(td.Microseconds)/1e6
}

func wrapTimedelta(td *PyTimedelta) *runtime.PyUserData {
	ud := runtime.NewUserData(td)
	ud.Metatable = runtime.NewDict()
	ud.Metatable.Items[runtime.NewString("__type__")] = runtime.NewString("timedelta")
	return ud
}

// timedelta(days=0, seconds=0, microseconds=0, milliseconds=0, minutes=0, hours=0, weeks=0)
func timedeltaNew(vm *runtime.VM) int {
	var days, seconds, microseconds int64

	if vm.GetTop() >= 1 && !runtime.IsNone(vm.Get(1)) {
		days = vm.CheckInt(1)
	}
	if vm.GetTop() >= 2 && !runtime.IsNone(vm.Get(2)) {
		seconds = vm.CheckInt(2)
	}
	if vm.GetTop() >= 3 && !runtime.IsNone(vm.Get(3)) {
		microseconds = vm.CheckInt(3)
	}
	if vm.GetTop() >= 4 && !runtime.IsNone(vm.Get(4)) {
		milliseconds := vm.CheckInt(4)
		microseconds += milliseconds * 1000
	}
	if vm.GetTop() >= 5 && !runtime.IsNone(vm.Get(5)) {
		minutes := vm.CheckInt(5)
		seconds += minutes * 60
	}
	if vm.GetTop() >= 6 && !runtime.IsNone(vm.Get(6)) {
		hours := vm.CheckInt(6)
		seconds += hours * 3600
	}
	if vm.GetTop() >= 7 && !runtime.IsNone(vm.Get(7)) {
		weeks := vm.CheckInt(7)
		days += weeks * 7
	}

	// Normalize
	seconds += microseconds / 1000000
	microseconds = microseconds % 1000000
	if microseconds < 0 {
		microseconds += 1000000
		seconds--
	}

	days += seconds / 86400
	seconds = seconds % 86400
	if seconds < 0 {
		seconds += 86400
		days--
	}

	td := &PyTimedelta{
		Days:         days,
		Seconds:      seconds,
		Microseconds: microseconds,
	}
	vm.Push(wrapTimedelta(td))
	return 1
}

// timedelta.total_seconds() -> float
func timedeltaTotalSeconds(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	if ud == nil {
		vm.RaiseError("expected timedelta object")
		return 0
	}
	td, ok := ud.Value.(*PyTimedelta)
	if !ok {
		vm.RaiseError("expected timedelta object")
		return 0
	}

	vm.Push(runtime.NewFloat(td.TotalSeconds()))
	return 1
}

// Property accessors for timedelta
func timedeltaDays(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	td := ud.Value.(*PyTimedelta)
	vm.Push(runtime.NewInt(td.Days))
	return 1
}

func timedeltaSeconds(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	td := ud.Value.(*PyTimedelta)
	vm.Push(runtime.NewInt(td.Seconds))
	return 1
}

func timedeltaMicroseconds(vm *runtime.VM) int {
	ud := vm.ToUserData(1)
	td := ud.Value.(*PyTimedelta)
	vm.Push(runtime.NewInt(td.Microseconds))
	return 1
}

// =====================================
// Helper functions
// =====================================

// toOrdinal calculates the proleptic Gregorian ordinal for a date
func toOrdinal(year, month, day int) int {
	// Days in each month (non-leap year)
	daysInMonth := []int{0, 31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

	// Check for leap year
	isLeap := (year%4 == 0 && year%100 != 0) || (year%400 == 0)

	// Days before this year
	y := year - 1
	days := y*365 + y/4 - y/100 + y/400

	// Days in this year before this month
	for m := 1; m < month; m++ {
		days += daysInMonth[m]
		if m == 2 && isLeap {
			days++
		}
	}

	// Add day of month
	days += day

	return days
}

// convertStrftimeDatetime converts Python strftime format to formatted string
func convertStrftimeDatetime(format string, t time.Time) string {
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
			case 'f':
				result += fmt.Sprintf("%06d", t.Nanosecond()/1000)
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
			case 'z':
				result += t.Format("-0700")
			case 'j':
				result += fmt.Sprintf("%03d", t.YearDay())
			case 'w':
				result += fmt.Sprintf("%d", t.Weekday())
			case 'W':
				_, week := t.ISOWeek()
				result += fmt.Sprintf("%02d", week)
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
