This repository provided an adapter to forward standard DNS query to google HTTP DNS. 
you can use it as a local DNS server. if you could not access google directly, 
you can build and run the `forwarder.go` on your remote server, and modify the 
`HttpDnsApi` in `adapter.go` to your own forward server.

Currently it does not support `ANY` type query.

