package tools

import (
	"fmt"
	"math"
)

type Vec3f struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

func (s *Vec3f) String() string {
	return fmt.Sprintf("x:%v y:%v", s.X, s.Y)
}

func Add(l, r Vec3f) Vec3f {
	return Vec3f{
		X: l.X + r.X,
		Y: l.Y + r.Y,
		Z: l.Z + r.Z,
	}
}
func (s *Vec3f) Add(v Vec3f) {
	s.X += v.X
	s.Y += v.Y
	s.Z += v.Z
}
func Sub(l, r Vec3f) Vec3f {
	return Vec3f{
		X: l.X - r.X,
		Y: l.Y - r.Y,
		Z: l.Z - r.Z,
	}
}
func (s *Vec3f) Sub(v Vec3f) {
	s.X -= v.X
	s.Y -= v.Y
	s.Z -= v.Z
}

func Mul(v Vec3f, d float64) Vec3f {
	return Vec3f{
		X: v.X * d,
		Y: v.Y * d,
		Z: v.Z * d,
	}
}

func Div(v Vec3f, d float64) Vec3f {
	return Vec3f{
		X: v.X / d,
		Y: v.Y / d,
		Z: v.Z / d,
	}
}

var _invalid = Vec3f{-1, -1, -1}

func Valid(v Vec3f) bool {
	return v != _invalid
}

func Invalid() Vec3f {
	return _invalid
}

func SqrMagnitude(v Vec3f) float64 {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

func Magnitude(v Vec3f) float64 {
	return math.Sqrt(SqrMagnitude(v))
}

func SqrDistance(p0, p1 Vec3f) float64 {
	return SqrMagnitude(Sub(p1, p0))
}

func Distance(p0, p1 Vec3f) float64 {
	return Magnitude(Sub(p1, p0))
}
