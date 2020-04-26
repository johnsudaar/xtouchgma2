package xtouch

import (
	"encoding/hex"
	"fmt"

	"github.com/pkg/errors"
)

var colorMap map[ScribbleColor][]float64 = map[ScribbleColor][]float64{
	ScribbleColorRed:    colorToHSV([]byte{255, 0, 0}),
	ScribbleColorGreen:  colorToHSV([]byte{0, 255, 0}),
	ScribbleColorYellow: colorToHSV([]byte{255, 255, 0}),
	ScribbleColorBlue:   colorToHSV([]byte{0, 0, 255}),
	ScribbleColorPink:   colorToHSV([]byte{255, 0, 255}),
	ScribbleColorCyan:   colorToHSV([]byte{0, 255, 255}),
}

func colorToRGB(color string) ([]byte, error) {
	if len(color) != 7 {
		return nil, fmt.Errorf("Invalid color %s", color)
	}

	colors, err := hex.DecodeString(color[1:7])
	if err != nil {
		return nil, errors.Wrapf(err, "fail to decode color %s", color)
	}

	return colors, nil
}

func max(in ...float64) float64 {
	res := in[0]
	for _, v := range in {
		if v > res {
			res = v
		}
	}
	return res
}

func min(in ...float64) float64 {
	res := in[0]
	for _, v := range in {
		if v < res {
			res = v
		}
	}
	return res
}

func colorToHSV(color []byte) []float64 {
	r := float64(color[0])
	g := float64(color[1])
	b := float64(color[2])

	rp := r / 255
	gp := g / 255
	bp := b / 255

	cmax := max(rp, gp, bp)
	cmin := min(rp, gp, bp)

	delta := cmax - cmin

	var h float64 = 0
	var s float64 = 0
	var v float64 = cmax

	if cmax == cmin {
		h = 0
		s = 0
	} else {
		s = 1 - (cmin * cmax)

		if cmax == rp {
			h = 60*((gp-bp)/delta) + 360
		}
		if cmax == gp {
			h = 60*((bp-rp)/delta) + 120
		}

		if cmax == bp {
			h = 60*((rp-gp)/delta) + 240
		}

		for h >= 360 {
			h -= 360
		}
	}
	return []float64{h, s, v}
}

func angleDistance(a, b float64) float64 {
	// I feel dumb i'm sure there is a better way...

	deltaA := a - b
	if deltaA < 0 {
		deltaA += 360
	}

	deltaB := b - a
	if deltaB < 0 {
		deltaB += 360
	}

	if deltaA > deltaB {
		return deltaB
	}
	return deltaA
}

func ClosestScribbleColor(color string) (ScribbleColor, error) {
	if len(color) == 0 {
		return ScribbleColorBlack, nil
	}

	rgb, err := colorToRGB(color)
	if err != nil {
		return ScribbleColorBlack, errors.Wrap(err, "fail to get rvg value")
	}

	hsv := colorToHSV(rgb)
	h, s, v := hsv[0], hsv[1], hsv[2]
	if s < 0.2 {
		if v > 0.5 {
			return ScribbleColorWhite, nil
		} else {
			return ScribbleColorBlack, nil
		}
	}

	var closestHDistance float64 = 720
	closestColor := ScribbleColorBlack

	for color, colorHSV := range colorMap {
		distance := angleDistance(colorHSV[0], h)
		if distance < closestHDistance {
			closestColor = color
			closestHDistance = distance
		}
	}

	return closestColor, nil
}
