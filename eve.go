package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"encoding/binary"
)

func ieeeFloatToInt(b [10]byte) int {
	// from https://github.com/go-audio/audio/blob/master/conv.go
	// Apache License 2.0
	var i uint32
	// Negative number
	if (b[0] & 0x80) == 1 {
		return 0
	}

	// Less than 1
	if b[0] <= 0x3F {
		return 1
	}

	// Too big
	if b[0] > 0x40 {
		return 67108864
	}

	// Still too big
	if b[0] == 0x40 && b[1] > 0x1C {
		return 800000000
	}

	i = (uint32(b[2]) << 23) | (uint32(b[3]) << 15) | (uint32(b[4]) << 7) | (uint32(b[5]) >> 1)
	i >>= (29 - uint32(b[1]))

	return int(i)
}


func maya(inputdata []byte)(outputdata []byte) {
	// from http://www.cs.columbia.edu/~gskc/Code/AdvancedInternetServices/SoundNoiseRatio/dvi_adpcm.c
	// by Stichting Mathematisch Centrum, Amsterdam, The Netherlands.
	// ISC License 
	
	// from Python3/Modules/audioloop.c
	// PSF-2.0
	indexTable := [16]int{-1, -1, -1, -1, 2, 4, 6, 8, -1, -1, -1, -1, 2, 4, 6, 8}
	stepsizeTable := [89]int{7, 8, 9, 10, 11, 12, 13, 14, 16, 17, 19, 21, 23, 25, 28, 31, 34, 37, 41, 45, 50, 55, 60, 66, 73, 80, 88, 97, 107, 118, 130, 143, 157, 173, 190, 209, 230, 253, 279, 307, 337, 371, 408, 449, 494, 544, 598, 658, 724, 796, 876, 963, 1060, 1166, 1282, 1411, 1552, 1707, 1878, 2066, 2272, 2499, 2749, 3024, 3327, 3660, 4026, 4428, 4871, 5358, 5894, 6484, 7132, 7845, 8630, 9493, 10442, 11487, 12635, 13899, 15289, 16818, 18500, 20350, 22385, 24623, 27086, 29794, 32767}
	bufferstep := false
	inputbuffer := 0
	delta := 0
	index := 0
	vpdiff := 0
	valpred := 0
	sign := 0
	step := stepsizeTable[index]
	for i := 0; i < len(inputdata)*2; i++ {
		if bufferstep {
			inputbuffer = int(inputdata[(i-1)/2])
			delta = inputbuffer & 0xf
		} else {
			inputbuffer = int(inputdata[i/2]>>4)
			delta = inputbuffer & 0xf
		}
		bufferstep = !bufferstep
		index = index + indexTable[delta]
		if index < 0 {
			index = 0
		}
		if index > 88 {
			index = 88
		}

		sign = delta & 8
		delta = delta & 7
		
		vpdiff = int(int32(step) >> 3)

		if delta & 4 != 0 {
			vpdiff = vpdiff+step
		}
		if delta & 2 != 0 {
			vpdiff = vpdiff+int(int32(step) >> 1)
		}
		if delta & 1 != 0 {
			vpdiff = vpdiff+int(int32(step) >> 2)
		}

		if sign != 0 {
			valpred = valpred - vpdiff
		} else {
			valpred = valpred + vpdiff
		}

		if valpred > 32767 {
			valpred = 32767
		} else if valpred < -32768 {
			valpred = -32768
		}
		step = stepsizeTable[index]

		b := make([]byte, 2)
		binary.LittleEndian.PutUint16(b, uint16(valpred))
		outputdata = append(outputdata, b...)
	}
	return
}

