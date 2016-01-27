package bc2015

func (r *Replay) Stream(c chan interface{}, done chan bool) {
	close(c)
	return
}
