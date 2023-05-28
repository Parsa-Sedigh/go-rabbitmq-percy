Monolith vs event-driven arch:

In monolith, we have coupling and we have to re-deploy the whole monolith. But in event-driven arch, we leverage microservices and
these microservices push events to each other using events and rabbitMQ can deliver these events between the microservices.

The easiest way to start up rabbitmq is using docker.

Run an instance of rabbitmq:
```shell
# 5672 port is the amqp port for connections
# 15672 is the port used by admin(management) UI for rabbitmq
docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3.11-management
```

Now go to `localhost:15672` to see the admin UI. The credentials are: `u: guest p: guest`.

Tip: We don't want to use this preinstalled user and we want to remove it. But before doing that, we're gonna add our own user.

To run `rabbitmqctl`, you can install it on computer or since we already have rabbitmq running in a docker container, we can execute this
command line tool from that container:
```shell
# docker exec <container name> <command>
docker exec rabbitmq rabbitmqctl
```

Add a user to rabbitmq server:
```shell
# rabbitmqctl add_user <username> <password>
docker exec rabbitmq rabbitmqctl add_user parsa secret
```
Note: Use a secure password.

Right now we have added a user but we can't do anything with it. Because he has no permissions.

To make our new user an administrator:
```shell
docker exec rabbitmq rabbitmqctl set_user_tags parsa administrator
```

