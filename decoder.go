package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type Pattern struct {
	header header
	tracks []*track
}

//header is fixed size, so its easy to decode in one go with binary.Read
type header struct {
	Splice  [6]byte  // "SPLICE"
	_       [7]byte  //padding
	Size    uint8    // size of file - splice - padding, -> bytes following
	Version [32]byte // TODO not sure about padding, but works for now
	Tempo   float32
}

type track struct {
	name string
	data [16]bool
	id   uint32
}

func DecodeFile(path string) (*Pattern, error) {
	f, err := os.Open(path)
	defer f.Close()

	p := &Pattern{}
	err = binary.Read(f, binary.LittleEndian, &p.header)
	if err != nil {
		return nil, fmt.Errorf("Error parsing header: %s", err)
	}
	//header is fixed, rest is dynamic in size
	//only read max "announced" bytes
	data_size := int(p.header.Size) - len(p.header.Version) - 4 //-4 == float32 for Version
	lr := io.LimitReader(f, int64(data_size))
	for {
		t, err := parseTrack(lr)
		// we dont consider EOF an error if everything else worked
		if err == io.EOF {
			return p, nil
		}
		if err != nil {
			return nil, err
		}
		p.tracks = append(p.tracks, t)
	}
}

func parseTrack(r io.Reader) (*track, error) {
	t := &track{}

	err := binary.Read(r, binary.LittleEndian, &t.id)
	// we expect EOF here after last track
	if err == io.EOF {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("Error parsing track-id: %s", err)
	}

	var name_len uint8
	err = binary.Read(r, binary.LittleEndian, &name_len)
	if err != nil {
		return nil, fmt.Errorf("Error parsing track name length: %s", err)
	}

	name := make([]byte, name_len)
	err = binary.Read(r, binary.LittleEndian, &name)
	if err != nil {
		return nil, fmt.Errorf("Error parsing track name: %s", err)
	}
	t.name = string(name)

	var data [16]byte
	err = binary.Read(r, binary.LittleEndian, &data)
	if err != nil {
		return nil, fmt.Errorf("Error parsing track data: %s", err)
	}
	for k, v := range data {
		switch v {
		case 0:
		case 1:
			t.data[k] = true
		default:
			return nil, fmt.Errorf("error parsing track, non-boolean value: %b", v)
		}
	}

	return t, nil
}

func (t *track) String() string {
	track_line := "|"
	note := 0
	for i := 1; i < 20; i++ {
		if i%5 == 0 {
			track_line += "|"
		} else {
			if t.data[note] {
				track_line += "x"
			} else {
				track_line += "-"
			}
			note++
		}

	}
	track_line += "|"
	return fmt.Sprintf("(%d) %s\t%s", t.id, t.name, track_line)
}

func (p *Pattern) String() string {
	v := bytes.Trim(p.header.Version[:], "\x00")
	s := fmt.Sprintf(`Saved with HW Version: %s
Tempo: %g
`, v, p.header.Tempo)
	for _, t := range p.tracks {
		s += fmt.Sprintf("%s\n", t.String())
	}
	return s
}
