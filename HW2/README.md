# Homework 2: Hands-on with Hazelcast Database

Author: [**Shevchenko Ivan**](https://github.com/ishevche) 

GitHub: [https://github.com/ishevche/Software-Architecture/tree/hazelcast-basics/HW2](https://github.com/ishevche/Software-Architecture/tree/hazelcast-basics/HW2)

## Description

In this homework, I have created a Hazelcast cluster consisting of three nodes.
After that, I have tested the work of a distributed map by filling it with items,
simultaneously updating a value from three routines using different synchronization
methods for that. Lastly, I tested the work of the bounded queue based on the
distributed queue.

All the client side code is written using Golang, and the Hazelcast cluster is
wrapped in using Docker compose.

## Usage

To run the application do the following steps:

1. Clone the repository
2. Move to the project root directory
3. Run `docker compose up` to start the Hazelcast cluster alongside
   the Manager Center. This will do the following:
    - Create the dedicated network called `hw2_shevchenko_network`.
    - Start three Hazelcast instances in containers `hw2_shevchenko_node1`,
      `hw2_shevchenko_node2`, and `hw2_shevchenko_node3` connected to the
      created network. Nodes can be accessed on the ports `:5701`, `:5702`,
      and `:5703` respectively. The resulting cluster is named
      `hw2_shevchenko_hazelcast`.
    - Start the Manager Center in container `hw2_shevchenko_ui`, that can
      be accessed on the port `:8080`, so by the URL `http://localhost:8080`.
4. Run one of Golang scripts.

### Golang scripts:

In this homework, I have created three scripts using Golang for manipulations
with Hazelcast cluster. They are the following ones:

1. `insert.go` - Inserts 1000 key-value pairs in the specified distributed
   map of the specified Hazelcast cluster.
   ````
   Usage: insert [ <cluster_name> [ <map_name> ]]:
      <cluster_name>    Name of the Hazelcast cluster to connect 
                        (default: hw2_shevchenko_hazelcast) 
      <map_name>        Name of the distributed map to insert values into
                        (default: map)
   ````

2. `increment.go` - Sets the value of the specified value in the specified
   distributed map of the specified Hazelcast cluster to `0` and increments
   the value `10'000` times from 3 goroutines simultaneously. Three different
   methods of synchronization are used:
    - `no blocking` - no synchronization
    - `pessimistic blocking` - before incrementing each goroutine locks the
      key in the map and unlocks after increment is done.
    - `optimistic blocking` - each goroutine reads the value, increments, and
      writes it back only if the value is still the same as it read before,
      otherwise operation is repeated
    ````
     Usage: increment [ <cluster_name> [ <map_name> [ <key_name> ]]]:
        <cluster_name>    Name of the Hazelcast cluster to connect 
                          (default: hw2_shevchenko_hazelcast) 
        <map_name>        Name of the distributed map where value is incremented
                          (default: map)
        <key_name>        Key, which value is set and incremented
                          (default: key)
    ````

3. `queue.go` - Launches three goroutines. The first one puts values `1..100`
   in the specified queue of the specified Hazelcast cluster. At the end `-1`
   is put as a poison pill for consumers. Two other coroutines are reading
   values from the queue and save them in the local list. If the value `-1`
   is read, it is returned in the queue and consumer dies printing all
   values it consumed overall. After the exit of all goroutines, the queue
   is cleared.
   ````
   Usage: queue [ <cluster_name> [ <queue_name> ]]:
      <cluster_name>    Name of the Hazelcast cluster to connect 
                        (default: hw2_shevchenko_hazelcast) 
      <queue_name>      Name of the distributed queue to use
                        (default: queue)
   ````

## Results

First of all, I have started the cluster:

![Cluster is started][start_up]

After that, I have launched `insert` (Golang library for Hazelcast prints
some debug info in the terminal):

![Insertion in the cluster][insertion]

In the Manager Center we can see that map is created and contains `1000`
keys:

![Map is created][map_created]

Also, we can observe the distribution of keys in the nodes:

![Keys distribution][keys_distribution]

Now I will kill one of the nodes. As Hazelcast is a distributed system,
it has mechanisms to prevent data losses during such events. In this case,
data was cached on other nodes, so it can be restored without any loss:

![Data distribution after single node loss][first_node_loss]

We can repeat the same by killing one more node. Data still won't be lost,
as nodes had plenty of time to replicate the data to the corresponding
caches:

![Data distribution after second loss][second_node_loss]

This won't be so unfortunately if two nodes are killed almost simultaneously,
so they won't have enough time to replicate data, so after restarting and
refilling the cluster, I have tried to kill two nodes at the same time.
As expected, some data was lost:

![Data distribution after two simultaneous losses][double_node_loss]

To tackle this issue, we should not kill nodes, but shutdown them
gracefully, then, according to the documentation, nodes should send
the missing data if there is such.

After that, I have restarted the cluster one more time and launched
`increment` script, omitting all debug info printed by library:

![Output of the `increment` script][increment]

On the screenshot, we can see that all three methods have successfully
completed.

The first method (`no blocking`) produces the wrong result. The reason
is the time needed to fetch the value from the Database, increment it
and write it back. During this time, another client could update the value,
and we result in overwriting it. So in the end, we end up with the wrong
number.

The second method (`pessimistic blocking`) locks the key, reads the value,
increments it, writes the new one, and then unlocks the key. As the key can
be locked only once, using this method ensures that no updates would be lost.
And that is what we can see, as the resulting number is `30'000`.

The third method (`optimistic blocking`) works similarly to the first one.
The only difference is that new value is not written straight away in the map,
but only if the current value in the map equals the one we read. If it does not,
it repeats all steps (read, increment, try to write) until it succeeds. So
actually it performs atomic operation `compare-and-swap` or `CAS`.

I have also measured time taken for all methods to work. All times could be seen
on the screenshot. We can see that the fasted method is the one with `no blocking`,
what is expected, as there is no overhead for synchronization there. The idea was
to compare two blocking methods. In this case, we can see that `optimistic` one
works faster then `pessimistic`. The idea behind the first one is to repeat
operation on the object until it succeeds, and behind the second one is to wait
until someone else is done updating the object. In our case, operation is
straightforward - increment a value by one, so it is not so costly to repeat it
over and over. In case of more challenging update logic using locks could be
faster as repeating update could take much more time.

Lastly, I have launched `queue` script and it resulted in the following:

![Output of the `queue` script][queue]

We can see that all `100` values where consumed by the consumers, and moreover
they separated the values in half, so each one handled only `50`. This could help
for systems with high workload to distribute tasks between different instances of
the same program launched on the different servers to increase throughput.

In the `compose.yaml` that is responsible for cluster creation I have specified
the environmental variable for each node: `HZ_QUEUE_BOUNDEDQUEUE_MAXSIZE=10`.
This set a bound for queue `boundedqueue` to have maximum 10 elements in it.
So now I have commented the code responsible for launching the consumers, and
launched it one more time. The program never exited:

![Running `queue` script with only producer][producer_queue]

And in the Manager Center we can see the reason of this:

![Number of element in the queue][manager_center_producer_queue]

So the producer put 10 elements in the queue, so it filled up. Then it tried to
put the next one. However, Hazelcast didn't allow it to do so. As specified in
the documentation, code just fell into sleep, waiting for a free space to appear,
and as there are no consumers launched, it will wait forever, what we can actually
observe.


[start_up]: img/cluster_startup.png "Cluster is started"
[insertion]: img/insert.png "Insertion of the data in the cluster"
[map_created]: img/map_created.png "Map is created and populated"
[keys_distribution]: img/keys_distribution.png "Distribution of keys between nodes"
[first_node_loss]: img/first_loss.png "Keys distribution after loss of single node"
[second_node_loss]: img/second_loss.png "Keys distribution after loss of second node"
[double_node_loss]: img/double_loss.png "Keys distribution after loss of two nodes simultaneously"
[increment]: img/increment.png "Run of the increment script"
[queue]: img/queue.png "Run of the queue script"
[producer_queue]: img/queue_put.png "Run of the queue script only with only producer"
[manager_center_producer_queue]: img/queue_put_values.png "Values in the queue upon producer work"
