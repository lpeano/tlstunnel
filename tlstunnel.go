package main

import (
	"context"
	"crypto/tls"
	"io"
	"log"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/panlibin/gnet"
	"gopkg.in/yaml.v3"
)

type tcpServer struct {
	*gnet.EventServer

}

type outputdx struct {
	thid int
	dx  []byte
	Action gnet.Action
	
}

type conn_channel struct{
	
	Conn gnet.Conn
	data []byte
	outdx chan outputdx
}


var ctx = context.Background()


// func Connect( ServCon gnet.Conn) (*tls.Conn, error) {
func Connect() (*tls.Conn, error) {

	//gnet.Conn viene sostituito dalla mappa
	// Configurazione (sostituisci con i tuoi valori)
	remoteAddr := cf.Tlscfg.Proxyhost 				// Indirizzo del server remoto
	certFile := cf.Tlscfg.Certfile                  // File del certificato client
	keyFile := cf.Tlscfg.Keyfile                    // File della chiave privata client

	// Carica certificato client e chiave privata
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("Errore durante il caricamento del certificato client: %v", err)
		return nil, err
	} else {
		log.Printf("Certificates loaded!")
	}

	// Crea configurazione TLS - Trusting any server
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true, // Trust any server certificate (use with caution in production)
	}

	conn, err := tls.Dial("tcp", remoteAddr, tlsConfig)
	if err != nil {
		log.Fatalf("Error in tls dial %s",err)
	} else {
		log.Printf("Connected to %s",remoteAddr)
	}

	return conn, err
	
}


var PrxConn *tls.Conn
var PrxConnected bool
var ManageConnect bool

func (es *tcpServer) OnClosed(c gnet.Conn, err error) (action gnet.Action){
	log.Printf("Closing Connection")
	mu.Lock()
	defer mu.Unlock()
	conn,ok:=connections.Load(c)
	if ok {
		logger.Debug("Fuonded connection with context ",slog.Any("conn",conn))
		D,ok:=connections.Load(c)
		if ok {
			error:=D.(TLSConn).Close()
			if error != nil{
				log.Printf("Error Closing connection : \n\t%s\n",error)
			}
			
		}
		C:=conn.(TLSConn)
		error:=C.Close()
		if error != nil {
			logger.Debug("Error Closing tls connectio",slog.Any("error",error))
		}
		
		logger.Warn("Removing Connection ",slog.String("ReamoteAddr",C.RemoteAddr().String()))
		connections.Delete(c)
	
	}
	

	return gnet.Close
}
func (es *tcpServer) React(frame []byte, c gnet.Conn) (out []byte,action gnet.Action){
	// Write Data to Cahannel
	//mu.Lock()
	
	data:=conn_channel{
		Conn: c,
		data: frame,
		outdx: make( chan outputdx),
	}
	// Send data to io_thread
	connChan<-data
	logger.Debug("Exiting from reactor")
	return
	
}

