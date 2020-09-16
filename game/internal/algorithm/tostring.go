package algorithm

import (
	"fmt"
	"github.com/name5566/leaf/log"
	"strings"
)

func (this *Cards) Bytes() []byte {
	b := make([]byte, len(*this))
	for k, v := range *this {
		b[k] = byte(v)
	}
	return b
}
func (this *Cards) Len() int {
	return len(*this)
}
func (this *Cards) Take() byte {
	card := (*this)[0]
	(*this) = (*this)[1:]
	//fmt.Println("牌型剩余Take数量 ~ :", len(*this))
	return card
}
func (this *Cards) Append(cards ...byte) Cards {
	cs := make([]byte, 0, len(cards)+len(*this))
	cs = append(cs, (*this)...)
	cs = append(cs, cards...)
	return cs
}

func (this *Cards) Equal(cards []byte) bool {
	if len(*this) != len(cards) {
		return false
	}
	for k, v := range *this {
		if cards[k] != v {
			return false
		}
	}
	return true
}
func Color(color byte) (char string) {
	switch color {
	case 0:
		char = "♦"
	case 1:
		char = "♣"
	case 2:
		char = "♥"
	case 3:
		char = "♠"
	}
	return
}

func String2Num(c byte) (n byte) {
	switch c {
	case '2':
		n = 2
	case '3':
		n = 3
	case '4':
		n = 4
	case '5':
		n = 5
	case '6':
		n = 6
	case '7':
		n = 7
	case '8':
		n = 8
	case '9':
		n = 9
	case 'T':
		n = 0xA
	case 'J':
		n = 0xB
	case 'Q':
		n = 0xC
	case 'K':
		n = 0xD
	case 'A':
		n = 0xE
	}
	return
}
func Num2String(n byte) (c byte) {
	switch n {
	case 2:
		c = '2'
	case 3:
		c = '3'
	case 4:
		c = '4'
	case 5:
		c = '5'
	case 6:
		c = '6'
	case 7:
		c = '7'
	case 8:
		c = '8'
	case 9:
		c = '9'
	case 0xA:
		c = 'T'
	case 0xB:
		c = 'J'
	case 0xC:
		c = 'Q'
	case 0xD:
		c = 'K'
	case 0xE:
		c = 'A'
	}
	return
}

func (this *Cards) SetByString(str string) {
	array := strings.Split(str, " ")
	*this = make([]byte, len(array))
	for k, v := range array {
		(*this)[k] = String2Num(byte(v[0]))
	}

}
func (this *Cards) String() (str string) {
	for k, v := range *this {
		color := Color(v)
		value := Num2String(v)
		str += string(color) + string(value)
		if k < len(*this)-1 {
			str += " "
		}
	}
	return
}

func (this *Cards) Hex() string {
	//for _, val := range *this {
	//	fmt.Println(fmt.Sprintf("0x"+"%.2x", val))
	//}
	return fmt.Sprintf("%#v", *this)
}

func (this *Cards) HexInt() []int32 {
	hex := make([]int32, 0)
	for _, val := range *this {
		str := fmt.Sprintf("0x"+"%.2x", val)
		switch str {
		case "0x0e":
			hex = append(hex, 1)
		case "0x02":
			hex = append(hex, 2)
		case "0x03":
			hex = append(hex, 3)
		case "0x04":
			hex = append(hex, 4)
		case "0x05":
			hex = append(hex, 5)
		case "0x06":
			hex = append(hex, 6)
		case "0x07":
			hex = append(hex, 7)
		case "0x08":
			hex = append(hex, 8)
		case "0x09":
			hex = append(hex, 9)
		case "0x0a":
			hex = append(hex, 10)
		case "0x0b":
			hex = append(hex, 11)
		case "0x0c":
			hex = append(hex, 12)
		case "0x0d":
			hex = append(hex, 13)
		case "0x1e":
			hex = append(hex, 14)
		case "0x12":
			hex = append(hex, 15)
		case "0x13":
			hex = append(hex, 16)
		case "0x14":
			hex = append(hex, 17)
		case "0x15":
			hex = append(hex, 18)
		case "0x16":
			hex = append(hex, 19)
		case "0x17":
			hex = append(hex, 20)
		case "0x18":
			hex = append(hex, 21)
		case "0x19":
			hex = append(hex, 22)
		case "0x1a":
			hex = append(hex, 23)
		case "0x1b":
			hex = append(hex, 24)
		case "0x1c":
			hex = append(hex, 25)
		case "0x1d":
			hex = append(hex, 26)
		case "0x2e":
			hex = append(hex, 27)
		case "0x22":
			hex = append(hex, 28)
		case "0x23":
			hex = append(hex, 29)
		case "0x24":
			hex = append(hex, 30)
		case "0x25":
			hex = append(hex, 31)
		case "0x26":
			hex = append(hex, 32)
		case "0x27":
			hex = append(hex, 33)
		case "0x28":
			hex = append(hex, 34)
		case "0x29":
			hex = append(hex, 35)
		case "0x2a":
			hex = append(hex, 36)
		case "0x2b":
			hex = append(hex, 37)
		case "0x2c":
			hex = append(hex, 38)
		case "0x2d":
			hex = append(hex, 39)
		case "0x3e":
			hex = append(hex, 40)
		case "0x32":
			hex = append(hex, 41)
		case "0x33":
			hex = append(hex, 42)
		case "0x34":
			hex = append(hex, 43)
		case "0x35":
			hex = append(hex, 44)
		case "0x36":
			hex = append(hex, 45)
		case "0x37":
			hex = append(hex, 46)
		case "0x38":
			hex = append(hex, 47)
		case "0x39":
			hex = append(hex, 48)
		case "0x3a":
			hex = append(hex, 49)
		case "0x3b":
			hex = append(hex, 50)
		case "0x3c":
			hex = append(hex, 51)
		case "0x3d":
			hex = append(hex, 52)
		}
	}
	return hex
}

