#!/usr/bin/python

"""Custom topology example

Two directly connected switches plus a host for each switch:

            ==========         ===============         ===========
            |        |         |             |         |         |
            | client |---s1 ---|eth1 r0  eth3| ---s2---|  server |
            |        |         |             |         |         |
            |        |         |    eth2     |         |         |
            ==========         ===============         ===========
                                     |
                                     |                 =============
                                     |                 |           |
                                     ---------s3-------|   target  | 
                                                       |           |
                                                       =============


Adding the 'topos' dict with a key/value pair to generate our newly defined
topology enables one to pass in '--topo=mytopo' from the command line.
"""

from mininet.topo import Topo
from mininet.net import Mininet
from mininet.node import Node
from mininet.log import setLogLevel, info
from mininet.cli import CLI

class LinuxRouter( Node ):
    "A Node with IP forwoarding enabled."

    def config( self, **params ):
        super( LinuxRouter, self ).config( **params )
        # Enable forwarding on the router
        self.cmd( 'sysctl net.ipv4.ip_forward=1' )

    def terminate( self ):
        self.cmd( 'sysctl net.ipv4.ip_forward=0' )
        super( LinuxRouter, self ).terminate()

class SimTopo( Topo ):
    "tinyvpn's simulation network."

    def build( self, **_opts ):

        router = self.addNode('r0', cls=LinuxRouter, ip='10.0.1.1/24' )

        # Add three switches to connect to r0.
        s1, s2, s3 = [ self.addSwitch( s ) for s in 's1', 's2', 's3' ]
        self.addLink( s1, router, intfName2='r0-eth1',
                      params2={ 'ip' : '10.0.1.1/24' } )
        self.addLink( s2, router, intfName2='r0-eth2',
                      params2={ 'ip' : '10.0.3.1/24' } )
        self.addLink( s3, router, intfName2='r0-eth3',
                      params2={ 'ip' : '10.0.5.1/24' } )

        # Add hosts with IP config 
        client = self.addHost( 'client', ip="10.0.1.100/24",
                               defaultRoute="via 10.0.1.1")
        server = self.addHost( 'server', ip="10.0.3.100/24", 
                               defaultRoute="via 10.0.3.1")
        target = self.addHost( 'target', ip="10.0.5.100/24", 
                               defaultRoute="via 10.0.5.1")

        # Add links between hosts and switches.
        for h, s in [ (client, s1), (server, s2), (target, s3) ]:
            self.addLink(h, s)

def run():
    "Test tinyvpn simulation network"
    topo = SimTopo()
    net = Mininet( topo=topo )
    net.start()
    info( '*** Routing table on Router:\n' )
    print net[ 'r0' ].cmd( 'route' )
    CLI( net )
    net.stop()

if __name__=='__main__':
    setLogLevel( 'info' )
    run()

topos = {'SimTopo': ( lambda: SimTopo() ) }
