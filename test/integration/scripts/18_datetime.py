# Test: datetime module
# Tests datetime, date, time, timedelta classes

results = {}

import datetime

# =====================================
# Constants
# =====================================
results["datetime_minyear"] = datetime.MINYEAR
results["datetime_maxyear"] = datetime.MAXYEAR

# =====================================
# datetime class - constructor
# =====================================
dt = datetime.datetime(2024, 6, 15, 14, 30, 45, 123456)
results["datetime_year"] = dt.year()
results["datetime_month"] = dt.month()
results["datetime_day"] = dt.day()
results["datetime_hour"] = dt.hour()
results["datetime_minute"] = dt.minute()
results["datetime_second"] = dt.second()
results["datetime_microsecond"] = dt.microsecond()

# datetime with defaults
dt2 = datetime.datetime(2024, 1, 1)
results["datetime_default_hour"] = dt2.hour()
results["datetime_default_minute"] = dt2.minute()
results["datetime_default_second"] = dt2.second()

# =====================================
# datetime methods
# =====================================

# weekday (Monday=0)
dt3 = datetime.datetime(2024, 6, 17)  # Monday
results["datetime_weekday_monday"] = dt3.weekday()

dt4 = datetime.datetime(2024, 6, 22)  # Saturday
results["datetime_weekday_saturday"] = dt4.weekday()

# isoweekday (Monday=1)
results["datetime_isoweekday_monday"] = dt3.isoweekday()
results["datetime_isoweekday_saturday"] = dt4.isoweekday()

# isoformat
dt5 = datetime.datetime(2024, 3, 14, 9, 26, 53)
results["datetime_isoformat"] = dt5.isoformat()
results["datetime_isoformat_space"] = dt5.isoformat(" ")

# isoformat with microseconds
dt6 = datetime.datetime(2024, 3, 14, 9, 26, 53, 500000)
results["datetime_isoformat_micro"] = dt6.isoformat()

# strftime
results["datetime_strftime_ymd"] = dt5.strftime("%Y-%m-%d")
results["datetime_strftime_hms"] = dt5.strftime("%H:%M:%S")
results["datetime_strftime_full"] = dt5.strftime("%Y-%m-%d %H:%M:%S")
results["datetime_strftime_weekday"] = dt5.strftime("%A")
results["datetime_strftime_month"] = dt5.strftime("%B")

# ctime
results["datetime_ctime"] = dt5.ctime()

# timetuple
tt = dt5.timetuple()
results["datetime_timetuple_year"] = tt[0]
results["datetime_timetuple_month"] = tt[1]
results["datetime_timetuple_day"] = tt[2]
results["datetime_timetuple_hour"] = tt[3]
results["datetime_timetuple_minute"] = tt[4]
results["datetime_timetuple_second"] = tt[5]

# toordinal
results["datetime_toordinal"] = dt5.toordinal()

# timestamp (for a known date)
dt_epoch = datetime.datetime(1970, 1, 1, 0, 0, 0)
# Note: timestamp depends on timezone, just check it's a number
ts = dt_epoch.timestamp()
results["datetime_timestamp_is_number"] = ts == ts  # True if not NaN

# replace
dt7 = dt5.replace(2025)
results["datetime_replace_year"] = dt7.year()
dt8 = dt5.replace(None, 12)
results["datetime_replace_month"] = dt8.month()

# date() method - extracts date portion
d = dt5.date()
results["datetime_date_year"] = d.year()
results["datetime_date_month"] = d.month()
results["datetime_date_day"] = d.day()

# time() method - extracts time portion
t = dt5.time()
results["datetime_time_hour"] = t.hour()
results["datetime_time_minute"] = t.minute()
results["datetime_time_second"] = t.second()

# =====================================
# date class - constructor
# =====================================
d1 = datetime.date(2024, 12, 25)
results["date_year"] = d1.year()
results["date_month"] = d1.month()
results["date_day"] = d1.day()

# date methods
results["date_weekday"] = d1.weekday()
results["date_isoweekday"] = d1.isoweekday()
results["date_isoformat"] = d1.isoformat()
results["date_toordinal"] = d1.toordinal()

