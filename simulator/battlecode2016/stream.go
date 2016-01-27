package bc2016

type Message struct {
	MessageType string
	Data        interface{}
}

func (r *Replay) Stream(c chan interface{}, done chan bool) {
	defer close(c)
	c <- Message{"StoredConstants", r.StoredConstants}
	if isDone := <-done; isDone {
		return
	}
	c <- Message{"Header", r.Header}
	if isDone := <-done; isDone {
		return
	}
	c <- Message{"Metadata", r.Metadata}
	if isDone := <-done; isDone {
		return
	}
	for i := 0; i < len(r.Round); i++ {
		c <- Message{"Round", r.Round[i]}
		if isDone := <-done; isDone {
			return
		}
	}
	c <- Message{"GameStats", r.GameStats}
	if isDone := <-done; isDone {
		return
	}
	c <- Message{"Footer", r.Footer}
	if isDone := <-done; isDone {
		return
	}
}