We also want to always delete the guest user(by default it's always present):
```shell
docker exec rabbitmq rabbitmqctl delete_user guest
```

Login to the UI with the newly created user.

### Virtual hosts
In rabbitmq, the resources(channels, exchanges and ...) are contained(grouped) in a virtual host(sorta like a namespace).

So use virtual hosts to kinda limit and separate resources in a logical way. It's called **virtual** because it's done in the logical layer.
It's a soft restriction between what resources can reach which resources.

By default there's the `/` virtual host which is the global one. We don't want to operate in that one, we wanna create our own virtual host.

To create a virtual host:
```shell
# rabbitmqctl add_vhost <vhost name>
docker exec rabbitmq rabbitmqctl add_vhost customers
```
The customers virtual host will hold all of our future resources which are related to anything working on customers.

Once we have our virtual host, we need to make sure that we have permissions to operate on that virtual host.

A user wants to communicate with the resources inside a virtual host. So he needs the permissions to do this.
When we specify permissions, we need to specify 3 different permissions:
- configuration permissions: What the user is allowed to configure?
- write permissions: On what resources the user is allowed to write?
- read permissions: On what resources the user is allowed to read?

The way we specify these, is by using a regexp.

For example if you want to allow **parsa** to only configure resources starting with `customer`,  the regexp would be: `"^customer.*"`.
This will allow paras to configure every resource beginning with the name `customer`, inside customers virtual host(assuming this command is
executing for `customers` vhost). To allow him to configure all the resources inside that vhost, we say: `.*` .

We wanna give the user `parsa` permissions to everything inside `customers` vhost. With this, he's able to configure, read and write on everything
inside of that vhost.
```shell
docker exec rabbitmq rabbitmqctl set_permissions -p customers parsa ".*" ".*" ".*"
```

### Queues, producers, exchanges and consumers
- Producers: are any piece of software that is sending messages. Producers send their messages to exchanges(not queues) which then routes 
the messages where they should go.
- Consumers: any piece of software that is **receiving** messages. How do they receive messages? With exchanges and queues. 
- exchanges: Think of exchange as a broker or a router. The exchanges knows which queues are bound to the exchange. 
- To bind sth to an exchange, we use a binding. A binding is a rule or set of rules. The exchange is bound to a queue by a set of rules(bindings). 
- Queue is a buffer for messages. It's usually FIFO.

### Building the rabbitmq client and a producer
We would have a program called producer, so create a folder inside `cmd` called `producer`.

`cmd` folder will hold different commands that we can execute.

The `producer` microservice inside cmd folder will send some messages to an exchange.

`github.com/rabbitmq/amqp091-go` is the official supported client for rabbitmq.

A connection in rabbitmq is a TCP connection and you should reuse the connection across your whole app and you should spawn new channels on every
concurrent task that is running.

**Q:** What's a channel?

A channel is a multiplexed connection(sub connection) over the TCP connection. Think of it as a separate connection, but it's using the same TCP connection.
So the *amqp.Connection is the TCP connection. `*amqp.Connection` should be reused.

**tip:** You should recreate the channel for each concurrent task but always have one connection. If you spawn connections, you will create
so many TCP connections and that doesn't scale well.

### Queues, durability and auto-delete
A durable queue will be persisted whenever the broker restarts. It means when rabbitmq restarts, if you want your queue to persist, you need to
set the related option when declaring the queue.

When **auto delete** option is enabled, the queue will be deleted whenever the software that created it, shuts down. In our case,
whenever the producer shuts down, the queue will be deleted(because the producer created that queue).

The `exclusive` flag will make the queue only available to the connection that created it. So if we only expect the queue to be used from
this particular piece of software, we can set `exclusive` to true. Nobody else will be able to use the queue.

These options will be reused when we create messages and exchanges.

To test things:
```shell
go run cmd/producer/main.go
```

THen, before the time.Sleep ends, go to queues in the UI and see them.

To test durability:
```shell
docker restart rabbitmq
```

The durable queues should still be visible in UI after command runs.

So if you want the messages to survive between restarts or crashes, make that queue durable.

### Bindings
Note: You're not sending messages on queues, we're sending messages to **exchanges**.

**To start receiving or sending messages on a queue, you need to bind that queue to an exchange, this is called a binding.**

The binding is a routing rule.

A queue can be bound to multiple exchanges. You can even have exchanges being bound to exchanges.

Whenever you send message, you have to add a routing key(topic) which will be used by exchange.

#### Types of exchanges

- direct: producer sends a message named `customer_created`. The exchange will send the massage to a queue named exactly `customer_created`. Other
queues with different name won't get the message.
- fanout: it sends the message to every queue present. This type of exchanges ignores the routing key.
- topic: allows you to create routing keys that are delimited by dots. For example you set the routing key to `customers.created.Feburary` and the
exchange will then send out the messages to for example `customers.created.#`.
- header: we can have routing based on the header of messages(independent of the routing key)

\# means zero or more matches. For example `customers.created.#` will match with `customers.#.created` .

To create exchanges, we can use `rabbitmqadmin` command line tool instead of `rabbitmqctl`.

Note: You can create queues, exchanges and ... inside of code, but we like to create our exchanges and ... , beforehand.
```shell
docker exec rabbitmq rabbitmqadmin declare exchange --vhost=customers name=customer_events type=topic -u parsa -p <password> durable=true
```

We also need to set permissions for the user. We want to allow our user to read and write any messages starting with `customers.*`(so basically
anything sent on customers can be read and write by this user). For example if the exchange(customer_events) would send billing information on
`billing.<...>` topic, this person wouldn't be able to listen on those, because we're only allowed him to listen on any topic called `customers.*`.

```shell
docker exec rabbitmq rabbitmqctl set_topic_permissions -p customers parsa customer_events "^customers.*" "^customers.*"
```

### Publishing messages
Now that we have set up exchanges, let's publish messages onto the exchanges.

First we have to bind queues to exchanges, so we need bindings. For this, create `CreateBinding` func which is just a simple wrapper.

**Note:** Setting noWait to false, will make the channel return an error if it fails to bind.

You can also create bindings in the UI, but doing it in code is more apparent and reusable. But for testing things, UI is suitable.

Now that we have bound the queues to exchanges, we can publish messages. 

It's good to use PublishWithContext instead of Publish because that allows us to add timeouts to the messages.

In `PublishWithContext`, `mandatory` arg is used to determine if a failure should drop the message that you're sending or if it should return an error?
If you set `mandatory` to false, you won't get an error. It will fire and forget. But if it's true, it will make sure that if it fails to
publish the message, it will return an error.

`immediate` you can leave it to false because you would never use that if you're using the rabbitmq package that we used in this project. Because
`immediate` flag is removed in rabbitmq3. So this flag is deprecated unless you're using an old version of rabbitmq.

#### Delivery mode
If you want to have your messages persist which means if you send a message and no consumer consumes it and your rabbitmq server restarts, that message
will be deleted if it's not a persistent message and those messages are called **transient**.

Why would you not want your messages to be persistent?

It's a matter of performance and also it's up to you, for example if there's no reason that the event will happen if the server comes up,
there's no reason to persist it, so they should be transient to increase perf. Because by making things durable in rabbitmq, there will be
overhead to it.

**Note:** If you want to send persistent messages, your queue also needs to be durable. There's no point sending persistent messages on a queue that isn't
persistent itself.

If you acknowledge the message, the message will be gone.

To test things, after sending the messages, do:
```shell
docker restart rabbitmq
```

The transient messages should be gone.

Note: If you only want the messages to be just sent and don't care if they are actually consumed, set them as transient.

### consuming messages and acknowledgment
The exchange needs to know that the client(consumer) wanted the message, actually received it. To do this, an acknowledge is sent back to the
exchange. Then the exchange will know that it can drop the message.

Note: The tricky part is if you have a consumer that acknowledge the message but then fails to process it, that message will get lost. Because the server
has delivered it and we acked it. So you might not always want `autoAck`. Sometimes you want to acknowledge the message **manually** whenever 
that consumer is done processing that message. So if your consumer can fail when processing the message, don't use autoAck, unless you're sure
that's what you want.

If `exclusive` flag when consuming is set to true, it means this will be the one and only consumer consuming that queue. If it's set to false,
the server will distribute messages using a load balancing technique. So if you only have one consumer that you want to consume all messages,
set `exclusive` to true.

In Consume method, we're not gonna even let the user of method to set the `exclusive` flag.

`noLocal` is not supported in rabbitmq. It's supported in amqp but not in rabbitmq.

**Note:** If rabbitmq never receives an acknowledgement, it won't drop the message unless it expires.

If you ack the message, you will never see that message again.

To test things, you can run the producer to send message:
```shell
go run cmd/producer/main.go
```

**Note:** autoAck will send the ack to server as soon as consumer receives the message and if after receiving our program crashes or ... , we can't
process that message again because it has been dropped. So autoAck can be dangerous.

### Nacking and retries
When there's an actual failure and we want rabbitmq to know that we can send a NACK which will tell rabbitmq that it failed to process and by doing this,
rabbitmq can retry and requeue the message. Like sending the message again or it can throw it away.

Q: When to use `multiple` flag(when consuming - ACK and NACK)?

You wanna use this if you have a high volume of traffic and what rabbitmq will do then is the client will buffer a few messages before sending the
ACKs or NACKs which will reduce the network load.

The requeue flag in Nack() means if the server should retry sending the message out again.

So whenever your service fails, you can send back a Nack to the exchange telling it to re-send that message again.

Right now we can only receive one message at a time, because everything is single threaded(just 1 goroutine where we're ranging through `messageBuss` in
consumer).

### multithreading using ErrGroup
We will be setting the amount of concurrency in go right now, but we can also do this inside rabbitmq. We will look at this later.

To test things, run 2 consumers and 1 publishers(so we need 3 terminal windows):
```shell
go run cmd/consumer/main.go # in 2 windows
go run cmd/producer/main.go
```
Now our 10 messages will be load balanced onto the consumers.

Currently, the producer sends the messages and then immediately exits. What if we wanted the producer to **wait** until the work is done?
Because we know each message takes 10 seconds to complete(OFC they're being processed concurrently).

### Deferred confirm and confirm mode
So we want to wait in producer, because for example if the message processing fails, maybe the producer wants to do sth else.

To do this, we can use `PublishWithDeferredConfirmWithContext` instead of `PublishWithContext`.

`PublishWithDeferredConfirmWithContext` will always return nil on confirmation, if the queue isn't set to `confirmMode`.

**Note:** `confirmation.Wait()` won't wait until the work is completed, but it will at least wait until the server acknowledges that 
it has received the message.

So we need to enable confirmMode.

**Note:** This is not the same as when the consumer acknowledges the message. This is when the server acknowledges that it has published the message
on the exchange. They are different. This way, we can make sure that the message we send, is actually sent(so if it's really important that the
message is actually produced, use this).

Up until now we've been using FIFO queues.

### Fanout and publish and subscribe
In a pub/sub schema you want **all** the consumers to receive all the same messages(you don't want to load balance the messages).
To do this, we use a fanout exchange which will skip topics(routing keys) and will push messages to all the consumers.

Let's delete the current stage. Because it's the wrong type. So run:
```shell
docker exec rabbitmq rabbitmqadmin delete exchange name=customer_events --vhost=customers -u parsa -p <password>
```

Note: You can't change the type of an exchange. You have to delete it and then recreate a new one.
```shell
docker exec rabbitmq rabbitmqadmin declare exchange --vhost=customers name=customers_events type=fanout -u parsa -p <password> durable=true
```

Update the permissions for the new created exchange:
```shell
docker exec rabbitmq rabbitmqctl set_topic_permissions -p customers <username> customer_events ".*" ".*"
```

Note: In a pub/sub , it's most likely that the subscriber is creating the queues and bindings. Because the publisher won't know what queues exist. So put the
code for creating the queue in consumer, not publisher(producer).

In pub/sub, when creating queues in consumer, we don't specify the name of the queue because the consumer doesn't care about the name, rabbitmq will
generate those names. You **can** use known names(specify name) but usually when you have pub/sub you might end up with many subscribers(consumers) and
you don't know the subscribers.

The reason we return the queue in `CreateQueue` when calling it in consumer is because when we create the binding, we need the name of the queue which
in this case is randomly created by rabbitmq itself and not us.

Now run 2 instances of consumers(using 2 terminal windows) and then run the producer. We will see that all the 10 messages are sent to both consumers(instead
of having a load balanced strategy). So in the case where you want all the consumers receive all the messages, use fanout.

### RPC procedures
Producer will send a replyTo: <queue name that the producer is listening on> with each message and the service knows whenever it's done, it will replyTo
that queue which the producer is listening on