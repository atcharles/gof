package g2util

import (
	"image/color"

	"github.com/mojocn/base64Captcha"
)

// Captcha ...
type Captcha struct {
	cap *base64Captcha.Captcha
}

// Cap ...
func (c *Captcha) Cap() *base64Captcha.Captcha { return c.cap }

// SetCap ...
func (c *Captcha) SetCap(cap *base64Captcha.Captcha) { c.cap = cap }

// Constructor ...
func (c *Captcha) Constructor() {
	//driver := base64Captcha.NewDriverDigit(70, 150, 4, 0.0, 1)
	driver := base64Captcha.NewDriverString(60, 150, 0,
		base64Captcha.OptionShowHollowLine,
		4, "1234567890",
		&color.RGBA{R: 0, G: 0, B: 0, A: 0},
		base64Captcha.DefaultEmbeddedFonts,
		[]string{"wqy-microhei.ttc"},
	)
	c.cap = base64Captcha.NewCaptcha(driver, base64Captcha.DefaultMemStore)
}

// Generate ...
func (c *Captcha) Generate() (id, b64s string, err error) {
	id, b64s, _, err = c.cap.Generate()
	return
}

// Verify ...
func (c *Captcha) Verify(id, answer string) (match bool) { return c.cap.Verify(id, answer, true) }