func CardString(cards []int32) []string {
	var str []string
	for _, num := range cards {
		switch num {
		case 1:
			str = append(str, "♠A")
		case 2:
			str = append(str, "♠2")
		case 3:
			str = append(str, "♠3")
		case 4:
			str = append(str, "♠4")
		case 5:
			str = append(str, "♠5")
		case 6:
			str = append(str, "♠6")
		case 7:
			str = append(str, "♠7")
		case 8:
			str = append(str, "♠8")
		case 9:
			str = append(str, "♠9")
		case 10:
			str = append(str, "♠10")
		case 11:
			str = append(str, "♠J")
		case 12:
			str = append(str, "♠Q")
		case 13:
			str = append(str, "♠K")
		case 14:
			str = append(str, "♣A")
		case 15:
			str = append(str, "♣2")
		case 16:
			str = append(str, "♣3")
		case 17:
			str = append(str, "♣4")
		case 18:
			str = append(str, "♣5")
		case 19:
			str = append(str, "♣6")
		case 20:
			str = append(str, "♣7")
		case 21:
			str = append(str, "♣8")
		case 22:
			str = append(str, "♣9")
		case 23:
			str = append(str, "♣10")
		case 24:
			str = append(str, "♣J")
		case 25:
			str = append(str, "♣Q")
		case 26:
			str = append(str, "♣K")
		case 27:
			str = append(str, "♥A")
		case 28:
			str = append(str, "♥2")
		case 29:
			str = append(str, "♥3")
		case 30:
			str = append(str, "♥4")
		case 31:
			str = append(str, "♥5")
		case 32:
			str = append(str, "♥6")
		case 33:
			str = append(str, "♥7")
		case 34:
			str = append(str, "♥8")
		case 35:
			str = append(str, "♥9")
		case 36:
			str = append(str, "♥10")
		case 37:
			str = append(str, "♥J")
		case 38:
			str = append(str, "♥Q")
		case 39:
			str = append(str, "♥K")
		case 40:
			str = append(str, "♦A")
		case 41:
			str = append(str, "♦2")
		case 42:
			str = append(str, "♦3")
		case 43:
			str = append(str, "♦4")
		case 44:
			str = append(str, "♦5")
		case 45:
			str = append(str, "♦6")
		case 46:
			str = append(str, "♦7")
		case 47:
			str = append(str, "♦8")
		case 48:
			str = append(str, "♦9")
		case 49:
			str = append(str, "♦10")
		case 50:
			str = append(str, "♦J")
		case 51:
			str = append(str, "♦Q")
		case 52:
			str = append(str, "♦K")
		}
	}
	return str
}

func ShowCards(kind uint8, cards []int32) []int32 {

	switch kind {
	case 1: // 高牌
	case 2: // 一对
	case 3: // 两对
		cardShow := ShowTwoPairs(cards)
		return cardShow
	case 4: // 三条
	case 5: // 顺子
	case 6: // 同花
	case 7: // 葫芦
	case 8: // 四条
	case 9: // 同花顺
	case 10: // 皇家同花顺
	}
	return cards
}

func NewString(str string) string {
	str = strings.TrimPrefix(str, "♥")
	str = strings.TrimPrefix(str, "♠")
	str = strings.TrimPrefix(str, "♣")
	str = strings.TrimPrefix(str, "♦")
	fmt.Println(str)
	return str
}

func ShowTwoPairs(cards []int32) []int32 {
	str := CardString(cards)
	log.Debug("str:%v", str)
	num0 := str[0]
	num1 := str[1]
	num2 := str[2]
	num3 := str[3]
	num4 := str[4]
	num5 := str[5]
	num6 := str[6]
	num0 = NewString(num0)
	num1 = NewString(num1)
	num2 = NewString(num2)
	num3 = NewString(num3)
	num4 = NewString(num4)
	num5 = NewString(num5)
	num6 = NewString(num6)
	return cards
}
