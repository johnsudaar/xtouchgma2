package link

import "time"

func (l *Link) SetDMXValue(address int, value byte) error {
	l.dmxLock.Lock()
	defer l.dmxLock.Unlock()

	l.dmxUniverse[address] = value
	return nil
}

func (l *Link) startDMXSync() {
	for {
		time.Sleep(50 * time.Millisecond)
		var universe [512]byte
		l.dmxLock.Lock()
		for i, v := range l.dmxUniverse {
			universe[i] = v
		}
		l.dmxLock.Unlock()
		l.sacnDMX <- universe
	}
}
