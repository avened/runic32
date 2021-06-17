// This is a base32-like encoder/decoder with Anglo-Saxon runes used 
// as alphabet (Unicode). (Of course, any alphabet will do, if 
// sufficiently large.)
//
// Author: Alexander Venedioukhin
// License: Public domain, distribution unlimited
//
// https://github.com/avened/runic32
//
// Some known issues:
// 1. just a hack, not optimized;
// 2. reads everything into memory, then encodes/decodes;
// 3. interprets Unicode in weird way, expecting UTF-8 
//    and three bytes per symbol;
// 4. probably, has more than two errors;
// 5. implements an actual joke.

package main

import(
	"fmt"
	"os"
	"bufio"
	"io/ioutil"
)

var baseRunes = "ᚠᚢᚦᚩᚱᚳᚷᚹᚻᚾᛁᛡᛇᛈᛉᛋᛏᛒᛖᛗᛚᛝᛟᛞᚪᚫᛠᚣᛣᛤᚸᛢ"
var paddingRune = "ᛥ"

func numRune(c string) byte {
	for i, v := range(baseRunes){
		if string(v) == c {
			return byte(i/3)
		}
	}
	return 0xFF
}

func decodeEight(in string) ([]byte, bool) {
var q, r []byte
	if len(in) != 3*8 {
		return r, false
	}
	
	padding := false
	runesSeen := false
	paddingCount := 0
	for _, v := range(in){
		if string(v) == paddingRune {
			if !runesSeen {
				return r, false
			}
			if padding {
				paddingCount = paddingCount + 1
			}else{
				padding = true
				paddingCount = 1
			}
		}else{
			if padding {
				return r, false
			}
			runesSeen = true
			t := numRune(string(v))
			if t <= 0x20 {
				q = append(q, t)
			}else{
				return r, false
			}
		}
		
	}

	switch paddingCount{
		case 0:
			if len(q) < 8 {
				return r, false
			}
			r =	append(r,
				((q[0]) << 3) + ((q[1]) >> 2),
				((q[1]) << 6) + ((q[2]) << 1) + ((q[3]) >> 4),
				((q[3]) << 4) + ((q[4]) >> 1),
				((q[4]) << 7) + ((q[5]) << 2) + ((q[6]) >> 3),
				((q[6]) << 5) + q[7])
		case 1:
			if len(q) < 7 {
				return r, false
			}
			r =	append(r,
				((q[0]) << 3) + ((q[1]) >> 2),
				((q[1]) << 6) + ((q[2]) << 1) + ((q[3]) >> 4),
				((q[3]) << 4) + ((q[4]) >> 1),
				((q[4]) << 7) + ((q[5]) << 2) + ((q[6]) >> 3))
		case 3:
			if len(q) < 5 {
				return r, false
			}
			r =	append(r,
				((q[0]) << 3) + ((q[1]) >> 2),
				((q[1]) << 6) + ((q[2]) << 1) + ((q[3]) >> 4),
				((q[3]) << 4) + ((q[4]) >> 1))
		case 4:
			if len(q) < 4 {
				return r, false
			}
			r =	append(r,
				((q[0]) << 3) + ((q[1]) >> 2),
				((q[1]) << 6) + ((q[2]) << 1) + ((q[3]) >> 4))
		case 6:
			if len(q) < 2 {
				return r, false
			}
			r = append(r,
				((q[0]) << 3) + ((q[1]) >> 2))
		default:
			return r, false
	}
	
	return r, true
}

func convertFive(in []byte) (ret []byte) {
var r []byte

	defer func(){
		if r := recover(); r != nil {
			ret = nil
			return
		}
	}()
	if len(in) != 5 {
		return r
	}
	r = append(r, in[0]>>3, ((in[0]&0x07)<<2)+(in[1]>>6), ((in[1]&0x3E)>>1), ((in[1]&0x01)<<4)+(in[2]>>4), ((in[2]&0x0F)<<1)+((in[3]&0x80)>>7), ((in[3]&0x7C)>>2), ((in[3]&3)<<3)+(in[4]&0xE0>>5), in[4]&0x1F)
	return r
}

