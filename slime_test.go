package main

import "testing"

func TestLerpGivenZero(t *testing.T) {
	a := lerp(0, 255, 0)
	b := lerp(255, 0, 0)

	if a != 0 || b != 0 {
		t.Log("values should be 0")
		t.Fail()
	}
}

func TestLerp(t *testing.T) {
	cases := []struct {
		a, b     uint8
		t        float32
		expected uint8
	}{
		{220, 230, 0, 220},
		{220, 230, 1, 230},
		{220, 230, 0.5, 225},
		{220, 240, 0.5, 230},
		{255, 255, 1, 255},
	}

	for _, c := range cases {
		val := lerp(c.a, c.b, c.t)

		if val != c.expected {
			t.Log(
				"Expected value was not", c.expected,
				"actual:", val)
			t.Fail()
		}
	}
}

func TestAverage(t *testing.T) {
	frame := make([]uint8, 9)

	frame[4] = 255

	for i := range frame {
		avg := getAverage(frame, i, 3, 3)

		if avg != 28 {
			t.Log("Average was not expected 28, actual:", avg)
			t.Fail()
		}
	}
}
