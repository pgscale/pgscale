// Copyright 2021 Burak Sezer
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package protocol

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
)

const HeaderLen = 5

type Reader struct {
	src net.Conn
}

func New(src net.Conn) (*Reader, error) {
	return &Reader{
		src: src,
	}, nil
}

type DataPacket struct {
	Identifier byte
	Header     []byte
	Payload    []byte
}

func (d *DataPacket) Encode() []byte {
	data := make([]byte, HeaderLen+len(d.Payload))
	copy(data, d.Header)
	copy(data[HeaderLen:], d.Payload)
	return data
}

func (c *Reader) Read() (*DataPacket, error) {
	header := make([]byte, HeaderLen)
	_, err := c.src.Read(header)
	if err != nil {
		return nil, err
	}

	packetLen := binary.BigEndian.Uint32(header[1:]) - 4
	payload := make([]byte, packetLen)
	nr, err := c.src.Read(payload)
	if errors.Is(err, io.EOF) {
		err = nil
	}
	if err != nil {
		return nil, err
	}

	return &DataPacket{
		Identifier: header[0],
		Header:     header,
		Payload:    payload[:nr],
	}, nil
}
