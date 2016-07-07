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

### Logic Layers
1. Authentication Service: 
    1. Exchange username and password for session key
    2. Put session key into client-session map
2. Connection Layer Listener: 
    1. TCP or UDP socket listener, accepts connection/packets from the client
    2. Extract session key to create / load corresponding client, maintains client-session map
    3. Put packets into a encrypted packets channel
    4. Accepts packets from packets return channel, lookup the corresponding client connection and return the packets
3. Decryption/Encryption Layer: 
    1. Decrypt the packet content from packets channel to tun
    2. Encrypt the packets from tun to return packets channel
4. Packet forward layer: 
    1. Foward the decrypted packet to the tun interface
    2. Maintains IP-session map
    3. Accepts packets coming from the tun interface

### Data Structures

##### Packet
````golang
type Packet struct {
    iv [4]byte
    session [6]byte // Session key, used as MAC address for tun device.
    length uint16
    data []byte
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
  * $make build
  * $make run    
  * You should be able to see both wireshark and a mininet console. In the mininet console, run below commands to start **petrel**:
    * mininet>server ./tinyvpn/server-start.sh
    * mininet>client ./tinyvpn/client-start.sh
    * You should be able to observe the connection established.
  * The network qos can be adjusted in the Makefile, by changing $(CMD_START_MN) with qos settings.

##### How to observe
  * Open another console of the mininet container
    * $docker attach tinyvpn_mininet
  * Observe the network status of each host
    * mininet>client ip route
    * mininet>client ifconfig -a
  * You can filter the packets in Wireshark by adding filers.
    For example:
    ip.src == 10.0.1.100
  


