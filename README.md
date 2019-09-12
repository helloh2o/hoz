## HTTP & SOCKS5 fast security proxy
Currently implemented HTTP,SOCKS5 protocol
### Transport Protocol
The package <b>[head, body]</b>, head is 4 bytes containing the current encrypted packet length, and the body is the encrypted bytes. After the read is completed, the decryption is performed.
#### implement your Transport Protocol by implement pkg.PackageReader & pkg.PackageWriter

### About encryption
The default is OORR which is implemented by myself. It is mainly used to perform or perform operations on bytes. Modify all bytes to achieve the purpose of encryption. It is fast and consumes less resources.

### How to use it
./client_side -addr ":1080" -remote "127.0.0.1:10800" -password "oor://your-password-xxx" <br>
./server_side -addr ":10800" -password "oor://your-password-xxx"
####  support oor -password "oor://password-xxx", salsa20 -password "sal://password-xxx", aes -password "aes://password-xxx"
#### [releases download](https://github.com/helloh2o/hoz/releases)