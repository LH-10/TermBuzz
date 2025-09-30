package main;

import ("fmt"
"log"
"net/http"
"context"
"time"
"github.com/coder/websocket"
"github.com/coder/websocket/wsjson"
)

type Client struct{
	conn *websocket.Conn
	id int
	name string
}
var clients []*Client

func ClientToClient(sender *websocket.Conn,recievername string,ctx context.Context,message string){
	var sendername string
	var reciever *websocket.Conn
	for i:=range clients{
		if sender==clients[i].conn {
			sendername=clients[i].name
			continue
		}
		if recievername==clients[i].name {
			reciever=clients[i].conn
		}
	}
	fmt.Println("Sender",sendername,"\n reciever conn",reciever)
	message+=sendername
	wsjson.Write(ctx,reciever,message)
}

func main(){
	http.Handle("/",http.HandlerFunc(func (w http.ResponseWriter,r *http.Request){
			c,err:=websocket.Accept(w,r,&websocket.AcceptOptions{
				OriginPatterns:[]string{"*"},
			})
			currentClient:=&Client{conn:c,id:int(time.Now().Unix())}
			clients=append(clients,currentClient)
			fmt.Println("New connection",clients)
			if err!=nil{
				log.Println(err)
			}
			defer c.CloseNow()
			ctx:=context.Background()
			var name string
			time.Sleep(time.Second*5)
			fmt.Println("Woke up")
			err=wsjson.Read(ctx,c,&name)
			
			if err!=nil{
				log.Println(err)
				return
			}
			(currentClient).name=string(name)
			fmt.Println(currentClient.name)

			fmt.Println("connection with name",clients)
			
			for{
				
				
				var v any
				err=wsjson.Read(ctx,c,&v)
				if err!=nil{
					log.Println(err)
					break
				}
				log.Printf("recieved: %v",v)
				wsjson.Write(ctx,c,fmt.Sprintf("recieved %v",v))
			}
				
			c.Close(websocket.StatusNormalClosure, "cross origin WebSocket accepted")
		}))
		
		
		err:=http.ListenAndServe("localhost:8081",nil)
		if err!=nil{
			fmt.Println(err)
		}
		// fmt.Println("Hello",wsfn)
		
	}