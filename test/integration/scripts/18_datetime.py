# Test: datetime module
# Tests datetime, date, time, timedelta classes

from test_framework import test, expect

import datetime

def test_datetime_constants():
    expect(datetime.MINYEAR).to_be(1)
    expect(datetime.MAXYEAR).to_be(9999)

def test_datetime_constructor():
    dt = datetime.datetime(2024, 6, 15, 14, 30, 45, 123456)
    expect(dt.year()).to_be(2024)
    expect(dt.month()).to_be(6)
    expect(dt.day()).to_be(15)
    expect(dt.hour()).to_be(14)
    expect(dt.minute()).to_be(30)
    expect(dt.second()).to_be(45)
    expect(dt.microsecond()).to_be(123456)

    dt2 = datetime.datetime(2024, 1, 1)
    expect(dt2.hour()).to_be(0)
    expect(dt2.minute()).to_be(0)
    expect(dt2.second()).to_be(0)

def test_datetime_weekday():
    dt3 = datetime.datetime(2024, 6, 17)  # Monday
    expect(dt3.weekday()).to_be(0)
    expect(dt3.isoweekday()).to_be(1)

    dt4 = datetime.datetime(2024, 6, 22)  # Saturday
    expect(dt4.weekday()).to_be(5)
    expect(dt4.isoweekday()).to_be(6)

def test_datetime_isoformat():
    dt5 = datetime.datetime(2024, 3, 14, 9, 26, 53)
    expect(dt5.isoformat()).to_be("2024-03-14T09:26:53")
    expect(dt5.isoformat(" ")).to_be("2024-03-14 09:26:53")

    dt6 = datetime.datetime(2024, 3, 14, 9, 26, 53, 500000)
    expect(dt6.isoformat()).to_be("2024-03-14T09:26:53.500000")

def test_datetime_strftime():
    dt5 = datetime.datetime(2024, 3, 14, 9, 26, 53)
    expect(dt5.strftime("%Y-%m-%d")).to_be("2024-03-14")
    expect(dt5.strftime("%H:%M:%S")).to_be("09:26:53")
    expect(dt5.strftime("%Y-%m-%d %H:%M:%S")).to_be("2024-03-14 09:26:53")
    expect(dt5.strftime("%A")).to_be("Thursday")
    expect(dt5.strftime("%B")).to_be("March")

def test_datetime_timetuple():
    dt5 = datetime.datetime(2024, 3, 14, 9, 26, 53)
    tt = dt5.timetuple()
    expect(tt[0]).to_be(2024)
    expect(tt[1]).to_be(3)
    expect(tt[2]).to_be(14)
    expect(tt[3]).to_be(9)
    expect(tt[4]).to_be(26)
    expect(tt[5]).to_be(53)

def test_datetime_toordinal():
    dt5 = datetime.datetime(2024, 3, 14, 9, 26, 53)
    # Check ordinal is a reasonable value (around 739000 for 2024)
    ordinal = dt5.toordinal()
    expect(738000 < ordinal < 740000).to_be(True)

def test_datetime_timestamp():
    dt_epoch = datetime.datetime(1970, 1, 1, 0, 0, 0)
    ts = dt_epoch.timestamp()
    expect(ts == ts).to_be(True)  # Check not NaN

def test_datetime_replace():
    dt5 = datetime.datetime(2024, 3, 14, 9, 26, 53)
    dt7 = dt5.replace(2025)
    expect(dt7.year()).to_be(2025)
    dt8 = dt5.replace(None, 12)
    expect(dt8.month()).to_be(12)

def test_datetime_date_time():
    dt5 = datetime.datetime(2024, 3, 14, 9, 26, 53)
    d = dt5.date()
    expect(d.year()).to_be(2024)
    expect(d.month()).to_be(3)
    expect(d.day()).to_be(14)

    t = dt5.time()
    expect(t.hour()).to_be(9)
    expect(t.minute()).to_be(26)
    expect(t.second()).to_be(53)