func runic32enc(in []byte) (out string, state bool){
var r string

	r = ""
	blocksNum := len(in)/5
	for i := 0; i < blocksNum; i++ {
		var nums []byte
		block := in[5*i:5*i+5]
		nums = convertFive(block)
		if len(nums)>0{
			for _, v := range(nums){
				r = r + string(baseRunes[int(v*3):int(v*3)+3])
			}
		}
	}
	if (len(in) - blocksNum*5) > 0 {
		var nums []byte
		block := in[5*blocksNum:len(in)]
		padLen := len(in)%5
		block = append(block, make([]byte, 5-padLen)[0:]...)
		nums = convertFive(block)
		
		switch padLen{
			case 1:
				r = r + string(baseRunes[3*int(nums[0]):3*int(nums[0])+3]) + string(baseRunes[3*int(nums[1]):3*int(nums[1])+3])
				for i := 0; i < 6; i++ {
					r = r + paddingRune
				}
			case 2:
				for i := 0; i < 4; i++ {
					r = r + string(baseRunes[3*int(nums[i]):3*int(nums[i])+3])
				}
				for i := 0; i < 4; i++ {
					r = r + paddingRune
				}
			case 3:
				for i := 0; i < 5; i++ {
					r = r + string(baseRunes[3*int(nums[i]):3*int(nums[i])+3])
				}
				for i := 0; i < 3; i++ {
					r = r + paddingRune
				}
			case 4:
				for i := 0; i < 7; i++ {
					r = r + string(baseRunes[3*int(nums[i]):3*int(nums[i])+3])
				}
				r = r + paddingRune
		}
	}
	return r, true
}

func runic32dec(in string) ([]byte, bool){
var r []byte
	if len(in)%24 != 0 {
		return r, false
	}
	
	blocksNum := len(in)/24
	
	for n := 0; n < blocksNum; n++ {
		ret, state := decodeEight(in[24*n:24*n+24])
		if state {
			r = append(r, ret[0:]...)
		}else{
			return r, false
		}
	}
	
	return r, true
}

func filterInput(in string) string{
var r []byte
	for i:=0; i < len(in); i++{
		if (in[i] != 0x0a) && (in[i] != 0x0d) {
			r = append(r, byte(in[i]))
		}
	}
	
	return string(r)
}

func main(){
var wantHelp, wantDecode bool
var gotFname string
var data []byte

	if (len(os.Args) > 1){
		if os.Args[1] == "-h" {
			wantHelp = true
		}
	}
	if (len(os.Args) > 3) || wantHelp {
		errDesc := "Usage: " + os.Args[0] + " [-d|-h] [file.name]\n\t-h - help;\n\t-d - to decode;\n\tfile.name - input file name (will read STDIN if empty).\nOutputs result to STDOUT (so watch your command line).\nIgnores \\n,\\r in encoded input.\n\n"
		if !wantHelp {
			errDesc = "Bad option!\n" + errDesc
		}
		fmt.Print(errDesc)
		os.Exit(1)
	}
	
	if len(os.Args) < 2 {
		wantDecode = false
		gotFname = ""
	}else{
	
		if os.Args[1] == "--d" {
			if len(os.Args) == 2 {
				gotFname = "-d"
			}else{
				fmt.Print("Bad options!\n")
				os.Exit(1)
			}
		}else{
			if os.Args[1] == "-d" {
				wantDecode = true
				if len(os.Args) > 2 {
					gotFname = os.Args[2]
				}
			}else{
				if len(os.Args) > 2 {
					fmt.Print("Bad options!\n")
					os.Exit(1)
				}
				gotFname = os.Args[1]
			}
		}
	
	}
	if gotFname == "" {
		var err error
		reader := bufio.NewReader(os.Stdin)
		data, err = ioutil.ReadAll(reader)
		if err != nil {
			panic("Could not read stdin!")
		}
	}else{
		var err error
		data, err = ioutil.ReadFile(gotFname)
		if err != nil {
			panic("Could not read!")
		}
	}
	
	if wantDecode {
		p, good := runic32dec(filterInput(string(data)))
		if good {
			writer := bufio.NewWriter(os.Stdout)
			_, err := writer.Write(p)
			if err != nil {
				panic("Could not write stdout!")
			}
			writer.Flush()
		}else{
			fmt.Print("Bad decode!\n")
		}
	}else{
		p, good := runic32enc(data)
		if good {
			fmt.Printf("%s\n", p)
		}else{
			fmt.Print("Bad encode!\n")
		}
	}
}
