# Test: datetime module
# Tests datetime, date, time, timedelta classes

import datetime

def test_datetime_constants():
    expect(1, datetime.MINYEAR)
    expect(9999, datetime.MAXYEAR)

def test_datetime_constructor():
    dt = datetime.datetime(2024, 6, 15, 14, 30, 45, 123456)
    expect(2024, dt.year())
    expect(6, dt.month())
    expect(15, dt.day())
    expect(14, dt.hour())
    expect(30, dt.minute())
    expect(45, dt.second())
    expect(123456, dt.microsecond())

    dt2 = datetime.datetime(2024, 1, 1)
    expect(0, dt2.hour())
    expect(0, dt2.minute())
    expect(0, dt2.second())

def test_datetime_weekday():
    dt3 = datetime.datetime(2024, 6, 17)  # Monday
    expect(0, dt3.weekday())
    expect(1, dt3.isoweekday())

    dt4 = datetime.datetime(2024, 6, 22)  # Saturday
    expect(5, dt4.weekday())
    expect(6, dt4.isoweekday())

def test_datetime_isoformat():
    dt5 = datetime.datetime(2024, 3, 14, 9, 26, 53)
    expect("2024-03-14T09:26:53", dt5.isoformat())
    expect("2024-03-14 09:26:53", dt5.isoformat(" "))

    dt6 = datetime.datetime(2024, 3, 14, 9, 26, 53, 500000)
    expect("2024-03-14T09:26:53.500000", dt6.isoformat())

def test_datetime_strftime():
    dt5 = datetime.datetime(2024, 3, 14, 9, 26, 53)
    expect("2024-03-14", dt5.strftime("%Y-%m-%d"))
    expect("09:26:53", dt5.strftime("%H:%M:%S"))
    expect("2024-03-14 09:26:53", dt5.strftime("%Y-%m-%d %H:%M:%S"))
    expect("Thursday", dt5.strftime("%A"))
    expect("March", dt5.strftime("%B"))

def test_datetime_timetuple():
    dt5 = datetime.datetime(2024, 3, 14, 9, 26, 53)
    tt = dt5.timetuple()
    expect(2024, tt[0])
    expect(3, tt[1])
    expect(14, tt[2])
    expect(9, tt[3])
    expect(26, tt[4])
    expect(53, tt[5])

def test_datetime_toordinal():
    dt5 = datetime.datetime(2024, 3, 14, 9, 26, 53)
    # Check ordinal is a reasonable value (around 739000 for 2024)
    ordinal = dt5.toordinal()
    expect(True, 738000 < ordinal < 740000)

def test_datetime_timestamp():
    dt_epoch = datetime.datetime(1970, 1, 1, 0, 0, 0)
    ts = dt_epoch.timestamp()
    expect(True, ts == ts)  # Check not NaN

def test_datetime_replace():
    dt5 = datetime.datetime(2024, 3, 14, 9, 26, 53)
    dt7 = dt5.replace(2025)
    expect(2025, dt7.year())
    dt8 = dt5.replace(None, 12)
    expect(12, dt8.month())

def test_datetime_date_time():
    dt5 = datetime.datetime(2024, 3, 14, 9, 26, 53)
    d = dt5.date()
    expect(2024, d.year())
    expect(3, d.month())
    expect(14, d.day())

    t = dt5.time()
    expect(9, t.hour())
    expect(26, t.minute())
    expect(53, t.second())

def test_date_class():
    d1 = datetime.date(2024, 12, 25)
    expect(2024, d1.year())
    expect(12, d1.month())
    expect(25, d1.day())
    expect(2, d1.weekday())
    expect(3, d1.isoweekday())
    expect("2024-12-25", d1.isoformat())
    expect("2024/12/25", d1.strftime("%Y/%m/%d"))

    d2 = d1.replace(2025)
    expect(2025, d2.year())

    dtt = d1.timetuple()
    expect(2024, dtt[0])
    expect(0, dtt[3])  # Hour should be 0 for date

