<a name="readme-top"></a>

<!-- PROJECT LOGO -->
<div align="center">
  <h3 align="center">P2P Messaging</h3>

  <p align="center">
   Developed by: Sebastian Corcino and Gaozong Lo
  </p>
</div>

<!-- ABOUT THE PROJECT -->

## About The Project

We created a Peer to Peer messaging chat application inside the terminal utilizing the [Libp2p Library](https://docs.libp2p.io/concepts/discovery-routing/mdns/). Libp2p is a modular library that includes modules like networking protocols and peer discovery used for creating distributed applications. The core concepts that is being applied is [Publishers and Subscribers](https://docs.libp2p.io/concepts/pubsub/overview/). Chat rooms follows this architecture due to the idea that everybody in the room (topic) is receiving all messages from other subscribers.

What can the System do?

- A user can login with there username. (Automatically generate username if not provided and generate ID)
- A user can friend existing users. (Accepting friend invitations **DOESN'T EXIST**)
- A user can send messages to online friends.
- A user can create chat rooms that anybody can join and see who is currently viewing that chat room.
- Pure P2P application: Nodes may be added or removed as needed without configurations.

To run the program without setting a name:
```
go run .
```

To run the program with a name:
```
go run . -nick=<name>
```

To run the program and enter a private chatroom:
```
go run . -room=<room name>
```
To exit the chatroom run this in the command line:
```
/quit
```
When you run the command to get into the chatroom, you should see the general chatroom where incoming messages will display. A peers list is located on the right side displaying detected peers in numerical order. 
Peers that are shown are considered online, however, you can also test this command to check if the peer is online:
```
/status <full peer ID>
```
The full peer ID will be listed in the peers column. Note: Status does not check for self therefore users can only check the status of other users. 

Base Project Reference: https://github.com/libp2p/go-libp2p/tree/master/examples/pubsub/chat
Terminal Prettified References: [tcell](https://github.com/gdamore/tcell) and [tview](https://github.com/rivo/tview)

<p align="right">(<a href="#readme-top">back to top</a>)</p>
