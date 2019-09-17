package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	qrcodeTerminal "github.com/Baozisoftware/qrcode-terminal-go"
	"github.com/Rhymen/go-whatsapp"
)

type waHandler struct {
	c *whatsapp.Conn
}

//HandleError needs to be implemented to be a valid WhatsApp handler
func (h *waHandler) HandleError(err error) {

	if e, ok := err.(*whatsapp.ErrConnectionFailed); ok {
		log.Printf("Connection failed, underlying error: %v", e.Err)
		log.Println("Waiting 15sec...")
		<-time.After(15 * time.Second)
		log.Println("Reconnecting...")
		err := h.c.Restore()
		if err != nil {
			log.Fatalf("Restore failed: %v", err)
		}
	} else {
		log.Printf("error occoured: %v\n", err)
	}
}

var prevDate uint64

var isReplyDetected bool

var globalSelisih int

func getReplyDetected() bool {
	return isReplyDetected
}

//Optional to be implemented. Implement HandleXXXMessage for the types you need.
func (*waHandler) HandleTextMessage(message whatsapp.TextMessage) {

	if strings.Contains(message.Info.RemoteJid, "6281250002655") && !message.Info.FromMe && isLoaded {
		fmt.Printf("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++\n")

		fmt.Printf("%v %v %v %v\n\t%v\n", message.Info.Timestamp, message.Info.Id, message.Info.RemoteJid, message.Info.QuotedMessageID, message.Text)
		selisih := message.Info.Timestamp - prevDate
		globalSelisih = int(selisih)

		fmt.Printf("Yuhuu! Selisihnya adalah %v detik\n", selisih)

		// from thizzz

		// if selisih >= 10 {
		// 	resp, err := http.Get("http://168.235.67.17/uptime/send2wa.php?group=uji+whatsmate&msg=bot+whatsapp+lebih+dari+10+detik")
		// 	if err != nil {
		// 		log.Fatalln(err)
		// 	}
		// 	fmt.Printf("Message sent to WhatsApp Group!")
		// 	defer resp.Body.Close()
		// } else {
		// 	fmt.Printf("Selisihnya tidak lebih dari 10, masih aman, yaitu %v\n", selisih)
		// }
		// to thizzz
		prevDate = message.Info.Timestamp
		isReplyDetected = true
		fmt.Printf("-------------------------------------end--------------------------------\n\n\n")
	} else if strings.Contains(message.Info.RemoteJid, "6281250002655") && message.Info.FromMe && isLoaded {
		fmt.Println(("Rizal sudah mengirimkan pesan"))
		// <-time.After((10 * time.Second))
		// log.Println("Sudah 10 detik nih")
		prevDate = message.Info.Timestamp
	}
}

/*//Example for media handling. Video, Audio, Document are also possible in the same way
func (*waHandler) HandleImageMessage(message whatsapp.ImageMessage) {
	data, err := message.Download()
	if err != nil {
		return
	}
	filename := fmt.Sprintf("%v/%v.%v", os.TempDir(), message.Info.Id, strings.Split(message.Type, "/")[1])
	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		return
	}
	_, err = file.Write(data)
	if err != nil {
		return
	}
	log.Printf("%v %v\n\timage reveived, saved at:%v\n", message.Info.Timestamp, message.Info.RemoteJid, filename)
}*/

var isLoaded bool

func main() {

	//create new WhatsApp connection
	wac, err := whatsapp.NewConn(5 * time.Second)
	if err != nil {
		log.Fatalf("error creating connection: %v\n", err)
	}

	//Add handler
	wac.AddHandler(&waHandler{wac})

	//login or restore
	if err := login(wac); err != nil {
		log.Fatalf("error logging in: %v\n", err)
	}

	//verifies phone connectivity
	pong, err := wac.AdminTest()

	if !pong || err != nil {
		log.Fatalf("error pinging in: %v\n", err)
	}

	isLoaded = false

	for i := 0; i < 5; i++ {
		<-time.After((3600 * time.Second))
		log.Printf("Sudah 3600 detik nih, gw dari main, ini yang ke %v \nSending now!", i)
		// whatsapp code
		msg := whatsapp.TextMessage{
			Info: whatsapp.MessageInfo{
				RemoteJid: "6281250002655@s.whatsapp.net",
			},
			Text: "Ya",
		}

		msgID, err := wac.Send(msg)
		prevDate = uint64(time.Now().Unix())
		if err != nil {
			os.Exit(1)
		} else {
			fmt.Println("Message Sent -> ID : " + msgID)
			isLoaded = true
		}
		// whatsapp end

		//
		// checking reply
		go checkReply()
		//
		// if globalSelisih >= 10 {
		// 	fmt.Printf("Selisihnya lebih BESAR dari 10, PARAAAHHHH %v\n", globalSelisih)
		// 	// resp, err := http.Get("http://168.235.67.17/uptime/send2wa.php?group=uji+whatsmate&msg=bot+whatsapp+lebih+dari+10+detik")
		// 	// if err != nil {
		// 	// 	log.Fatalln(err)
		// 	// }
		// 	// fmt.Printf("Message sent to WhatsApp Group!")
		// 	// defer resp.Body.Close()
		// } else {
		// 	fmt.Printf("Selisihnya tidak lebih dari 10, masih aman, yaitu %v\n", globalSelisih)
		// }
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	//Disconnect safe
	fmt.Println("Shutting down now.")
	session, err := wac.Disconnect()
	if err != nil {
		log.Fatalf("error disconnecting: %v\n", err)
	}
	if err := writeSession(session); err != nil {
		log.Fatalf("error saving session: %v", err)
	}
}

func checkReply() {
	c2 := make(chan string, 1)
	go func() {
		// time.Sleep(5 * time.Second)
		for isReplyDetected == false {
			fmt.Println("waiting for reply ...", isReplyDetected)
			time.Sleep(time.Second)
		}
		if isReplyDetected == true {
			c2 <- "ini hasil prosesnya berhasil"
		}
	}()
	select {
	case res := <-c2:
		fmt.Println(res)
	case <-time.After(10 * time.Second):
		fmt.Println("kena timeout")
		resp, err := http.Get("http://168.235.67.17/uptime/send2wa.php?group=uji+whatsmate&msg=bot+whatsapp+lebih+dari+10+detik")
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("Message sent to WhatsApp Group!")
		defer resp.Body.Close()
	}

	isReplyDetected = false
}

func login(wac *whatsapp.Conn) error {
	//load saved session
	session, err := readSession()
	if err == nil {
		//restore session
		session, err = wac.RestoreWithSession(session)
		if err != nil {
			return fmt.Errorf("restoring failed: %v\n", err)
		}
	} else {
		//no saved session -> regular login
		qr := make(chan string)
		go func() {
			terminal := qrcodeTerminal.New()
			terminal.Get(<-qr).Print()
		}()
		session, err = wac.Login(qr)
		if err != nil {
			return fmt.Errorf("error during login: %v\n", err)
		}
	}

	//save session
	err = writeSession(session)
	if err != nil {
		return fmt.Errorf("error saving session: %v\n", err)
	}
	return nil
}

func readSession() (whatsapp.Session, error) {
	session := whatsapp.Session{}
	file, err := os.Open(os.TempDir() + "/whatsappSession.gob")
	if err != nil {
		return session, err
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&session)
	if err != nil {
		return session, err
	}
	return session, nil
}

func writeSession(session whatsapp.Session) error {
	file, err := os.Create(os.TempDir() + "/whatsappSession.gob")
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(session)
	if err != nil {
		return err
	}
	return nil
}
