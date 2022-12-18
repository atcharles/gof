package g2util

import (
	"github.com/robfig/cron/v3"
)

// G2cron ...
type G2cron struct {
	Logger LevelLogger `inject:""`

	cron *cron.Cron
}

// Cron ...
func (c *G2cron) Cron() *cron.Cron { return c.cron }

// Constructor ...
func (c *G2cron) Constructor() { c.cron = cron.New(cron.WithSeconds()); c.OnProcessStart() }

// AddTask ...
func (c *G2cron) AddTask(spec string, fn func()) {
	if _, err := c.cron.AddFunc(spec, fn); err != nil {
		c.Logger.Debugf("[Add] %s", err.Error())
		return
	}
}

// execFunc ...
/*func (c *G2cron) execFunc(shellStr string) {
	out, e := StdExec(shellStr).CombinedOutput()
	if len(out) > 0 {
		c.Logger.Debugf("%s", out)
	}
	if e != nil {
		c.Logger.Warnf("%s", e.Error())
		return
	}
	c.Logger.Debugf("[Run] %s", shellStr)
}*/

// OnProcessStart ...启动定时任务
func (c *G2cron) OnProcessStart() {
	c.cron.Start()

	/*c.AddTask("1 0 * * * *", func() {
		//同步时间
		//shellCmd := "ntpdate -u ntp1.aliyun.com && hwclock --systohc"
		//shellCmd := `date -s "$(curl -s --head https://www.baidu.com | grep ^Date: | sed 's/Date: //g')" && hwclock --systohc`
		shellCmd := "ntpdate -u ntp1.aliyun.com"
		c.execFunc(shellCmd)
	})*/
}
