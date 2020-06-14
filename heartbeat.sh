while true
do
	echo -n "hello" | nc -q1 $1 $2
	sleep 30	
done