func eve(filename string) {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err.Error())
		//fmt.Println("File not found")
		return
	}
	sr := [10]byte{dat[28], dat[29], dat[30], dat[31], dat[32], dat[33], dat[34], dat[35], dat[36], dat[37]}
	hzint := ieeeFloatToInt(sr)
	fmt.Println(hzint)
	findch := dat[21]
	onech := byte(0x01)
	twoch := byte(0x02)
	if findch == onech {
		fmt.Println("1채널")
		adpcmdata := dat[54:]
		pcmdata:=maya(adpcmdata)
		wavdata:=[]byte{0x52, 0x49, 0x46, 0x46}
		riffchunksize := make([]byte, 4)
		riffchunksizeint := uint32(len(pcmdata)+36)
    	binary.LittleEndian.PutUint32(riffchunksize, riffchunksizeint)
		wavheader:=[]byte{0x57, 0x41, 0x56, 0x45, 0x66, 0x6D, 0x74, 0x20, 0x10, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00}
		sampleratesize := make([]byte, 4)
		sampleratesizeint := uint32(hzint)
		binary.LittleEndian.PutUint32(sampleratesize, sampleratesizeint)
		byteratesize := make([]byte, 4)
		byteratesizeint := uint32(hzint*2)
		binary.LittleEndian.PutUint32(byteratesize, byteratesizeint)
		wavheader2:=[]byte{0x02, 0x00, 0x10, 0x00, 0x64, 0x61, 0x74, 0x61}
		wavchunksize := make([]byte, 4)
		wavchunksizeint := uint32(len(pcmdata))
    	binary.LittleEndian.PutUint32(wavchunksize, wavchunksizeint)
		wavdata=append(wavdata, riffchunksize...)
		wavdata=append(wavdata, wavheader...)
		wavdata=append(wavdata, sampleratesize...)
		wavdata=append(wavdata, byteratesize...)
		wavdata=append(wavdata, wavheader2...)
		wavdata=append(wavdata, wavchunksize...)
		wavdata=append(wavdata, pcmdata...)
		err := ioutil.WriteFile(filename+".wav", wavdata, 0755)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

	} else if findch == twoch {
		fmt.Println("2채널")
		adpcmdata := dat[54:]
		leftadpcmdata := []byte{}
		rightadpcmdata := []byte{}
		frontbool := true
		var leftbit1 byte 
		var leftbit2 byte 
		var rightbit1 byte 
		var rightbit2 byte 
		for i := 0; i < len(adpcmdata); i = i + 1 {
			if frontbool{
				onebyte1:=adpcmdata[i]
				leftbit1=onebyte1 & 240
				rightbit1=(onebyte1 & 15) << 4
				frontbool = false
			} else {
				onebyte2:=adpcmdata[i]
				leftbit2=(onebyte2 & 240) >> 4
				rightbit2=onebyte2 & 15
				frontbool = true
				leftbit:=leftbit1|leftbit2
				rightbit:=rightbit1|rightbit2
				leftadpcmdata=append(leftadpcmdata, leftbit)
				rightadpcmdata=append(rightadpcmdata, rightbit)
			} 
		}
		if !frontbool {
			leftbit2=byte(0)
			rightbit2=byte(0)
			leftbit:=leftbit1|leftbit2
			rightbit:=rightbit1|rightbit2
			leftadpcmdata=append(leftadpcmdata, leftbit)
			rightadpcmdata=append(rightadpcmdata, rightbit)
		}	
		leftpcmdata:=maya(leftadpcmdata)
		rightpcmdata:=maya(rightadpcmdata)

		pcmdata := []byte{}
		for i := 0; i < len(leftpcmdata); i += 2 {
			pcmdata=append(pcmdata, leftpcmdata[i:i+2]...)
			pcmdata=append(pcmdata, rightpcmdata[i:i+2]...)
		}
		wavdata:=[]byte{0x52, 0x49, 0x46, 0x46}
		riffchunksize := make([]byte, 4)
		riffchunksizeint := uint32(len(pcmdata)+36)
    	binary.LittleEndian.PutUint32(riffchunksize, riffchunksizeint)
		wavheader:=[]byte{0x57, 0x41, 0x56, 0x45, 0x66, 0x6D, 0x74, 0x20, 0x10, 0x00, 0x00, 0x00, 0x01, 0x00, 0x02, 0x00}
		sampleratesize := make([]byte, 4)
		sampleratesizeint := uint32(hzint)
		binary.LittleEndian.PutUint32(sampleratesize, sampleratesizeint)
		byteratesize := make([]byte, 4)
		byteratesizeint := uint32(hzint*4)
		binary.LittleEndian.PutUint32(byteratesize, byteratesizeint)
		wavheader2:=[]byte{0x04, 0x00, 0x10, 0x00, 0x64, 0x61, 0x74, 0x61}
		wavchunksize := make([]byte, 4)
		wavchunksizeint := uint32(len(pcmdata))
    	binary.LittleEndian.PutUint32(wavchunksize, wavchunksizeint)
		wavdata=append(wavdata, riffchunksize...)
		wavdata=append(wavdata, wavheader...)
		wavdata=append(wavdata, sampleratesize...)
		wavdata=append(wavdata, byteratesize...)
		wavdata=append(wavdata, wavheader2...)
		wavdata=append(wavdata, wavchunksize...)
		wavdata=append(wavdata, pcmdata...)
		
		err := ioutil.WriteFile(filename+".wav", wavdata, 0755)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

}

func main() {
	file := flag.String("file", "01722.KYC1.ICM", "Input file")
	flag.Parse()
	eve(*file)
}