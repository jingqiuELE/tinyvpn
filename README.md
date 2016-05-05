TinyVPN
=======
Code named petrel.

Server Architecture
-------------------
![Architecture Diagram](TinyVPN.png)
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
````
type Packet struct {
    iv [4]byte
    session [6]byte // Session key, used as MAC address for tun device.
    length uint16
    data []byte
}
````

##### Session
````
type Connection Interface {
    readPacket() Packet
    writePacket(p Packet)
}

type Session struct {
    c Connection
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
