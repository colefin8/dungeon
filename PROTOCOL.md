# Client/server protocol
Following is the specification for the binary format in which the server and client applications communicate with one another over the Unix socket.

All multi-byte numbers are little-endian.

## Server → client

### Login
Sent from the server to all logged-in clients when a user logs in
| Length | Contents |
| - | - |
| 1 | `ResponseTypeLogin` |
| 2 | `len(username)` |
| `len(username)` | `username` |

### Say
Sent from the server to all logged-in clients when a user uses the `say` command
| Length | Contents |
| - | - |
| 1 | `ResponseTypeSay` |
| 2 | `len(username)` |
| `len(username)` | `username` |
| 2 | `len(message)` |
| `len(message)` | `message` |

### Logged-in users
Sent from the server to any client who asks, and to every user on the Welcome screen any time `numLoggedInUsers` is updated
| Length | Contents |
| - | - |
| 1 | `ResponseTypeLoggedInUsers` |
| 2 | `len(username)` |
| `len(username)` | `username` |
| 2 | `len(message)` |
| `len(message)` | `message` |