// TLS Reader asyncronously manage TLS connection
func tls_reader(thread int ,connectionTLS TLSConn ,conn gnet.Conn){
	log.Printf("Thread: %d - Reading from tls\n",thread)
	// Reduce race conditions on socket management
	// mu.Lock()
	// defer mu.Unlock()
	
	for{
		logger.Debug("",slog.Int("Thread",thread),"- tls_reader - Managing data on Context",slog.Any("sessionID",connectionTLS.ctx.Value("sessionID")))
		// Set DeadLine On TLS
		connectionTLS.SetDeadline(time.Now().Add(deadline))
		connectionTLS.SetWriteDeadline(time.Now().Add(deadline))
		connectionTLS.SetReadDeadline(time.Now().Add(deadline))
		
		
		logger.Debug("",slog.Int("Thread",thread),"- tls_reader - Reading from tls","")
		// Read TLS connection and forward data tu Client connection
		data := make([]byte, 4096)
		size,err:=connectionTLS.Read(data)
		//log.Printf("Thread: %d - tls_reader - Readed %d bytes from tls",thread,size)
		
		logger.Debug("",slog.Int("Thread",thread),"- tls_reader - Readed",slog.Int("bytes",size)," from tls","")
		if err == io.EOF {
			// Closed Socket
			//log.Printf("Thread: %d - tls_reader - reading closed socket tls: %s",thread,err)
			logger.Debug("Thread: %d - tls_reader - reading closed socket tls: %s",thread,err)
			return	
		} else {
			if err == nil {
				// Write data to Client
				//log.Printf("Thread: %d - tls_reader - Sending %d bytes to Client",thread,size)
				logger.Debug("Thread: ",thread," - tls_reader - Sending ",size," bytes to Client")
				error:=conn.AsyncWrite(data[:size])
				if error != nil {
					//log.Printf("\nThread: %d - tls_reader - Error writing to client:\n\t%s\n",thread,error)
					logger.Debug("",slog.Int("Thread",thread)," - tls_reader - Error writing to client:\n\t",error)
					return
				} else {
					//log.Printf("\nThread: %d - tls_reader - Written datas\n",thread)
					logger.Debug("",slog.Int("Thread",thread)," tls_reader - Written datas","\n")
				}

			} else{
				// Timeout Reaced
				//log.Printf("Thread: %d - tls_reader - IO ERROR reading socket tls: %s",thread,err)
				logger.Warn("",slog.Int("Thread",thread)," - tls_reader - reading socket tls",err)
				conn.AsyncWrite(data[:size])
				return	
			}
		}
		if err != nil {
			//log.Printf("Thread: %d - tls_reader - IO ERROR reading socket tls: %s",thread,err)
			logger.Warn("",slog.Int("Thread",thread)," - tls_reader - reading socket tls",err)
			conn.AsyncWrite(data[:size])
			return			
		}
	}
}
type TLSConn struct{
	*tls.Conn
	ctx context.Context
}
// IO Thread for client Connection
func io_thread(conChannel chan conn_channel,id int){
	thread:=id
	//var connectionTLS *tls.Conn
	var connectionTLS TLSConn
	// Take client requests and manage it
	for conn := range conChannel{
		logger.Debug("Contexts dump ",slog.Any("ctx",ctx))
		//mu.Lock()
	    retrievedTLSConn, ok := connections.Load(conn.Conn)
		if ok {
			// TLS connection allready establisched
			//log.Printf("Thread: %d - TLS connection founded % v",thread,conn.Conn)
			logger.Debug("",slog.Int("Thread",thread),"- TLS connection founded","\n")
			//connectionTLS=retrievedTLSConn.(*tls.Conn)
			connectionTLS=retrievedTLSConn.(TLSConn)
		} else {
			// Take a TLS connection 
			logger.Debug("",slog.Int("Thread",thread)," - ","io_thread - gnet.Conn not found in map")
	
			tlsConnection,err:=Connect()
			if err != nil {
				// Failed to take a connection client socket will be closed
				logger.Debug("",slog.Int("Thread",thread)," - ","io_thread - Error Creating tls Connections, closign Client")
				conn.Conn.Close()
				continue
			}
	
			// Save to map
			mu.Lock()
			connectionTLS.Conn=tlsConnection
			//Generate sessionID
			sessionID := uuid.New().String()
			connectionTLS.ctx=context.WithValue(ctx,"sessionID", sessionID)
			connections.Store(conn.Conn,connectionTLS)
			mu.Unlock()
			go tls_reader(thread,connectionTLS,conn.Conn)
		}
			//Reading TLS
		
		// Write to TLS 
		size,err:=connectionTLS.Write(conn.data)
		if err != nil {
			// If write failed reestablish connection and retry
			logger.Debug("",slog.Int("Thread",thread)," - ","io_thread - Error writnig tls:",slog.Any("error",err))
			connections.Delete(conn.Conn)
			tlsConnection,err:=Connect()
			if err != nil {
				logger.Warn("",slog.Int("Thread",thread)," - ","io_thread - Error Creating tls Connections, closign Client")
				conn.Conn.Close()
				continue
			} else {
				mu.Lock()
				connectionTLS.Conn=tlsConnection
				//Generate sessionID
				sessionID := uuid.New().String()
				connectionTLS.ctx=context.WithValue(ctx,"sessionID", sessionID)
				connections.Store(conn.Conn,connectionTLS)
				mu.Unlock()		
				go tls_reader(thread,connectionTLS,conn.Conn)	
				size,err:=connectionTLS.Write(conn.data)	
				if err != nil {
					logger.Warn("",slog.Int("Thread",thread)," - ","io_thread - retry 2 Error writnig tls:",slog.Any("error",err))
					conn.Conn.Close()
				} else {
					log.Printf("Thread: %d - Sent %d bytes to tls",thread,size)
					logger.Debug("",slog.Int("Thread",thread)," - ","io_thread - Sent ",slog.Int("bytes",size),"to tls","" )
				}
			}				
			} else {
			logger.Debug("",slog.Int("Thread",thread)," io_thread - Sent ",slog.Int("Bytes",size),"to tls","")
			}
		
	}
}

var(
 // Map of connections
 connections sync.Map 
 // 
 connChan chan conn_channel 
 io_thread_pool_size=1
 mu sync.Mutex
 deadline time.Duration
 keepalive time.Duration
)

func io_thread_pool( channel chan conn_channel){
	io_thread_pool_size=cf.Poolsize
	for i := 0; i < io_thread_pool_size; i++ {
		go io_thread(channel,i) 
	} 

}
type Netcfg struct {
	Deadline time.Duration `yaml:"deadline"`
	Keepalive time.Duration `yaml:"keepalive"`
}
type Tlscfg struct{
	Certfile string `yaml:"certfile"`
	Keyfile  string `yaml:"keyfile"`
	Proxyhost string `yaml:"proxyhost"`
} 
type Conf struct{
	Loglevel string `yaml:"loglevel"`
	Poolsize int `yaml:"poolsize"`
	Tlscfg Tlscfg `yaml:"tlscfg"`
	Netcfg Netcfg `yaml:"netcfg"`
}
var cf Conf

func (c *Conf) ReadConfig() {

	 yamlFile, err := os.ReadFile("config.yaml")
	 if err != nil {
		log.Panicf("Error readinf fil %s",err)
	}
	//cf:=make(map[string]interface{})
	
	error:=yaml.Unmarshal(yamlFile,&cf)
	if error != nil {
		log.Printf("Errore in Marshal %s\n",error)
	}
	log.Printf("Values %v\n",cf)
}
var logger *slog.Logger
var logLevel slog.Level
func main(){
	
	var config Conf
	config.ReadConfig()
    log.Printf("Config -> :%v",config)
	deadline=cf.Netcfg.Deadline
	keepalive=cf.Netcfg.Keepalive
	connChan = make(chan conn_channel)
	io_thread_pool(connChan)
	PrxConnected=false
	ManageConnect=true
	echo := new(tcpServer)
    if err := logLevel.UnmarshalText([]byte(cf.Loglevel)); err != nil {
        // Handle error, e.g., fallback to a default log level
        logLevel = slog.LevelInfo
    }
	logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))


	//(gnet.Serve(echo, "tcp://:8080", gnet.WithMulticore(true)))
	
    err:=gnet.Serve(echo, "tcp://:8080", gnet.WithMulticore(true),gnet.WithTCPKeepAlive(keepalive))
	log.Fatalf("Error:%s\n",err)

}
