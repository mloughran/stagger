package metric

import "testing"

func Test_AddingValuesToDists(t *testing.T) {
	d := NewDistFromValue(1)
	d.AddEntry(3)
	d.AddEntry(5)

	if d.Mean() != 3 {
		t.Fail()
	}
	if d.Min != 1 {
		t.Fail()
	}
	if d.Max != 5 {
		t.Fail()
	}
	if d.N != 3 {
		t.Fail()
	}
}

func Test_AddingDistsTogether(t *testing.T) {
	d := Dist{3, 1, 5, 9, 35}
	d2 := Dist{3, 1, 5, 9, 35}

	d.Add(&d2)

	if d.Mean() != 3 {
		t.Fail()
	}
	if d.Min != 1 {
		t.Fail()
	}
	if d.Max != 5 {
		t.Fail()
	}
	if d.N != 6 {
		t.Fail()
	}
}
