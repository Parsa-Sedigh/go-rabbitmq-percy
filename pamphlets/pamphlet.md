Monolith vs event-driven arch:

In monolith, we have coupling and we have to re-deploy the whole monolith. But in event-driven arch, we leverage microservices and
these microservices push events to each other using events and rabbitMQ can deliver these events between the microservices.

The easiest way to start up rabbitmq is using docker.

Run an instance of rabbitmq:
```shell
# 5672 port is the amqp port for connections
# 15672 is the port used by admin(management) UI for rabbitmq
docker run -d --name rabbitmq 5672:5672 -p 15672:15672 rabbitmq:3.11-management
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
docker exec rabbitmq rabbitmqctl set_user_tags percy adminstrator
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

By default there's the / virtual host which is the global one. We don't want to operate in that one, we wanna create our own virtual host.

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