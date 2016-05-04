MININET_NAME = tinyvpn_mininet 
TCPDUMP_NAME = tinyvpn_tcpdump

build: docker-mininet docker-tcpdump

docker-mininet:
	git clone https://github.com/jingqiuELE/docker-mininet.git
	cd docker-mininet; docker build -t docker-mininet . 

docker-tcpdump:
	git clone https://github.com/jingqiuELE/docker-tcpdump.git
	cd docker-tcpdump; docker build -t docker-tcpdump .

run: clean
	docker run --privileged -d -it --name $(MININET_NAME) docker-mininet 
	docker run --rm --net=container:$(MININET_NAME) --name $(TCPDUMP_NAME) docker-tcpdump -i any --immediate-mode -w - | wireshark -k -i -

clean:
	@docker rm -f $(TCPDUMP_NAME) 2>/dev/null || true
	@docker rm -f $(MININET_NAME) 2>/dev/null || true
