package g2util

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

/*func init() {
	//设定时区,shanghai
	_ = SetTimeZone()
}*/

const (
	//TimeZone ...
	TimeZone = "Asia/Shanghai"
	//Custom ...
	Custom = time.RFC3339
	//DateLayout ...
	DateLayout = "2006-01-02"
)

// TimeNowFunc ...
var TimeNowFunc = time.Now
var jsonTimeParseLayouts = []string{
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05.999999999",
	"2006-01-02 15:04:05.999999999",
	"2006-01-02T15:04:05.999999999Z07:00",
	"2006-01-02",
	"2006-01-02T15:04:05.999Z07:00",
	"2006-01-02T15:04:05.999999Z07:00",
	"2006-01-02 15:04:05.999",
	"2006-01-02 15:04:05.999999",
	time.RFC3339,
	time.RFC3339Nano,
	time.RFC1123,
	time.RFC1123Z,
	time.RFC822,
	time.RFC822Z,
	time.ANSIC,
}

// JSONTime ...
type JSONTime time.Time

// Add ...
func (p *JSONTime) Add(d time.Duration) *JSONTime { return NewJSONTimeOfTime(p.Time().Add(d)) }

// Convert2Time ...
func (p *JSONTime) Convert2Time() time.Time {
	if p == nil {
		return time.Time{}
	}
	return time.Time(*p).Local()
}

// Date ...返回一个日期0点的时间
func (p *JSONTime) Date() *JSONTime {
	if p == nil {
		return nil
	}
	y, m, d := p.Time().Date()
	dt := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	t := JSONTime(dt)
	return &t
}

// FromDB ...
func (p *JSONTime) FromDB(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if p == nil {
		return nil
	}
	*p = *parseInlocation2jsonTime(string(data))
	return nil
}

// GobDecode implements the gob.GobDecoder interface.
func (p *JSONTime) GobDecode(data []byte) error {
	if p == nil {
		return nil
	}
	s := p.Convert2Time()
	if err := (&s).UnmarshalBinary(data); err != nil {
		return err
	}
	*p = JSONTime(s)
	return nil
}

// GobEncode implements the gob.GobEncoder interface.
func (p *JSONTime) GobEncode() ([]byte, error) {
	if p == nil {
		return nil, nil
	}
	return p.Convert2Time().MarshalBinary()
}

// MarshalJSON ...
func (p *JSONTime) MarshalJSON() ([]byte, error) {
	if p == nil {
		return []byte(`""`), nil
	}
	if time.Time(*p).IsZero() {
		return []byte(`""`), nil
	}
	data := make([]byte, 0)
	data = append(data, '"')
	data = p.Convert2Time().AppendFormat(data, Custom)
	data = append(data, '"')
	return data, nil
}

// Scan the value of time.Time
func (p *JSONTime) Scan(v interface{}) error {
	if v == nil {
		return nil
	}
	value, ok := v.(time.Time)
	if ok {
		*p = JSONTime(value)
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}

// SetByTime ...
func (p *JSONTime) SetByTime(timeVal time.Time) {
	if p == nil {
		return
	}
	*p = JSONTime(timeVal)
}

// String ...
func (p *JSONTime) String() string {
	if p == nil {
		return ""
	}
	return p.Convert2Time().Format(Custom)
}

// StringFormat 转换为固定格式字符串
func (p *JSONTime) StringFormat(layout string) string {
	if p == nil {
		return ""
	}
	return p.Convert2Time().Format(layout)
}

// Time ...
func (p *JSONTime) Time() time.Time {
	if p == nil {
		return time.Time{}
	}
	return p.Convert2Time()
}

// ToDB ...
func (p *JSONTime) ToDB() (b []byte, err error) {
	if p == nil {
		return nil, nil
	}
	b = []byte(p.String())
	return
}

// UnmarshalJSON ...
func (p *JSONTime) UnmarshalJSON(data []byte) error {
	if p == nil {
		return nil
	}
	str := string(data)
	str = strings.Trim(str, `"`)
	*p = *parseInlocation2jsonTime(str)
	return nil
}

// Value insert timestamp into Mysql needs this function.
func (p *JSONTime) Value() (driver.Value, error) {
	if p == nil {
		return nil, nil
	}
	var zeroTime time.Time
	var ti = p.Convert2Time()
	if ti.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return ti, nil
}

// Must2JSONTimeAddr maybe panic
func Must2JSONTimeAddr(in string) *JSONTime {
	j, err := ToDatetime(in)
	if err != nil {
		panic(err)
	}
	return &j
}

// NewJSONTimeOfTime 时间转换为JSONTime
func NewJSONTimeOfTime(t time.Time) *JSONTime { return NewJSONTimePtr(t) }

func NewJSONTimePtr(t time.Time) *JSONTime {
	jt := JSONTime(t)
	return &jt
}

// Now 当前时间
func Now() *JSONTime { return NewJSONTimeOfTime(TimeNow()) }

// RetryDo 重试行为
func RetryDo(fn func() error, intervalSecond int64) error {
	return RetryDoTimes(10, intervalSecond, fn)
}

// RetryDoTimes ...
func RetryDoTimes(times, intervalSecond int64, fn func() error) (err error) {
	times = Clamp(times, 1, 20)
	intervalSecond = Clamp(intervalSecond, 1, 60)
	var a int64 = 1
	for {
		err = fn()
		if err == nil || a >= times {
			break
		}
		a++
		time.Sleep(time.Second * time.Duration(intervalSecond))
	}
	return err
}

// SetTimeZone ...Shanghai
func SetTimeZone() error {
	lc, err := time.LoadLocation(TimeZone)
	if err == nil {
		time.Local = lc
	}
	return err
}

// TimeExcWrap 包装执行时间
func TimeExcWrap(fn func()) time.Duration {
	n := TimeNow()
	fn()
	return time.Since(n)
}

func TimeNow() time.Time { return TimeNowFunc() }

func ToDatetime(in string) (JSONTime, error) {
	return *parseInlocation2jsonTime(in), nil
}

func ToDatetimeOK(in string) JSONTime {
	return *parseInlocation2jsonTime(in)
}

// Today ...今日日期
func Today() *JSONTime { return NewJSONTimeOfTime(TimeNow()).Date() }

func TodayDate() string { return TimeNow().Format(DateLayout) }

func parseInlocation2jsonTime(str string) *JSONTime {
	str = strings.TrimSpace(str)
	if str == "" || strings.EqualFold(str, "null") {
		return NewJSONTimePtr(time.Time{})
	}
	if t, e := time.Parse(time.RFC3339Nano, str); e == nil {
		return NewJSONTimePtr(t)
	}
	for _, layout := range jsonTimeParseLayouts {
		t, e := time.ParseInLocation(layout, str, time.Local)
		if e == nil {
			return NewJSONTimePtr(t)
		}
	}
	return NewJSONTimePtr(time.Time{})
}
