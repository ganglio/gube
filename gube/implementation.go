package gube

import (
	"fmt"
	"io"
	"math"
	"strings"
)

// Ensure interface compliance.
var _ Gube = (*GubeImpl)(nil)

type GubeImpl struct {
	name        string
	tableType   int64
	tableSize   int64
	tableData1D *[]RGB
	tableData3D *[][][]RGB
	domainMin   RGB
	domainMax   RGB
}

// NewFromReader returns a Gube instance created from the reader data.
func NewFromReader(r io.Reader) (*GubeImpl, error) {
	return parseFromReader(r)
}

func NewFromString(s string) (*GubeImpl, error) {
	return parseFromReader(strings.NewReader(s))
}

func (gi *GubeImpl) Diff(other *GubeImpl) float64 {
	if gi.TableType() != other.TableType() || gi.TableSize() != other.TableSize() {
		return -1.0
	}

	if gi.TableType() == LUT_1D {
		td1 := *gi.TableData1D()
		td2 := *other.TableData1D()
		if len(td1) != len(td2) {
			return -1.0
		}
		diffs := []float64{}
		for i := 0; i < len(td1); i++ {
			diff := math.Abs(td1[i][0]-td2[i][0]) + math.Abs(td1[i][1]-td2[i][1]) + math.Abs(td1[i][2]-td2[i][2])
			diffs = append(diffs, diff)
		}
		sum := 0.0
		for _, diff := range diffs {
			sum += diff * diff
		}
		return math.Sqrt(sum) / float64(len(td1))
	} else {
		td1 := *gi.TableData3D()
		td2 := *other.TableData3D()
		if len(td1) != len(td2) {
			return 1.0
		}
		diffs := []float64{}
		sum := 0.0
		for i := 0; i < len(td1); i++ {
			if len(td1[i]) != len(td2[i]) {
				return -11.0
			}
			for j := 0; j < len(td1[i]); j++ {
				if len(td1[i][j]) != len(td2[i][j]) {
					return -11.0
				}
				for k := 0; k < len(td1[i][j]); k++ {
					diff := math.Abs(td1[i][j][k][0]-td2[i][j][k][0]) + math.Abs(td1[i][j][k][1]-td2[i][j][k][1]) + math.Abs(td1[i][j][k][2]-td2[i][j][k][2])
					diffs = append(diffs, diff)
				}
				for _, diff := range diffs {
					sum += diff * diff
				}
			}
		}
		return math.Sqrt(sum) / (float64(len(td1) * len(td1) * len(td1)))
	}
}

func (gi *GubeImpl) LookUp(r float64, g float64, b float64) (RGB, error) {
	switch gi.tableType {
	case LUT_1D:
		return gi.lookUp1D(r, g, b)
	case LUT_3D:
		return gi.lookUp3D(r, g, b)
	default:
		return RGB{}, ErrInvalidLutType
	}
}

func (gi *GubeImpl) Name() string {
	return gi.name
}

func (gi *GubeImpl) TableType() int64 {
	return gi.tableType
}

func (gi *GubeImpl) TableSize() int64 {
	return gi.tableSize
}

func (gi *GubeImpl) TableData1D() *[]RGB {
	return gi.tableData1D
}

func (gi *GubeImpl) TableData3D() *[][][]RGB {
	return gi.tableData3D
}

func (gi *GubeImpl) Domain() (RGB, RGB) {
	return gi.domainMin, gi.domainMax
}

