package rans

var (
	s_prec = 64
	t_prec = 32
	t_mask = (1 << t_prec) - 1
	s_min  = 1 << (s_prec - t_prec)
	s_max  = 1 << s_prec
)

type message struct {
	data uint
	next *message
}
