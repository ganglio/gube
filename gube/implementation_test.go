package gube

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestWithinDomain(t *testing.T) {
	testData := []struct {
		name string
		r    float64
		g    float64
		b    float64
		lut  *GubeImpl
		want bool
	}{
		{
			name: "All values ok",
			lut: &GubeImpl{
				domainMax: RGB{1.0, 1.0, 1.0},
			},
			want: true,
		},
		{
			name: "R is outside of range (too high)",
			r:    8.0,
			lut: &GubeImpl{
				domainMax: RGB{1.0, 1.0, 1.0},
			},
		},
		{
			name: "G is outside of range (too high)",
			g:    3.0,
			lut: &GubeImpl{
				domainMax: RGB{1.0, 1.0, 1.0},
			},
		},
		{
			name: "B is outside of range (too high)",
			g:    100.0,
			lut: &GubeImpl{
				domainMax: RGB{1.0, 1.0, 1.0},
			},
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {

			got := test.lut.withinDomain(test.r, test.g, test.b)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("withinDomain() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestLookUp1D(t *testing.T) {
	testData := []struct {
		name    string
		r       float64
		g       float64
		b       float64
		lut     *GubeImpl
		want    RGB
		wantErr bool
	}{
		{
			name: "Fetch some values from the table",
			r:    1.0,
			g:    0.5,
			b:    0,
			lut: &GubeImpl{
				domainMax: RGB{1.0, 1.0, 1.0},
				tableSize: 10,
				tableData1D: &[]RGB{
					{0, 0, 1},
					{0.1, 0.05, .9},
					{0.2, 0.1, .8},
					{0.3, 0.15, .7},
					{0.4, 0.2, .6},
					{0.5, 0.25, .5},
					{0.6, 0.3, .4},
					{0.7, 0.35, .3},
					{0.9, 0.4, .2},
					{1, 0.5, .1},
				},
			},
			want: RGB{1, 0.025, 1},
		},
		{
			name: "Interpolate",
			r:    0.5,
			g:    0.5,
			b:    0.5,
			lut: &GubeImpl{
				domainMax: RGB{1.0, 1.0, 1.0},
				tableSize: 2,
				tableData1D: &[]RGB{
					{0, 0, 0},
					{1, 0.5, .1},
				},
			},
			want: RGB{0.5, 0.25, 0.05},
		},
		{
			name: "Outside of domain values",
			r:    0.5,
			g:    4.5,
			b:    0.5,
			lut: &GubeImpl{
				domainMax: RGB{1.0, 1.0, 1.0},
				tableSize: 2,
				tableData1D: &[]RGB{
					{0, 0, 0},
					{1, 0.5, .1},
				},
			},
			wantErr: true,
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {

			got, gotErr := test.lut.lookUp1D(test.r, test.g, test.b)
			if (gotErr != nil) != test.wantErr {
				t.Errorf("Test: %q :  Got error %v, wanted err=%v", test.name, gotErr, test.wantErr)
			}
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("lookUp1D() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestString(t *testing.T) {
	testData := []struct {
		name    string
		lut     *GubeImpl
		want    string
		wantErr bool
	}{
		{
			name: "string_1",
			lut: &GubeImpl{
				name:      "string_1",
				domainMax: RGB{1.0, 1.0, 1.0},
				tableType: LUT_1D,
				tableSize: 2,
				tableData1D: &[]RGB{
					{0, 0, 0},
					{1, 1, 1},
				},
			},
			want: `TITLE: "string_1"

LUT_1D_SIZE: 2

DOMAIN_MIN: 0.000000 0.000000 0.000000
DOMAIN_MAX: 1.000000 1.000000 1.000000

0.000000 0.000000 0.000000
1.000000 1.000000 1.000000`,
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {

			got := test.lut.String()
			if test.want != got {
				t.Errorf("String() mismatch (want:\"%s\" got:\"%s\")", test.want, got)
			}
		})
	}
}

func TestNewFromString(t *testing.T) {
	testData := []struct {
		name    string
		input   string
		want    *GubeImpl
		wantErr bool
	}{
		{
			name: "parseFromString_1",
			input: `TITLE: "string_1"

LUT_3D_SIZE: 2

DOMAIN_MIN: 0.000000 0.000000 0.000000
DOMAIN_MAX: 1.000000 1.000000 1.000000

0.000000 0.000000 0.000000
0.000000 0.000000 1.000000
0.000000 1.000000 0.000000
0.000000 1.000000 1.000000
1.000000 0.000000 0.000000
1.000000 0.000000 1.000000
1.000000 1.000000 0.000000
1.000000 1.000000 1.000000`,
			want: &GubeImpl{
				name:      "string_1",
				domainMax: RGB{1.0, 1.0, 1.0},
				tableType: LUT_3D,
				tableSize: 2,
				tableData3D: &[][][]RGB{
					{
						{
							{0, 0, 0},
							{1, 0, 0},
						},
						{
							{0, 1, 0},
							{1, 1, 0},
						},
					},
					{
						{
							{0, 0, 1},
							{1, 0, 1},
						},
						{
							{0, 1, 1},
							{1, 1, 1},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {

			got, err := NewFromString(test.input)
			if err != nil && test.wantErr == false {
				t.Errorf("NewFromString Returned error: %v", err)
			}
			if test.want.String() != got.String() {
				t.Errorf("String() mismatch (want:\"%s\" got:\"%s\")", test.want.String(), got.String())
			}
		})
	}
}

func TestResample(t *testing.T) {
	testData := []struct {
		name     string
		input    string
		resample int
		want     string
		wantErr  bool
	}{
		{
			name: "resample 3->2",
			input: `TITLE: "resample 3->2"

LUT_3D_SIZE: 3

DOMAIN_MIN: 0.000000 0.000000 0.000000
DOMAIN_MAX: 1.000000 1.000000 1.000000

0.000000 0.000000 0.000000
0.000000 0.000000 0.500000
0.000000 0.000000 1.000000
0.000000 0.500000 0.000000
0.000000 0.500000 0.500000
0.000000 0.500000 1.000000
0.000000 1.000000 0.000000
0.000000 1.000000 0.500000
0.000000 1.000000 1.000000
0.500000 0.000000 0.000000
0.500000 0.000000 0.500000
0.500000 0.000000 1.000000
0.500000 0.500000 0.000000
0.500000 0.500000 0.500000
0.500000 0.500000 1.000000
0.500000 1.000000 0.000000
0.500000 1.000000 0.500000
0.500000 1.000000 1.000000
1.000000 0.000000 0.000000
1.000000 0.000000 0.500000
1.000000 0.000000 1.000000
1.000000 0.500000 0.000000
1.000000 0.500000 0.500000
1.000000 0.500000 1.000000
1.000000 1.000000 0.000000
1.000000 1.000000 0.500000
1.000000 1.000000 1.000000`,
			resample: 2,
			want: `TITLE: "resample 3->2"

LUT_3D_SIZE: 2

DOMAIN_MIN: 0.000000 0.000000 0.000000
DOMAIN_MAX: 1.000000 1.000000 1.000000

0.000000 0.000000 0.000000
0.000000 0.000000 1.000000
0.000000 1.000000 0.000000
0.000000 1.000000 1.000000
1.000000 0.000000 0.000000
1.000000 0.000000 1.000000
1.000000 1.000000 0.000000
1.000000 1.000000 1.000000`,
			wantErr: false,
		},
		{
			name: "resample 3->4",
			input: `TITLE: "resample 3->4"

LUT_3D_SIZE: 3

DOMAIN_MIN: 0.000000 0.000000 0.000000
DOMAIN_MAX: 1.000000 1.000000 1.000000

0.000000 0.000000 0.000000
0.000000 0.000000 0.500000
0.000000 0.000000 1.000000
0.000000 0.500000 0.000000
0.000000 0.500000 0.500000
0.000000 0.500000 1.000000
0.000000 1.000000 0.000000
0.000000 1.000000 0.500000
0.000000 1.000000 1.000000
0.500000 0.000000 0.000000
0.500000 0.000000 0.500000
0.500000 0.000000 1.000000
0.500000 0.500000 0.000000
0.500000 0.500000 0.500000
0.500000 0.500000 1.000000
0.500000 1.000000 0.000000
0.500000 1.000000 0.500000
0.500000 1.000000 1.000000
1.000000 0.000000 0.000000
1.000000 0.000000 0.500000
1.000000 0.000000 1.000000
1.000000 0.500000 0.000000
1.000000 0.500000 0.500000
1.000000 0.500000 1.000000
1.000000 1.000000 0.000000
1.000000 1.000000 0.500000
1.000000 1.000000 1.000000`,
			resample: 4,
			want: `TITLE: "resample 3->4"

LUT_3D_SIZE: 4

DOMAIN_MIN: 0.000000 0.000000 0.000000
DOMAIN_MAX: 1.000000 1.000000 1.000000

0.000000 0.000000 0.000000
0.000000 0.000000 0.333333
0.000000 0.000000 0.666667
0.000000 0.000000 1.000000
0.000000 0.333333 0.000000
0.000000 0.333333 0.333333
0.000000 0.333333 0.666667
0.000000 0.333333 1.000000
0.000000 0.666667 0.000000
0.000000 0.666667 0.333333
0.000000 0.666667 0.666667
0.000000 0.666667 1.000000
0.000000 1.000000 0.000000
0.000000 1.000000 0.333333
0.000000 1.000000 0.666667
0.000000 1.000000 1.000000
0.333333 0.000000 0.000000
0.333333 0.000000 0.333333
0.333333 0.000000 0.666667
0.333333 0.000000 1.000000
0.333333 0.333333 0.000000
0.333333 0.333333 0.333333
0.333333 0.333333 0.666667
0.333333 0.333333 1.000000
0.333333 0.666667 0.000000
0.333333 0.666667 0.333333
0.333333 0.666667 0.666667
0.333333 0.666667 1.000000
0.333333 1.000000 0.000000
0.333333 1.000000 0.333333
0.333333 1.000000 0.666667
0.333333 1.000000 1.000000
0.666667 0.000000 0.000000
0.666667 0.000000 0.333333
0.666667 0.000000 0.666667
0.666667 0.000000 1.000000
0.666667 0.333333 0.000000
0.666667 0.333333 0.333333
0.666667 0.333333 0.666667
0.666667 0.333333 1.000000
0.666667 0.666667 0.000000
0.666667 0.666667 0.333333
0.666667 0.666667 0.666667
0.666667 0.666667 1.000000
0.666667 1.000000 0.000000
0.666667 1.000000 0.333333
0.666667 1.000000 0.666667
0.666667 1.000000 1.000000
1.000000 0.000000 0.000000
1.000000 0.000000 0.333333
1.000000 0.000000 0.666667
1.000000 0.000000 1.000000
1.000000 0.333333 0.000000
1.000000 0.333333 0.333333
1.000000 0.333333 0.666667
1.000000 0.333333 1.000000
1.000000 0.666667 0.000000
1.000000 0.666667 0.333333
1.000000 0.666667 0.666667
1.000000 0.666667 1.000000
1.000000 1.000000 0.000000
1.000000 1.000000 0.333333
1.000000 1.000000 0.666667
1.000000 1.000000 1.000000`,
			wantErr: false,
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {

			got, err := NewFromString(test.input)
			if err != nil && test.wantErr == false {
				t.Errorf("NewFromString Returned error: %v", err)
			}
			res := got.Resample(test.resample)
			if test.want != res.String() {
				t.Errorf("String() mismatch (want:\"%s\" got:\"%s\")", test.want, got.String())
			}
		})
	}
}

func TestResampleActual(t *testing.T) {
	lut8, err := os.Open("./testdata/Test_LUT8.cube")
	if err != nil {
		t.Errorf("Error opening file: %v", err)
	}
	cube8, err := NewFromReader(lut8)
	if err != nil {
		t.Errorf("Error parsing file: %v", err)
	}

	lut32, err := os.Open("./testdata/Test_LUT32.cube")
	if err != nil {
		t.Errorf("Error opening file: %v", err)
	}
	cube32, err := NewFromReader(lut32)
	if err != nil {
		t.Errorf("Error parsing file: %v", err)
	}

	testData := []struct {
		name     string
		input    *GubeImpl
		resample int
		want     *GubeImpl
	}{
		{
			name:     "resample 32->8",
			input:    cube32,
			resample: 8,
			want:     cube8,
		},
		{
			name:     "resample 8->32",
			input:    cube8,
			resample: 32,
			want:     cube32,
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {

			got := test.input.Resample(test.resample)
			if diff := test.want.Diff(got); diff > 0.01 {
				t.Errorf("Resample() mismatch (-want +got):\n%f", diff)
			}
		})
	}
}
