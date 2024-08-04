arrivalRate=$1
round_trip_time=$2

replica_path="replica/bin/replica"
ctl_path="client/bin/client"
output_path="logs/"

go build -v -o ./client/bin/client ./client/
go build -v -o ./replica/bin/replica ./replica/

rm -r ${output_path}
mkdir -p ${output_path}

pkill replica; pkill replica; pkill replica; pkill replica; pkill replica
pkill client; pkill client; pkill client; pkill client; pkill client

echo "Killed previously running instances"

nohup ./${replica_path} --name 1 --debugOn --debugLevel 0 --roundTripTime "${round_trip_time}"  --logFilePath ${output_path} >${output_path}1.log &
nohup ./${replica_path} --name 2 --debugOn --debugLevel 0 --roundTripTime "${round_trip_time}"  --logFilePath ${output_path} >${output_path}2.log &
nohup ./${replica_path} --name 3 --debugOn --debugLevel 0 --roundTripTime "${round_trip_time}"  --logFilePath ${output_path} >${output_path}3.log &
nohup ./${replica_path} --name 4 --debugOn --debugLevel 0 --roundTripTime "${round_trip_time}"  --logFilePath ${output_path} >${output_path}4.log &
nohup ./${replica_path} --name 5 --debugOn --debugLevel 0 --roundTripTime "${round_trip_time}"  --logFilePath ${output_path} >${output_path}5.log &

echo "Started 5 replicas"

sleep 5

./${ctl_path} --name 51 --logFilePath ${output_path} --requestType status --operationType 1  --debugOn --debugLevel 0 >${output_path}status1.log

sleep 5

echo "sent initial status"

sleep 10

echo "starting clients"

nohup ./${ctl_path} --name 51 --logFilePath ${output_path} --requestType request --debugOn --debugLevel 0 --arrivalRate "${arrivalRate}"  >${output_path}51.log &
nohup ./${ctl_path} --name 52 --logFilePath ${output_path} --requestType request --debugOn --debugLevel 0 --arrivalRate "${arrivalRate}"  >${output_path}52.log &
nohup ./${ctl_path} --name 53 --logFilePath ${output_path} --requestType request --debugOn --debugLevel 0 --arrivalRate "${arrivalRate}"  >${output_path}53.log &
nohup ./${ctl_path} --name 54 --logFilePath ${output_path} --requestType request --debugOn --debugLevel 0 --arrivalRate "${arrivalRate}"  >${output_path}54.log &
nohup ./${ctl_path} --name 55 --logFilePath ${output_path} --requestType request --debugOn --debugLevel 0 --arrivalRate "${arrivalRate}"  >${output_path}55.log &

sleep 10

# python3 integration-test/python/crash-recovery-test.py ${output_path}/1.log ${output_path}/2.log ${output_path}/3.log ${output_path}/4.log ${output_path}/5.log > ${output_path}crash_recovery.log


sleep 120

echo "finished running clients"


nohup ./${ctl_path} --name 51 --logFilePath ${output_path} --requestType status --operationType 2  --debugOn --debugLevel 0 >${output_path}status2.log &


echo "sent status to print logs"

sleep 30

pkill replica; pkill replica; pkill replica; pkill replica; pkill replica
pkill client; pkill client; pkill client; pkill client; pkill client


python3 integration-test/python/overlay-test.py ${output_path}/1-consensus.txt ${output_path}/2-consensus.txt ${output_path}/3-consensus.txt ${output_path}/4-consensus.txt ${output_path}/5-consensus.txt > ${output_path}consensus_correctness.log

echo "Killed instances"

echo "Finish test"
