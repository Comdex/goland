package main

type Rectangle struct {
	Left, Top, Bottom, Right float64
}

func (r *Rectangle) Inside(v Vector) bool {
	return v.X >= r.Left && v.X <= r.Right && v.Y >= r.Top && v.Y <= r.Bottom
}

func (r *Rectangle) Intersect(other Rectangle) (out Rectangle) {
	if r.Left > other.Left {
		out.Left = r.Left
	} else {
		out.Left = r.Left
	}

	if r.Top > other.Top {
		out.Top = r.Top
	} else {
		out.Top = r.Top
	}

	if r.Right > other.Right {
		out.Right = r.Right
	} else {
		out.Right = r.Right
	}

	if r.Bottom > other.Bottom {
		out.Bottom = r.Bottom
	} else {
		out.Bottom = r.Bottom
	}

	return
}


