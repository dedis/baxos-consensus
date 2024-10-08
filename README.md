# Baxos Consensus

This repository implements [Baxos Consensus](https://arxiv.org/abs/2204.10934).

Note: This repository deviates from the prototype used in the [Baxos Paper](https://arxiv.org/abs/2204.10934) in 
several ways: (1) this repository does not implement pipelining of 2 phases of Baxos,
(2) this repository implements client side batching, and (3) this repository does not use gRPC.

```Baxos``` is a variant of ```Paxos```. 
In ```leader-based Paxos``` (a.k.a. ```Multi-Paxos```), a leader is elected and all the proposals are initiated by the leader.
In contrast, in ```Baxos```, each replica acts as a leader and proposes simultaneously.
When replicas experience contention, in the event when more than 1 proposer is proposing at the same time, ```Baxos``` uses
random exponential backoff to avoid further collisions.

This repository uses [Protocol Buffers](https://developers.google.com/protocol-buffers/).
It requires the ```protoc``` compiler with the ```go``` output plugin installed.

This repository uses [Redis](https://redis.io/topics/quickstart) and it should be installed with default options.

All implementations are tested in ```Ubuntu 20.04.3 LTS```

run ```go get -u github.com/golang/protobuf/protoc-gen-go``` and ```go get -u google.golang.org/grpc``` to install ```protobuff``` and ```grpc```

run ```go mod vendor``` to install dependencies

run ```go build -v -o ./client/bin/client ./client/``` and ```go build -v -o ./replica/bin/replica ./replica/``` to build the client and the replica

All the commands to run replicas and the clients are available in the ```integration-test/``` directory