func (gi *GubeImpl) String() string {
	out := []string{}

	out = append(out, fmt.Sprintf("TITLE: \"%s\"", gi.Name()))
	out = append(out, "")
	if gi.TableType() == LUT_1D {
		out = append(out, fmt.Sprintf("LUT_1D_SIZE: %d", gi.TableSize()))
	} else {
		out = append(out, fmt.Sprintf("LUT_3D_SIZE: %d", gi.TableSize()))
	}
	out = append(out, "")
	d_min, d_max := gi.Domain()
	out = append(out, fmt.Sprintf("DOMAIN_MIN: %f %f %f", d_min[0], d_min[1], d_min[2]))
	out = append(out, fmt.Sprintf("DOMAIN_MAX: %f %f %f", d_max[0], d_max[1], d_max[2]))
	out = append(out, "")

	if gi.TableType() == LUT_1D {
		td := gi.TableData1D()
		for _, rgb := range *td {
			out = append(out, fmt.Sprintf("%f %f %f", rgb[0], rgb[1], rgb[2]))
		}
	} else {
		td := gi.TableData3D()
		for k := int64(0); k < gi.TableSize(); k++ {
			for j := int64(0); j < gi.TableSize(); j++ {
				for i := int64(0); i < gi.TableSize(); i++ {
					out = append(out, fmt.Sprintf("%f %f %f", (*td)[i][j][k][0], (*td)[i][j][k][1], (*td)[i][j][k][2]))
				}
			}
		}
	}

	return strings.Join(out, "\n")
}

func (c *GubeImpl) Resample(ts int) *GubeImpl {
	d_min, d_max := c.Domain()
	if c.TableType() == LUT_1D {
		return nil
	} else {
		td := make([][][]RGB, ts)
		for i := 0; i < ts; i++ {
			td[i] = make([][]RGB, ts)
			for j := 0; j < ts; j++ {
				td[i][j] = make([]RGB, ts)
			}
		}
		for i := 0; i < ts; i++ {
			for j := 0; j < ts; j++ {
				for k := 0; k < ts; k++ {
					r := (d_max[0]-d_min[0])*float64(i)/float64(ts-1) + d_min[0]
					g := (d_max[1]-d_min[1])*float64(j)/float64(ts-1) + d_min[1]
					b := (d_max[2]-d_min[2])*float64(k)/float64(ts-1) + d_min[2]
					rgb, err := c.LookUp(r, g, b)
					if err != nil {
						return nil
					}
					td[i][j][k] = rgb
				}
			}
		}
		return &GubeImpl{
			name:        c.Name(),
			tableType:   c.TableType(),
			domainMin:   d_min,
			domainMax:   d_max,
			tableSize:   int64(ts),
			tableData3D: &td,
		}
	}
}

func (gi *GubeImpl) lookUp3D(r, g, b float64) (RGB, error) {
	var res RGB

	if !gi.withinDomain(r, g, b) {
		return res, ErrOutsideOfDomain
	}

	return gi.trilinear(r*float64(gi.tableSize-1), g*float64(gi.tableSize-1), b*float64(gi.tableSize-1)), nil
}

func (gi *GubeImpl) lookUp1D(r, g, b float64) (RGB, error) {
	var res RGB

	if !gi.withinDomain(r, g, b) {
		return res, ErrOutsideOfDomain
	}

	res[0] = gi.lookUp1DSingleValue(r, 0)
	res[1] = gi.lookUp1DSingleValue(g, 1)
	res[2] = gi.lookUp1DSingleValue(b, 2)

	return res, nil
}

// For 1D LUTs we perform a linear interpolation if necessary.
func (gi *GubeImpl) lookUp1DSingleValue(v float64, index int) float64 {
	vInt, t := math.Modf(v)
	v1Index := int(vInt * float64(gi.tableSize-1))
	if t == 0 {
		return (*gi.tableData1D)[v1Index][index]
	} else {
		v2Index := v1Index + 1
		v0 := (*gi.tableData1D)[v1Index][index]
		v1 := (*gi.tableData1D)[v2Index][index]
		return lerp(v0, v1, t)
	}
}

func (gi *GubeImpl) withinDomain(r, g, b float64) bool {
	return r >= gi.domainMin[0] && g >= gi.domainMin[1] && b >= gi.domainMin[2] &&
		r <= gi.domainMax[0] && g <= gi.domainMax[1] && b <= gi.domainMax[2]
}
