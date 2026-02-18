# Client/server protocol
Following is the specification for the binary format in which the server and client applications communicate with one another over the Unix socket.

All multi-byte numbers are little-endian.

## Client → server

### Login
Sent from a client to the server when logging in
| Length | Type | Content |
| - | - | - |
| 1 | `byte` | `RequestTypeLogin` |
| variable | `string` | `username` |
| 1 | `byte` | `\n` |

## Server → client

### Login
Sent from the server to all logged-in clients when a user logs in
| Length | Type | Content |
| - | - | - |
| 1 | `byte` | `ResponseTypeLogin` |
| 2 | `uint16` | `len(username)` |
| `len(username)` | `string` | `username` |

### Logout
Sent from the server to all logged-in clients when a user logs out
| Length | Type | Content |
| - | - | - |
| 1 | `byte` | `ResponseTypeLogout` |
| 2 | `uint16` | `len(username)` |
| `len(username)` | `string` | `username` |

### Logged-in users
Sent from the server to any client who asks, and to every user on the Welcome screen any time `numLoggedInUsers` is updated
| Length | Type | Content |
| - | - | - |
| 1 | `byte` | `ResponseTypeLoggedInUsers` |
| 2 | `uint16` | `numLoggedInUsers` |
| 2 | `uint16` | `len(username1)` |
| `len(username1)` | `string` | `username1` |
| 2 | `uint16` | `len(username2)` |
| `len(username2)` | `string` | `username2` |
| ... | | |

### Say
Sent from the server to all logged-in clients when a user uses the `say` command
| Length | Type | Content |
| - | - | - |
| 1 | `byte` | `ResponseTypeSay` |
| 2 | `uint16` | `len(username)` |
| `len(username)` | `string` | `username` |
| 2 | `uint16` | `len(message)` |
| `len(message)` | `string` | `message` |

### Look
Sent from the server to any client who asks; gives description of their character's surroundings
| Length | Type | Content |
| - | - | - |
| 1 | `byte` | `ResponseTypeLook` |
| 2 | `uint16` | `len(roomTitle)` |
| `len(roomTitle)` | `string` | `roomTitle` |
| 2 | `uint16` | `len(roomDescription)` |
| `len(roomDescription)` | `string` | `roomDescription` |
