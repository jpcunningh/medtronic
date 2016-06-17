package medtronic

import (
	"bytes"
)

const (
	CurrentPage CommandCode = 0x9D
	History     CommandCode = 0x80
)

func (pump *Pump) CurrentPage() int {
	result := pump.Execute(CurrentPage, func(data []byte) interface{} {
		if len(data) < 5 || data[0] != 4 {
			return nil
		}
		return fourByteInt(data[1:5])
	})
	if pump.Error() != nil {
		return 0
	}
	return result.(int)
}

func (pump *Pump) History(page int) [][]byte {
	result := pump.Execute(History, func(data []byte) interface{} {
		return data
	}, byte(page))
	if pump.Error() != nil {
		return nil
	}
	data := result.([]byte)
	results := [][]byte{}
	ack := commandPacket(Ack, nil)
	for {
		done := data[0]&0x80 != 0
		data[0] &^= 0x80
		// Skip duplicate responses.
		n := len(results)
		if n == 0 || !bytes.Equal(results[n-1], data) {
			results = append(results, data)
		}
		if done {
			break
		}
		pump.Radio.Send(ack)
		next, _ := pump.Radio.Receive(pump.timeout)
		if next == nil {
			pump.SetError(nil)
			continue
		}
		data = pump.DecodePacket(next)
		if pump.Error() != nil {
			pump.SetError(nil)
			continue
		}
		if pump.Error() == nil && !expected(History, data) {
			pump.SetError(BadResponseError{command: History, data: data})
			break
		}
		data = data[5:]
	}
	return results
}