def test_time_class():
    t1 = datetime.time(10, 30, 45, 123456)
    expect(10, t1.hour())
    expect(30, t1.minute())
    expect(45, t1.second())
    expect(123456, t1.microsecond())
    expect("10:30:45.123456", t1.isoformat())
    expect("10:30:45", t1.strftime("%H:%M:%S"))

    t2 = datetime.time(12, 30)
    expect(0, t2.second())
    expect(0, t2.microsecond())
    expect("12:30:00", t2.isoformat())

    t3 = t1.replace(15)
    expect(15, t3.hour())

def test_timedelta_class():
    td1 = datetime.timedelta(5, 3600, 500000)
    expect(5, td1.days())
    expect(3600, td1.seconds())
    expect(500000, td1.microseconds())
    expect(435600.5, td1.total_seconds())

    td2 = datetime.timedelta(1, 0, 0, 0, 30, 2, 1)  # 1 day + 30 min + 2 hours + 1 week
    expect(True, td2.total_seconds() > 0)

    td3 = datetime.timedelta(0, 0, 0, 0, 0, 0, 2)  # 2 weeks
    expect(14, td3.days())

    td4 = datetime.timedelta(0, 90061)  # 90061 seconds = 1 day + 1 hour + 1 minute + 1 second
    expect(1, td4.days())
    expect(3661, td4.seconds())

def test_datetime_now_today():
    now = datetime.now()
    expect(True, now.year() >= 2024)
    expect(True, 1 <= now.month() <= 12)
    expect(True, 1 <= now.day() <= 31)

    today = datetime.today()
    expect(True, today.year() >= 2024)
    expect(True, 1 <= today.month() <= 12)

def test_datetime_fromtimestamp():
    dt_from_ts = datetime.fromtimestamp(1000000000)  # 2001-09-09
    expect(2001, dt_from_ts.year())
    expect(9, dt_from_ts.month())
    expect(True, dt_from_ts.day() >= 8)  # May vary by timezone

def test_datetime_fromisoformat():
    dt_parsed = datetime.fromisoformat("2024-07-04T12:00:00")
    expect(2024, dt_parsed.year())
    expect(7, dt_parsed.month())
    expect(4, dt_parsed.day())
    expect(12, dt_parsed.hour())

    dt_parsed2 = datetime.fromisoformat("2024-07-04 15:30:00")
    expect(15, dt_parsed2.hour())
    expect(30, dt_parsed2.minute())

    dt_parsed3 = datetime.fromisoformat("2024-07-04")
    expect(2024, dt_parsed3.year())
    expect(0, dt_parsed3.hour())

def test_datetime_combine():
    d_combine = datetime.date(2024, 1, 1)
    t_combine = datetime.time(12, 30, 45)
    dt_combined = datetime.combine(d_combine, t_combine)
    expect(2024, dt_combined.year())
    expect(12, dt_combined.hour())
    expect(30, dt_combined.minute())

test("datetime_constants", test_datetime_constants)
test("datetime_constructor", test_datetime_constructor)
test("datetime_weekday", test_datetime_weekday)
test("datetime_isoformat", test_datetime_isoformat)
test("datetime_strftime", test_datetime_strftime)
test("datetime_timetuple", test_datetime_timetuple)
test("datetime_toordinal", test_datetime_toordinal)
test("datetime_timestamp", test_datetime_timestamp)
test("datetime_replace", test_datetime_replace)
test("datetime_date_time", test_datetime_date_time)
test("date_class", test_date_class)
test("time_class", test_time_class)
test("timedelta_class", test_timedelta_class)
test("datetime_now_today", test_datetime_now_today)
test("datetime_fromtimestamp", test_datetime_fromtimestamp)
test("datetime_fromisoformat", test_datetime_fromisoformat)
test("datetime_combine", test_datetime_combine)

print("datetime module tests completed")
