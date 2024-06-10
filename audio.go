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
