MININET_NAME = tinyvpn_mininet
TCPDUMP_NAME = tinyvpn_tcpdump
CMD_START_OVS = service openvswitch-switch start
CMD_START_MN = sudo mn --custom /projects/tinyvpn_topo.py --topo SimTopo --link tc1

GO_BIN = ~/Projects/golang/bin
SERVER = ../petrel/petrel
CLIENT = ../petrelc/petrelc
SERVER_START = ./server-start.sh
CLIENT_START = ./client-start.sh

build: docker-mininet docker-tcpdump
	cd docker-mininet; docker build -t docker-mininet . 
	cd docker-tcpdump; docker build -t docker-tcpdump .

docker-mininet:
	git clone https://github.com/jingqiuELE/docker-mininet.git

docker-tcpdump:
	git clone https://github.com/jingqiuELE/docker-tcpdump.git

run: clean start_container
	docker exec $(MININET_NAME) mkdir /projects/tinyvpn
	docker cp $(SERVER) $(MININET_NAME):/projects/tinyvpn/
	docker cp $(CLIENT) $(MININET_NAME):/projects/tinyvpn/
	docker cp $(SERVER_START) $(MININET_NAME):/projects/tinyvpn/
	docker cp $(CLIENT_START) $(MININET_NAME):/projects/tinyvpn/
	docker exec -it $(MININET_NAME) $(CMD_START_MN) 

start_container:
	docker run --privileged --cap-add=ALL -v /lib/modules:/lib/modules -d -it --name $(MININET_NAME) docker-mininet 
	docker cp tinyvpn_topo.py $(MININET_NAME):/projects/
	docker exec $(MININET_NAME) $(CMD_START_OVS)
	docker run --rm --net=container:$(MININET_NAME) --name $(TCPDUMP_NAME) docker-tcpdump -i any --immediate-mode -w - | wireshark -k -i - &

clean:
	@docker rm -f $(TCPDUMP_NAME) 2>/dev/null || true
	@docker rm -f $(MININET_NAME) 2>/dev/null || true
