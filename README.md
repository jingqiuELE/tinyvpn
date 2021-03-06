TinyVPN
=======
Code named petrel.

Petrel Architecture
-------------------
### Server:


                         Set      +----------------------+
                      +----------->    secretMap         |    +----------+
                      |           +----------+-----------+    |   tun    |
                      |           |  session |  session  |    +----+-^---+
                +-----+-------+   |   Index  |  Secret   |         | |
       +------> |  AuthServer |   +----------+-----------+        Data
       |        +-------------+               |Get                 | |
     authData                       +-----------------+      +-----v-+-----+
       |        +-------------+ eIn | +-------v-----+ | pIn  |             |
       |        |             +-----> |  Decryption +-------->             |
    +--v--+ data| ConnServer  |     | +-------------+ |      |  bookServer |
    | Wan <-----> (TCP,UDP)   | eOut| +-------------+ | pOut |             |
    +-----+     |             <-----+ |  Encrytion  <--------+             |
                +--------^----+     | +-------------+ |      |             |
                         |          +-----------------+      +------^------+
                      Set|Get                                    Set|Get
                +--------v-------------+  +---------------------------------+
                |      connMap         |  | +-----------------------v-----+ |
                +---------+------------+  | |      ipToSession            | |
                | session | Connection |  | +-------------+---------------+ |
                |  Index  |            |  | | ip.String() | session.Index | |
                +---------+------------+  | +-----------------------------+ |
                                          | +-----------------------------+ |
                                          | |    sessionToIp              | |
                                          | +---------------+-------------+ |
                                          | | session.Index | ip.String() | |
                                          | +-----------------------------+ |
                                          +---------------------------------+

#### Logic Layers
1. AuthServer: 
    1. Exchange username and password for session key
    2. Put session key into client-session map
2. ConnServer: 
    1. TCP or UDP socket listener, accepts connection/packets from the client
    2. Extract session key to create / load corresponding client, maintains client-session map
    3. Put packets into a eIn channel
    4. Accepts packets from eOut channel, lookup the corresponding client connection and return the packets
3. EncryptServer: 
    1. Decrypt the packet content from eIn channel to pIn channel. 
    2. Encrypt the packets from pOut to eOut channel.
4. bookServer: 
    1. Foward the decrypted packet from pIn channel to the tun interface
    2. Maintains IP-session map
    3. Accepts packets coming from the tun interface to pOut channel

### Client:        

    
                                                             +--------+
                                                             |  apps  |
                +---------+      session                     +---^-+--+
       +------> |  Auth   +-------------------+                  | |
       +        +---------+                   |                 Data
     auth                           +-----------------+          | |
       +        +-------------+ eIn | +-------v-----+ | pIn  +---+-v----+
       |        |             +-----> |  Decryption +-------->          |
    +--v--+ data| ConnServer  |     | +-------------+ |      |   tun    |
    | Wan <-----> (TCP,UDP)   | eOut| +-------------+ | pOut |          |
    +-----+     |             <-----+ |  Encrytion  <--------+          |
                +-------------+     | +-------------+ |      +----------+
                                    +-----------------+


### Common Data Structures

##### Packet
````golang
/* Fixed-size PacketHeader to ease marshal and unmarshal */
type Iv [IvLen]byte
type PacketHeader struct {
	Iv  Iv
	Sk  session.Index
	Len uint16
}

type Packet struct {
    Header PacketHeader
    Data []byte
}
````

##### Session
````golang
type Connection Interface {
    writePacket(p Packet)
}

type Session struct {
    conn Connection
    secret []byte
}
````

##### Sessionkey-Session Map
A map with sessionkey as the key, and Session as the value. When first
authenticated, a sessionkey is placed in the map, with nil as the value. Until
the first connection from the client with the corresponding sessionkey is made,
the newly created client is placed into the map.

##### IP-Session Map
A bi-directional map facilitates lookups using sessionkey and IP address as
keys. A better than iterate through the list way of handling boardcast packets
needs to be find.

### Usage
  * Petrel contains two programs: _petrel_ as the server, and _petrelc_ as the client.
  * Start the server:
      * Please refer to _test/scripts/server-start.sh_
  * Start the client:
      * Please refer to _test/scripts/client-start.sh_
  
### Testing

##### Goal
  The test would run **petrel** in a simulated network created with mininet running in Docker container. The purpose of this approach is to test the program in a sandbox, changable networking environment.
  The topology of the network is:
 
  
     10.0.1.100                                                    10.0.3.100
     ==========                 ===============                   ===========
     |        |                 |             |10.0.3.1           |         |
     | client |-------s1--------|eth1 r0  eth3|----------s2-------|  server |
     |        |        10.0.1.1 |             |                   |         |
     |        |                 |    eth2     |                   |         |
     ==========                 ===============                   ===========
                                      |10.0.5.1                
                                      |                            10.0.5.100
                                      |                           =============
                                      |                           |           |
                                      -------------s3-------------|   target  | 
                                                                  |           |
                                                                  =============
##### Dependency
  The test program relies on:
  * Docker
  * Wireshark
  * Only supports Linux operating system.

##### How to run
  In the test directory, please follow below steps:
  * Change mininet topo to suit your needs. 
  * Build the docker container of mininet with either of below commands:
      * $make build
      * $docker pull jingqiu/docker-mininet-auto
  * $make run    
  * You should be able to see both wireshark and a mininet console. In the mininet console, run below commands to start **petrel**:
    * mininet>server ./scripts/server-start.sh
    * mininet>client ./scripts/client-start.sh
    * You should be able to observe the connection established.
  * The network qos can be adjusted in the Makefile, by changing $(CMD\_START\_MN) with qos settings.

##### How to observe
  * Start wireshark to observe traffic of the mininet network (You need to have wireshark installed):
    * ./test/start\_capture.sh  
    Please note that there will be two wireshark window opened. One for the traffic to/from target, the other for the traffic to/from the server. This approach tries to avoid duplication of packets reported on mininet switches.
  * Open another console of the mininet container:
    * $docker exec -it tinyvpn\_mininet bash
  * Observe the network status of each host:
    * mininet>client ip route
    * mininet>client ifconfig -a
    * mininet>client ping 10.0.5.100
  * You can filter the packets in Wireshark by adding filers:
    For example:
    ip.src == 10.0.1.100
  * You can try _nc_ command and _iperf_ command to do bench mark test. Please note that the link traffic control can be set in mininet's config script:  
    tinyvpn_topo.py  
    Please read mininet mannual for the details. The example test scripts can be found in:  
      * test/scripts/nc_test
      * test/scripts/iperf_test

