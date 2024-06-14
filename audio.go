package main

// using gopkg.in/hraban/opus.v2

// func loadSound() {
// 	s, err := opus.NewStream(file)
// 	if err != nil {
// 		fmt.Println("Couldnt open opus stream")
// 		return errors.New("couldnt open Opus stream")
// 	}
// 	defer s.Close()
// 	pcmbuf := make([]int16, 16384)
// 	for {
// 		n, err := s.Read(pcmbuf)
// 		if err == io.EOF {
// 			break
// 		} else if err != nil {
// 			return err
// 		}
// 		pcm := pcmbuf[:n*channels]

// 		// send pcm to audio device here, or write to a .wav file
// 		// Append encoded pcm data to the buffer.
// 		temp := make([]byte, 2)
// 		for _, p := range pcm {
// 			binary.BigEndian.PutUint16(temp, uint16(p))
// 		}
// 		buffer = append(buffer, temp)
// 	}
// }

// encoded using jonas747/dca cli, reading raw dca

// func loadSound(filename string) error {

// 	file, err := os.Open("downloads/" + filename)
// 	if err != nil {
// 		fmt.Println("Error opening dca file :", err)
// 		return err
// 	}

// 	var opuslen int16
// 	buffer = make([][]byte, 0)

// 	for {
// 		// Read opus frame length from dca file.
// 		err = binary.Read(file, binary.LittleEndian, &opuslen)

// 		// If this is the end of the file, just return.
// 		if err == io.EOF || err == io.ErrUnexpectedEOF {
// 			err := file.Close()
// 			if err != nil {
// 				return err
// 			}
// 			return nil
// 		}

// 		if err != nil {
// 			fmt.Println("Error reading from dca file :", err)
// 			return err
// 		}

// 		if opuslen < 0 {
// 			return errors.New("frame size is negative, possibly corrupted")
// 		}

// 		// Read encoded pcm from dca file.
// 		InBuf := make([]byte, opuslen)
// 		err = binary.Read(file, binary.LittleEndian, &InBuf)

// 		// Should not be any end of file errors
// 		if err != nil {
// 			fmt.Println("Error reading from dca file :", err)
// 			return err
// 		}

// 		// Append encoded pcm data to the buffer.
// 		buffer = append(buffer, InBuf)
// 	}
// }

// Using jonas747/dca lib

// func loadSound(filename string) error {

// 	file, err := os.Open("downloads/" + filename)
// 	if err != nil {
// 		fmt.Println("Error opening dca file :", err)
// 		return err
// 	}

// 	buffer = make([][]byte, 0)
// 	decoder := dca.NewDecoder(file)

// 	for {
// 		frame, err2 := decoder.OpusFrame()
// 		if err2 != nil {
// 			if err2 != io.EOF || err != io.ErrUnexpectedEOF {
// 				return err2
// 			}

// 			file.Close()
// 			break
// 		}

// 		// Append encoded pcm data to the buffer.
// 		buffer = append(buffer, frame)
// 	}

// 	return nil
// }
