package main;

import (
	"log"
	"context"
	"time"
	"fmt"
	"bufio"
	_"encoding/json"
	"os"
"github.com/coder/websocket"
 "github.com/coder/websocket/wsjson"
)

func main(){

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	
	c, _, err := websocket.Dial(ctx, "ws://localhost:8081", nil)
	if err != nil {
		log.Println(err)
	}
	defer c.CloseNow()
	var v any
	inps:=bufio.NewScanner(os.Stdin)
	var myname string
	for inps.Scan(){
		myname=inps.Text()
		err = wsjson.Write(ctx, c, fmt.Sprintf(myname))

	}
	// jsonWithName,err:=json.Marshal(map[string]interface{}{
	// 	"name": 
	// })
	err = wsjson.Write(ctx, c, fmt.Sprintf("hi "))
	for{
		
		time.Sleep(time.Second*1)
		err = wsjson.Read(ctx, c, &v)
		log.Println(v)
		if err != nil {
			log.Println(err)
		}
	}
	
	c.Close(websocket.StatusNormalClosure, "closed")
}