# date strftime
results["date_strftime"] = d1.strftime("%Y/%m/%d")

# date replace
d2 = d1.replace(2025)
results["date_replace_year"] = d2.year()

# date timetuple
dtt = d1.timetuple()
results["date_timetuple_year"] = dtt[0]
results["date_timetuple_hour"] = dtt[3]  # Should be 0 for date

# =====================================
# time class - constructor
# =====================================
t1 = datetime.time(10, 30, 45, 123456)
results["time_hour"] = t1.hour()
results["time_minute"] = t1.minute()
results["time_second"] = t1.second()
results["time_microsecond"] = t1.microsecond()

# time with defaults
t2 = datetime.time(12, 30)
results["time_default_second"] = t2.second()
results["time_default_microsecond"] = t2.microsecond()

# time methods
results["time_isoformat"] = t1.isoformat()
results["time_isoformat_no_micro"] = t2.isoformat()

# time strftime
results["time_strftime"] = t1.strftime("%H:%M:%S")

# time replace
t3 = t1.replace(15)
results["time_replace_hour"] = t3.hour()

# =====================================
# timedelta class - constructor
# =====================================
td1 = datetime.timedelta(5, 3600, 500000)
results["timedelta_days"] = td1.days()
results["timedelta_seconds"] = td1.seconds()
results["timedelta_microseconds"] = td1.microseconds()

# total_seconds
results["timedelta_total_seconds"] = td1.total_seconds()

# timedelta with various units
# timedelta(days=0, seconds=0, microseconds=0, milliseconds=0, minutes=0, hours=0, weeks=0)
td2 = datetime.timedelta(1, 0, 0, 0, 30, 2, 1)  # 1 day + 30 min + 2 hours + 1 week
results["timedelta_complex_total"] = td2.total_seconds()

# timedelta with just weeks
td3 = datetime.timedelta(0, 0, 0, 0, 0, 0, 2)  # 2 weeks
results["timedelta_weeks_days"] = td3.days()

# timedelta normalization
td4 = datetime.timedelta(0, 90061)  # 90061 seconds = 1 day + 1 hour + 1 minute + 1 second
results["timedelta_normalize_days"] = td4.days()
results["timedelta_normalize_seconds"] = td4.seconds()

# =====================================
# Module-level functions
# =====================================

# now() - just verify it returns something with correct attributes
now = datetime.now()
results["now_has_year"] = now.year() >= 2024
results["now_has_month"] = 1 <= now.month() <= 12
results["now_has_day"] = 1 <= now.day() <= 31

# today() - same
today = datetime.today()
results["today_has_year"] = today.year() >= 2024
results["today_has_month"] = 1 <= today.month() <= 12

# fromtimestamp
dt_from_ts = datetime.fromtimestamp(1000000000)  # 2001-09-09
results["fromtimestamp_year"] = dt_from_ts.year()
results["fromtimestamp_month"] = dt_from_ts.month()
results["fromtimestamp_day"] = dt_from_ts.day()

# fromisoformat
dt_parsed = datetime.fromisoformat("2024-07-04T12:00:00")
results["fromisoformat_year"] = dt_parsed.year()
results["fromisoformat_month"] = dt_parsed.month()
results["fromisoformat_day"] = dt_parsed.day()
results["fromisoformat_hour"] = dt_parsed.hour()

# fromisoformat with space separator
dt_parsed2 = datetime.fromisoformat("2024-07-04 15:30:00")
results["fromisoformat_space_hour"] = dt_parsed2.hour()
results["fromisoformat_space_minute"] = dt_parsed2.minute()

# fromisoformat date only
dt_parsed3 = datetime.fromisoformat("2024-07-04")
results["fromisoformat_dateonly_year"] = dt_parsed3.year()
results["fromisoformat_dateonly_hour"] = dt_parsed3.hour()  # Should be 0

# combine
d_combine = datetime.date(2024, 1, 1)
t_combine = datetime.time(12, 30, 45)
dt_combined = datetime.combine(d_combine, t_combine)
results["combine_year"] = dt_combined.year()
results["combine_hour"] = dt_combined.hour()
results["combine_minute"] = dt_combined.minute()

print("datetime module tests completed")