def test_date_class():
    d1 = datetime.date(2024, 12, 25)
    expect(d1.year()).to_be(2024)
    expect(d1.month()).to_be(12)
    expect(d1.day()).to_be(25)
    expect(d1.weekday()).to_be(2)
    expect(d1.isoweekday()).to_be(3)
    expect(d1.isoformat()).to_be("2024-12-25")
    expect(d1.strftime("%Y/%m/%d")).to_be("2024/12/25")

    d2 = d1.replace(2025)
    expect(d2.year()).to_be(2025)

    dtt = d1.timetuple()
    expect(dtt[0]).to_be(2024)
    expect(dtt[3]).to_be(0)  # Hour should be 0 for date

def test_time_class():
    t1 = datetime.time(10, 30, 45, 123456)
    expect(t1.hour()).to_be(10)
    expect(t1.minute()).to_be(30)
    expect(t1.second()).to_be(45)
    expect(t1.microsecond()).to_be(123456)
    expect(t1.isoformat()).to_be("10:30:45.123456")
    expect(t1.strftime("%H:%M:%S")).to_be("10:30:45")

    t2 = datetime.time(12, 30)
    expect(t2.second()).to_be(0)
    expect(t2.microsecond()).to_be(0)
    expect(t2.isoformat()).to_be("12:30:00")

    t3 = t1.replace(15)
    expect(t3.hour()).to_be(15)

def test_timedelta_class():
    td1 = datetime.timedelta(5, 3600, 500000)
    expect(td1.days()).to_be(5)
    expect(td1.seconds()).to_be(3600)
    expect(td1.microseconds()).to_be(500000)
    expect(td1.total_seconds()).to_be(435600.5)

    td2 = datetime.timedelta(1, 0, 0, 0, 30, 2, 1)  # 1 day + 30 min + 2 hours + 1 week
    expect(td2.total_seconds() > 0).to_be(True)

    td3 = datetime.timedelta(0, 0, 0, 0, 0, 0, 2)  # 2 weeks
    expect(td3.days()).to_be(14)

    td4 = datetime.timedelta(0, 90061)  # 90061 seconds = 1 day + 1 hour + 1 minute + 1 second
    expect(td4.days()).to_be(1)
    expect(td4.seconds()).to_be(3661)

def test_datetime_now_today():
    now = datetime.now()
    expect(now.year() >= 2024).to_be(True)
    expect(1 <= now.month() <= 12).to_be(True)
    expect(1 <= now.day() <= 31).to_be(True)

    today = datetime.today()
    expect(today.year() >= 2024).to_be(True)
    expect(1 <= today.month() <= 12).to_be(True)

def test_datetime_fromtimestamp():
    dt_from_ts = datetime.fromtimestamp(1000000000)  # 2001-09-09
    expect(dt_from_ts.year()).to_be(2001)
    expect(dt_from_ts.month()).to_be(9)
    expect(dt_from_ts.day() >= 8).to_be(True)  # May vary by timezone

def test_datetime_fromisoformat():
    dt_parsed = datetime.fromisoformat("2024-07-04T12:00:00")
    expect(dt_parsed.year()).to_be(2024)
    expect(dt_parsed.month()).to_be(7)
    expect(dt_parsed.day()).to_be(4)
    expect(dt_parsed.hour()).to_be(12)

    dt_parsed2 = datetime.fromisoformat("2024-07-04 15:30:00")
    expect(dt_parsed2.hour()).to_be(15)
    expect(dt_parsed2.minute()).to_be(30)

    dt_parsed3 = datetime.fromisoformat("2024-07-04")
    expect(dt_parsed3.year()).to_be(2024)
    expect(dt_parsed3.hour()).to_be(0)

def test_datetime_combine():
    d_combine = datetime.date(2024, 1, 1)
    t_combine = datetime.time(12, 30, 45)
    dt_combined = datetime.combine(d_combine, t_combine)
    expect(dt_combined.year()).to_be(2024)
    expect(dt_combined.hour()).to_be(12)
    expect(dt_combined.minute()).to_be(30)

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
