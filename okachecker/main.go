package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	qrcodeTerminal "github.com/Baozisoftware/qrcode-terminal-go"
	"github.com/Rhymen/go-whatsapp"
)

var isReplyDetected bool
var isLoaded bool
var prevDate uint64

func main() {
	fmt.Println("WhatsApp Bot Checker started...")
	//create new WhatsApp connection
	wac, err := whatsapp.NewConn(5 * time.Second)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating connection: %v\n", err)
		return
	}

	//Add handler
	wac.AddHandler(&waHandler{wac})

	err = login(wac)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error logging in: %v\n", err)
		return
	}

	<-time.After(10 * time.Second)

	msg := whatsapp.TextMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: "6281250002655@s.whatsapp.net",
		},
		Text: "Hallo",
	}

	msgID, err := wac.Send(msg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error sending message: %v", err)
		os.Exit(1)
	} else {
		fmt.Println("Message Sent -> ID : " + msgID)
		isLoaded = true
	}

	//action after sending message
	c2 := make(chan string, 1)
	i := 1
	go func() {
		for isReplyDetected == false {
			time.Sleep(time.Second)
			fmt.Printf("[ %v ] waiting for reply ...\n", i)
			i++
		}
		if isReplyDetected == true {
			c2 <- "reply found!"
		}
	}()
	select {
	case res := <-c2:
		fmt.Println(res)
	case <-time.After(20 * time.Second):
		fmt.Println("20 seconds timeout reached")
		resp, err := http.Get("http://168.235.67.17/uptime/send2wa.php?group=Onboarding+Okadoc+ID&msg=Monitor%20is%20DOWN%3A%20%5BPROD%5D%20Whatsapp%20Bot%20RS%20Permata%20Pamulang%20-%20Reason%3A%20Responding%20more%20than%2020%20seconds")
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("Message sent to WhatsApp Group!")
		defer resp.Body.Close()
	}

	isReplyDetected = false
}

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

//Optional to be implemented. Implement HandleXXXMessage for the types you need.
func (*waHandler) HandleTextMessage(message whatsapp.TextMessage) {

	if strings.Contains(message.Info.RemoteJid, "6281250002655") && !message.Info.FromMe && isLoaded {
		fmt.Printf("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++\n")

		fmt.Printf("%v %v %v %v\n\t%v\n", message.Info.Timestamp, message.Info.Id, message.Info.RemoteJid, message.Info.QuotedMessageID, message.Text)
		selisih := message.Info.Timestamp - prevDate

		fmt.Printf("Yuhuu! Selisihnya adalah %v detik\n", selisih)

		prevDate = message.Info.Timestamp
		isReplyDetected = true
		fmt.Printf("-------------------------------------end--------------------------------\n\n\n")
	} else if strings.Contains(message.Info.RemoteJid, "6281250002655") && message.Info.FromMe && isLoaded {
		fmt.Println(("Rizal sudah mengirimkan pesan"))

		prevDate = message.Info.Timestamp
	}
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
