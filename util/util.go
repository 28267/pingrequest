package util

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"time"
)

var (
	time_out     int64
	size         int
	count        int
	typ          uint8 = 8
	cod          uint8 = 0
	senCount     int
	successCount int
	failCount    int
	minTs        int64 = math.MaxInt32
	maxTs        int64
	totalTs      int64
)

type ICMP struct {
	Type        uint8
	Code        uint8
	CheckSum    uint16
	ID          uint16
	SequenceNum uint16
}

func SendRequest() {

	GetCommandArgs()
	desIp := os.Args[len(os.Args)-1]
	t1 := time.Now()
	conn, err := net.DialTimeout("ip:icmp", desIp, time.Duration(time_out)*time.Millisecond)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer conn.Close()

	fmt.Printf("正在ping[%s][%s]具有%d字节的数据\n", desIp, conn.RemoteAddr(), size)
	for i := 0; i < count; i++ {
		senCount++
		var icmp *ICMP = &ICMP{
			Type:        typ,
			Code:        cod,
			CheckSum:    0,
			ID:          1,
			SequenceNum: 1,
		}

		data := make([]byte, size)
		var buffer bytes.Buffer //实现了读写方法的字节缓冲
		binary.Write(&buffer, binary.LittleEndian, icmp)
		buffer.Write(data)
		data = buffer.Bytes()
		checkSum := CheckSum(data)
		data[2] = byte(checkSum >> 8)
		data[3] = byte(checkSum)
		conn.SetDeadline(time.Now().Add(time.Duration(time_out) * time.Millisecond))
		_, err := conn.Write(data)
		if err != nil {
			failCount++
			log.Println(err)
			continue
		}
		// fmt.Println(n)
		buf := make([]byte, 65535)
		n, err := conn.Read(buf)
		if err != nil {
			log.Println(err)
			continue
		}
		successCount++
		t2 := time.Since(t1).Milliseconds()
		if minTs > t2 {
			minTs = t2
		}
		if maxTs < t2 {
			maxTs = t2
		}
		totalTs += t2
		fmt.Printf("来自 %d.%d.%d.%d的回复:字节=%d 时间=%dms TTL=%d\n",
			buf[12], buf[13], buf[14], buf[15], n-28, t2, buf[8])
	}
	fmt.Printf("来自%s的回复 ping的统计信息:\n 数据包: 已发送%d,已接收%d,丢失%d (%.2f%%丢失),\n往返形行程的估计时间(以毫秒为单位):%dms\n 单次请求最短时间%dms\n单次请求最长时间%dms", conn.RemoteAddr(), senCount, successCount, failCount, float64(failCount/senCount), totalTs, minTs, maxTs)
	//n-28代表IP数据报的首部占20字节+ICMP数据报的首部占8字节

}

func GetCommandArgs() {
	flag.Int64Var(&time_out, "w", 1000, "请求超时时长，单位毫秒")
	flag.IntVar(&size, "l", 32, "请求发送的缓冲区大小，单位字节")
	flag.IntVar(&count, "n", 4, "发送请求数")
	flag.Parse()
}

func CheckSum(data []byte) uint16 {
	length := len(data)
	index := 0
	var sum uint32 = 0
	for length > 1 {
		sum += uint32(data[index])<<8 + uint32(data[index+1])
		length -= 2
		index += 2
	}
	if length != 0 {
		sum += uint32(data[index])
	}
	hi16 := sum >> 16
	if hi16 != 0 {
		for {
			sum = hi16 + uint32(uint16(sum))
			hi16 = sum >> 16
			if hi16 == 0 {
				break
			}
		}
	}
	// sum = uint16(sum >> 16) + uint16(sum)

	return uint16(^sum)
}
