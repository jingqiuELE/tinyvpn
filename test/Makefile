MININET_NAME = tinyvpn_mininet
CMD_START_OVS = service openvswitch-switch start
CMD_START_MN = sudo mn --custom /projects/tinyvpn_topo.py --topo SimTopo --link tc1

GO_BIN = ~/Projects/golang/bin
EXECUTABLE = ../petrel/petrel
EXECUTABLE_SRC = ../petrel/*.go

WORK_DIR = /projects/tinyvpn
SCRIPTS = ./scripts
KEY = ./keys

build: docker-mininet docker-tcpdump
	cd docker-mininet; docker build -t docker-mininet .
	cd docker-tcpdump; docker build -t docker-tcpdump .

docker-mininet:
	git clone https://github.com/jingqiuELE/docker-mininet.git

docker-tcpdump:
	git clone https://github.com/jingqiuELE/docker-tcpdump.git

run: clean start_container $(EXECUTABLE)
	docker exec $(MININET_NAME) mkdir -p $(WORK_DIR)
	docker cp $(EXECUTABLE) $(MININET_NAME):$(WORK_DIR)
	docker cp $(SCRIPTS) $(MININET_NAME):$(WORK_DIR)
	docker cp $(KEY) $(MININET_NAME):$(WORK_DIR)
	docker exec -it $(MININET_NAME) $(CMD_START_MN)

$(EXECUTABLE): $(EXECUTABLE_SRC)
	cd ../petrel; go build

start_container:
	docker run --privileged --cap-add=ALL -v /lib/modules:/lib/modules -d -it --name $(MININET_NAME) docker-mininet
	docker cp tinyvpn_topo.py $(MININET_NAME):/projects/
	docker exec $(MININET_NAME) $(CMD_START_OVS)

clean:
	@docker rm -f $(MININET_NAME) 2>/dev/null || true
