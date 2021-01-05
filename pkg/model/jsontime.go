package model

import "time"

const (
	DateFormat  = "2006-01-02"
	TimeFormat  = "2006-01-02 15:04:05"
	TimeTFormat = "2006-01-02T15:04:05"
)

type JsonTime time.Time

func Now() JsonTime {
	return JsonTime(time.Now())
}

func (t *JsonTime) UnmarshalJSON(data []byte) (err error) {
	now, err := time.ParseInLocation(`"`+TimeTFormat+`"`, string(data), time.Local)
	*t = JsonTime(now)
	return
}

func (t JsonTime) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(TimeTFormat)+2)
	b = append(b, '"')
	b = time.Time(t).AppendFormat(b, TimeTFormat)
	b = append(b, '"')
	return b, nil
}

func (t JsonTime) String() string {
	return time.Time(t).Format(TimeTFormat)
}
