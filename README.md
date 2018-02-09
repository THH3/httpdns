This repository provided an adapter to forward standard DNS query to google HTTP DNS. 
You can use it as a local DNS server. If you could not access google directly, 
you can run the forwarder on your remote server, and set the `apiUrl` option of adapter 
to the forwarder server.

Currently it does not support `ANY` type of query.

