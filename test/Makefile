MININET_NAME = tinyvpn_mininet
TCPDUMP_NAME = tinyvpn_tcpdump
CMD_START_OVS = service openvswitch-switch start
CMD_START_MN = sudo mn --custom /projects/tinyvpn_topo.py --topo SimTopo

build: docker-mininet docker-tcpdump
	cd docker-mininet; docker build -t docker-mininet . 
	cd docker-tcpdump; docker build -t docker-tcpdump .

docker-mininet:
	git clone https://github.com/jingqiuELE/docker-mininet.git

docker-tcpdump:
	git clone https://github.com/jingqiuELE/docker-tcpdump.git

run: clean
	docker run --privileged -d -it --name $(MININET_NAME) docker-mininet 
	docker cp tinyvpn_topo.py $(MININET_NAME):/projects/
	docker exec $(MININET_NAME) $(CMD_START_OVS)
	docker run --rm --net=container:$(MININET_NAME) --name $(TCPDUMP_NAME) docker-tcpdump -i any --immediate-mode -w - | wireshark -k -i - &
	docker exec -it $(MININET_NAME) $(CMD_START_MN) 

clean:
	@docker rm -f $(TCPDUMP_NAME) 2>/dev/null || true
	@docker rm -f $(MININET_NAME) 2>/dev/null || true
